package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"tea-exam/internal/models"
	"tea-exam/internal/services"
	"tea-exam/pkg/csvutil"
	"tea-exam/pkg/response"
)

type AdminHandler struct {
	adminService *services.AdminService
}

func NewAdminHandler(adminService *services.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

func (h *AdminHandler) ListExamUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	users, total, err := h.adminService.ListExamUsers(page, pageSize, c.Query("keyword"))
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, response.NewPageData(users, total, page, pageSize))
}

func (h *AdminHandler) CreateExamUser(c *gin.Context) {
	var req services.ExamUserInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "答题人信息格式错误")
		return
	}
	user, err := h.adminService.CreateExamUser(req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "答题人创建成功", user)
}

type ReclassifyQuestionsRequest struct {
	QuestionIDs     []uint `json:"question_ids"`
	QuestionBankID  uint   `json:"question_bank_id" binding:"required"`
	AllUnclassified bool   `json:"all_unclassified"`
}

func (h *AdminHandler) ListQuestionBanks(c *gin.Context) {
	banks, err := h.adminService.ListQuestionBanks()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, banks)
}

func (h *AdminHandler) CreateQuestionBank(c *gin.Context) {
	var req services.QuestionBankTypeInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "题库信息格式错误")
		return
	}
	bank, err := h.adminService.CreateQuestionBank(req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "题库创建成功", bank)
}

func (h *AdminHandler) UpdateQuestionBank(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		response.BadRequest(c, "题库ID无效")
		return
	}
	var req services.QuestionBankTypeInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "题库信息格式错误")
		return
	}
	bank, err := h.adminService.UpdateQuestionBank(uint(id), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "题库更新成功", bank)
}

func (h *AdminHandler) GetExamRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	records, total, err := h.adminService.GetExamRecords(page, pageSize, c.Query("keyword"), c.Query("question_bank_id"))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, response.NewPageData(records, total, page, pageSize))
}

func (h *AdminHandler) GetBankStats(c *gin.Context) {
	stats, err := h.adminService.GetBankStats()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, stats)
}

func (h *AdminHandler) ListQuestions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	questions, total, err := h.adminService.ListQuestions(page, pageSize, c.Query("keyword"), c.Query("question_bank_id"))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, response.NewPageData(questions, total, page, pageSize))
}

func (h *AdminHandler) ReclassifyQuestions(c *gin.Context) {
	var req ReclassifyQuestionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请选择目标题库")
		return
	}
	count, err := h.adminService.ReclassifyQuestions(req.QuestionIDs, req.QuestionBankID, req.AllUnclassified)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "题目归类成功", gin.H{"updated_count": count})
}

func (h *AdminHandler) ImportQuestions(c *gin.Context) {
	mode := c.PostForm("mode")
	if mode != "replace" && mode != "append" {
		response.BadRequest(c, "导入模式无效")
		return
	}
	bankIDValue, err := strconv.ParseUint(c.PostForm("question_bank_id"), 10, 32)
	if err != nil || bankIDValue == 0 {
		response.BadRequest(c, "请选择导入的具体题库")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请上传文件")
		return
	}
	openedFile, err := file.Open()
	if err != nil {
		response.ServerError(c, "打开文件失败")
		return
	}
	defer openedFile.Close()

	records, err := csvutil.ReadCSVWithEncoding(openedFile)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if len(records) < 2 {
		response.BadRequest(c, "CSV 文件内容为空或格式错误")
		return
	}
	questions, result := csvutil.ParseQuestions(records, 1)
	if result.SuccessCount == 0 {
		response.BadRequest(c, "没有有效题目可导入")
		return
	}

	questionModels := make([]models.QuestionBank, 0, len(questions))
	for _, q := range questions {
		questionModels = append(questionModels, models.QuestionBank{
			QuestionNo:       q.QuestionNo,
			QuestionTypeCode: q.QuestionTypeCode,
			QuestionTypeName: q.QuestionTypeName,
			QuestionText:     q.QuestionText,
			OptionA:          q.OptionA,
			OptionB:          q.OptionB,
			OptionC:          q.OptionC,
			OptionD:          q.OptionD,
			OptionE:          q.OptionE,
			CorrectAnswer:    q.CorrectAnswer,
		})
	}
	if err := h.adminService.ImportQuestions(questionModels, mode, uint(bankIDValue)); err != nil {
		response.BadRequest(c, "导入失败: "+err.Error())
		return
	}
	response.SuccessWithMessage(c, "导入成功", gin.H{
		"success_count": result.SuccessCount,
		"fail_count":    result.FailCount,
		"fail_reasons":  result.FailReasons,
	})
}
