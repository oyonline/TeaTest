package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"tea-exam/internal/services"
	"tea-exam/pkg/response"
)

// ExamHandler 考试处理器
type ExamHandler struct {
	examService *services.ExamService
}

// NewExamHandler 创建考试处理器
func NewExamHandler(examService *services.ExamService) *ExamHandler {
	return &ExamHandler{examService: examService}
}

// StartExamRequest 开始考试请求
type StartExamRequest struct {
	UserID   uint   `json:"user_id"`
	UserName string `json:"user_name"`
}

// SubmitAnswerRequest 提交答案请求
type SubmitAnswerRequest struct {
	QuestionID uint   `json:"question_id" binding:"required"`
	Answer     string `json:"answer" binding:"required"`
}

// GetInProgressExam 获取进行中的考试
func (h *ExamHandler) GetInProgressExam(c *gin.Context) {
	userID, _ := c.Get("userID")
	userName, _ := c.Get("userName")

	record, err := h.examService.GetInProgressExam(userID.(uint))
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	if record == nil {
		response.Success(c, gin.H{
			"has_in_progress": false,
			"user_name":       userName,
		})
		return
	}

	// 获取考试统计
	stats, err := h.examService.GetExamStats(userID.(uint))
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"has_in_progress":  true,
		"exam_id":          record.ID,
		"start_time":       record.StartTime,
		"stats":            stats,
		"user_name":        userName,
	})
}

// StartExam 开始考试
func (h *ExamHandler) StartExam(c *gin.Context) {
	userID, _ := c.Get("userID")
	userName, _ := c.Get("userName")

	record, err := h.examService.StartExam(userID.(uint), userName.(string))
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, gin.H{
		"exam_id":    record.ID,
		"start_time": record.StartTime,
	})
}

// GetQuestions 获取分页题目
func (h *ExamHandler) GetQuestions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	examID, _ := strconv.ParseUint(c.Query("exam_id"), 10, 32)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	// 如果有 examID，返回带进度的题目
	if examID > 0 {
		result, err := h.examService.GetQuestionsWithProgress(page, pageSize, uint(examID))
		if err != nil {
			response.ServerError(c, err.Error())
			return
		}
		response.Success(c, result)
		return
	}

	// 否则只返回题目
	questions, total, err := h.examService.GetQuestions(page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, response.NewPageData(questions, total, page, pageSize))
}

// GetAllQuestionStatus 获取所有题目状态（用于进度网格）
func (h *ExamHandler) GetAllQuestionStatus(c *gin.Context) {
	examID, err := strconv.ParseUint(c.Query("exam_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "考试ID无效")
		return
	}

	status, err := h.examService.GetAllQuestionNosWithStatus(uint(examID))
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, status)
}

// SubmitAnswer 提交答案
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

	progress, err := h.examService.SubmitAnswer(uint(examID), req.QuestionID, req.Answer)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	// 检查是否完成考试
	record, completed, err := h.examService.CheckAndCompleteExam(uint(examID))
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	resp := gin.H{
		"is_correct":    progress.IsCorrect,
		"user_answer":   progress.UserAnswer,
		"is_completed": completed,
	}

	if completed {
		resp["result"] = gin.H{
			"total_score":    record.TotalScore,
			"correct_count":  record.CorrectCount,
			"wrong_count":    record.WrongCount,
			"accuracy_rate":  record.AccuracyRate,
			"duration_seconds": record.DurationSeconds,
		}
	}

	response.Success(c, resp)
}

// GetUnansweredQuestions 获取未答题的题目
func (h *ExamHandler) GetUnansweredQuestions(c *gin.Context) {
	examID, err := strconv.ParseUint(c.Param("exam_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "考试ID无效")
		return
	}

	unanswered, err := h.examService.GetUnansweredQuestions(uint(examID))
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"unanswered_count": len(unanswered),
		"unanswered_ids":   unanswered,
	})
}

// GetExamResult 获取考试结果
func (h *ExamHandler) GetExamResult(c *gin.Context) {
	examID, err := strconv.ParseUint(c.Param("exam_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "考试ID无效")
		return
	}

	record, err := h.examService.GetExamResult(uint(examID))
	if err != nil {
		response.NotFound(c, "考试记录不存在")
		return
	}

	// 验证权限（只能查看自己的考试记录）
	userID, _ := c.Get("userID")
	if record.UserID != userID.(uint) {
		response.Forbidden(c, "无权查看此考试记录")
		return
	}

	response.Success(c, gin.H{
		"id":               record.ID,
		"user_name":        record.UserName,
		"start_time":       record.StartTime,
		"end_time":         record.EndTime,
		"duration_seconds": record.DurationSeconds,
		"total_score":      record.TotalScore,
		"correct_count":    record.CorrectCount,
		"wrong_count":      record.WrongCount,
		"accuracy_rate":    record.AccuracyRate,
		"status":           record.Status,
	})
}

// GetExamStats 获取考试统计
func (h *ExamHandler) GetExamStats(c *gin.Context) {
	userID, _ := c.Get("userID")

	stats, err := h.examService.GetExamStats(userID.(uint))
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, stats)
}
