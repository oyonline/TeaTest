package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"tea-exam/internal/models"
	"tea-exam/internal/services"
	"tea-exam/pkg/csvutil"
	"tea-exam/pkg/response"
)

// AdminHandler 管理员处理器
type AdminHandler struct {
	adminService *services.AdminService
}

// NewAdminHandler 创建管理员处理器
func NewAdminHandler(adminService *services.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

// ImportQuestionsRequest 导入题目请求
type ImportQuestionsRequest struct {
	Mode string `json:"mode" binding:"required,oneof=replace append"`
}

// GetExamRecords 获取考试记录列表
func (h *AdminHandler) GetExamRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	records, total, err := h.adminService.GetExamRecords(page, pageSize, keyword)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, response.NewPageData(records, total, page, pageSize))
}

// GetBankStats 获取题库统计
func (h *AdminHandler) GetBankStats(c *gin.Context) {
	stats, err := h.adminService.GetBankStats()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, stats)
}

// ImportQuestions 导入题目
func (h *AdminHandler) ImportQuestions(c *gin.Context) {
	mode := c.PostForm("mode")
	if mode != "replace" && mode != "append" {
		mode = "replace" // 默认覆盖模式
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请上传文件")
		return
	}

	// 打开文件
	openedFile, err := file.Open()
	if err != nil {
		response.ServerError(c, "打开文件失败")
		return
	}
	defer openedFile.Close()

	// 读取 CSV 数据（自动处理编码）
	records, err := csvutil.ReadCSVWithEncoding(openedFile)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	if len(records) < 2 {
		response.Error(c, 400, "CSV 文件内容为空或格式错误")
		return
	}

	// 解析题目数据
	questions, result := csvutil.ParseQuestions(records, 1) // 从第2行开始解析（跳过标题行）

	if result.SuccessCount == 0 {
		response.BadRequest(c, "没有有效题目可导入")
		return
	}

	// 转换为模型
	var questionModels []models.QuestionBank
	for _, q := range questions {
		questionModels = append(questionModels, models.QuestionBank{
			QuestionNo:       q.QuestionNo,
			BankName:         q.BankName,
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

	// 导入数据库
	if err := h.adminService.ImportQuestions(questionModels, mode); err != nil {
		response.ServerError(c, "导入失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "导入成功", gin.H{
		"success_count": result.SuccessCount,
		"fail_count":    result.FailCount,
		"fail_reasons":  result.FailReasons,
	})
}
