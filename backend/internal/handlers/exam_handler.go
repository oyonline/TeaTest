package handlers

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"tea-exam/internal/services"
	"tea-exam/pkg/response"
)

type ExamHandler struct {
	examService *services.ExamService
}

func NewExamHandler(examService *services.ExamService) *ExamHandler {
	return &ExamHandler{examService: examService}
}

type StartExamRequest struct {
	QuestionBankID uint `json:"question_bank_id" binding:"required"`
}

type SubmitAnswerRequest struct {
	QuestionID uint   `json:"question_id" binding:"required"`
	Answer     string `json:"answer" binding:"required"`
}

func handleExamError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrExamNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, services.ErrExamForbidden):
		response.Forbidden(c, err.Error())
	default:
		response.Error(c, 400, err.Error())
	}
}

// GetQuestionBanks 获取用户可选择的具体题库。
func (h *ExamHandler) GetQuestionBanks(c *gin.Context) {
	userID, _ := c.Get("userID")
	banks, err := h.examService.GetQuestionBanks(userID.(uint))
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, banks)
}

// GetInProgressExam 获取指定题库的进行中考试；不传题库 ID 时查询历史未分类考试。
func (h *ExamHandler) GetInProgressExam(c *gin.Context) {
	userID, _ := c.Get("userID")
	userName, _ := c.Get("userName")
	var bankID *uint
	if value := c.Query("question_bank_id"); value != "" {
		parsed, err := strconv.ParseUint(value, 10, 32)
		if err != nil || parsed == 0 {
			response.BadRequest(c, "题库ID无效")
			return
		}
		id := uint(parsed)
		bankID = &id
	}
	record, err := h.examService.GetInProgressExam(userID.(uint), bankID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	if record == nil {
		response.Success(c, gin.H{"has_in_progress": false, "user_name": userName})
		return
	}
	_, total, err := h.examService.GetExamInfo(record.ID, userID.(uint))
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"has_in_progress":  true,
		"exam_id":          record.ID,
		"start_time":       record.StartTime,
		"question_bank_id": record.QuestionBankTypeID,
		"question_bank_name": func() string {
			if record.QuestionBankName == "" {
				return "历史未分类"
			}
			return record.QuestionBankName
		}(),
		"total_questions": total,
		"user_name":       userName,
	})
}

func (h *ExamHandler) StartExam(c *gin.Context) {
	var req StartExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请选择题库")
		return
	}
	userID, _ := c.Get("userID")
	userName, _ := c.Get("userName")
	record, err := h.examService.StartExam(userID.(uint), userName.(string), req.QuestionBankID)
	if err != nil {
		handleExamError(c, err)
		return
	}
	response.Success(c, gin.H{
		"exam_id":            record.ID,
		"start_time":         record.StartTime,
		"question_bank_id":   record.QuestionBankTypeID,
		"question_bank_name": record.QuestionBankName,
	})
}

func (h *ExamHandler) GetExamInfo(c *gin.Context) {
	examID, err := strconv.ParseUint(c.Param("exam_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "考试ID无效")
		return
	}
	userID, _ := c.Get("userID")
	record, total, err := h.examService.GetExamInfo(uint(examID), userID.(uint))
	if err != nil {
		handleExamError(c, err)
		return
	}
	bankName := record.QuestionBankName
	categoryName := services.QuestionBankCategoryName(record.QuestionBankCategory)
	if bankName == "" {
		bankName = "历史未分类"
	}
	response.Success(c, gin.H{
		"exam_id":            record.ID,
		"start_time":         record.StartTime,
		"status":             record.Status,
		"question_bank_id":   record.QuestionBankTypeID,
		"question_bank_name": bankName,
		"category_code":      record.QuestionBankCategory,
		"category_name":      categoryName,
		"total_questions":    total,
	})
}

func (h *ExamHandler) GetQuestions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	examID, err := strconv.ParseUint(c.Query("exam_id"), 10, 32)
	if err != nil || examID == 0 {
		response.BadRequest(c, "考试ID无效")
		return
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}
	userID, _ := c.Get("userID")
	result, err := h.examService.GetQuestionsWithProgress(page, pageSize, uint(examID), userID.(uint))
	if err != nil {
		handleExamError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ExamHandler) GetAllQuestionStatus(c *gin.Context) {
	examID, err := strconv.ParseUint(c.Query("exam_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "考试ID无效")
		return
	}
	userID, _ := c.Get("userID")
	status, err := h.examService.GetAllQuestionNosWithStatus(uint(examID), userID.(uint))
	if err != nil {
		handleExamError(c, err)
		return
	}
	response.Success(c, status)
}

func (h *ExamHandler) SubmitAnswer(c *gin.Context) {
	var req SubmitAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "题目ID和答案不能为空")
		return
	}
	examID, err := strconv.ParseUint(c.Param("exam_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "考试ID无效")
		return
	}
	userID, _ := c.Get("userID")
	progress, record, completed, correctAnswer, err := h.examService.SubmitAnswer(uint(examID), userID.(uint), req.QuestionID, req.Answer)
	if err != nil {
		handleExamError(c, err)
		return
	}
	resp := gin.H{
		"is_correct":     progress.IsCorrect,
		"user_answer":    progress.UserAnswer,
		"correct_answer": correctAnswer,
		"is_completed":   completed,
	}
	if completed {
		resp["result"] = gin.H{
			"total_score":      record.TotalScore,
			"correct_count":    record.CorrectCount,
			"wrong_count":      record.WrongCount,
			"accuracy_rate":    record.AccuracyRate,
			"duration_seconds": record.DurationSeconds,
		}
	}
	response.Success(c, resp)
}

func (h *ExamHandler) GetUnansweredQuestions(c *gin.Context) {
	examID, err := strconv.ParseUint(c.Param("exam_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "考试ID无效")
		return
	}
	userID, _ := c.Get("userID")
	unanswered, err := h.examService.GetUnansweredQuestions(uint(examID), userID.(uint))
	if err != nil {
		handleExamError(c, err)
		return
	}
	response.Success(c, gin.H{"unanswered_count": len(unanswered), "unanswered_ids": unanswered})
}

func (h *ExamHandler) GetExamResult(c *gin.Context) {
	examID, err := strconv.ParseUint(c.Param("exam_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "考试ID无效")
		return
	}
	userID, _ := c.Get("userID")
	record, err := h.examService.GetExamResult(uint(examID), userID.(uint))
	if err != nil {
		handleExamError(c, err)
		return
	}
	bankName := record.QuestionBankName
	if bankName == "" {
		bankName = "历史未分类"
	}
	response.Success(c, gin.H{
		"id":                     record.ID,
		"user_name":              record.UserName,
		"question_bank_name":     bankName,
		"question_bank_category": services.QuestionBankCategoryName(record.QuestionBankCategory),
		"start_time":             record.StartTime,
		"end_time":               record.EndTime,
		"duration_seconds":       record.DurationSeconds,
		"total_score":            record.TotalScore,
		"correct_count":          record.CorrectCount,
		"wrong_count":            record.WrongCount,
		"accuracy_rate":          record.AccuracyRate,
		"status":                 record.Status,
	})
}

func (h *ExamHandler) GetExamStats(c *gin.Context) {
	userID, _ := c.Get("userID")
	stats, err := h.examService.GetExamStats(userID.(uint))
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, stats)
}
