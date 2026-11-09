package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"tea-exam/internal/config"
	"tea-exam/internal/middleware"
	"tea-exam/internal/models"
	"tea-exam/internal/services"
	"tea-exam/internal/utils"
)

type stage2Envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func TestStage2MySQL(t *testing.T) {
	dsn := os.Getenv("TEA_STAGE2_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TEA_STAGE2_MYSQL_DSN 未设置；跳过真实 MySQL 集成测试")
	}
	dbName := dsnDatabaseName(dsn)
	if !strings.HasPrefix(dbName, "teatest_stage2_") {
		t.Fatalf("拒绝非 teatest_stage2_ 前缀的数据库: %q", dbName)
	}
	dsn = withDSNParam(dsn, "parseTime", "True")
	dsn = withDSNParam(dsn, "charset", "utf8mb4")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("连接测试库失败: %v", err)
	}
	if err := db.Exec(fmt.Sprintf("ALTER DATABASE `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbName)).Error; err != nil {
		t.Fatalf("设置测试库 utf8mb4 失败: %v", err)
	}
	if err := db.Exec("SET NAMES utf8mb4").Error; err != nil {
		t.Fatalf("设置 utf8mb4 失败: %v", err)
	}

	// 受控夹具：只清理专用测试库中的既有表。先造旧 varchar 列验证 AutoMigrate 扩容。
	db.Exec("SET FOREIGN_KEY_CHECKS=0")
	for _, table := range []string{"exam_progress", "exam_records", "question_bank", "question_bank_types", "exam_user_question_banks", "exam_users", "admin_config"} {
		if err := db.Exec("DROP TABLE IF EXISTS " + table).Error; err != nil {
			t.Fatalf("清理旧表 %s 失败: %v", table, err)
		}
	}
	db.Exec("SET FOREIGN_KEY_CHECKS=1")
	if err := db.Exec(`CREATE TABLE question_bank (
		id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
		question_no INT NOT NULL,
		question_text TEXT NOT NULL,
		correct_answer VARCHAR(10) NOT NULL
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`).Error; err != nil {
		t.Fatalf("创建旧 question_bank 失败: %v", err)
	}
	if err := db.Exec(`CREATE TABLE exam_progress (
		id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
		exam_record_id BIGINT UNSIGNED NOT NULL,
		question_id BIGINT UNSIGNED NOT NULL,
		user_answer VARCHAR(10) NULL
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`).Error; err != nil {
		t.Fatalf("创建旧 exam_progress 失败: %v", err)
	}
	if err := db.Exec(`INSERT INTO question_bank (question_no, question_text, correct_answer) VALUES (1, '历史题目', 'A')`).Error; err != nil {
		t.Fatalf("写入旧题目失败: %v", err)
	}

	if err := utils.InitDatabase(db); err != nil {
		t.Fatalf("AutoMigrate 失败: %v", err)
	}
	if err := utils.SeedData(db, "admin-pass", true); err != nil {
		t.Fatalf("种子数据失败: %v", err)
	}

	assertColumnType(t, db, "question_bank", "correct_answer", "text")
	assertColumnType(t, db, "exam_progress", "user_answer", "text")
	assertColumnCharset(t, db, "question_bank", "correct_answer", "utf8mb4")
	assertColumnCharset(t, db, "exam_progress", "user_answer", "utf8mb4")
	var migratedQuestion models.QuestionBank
	if err := db.Where("question_text = ?", "历史题目").First(&migratedQuestion).Error; err != nil || migratedQuestion.CorrectAnswer != "A" {
		t.Fatalf("旧题目/答案迁移后丢失: %+v err=%v", migratedQuestion, err)
	}

	adminSvc := services.NewAdminService(db)
	examSvc := services.NewExamService(db)

	// 一次性授权：已存在用户 × 已存在题库；重启/新用户/新库不自动补授。
	configRow, err := adminSvc.GetAdminConfig()
	if err != nil || !configRow.BankAccessInitialized {
		t.Fatalf("一次性授权标记未落库: %+v err=%v", configRow, err)
	}
	var initialLinks int64
	db.Model(&models.ExamUserQuestionBank{}).Count(&initialLinks)
	if initialLinks == 0 {
		t.Fatal("种子用户应获得存量题库授权")
	}
	var demoUser models.ExamUser
	if err := db.Order("id ASC").First(&demoUser).Error; err != nil {
		t.Fatalf("读取种子用户失败: %v", err)
	}
	if err := adminSvc.ReplaceUserQuestionBanks(demoUser.ID, nil); err != nil {
		t.Fatalf("撤销种子用户权限失败: %v", err)
	}
	var beforeReinitLinks int64
	db.Model(&models.ExamUserQuestionBank{}).Count(&beforeReinitLinks)
	afterInitBank, err := adminSvc.CreateQuestionBank(services.QuestionBankTypeInput{
		CategoryCode: models.QuestionBankCategoryGeneral, Name: "初始化后新通用题库", Status: 1,
	})
	if err != nil {
		t.Fatalf("创建新题库失败: %v", err)
	}
	afterInitUser, err := adminSvc.CreateExamUser(services.ExamUserInput{Name: "初始化后新用户", Password: "pass1234"})
	if err != nil {
		t.Fatalf("创建新用户失败: %v", err)
	}
	if err := utils.InitializeQuestionBankAccess(db); err != nil {
		t.Fatalf("重复初始化应成功: %v", err)
	}
	var afterLinks int64
	db.Model(&models.ExamUserQuestionBank{}).Count(&afterLinks)
	if afterLinks != beforeReinitLinks {
		t.Fatalf("重复初始化改变了授权数量: before=%d after=%d", beforeReinitLinks, afterLinks)
	}
	assertNoAccess(t, adminSvc, demoUser.ID)
	assertNoAccess(t, adminSvc, afterInitUser.ID)
	_ = afterInitBank

	// 两个入口、跨权、五题型、混合评分、撤销恢复、长度边界、NULL 旧考试。
	generalFive := mustBank(t, adminSvc, models.QuestionBankCategoryGeneral, "通用五题型")
	mixedBank := mustBank(t, adminSvc, models.QuestionBankCategoryGeneral, "通用混合评分")
	allShortBank := mustBank(t, adminSvc, models.QuestionBankCategoryGeneral, "通用全简答")
	lengthBank := mustBank(t, adminSvc, models.QuestionBankCategoryGeneral, "通用长度边界")
	teaFive := mustBank(t, adminSvc, models.QuestionBankCategoryCompetition, "竞赛五题型")
	csvBank := mustBank(t, adminSvc, models.QuestionBankCategoryGeneral, "CSV导入")

	userA := mustUser(t, adminSvc, "阶段2用户甲")
	userB := mustUser(t, adminSvc, "阶段2用户乙")
	grantBanks(t, adminSvc, userA.ID, generalFive, mixedBank, allShortBank, lengthBank)
	grantBanks(t, adminSvc, userB.ID, teaFive)

	// 新用户/新题库不自动授权（afterInitUser 未授权）。
	assertNoAccess(t, adminSvc, afterInitUser.ID)

	// 入口列表：只看授权 ∩ 分类。
	listPractice, err := examSvc.GetQuestionBanksForEntry(userA.ID, services.EntryPractice)
	if err != nil {
		t.Fatalf("practice 题库列表失败: %v", err)
	}
	if len(listPractice) != 4 {
		t.Fatalf("practice 应只看到 4 个已授权通用题库，实际 %d", len(listPractice))
	}
	for _, bank := range listPractice {
		if bank.CategoryCode != models.QuestionBankCategoryGeneral {
			t.Fatalf("practice 列表混入其他分类: %+v", bank)
		}
	}
	listTeaA, _ := examSvc.GetQuestionBanksForEntry(userA.ID, services.EntryTea)
	if len(listTeaA) != 0 {
		t.Fatalf("用户甲未授权茶题库，tea 列表应为空: %+v", listTeaA)
	}
	listTeaB, _ := examSvc.GetQuestionBanksForEntry(userB.ID, services.EntryTea)
	if len(listTeaB) != 1 || listTeaB[0].ID != teaFive.ID {
		t.Fatalf("用户乙 tea 列表应仅含竞赛五题型: %+v", listTeaB)
	}
	if _, err := examSvc.GetQuestionBanksForEntry(userA.ID, "invalid"); !errors.Is(err, services.ErrInvalidEntry) {
		t.Fatalf("非法 entry 应被拒绝: %v", err)
	}

	// 跨权开始考试不写记录。
	if _, err := examSvc.StartExamForEntry(userA.ID, userA.Name, teaFive.ID, services.EntryTea); !errors.Is(err, services.ErrBankAccessDenied) {
		t.Fatalf("跨权开始考试应拒绝: %v", err)
	}
	var crossCount int64
	db.Model(&models.ExamRecord{}).Where("user_id = ? AND question_bank_type_id = ?", userA.ID, teaFive.ID).Count(&crossCount)
	if crossCount != 0 {
		t.Fatalf("跨权开始考试写入了 %d 条记录", crossCount)
	}

	// 旧 NULL 题库考试：明确拒绝，不写进度，不纳入统计。
	legacy := models.ExamRecord{UserID: userA.ID, UserName: userA.Name, StartTime: time.Now(), Status: "in_progress"}
	if err := db.Create(&legacy).Error; err != nil {
		t.Fatalf("创建旧 NULL 考试失败: %v", err)
	}
	oldEnd := time.Now().Add(-time.Hour)
	legacyCompleted := models.ExamRecord{
		UserID: userA.ID, UserName: userA.Name, StartTime: oldEnd.Add(-time.Minute), EndTime: &oldEnd,
		Status: "completed", CompletedCount: 5, CorrectCount: 3, WrongCount: 2, TotalScore: 7, AccuracyRate: 70,
	}
	if err := db.Create(&legacyCompleted).Error; err != nil {
		t.Fatalf("创建旧 NULL 已完成考试失败: %v", err)
	}
	if _, _, err := examSvc.GetExamInfo(legacy.ID, userA.ID); !errors.Is(err, services.ErrLegacyExamUnsupported) {
		t.Fatalf("旧 NULL 考试信息应拒绝: %v", err)
	}
	if _, _, _, _, err := examSvc.SubmitAnswer(legacy.ID, userA.ID, 1, "A"); !errors.Is(err, services.ErrLegacyExamUnsupported) {
		t.Fatalf("旧 NULL 考试提交应拒绝: %v", err)
	}
	var legacyProgress int64
	db.Model(&models.ExamProgress{}).Where("exam_record_id = ?", legacy.ID).Count(&legacyProgress)
	if legacyProgress != 0 {
		t.Fatalf("旧 NULL 考试写入了进度: %d", legacyProgress)
	}
	statsPractice, err := examSvc.GetExamStatsForEntry(userA.ID, services.EntryPractice)
	if err != nil {
		t.Fatalf("practice 统计失败: %v", err)
	}
	if statsPractice["in_progress_count"].(int64) != 0 {
		t.Fatalf("旧 NULL 考试不应纳入统计: %+v", statsPractice)
	}
	if statsPractice["completed_exams"].(int64) != 0 {
		t.Fatalf("旧 NULL 已完成考试不应纳入统计: %+v", statsPractice)
	}
	adminLegacy, totalLegacy, err := adminSvc.GetExamRecords(1, 100, "", "unclassified")
	if err != nil || totalLegacy < 2 || len(adminLegacy) < 2 {
		t.Fatalf("管理员应仍能看到旧 NULL 考试: total=%d err=%v", totalLegacy, err)
	}
	foundLegacyCompleted := false
	for _, record := range adminLegacy {
		if record.ID == legacyCompleted.ID {
			foundLegacyCompleted = true
			if record.TotalScore != 7 || record.CorrectCount != 3 || record.WrongCount != 2 || record.AccuracyRate != 70 {
				t.Fatalf("旧 NULL 已完成成绩被改写: %+v", record)
			}
			if !record.HasGradedQuestions || record.GradedQuestions != 5 || record.UngradedCount != 0 {
				t.Fatalf("旧 NULL 已完成记录的评分标志错误: %+v", record)
			}
		}
	}
	if !foundLegacyCompleted {
		t.Fatal("管理员列表缺少旧 NULL 已完成考试")
	}

	// 五题型 × 两个入口，同一套答题流程。
	generalQuestions := seedFiveTypes(t, db, generalFive.ID)
	generalExam := mustStart(t, examSvc, userA, generalFive, services.EntryPractice)
	submitFiveTypes(t, examSvc, generalExam, userA.ID, generalQuestions)
	generalResult := mustResult(t, examSvc, generalExam.ID, userA.ID)
	assertResult(t, generalResult, 5, 4, 0, 1, 4, 4, 100)

	teaQuestions := seedFiveTypes(t, db, teaFive.ID)
	teaExam := mustStart(t, examSvc, userB, teaFive, services.EntryTea)
	submitFiveTypes(t, examSvc, teaExam, userB.ID, teaQuestions)
	teaResult := mustResult(t, examSvc, teaExam.ID, userB.ID)
	assertResult(t, teaResult, 5, 4, 0, 1, 4, 4, 100)

	// 2 填空（1 对 1 错）+ 2 简答：已答 4、对 1、错 1、未评 2、正确率 50%、score 1/2。
	fillCorrect := mustQuestion(t, db, mixedBank.ID, 1, models.QuestionTypeFillBlank, "填空一", "", "", "Alpha")
	fillWrong := mustQuestion(t, db, mixedBank.ID, 2, models.QuestionTypeFillBlank, "填空二", "", "", "Beta")
	shortOne := mustQuestion(t, db, mixedBank.ID, 3, models.QuestionTypeShortAnswer, "简答一", "", "", "参考一\n第二行")
	shortTwo := mustQuestion(t, db, mixedBank.ID, 4, models.QuestionTypeShortAnswer, "简答二", "", "", "参考二")
	mixedExam := mustStart(t, examSvc, userA, mixedBank, services.EntryPractice)
	for _, item := range []struct {
		question models.QuestionBank
		answer   string
	}{
		{fillCorrect, "Alpha"},
		{fillWrong, "beta"},
		{shortOne, "我的回答\n保留换行"},
		{shortTwo, "another answer"},
	} {
		if _, _, _, _, err := examSvc.SubmitAnswer(mixedExam.ID, userA.ID, item.question.ID, item.answer); err != nil {
			t.Fatalf("混合评分提交失败: %v", err)
		}
	}
	mixedResult := mustResult(t, examSvc, mixedExam.ID, userA.ID)
	assertResult(t, mixedResult, 4, 1, 1, 2, 1, 2, 50)
	var shortProgress models.ExamProgress
	db.Where("exam_record_id = ? AND question_id = ?", mixedExam.ID, shortOne.ID).First(&shortProgress)
	if shortProgress.IsGraded || shortProgress.IsCorrect {
		t.Fatalf("简答题不应被自动评分: %+v", shortProgress)
	}
	if shortProgress.UserAnswer != "我的回答\n保留换行" {
		t.Fatalf("简答题原文未保留: %q", shortProgress.UserAnswer)
	}

	// 全简答：完成但不显示 0 分/0%。
	shortA := mustQuestion(t, db, allShortBank.ID, 1, models.QuestionTypeShortAnswer, "全简答一", "", "", "参考甲")
	shortB := mustQuestion(t, db, allShortBank.ID, 2, models.QuestionTypeShortAnswer, "全简答二", "", "", "参考乙")
	allShortExam := mustStart(t, examSvc, userA, allShortBank, services.EntryPractice)
	examSvc.SubmitAnswer(allShortExam.ID, userA.ID, shortA.ID, "回答甲")
	examSvc.SubmitAnswer(allShortExam.ID, userA.ID, shortB.ID, "回答乙")
	allShortResult := mustResult(t, examSvc, allShortExam.ID, userA.ID)
	if allShortResult.CompletedCount != 2 || allShortResult.CorrectCount != 0 || allShortResult.WrongCount != 0 || allShortResult.UngradedCount != 2 {
		t.Fatalf("全简答计数错误: %+v", allShortResult)
	}
	if allShortResult.GradedQuestions != 0 || allShortResult.HasGradedQuestions || allShortResult.TotalScore != 0 {
		t.Fatalf("全简答不应有自动评分口径: %+v", allShortResult)
	}

	// 撤销授权后旧记录保留但立即不可访问；重新授权恢复。
	grantBanks(t, adminSvc, userA.ID, generalFive, allShortBank, lengthBank)
	if _, _, err := examSvc.GetExamInfo(mixedExam.ID, userA.ID); !errors.Is(err, services.ErrBankAccessDenied) {
		t.Fatalf("撤权后应拒绝旧考试: %v", err)
	}
	beforeRevokeProgress := countExamProgress(t, db, mixedExam.ID)
	if _, _, _, _, err := examSvc.SubmitAnswer(mixedExam.ID, userA.ID, fillCorrect.ID, "Alpha"); !errors.Is(err, services.ErrBankAccessDenied) {
		t.Fatalf("撤权后提交应拒绝: %v", err)
	}
	if after := countExamProgress(t, db, mixedExam.ID); after != beforeRevokeProgress {
		t.Fatalf("撤权请求写入了进度: before=%d after=%d", beforeRevokeProgress, after)
	}
	grantBanks(t, adminSvc, userA.ID, generalFive, mixedBank, allShortBank, lengthBank)
	if _, _, err := examSvc.GetExamInfo(mixedExam.ID, userA.ID); err != nil {
		t.Fatalf("重新授权后旧记录应恢复访问: %v", err)
	}

	// 4000/4001 含 emoji：4001 拒绝且 0 写入；4000 保存原长度。
	mustQuestion(t, db, lengthBank.ID, 1, models.QuestionTypeFillBlank, "长度题", "", "", "标准答案")
	lengthExam := mustStart(t, examSvc, userA, lengthBank, services.EntryPractice)
	lengthQuestion := models.QuestionBank{}
	db.Where("question_bank_type_id = ?", lengthBank.ID).First(&lengthQuestion)
	tooLong := strings.Repeat("😀", 4001)
	if _, _, _, _, err := examSvc.SubmitAnswer(lengthExam.ID, userA.ID, lengthQuestion.ID, tooLong); !errors.Is(err, services.ErrTextAnswerTooLong) {
		t.Fatalf("4001 字应拒绝: %v", err)
	}
	if countExamProgress(t, db, lengthExam.ID) != 0 {
		t.Fatal("4001 字拒绝时写入了进度")
	}
	exact := strings.Repeat("😀", 4000)
	if _, _, completed, _, err := examSvc.SubmitAnswer(lengthExam.ID, userA.ID, lengthQuestion.ID, exact); err != nil || !completed {
		t.Fatalf("4000 字应接受并完成: completed=%v err=%v", completed, err)
	}
	saved := models.ExamProgress{}
	db.Where("exam_record_id = ? AND question_id = ?", lengthExam.ID, lengthQuestion.ID).First(&saved)
	if utf8.RuneCountInString(saved.UserAnswer) != 4000 {
		t.Fatalf("4000 字保存后长度变化: %d", utf8.RuneCountInString(saved.UserAnswer))
	}

	// 后台记录口径一致。
	mixedRecords, mixedTotal, err := adminSvc.GetExamRecords(1, 10, "", strconv.Itoa(int(mixedBank.ID)))
	if err != nil || mixedTotal != 1 || len(mixedRecords) != 1 {
		t.Fatalf("后台混合记录查询失败: total=%d err=%v", mixedTotal, err)
	}
	if mixedRecords[0].GradedQuestions != 2 || mixedRecords[0].UngradedCount != 2 ||
		mixedRecords[0].CorrectCount != 1 || mixedRecords[0].WrongCount != 1 || mixedRecords[0].AccuracyRate != 50 {
		t.Fatalf("后台混合记录口径错误: %+v", mixedRecords[0])
	}
	shortRecords, _, _ := adminSvc.GetExamRecords(1, 10, "", strconv.Itoa(int(allShortBank.ID)))
	if len(shortRecords) != 1 || shortRecords[0].HasGradedQuestions || shortRecords[0].GradedQuestions != 0 || shortRecords[0].UngradedCount != 2 {
		t.Fatalf("后台全简答口径错误: %+v", shortRecords)
	}

	// 权限接口：普通用户禁止分配；非法 ID 整批拒绝；空数组撤销；管理员可恢复。
	secret := "stage2-test-secret"
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	adminGroup := api.Group("/admin", middleware.AdminAuth(secret))
	adminHandler := NewAdminHandler(adminSvc)
	adminGroup.GET("/users/:id/question-banks", adminHandler.GetUserQuestionBanks)
	adminGroup.PUT("/users/:id/question-banks", adminHandler.ReplaceUserQuestionBanks)
	adminGroup.POST("/questions/import", adminHandler.ImportQuestions)
	userOnly := api.Group("/exam", middleware.JWTAuth(secret), middleware.UserOnly())
	userOnly.GET("/question-banks", NewExamHandler(examSvc).GetQuestionBanks)

	adminToken := mustToken(t, 1, "admin", "admin", secret)
	userToken := mustToken(t, userB.ID, userB.Name, "user", secret)

	env := performJSON(t, router, http.MethodGet, fmt.Sprintf("/api/admin/users/%d/question-banks", userB.ID), nil, userToken)
	if env.Code != 403 {
		t.Fatalf("普通用户读取他人授权应 403: %+v", env)
	}
	env = performJSON(t, router, http.MethodPut, fmt.Sprintf("/api/admin/users/%d/question-banks", userB.ID),
		map[string]interface{}{"question_bank_ids": []uint{999999}}, adminToken)
	if env.Code != 400 {
		t.Fatalf("不存在题库 ID 应整批拒绝: %+v", env)
	}
	access, err := adminSvc.GetUserQuestionBanks(userB.ID)
	if err != nil || len(access.QuestionBankIDs) != 1 || access.QuestionBankIDs[0] != teaFive.ID {
		t.Fatalf("拒绝非法 ID 后原授权应保持: %+v err=%v", access, err)
	}
	for name, payload := range map[string]interface{}{
		"missing field": map[string]interface{}{},
		"null field":    json.RawMessage("null"),
	} {
		env = performJSON(t, router, http.MethodPut, fmt.Sprintf("/api/admin/users/%d/question-banks", userB.ID), payload, adminToken)
		if env.Code != 400 {
			t.Fatalf("%s 应拒绝且不撤销授权: %+v", name, env)
		}
		access, err = adminSvc.GetUserQuestionBanks(userB.ID)
		if err != nil || len(access.QuestionBankIDs) != 1 || access.QuestionBankIDs[0] != teaFive.ID {
			t.Fatalf("%s 后原授权应保持: %+v err=%v", name, access, err)
		}
	}
	env = performJSON(t, router, http.MethodPut, fmt.Sprintf("/api/admin/users/%d/question-banks", userB.ID),
		map[string]interface{}{"question_bank_ids": []uint{}}, adminToken)
	if env.Code != 0 {
		t.Fatalf("空数组撤销应成功: %+v", env)
	}
	access, _ = adminSvc.GetUserQuestionBanks(userB.ID)
	if len(access.QuestionBankIDs) != 0 {
		t.Fatalf("空数组应撤销全部授权: %+v", access)
	}
	env = performJSON(t, router, http.MethodPut, fmt.Sprintf("/api/admin/users/%d/question-banks", userB.ID),
		map[string]interface{}{"question_bank_ids": []uint{teaFive.ID}}, adminToken)
	if env.Code != 0 {
		t.Fatalf("管理员恢复授权失败: %+v", env)
	}

	// 用户接口必须 role=user；admin 固定 ID=1 不能当答题人。
	env = performJSON(t, router, http.MethodGet, "/api/exam/question-banks?entry=practice", nil, adminToken)
	if env.Code != 403 {
		t.Fatalf("admin token 访问答题接口应拒绝: %+v", env)
	}
	env = performJSON(t, router, http.MethodGet, "/api/exam/question-banks?entry=practice", nil, userToken)
	if env.Code != 0 {
		t.Fatalf("user token 应可访问题库列表: %+v", env)
	}

	// CSV：任一坏行整批拒绝且不 replace 旧库；合法五种题型整批导入。
	original := mustQuestion(t, db, csvBank.ID, 1, models.QuestionTypeSingleChoice, "原始题目", "甲", "乙", "A")
	badCSV := "题号,题库,题型编码,题型名称,题干,选项A,选项B,选项C,选项D,选项E,参考答案\n" +
		"1,CSV,multiple_choice,多选题,合法行,甲,乙,丙,,,AC\n" +
		"2,CSV,ordering,排序题,非法行,甲,乙,,,,\n"
	env = performCSVImport(t, router, adminToken, csvBank.ID, badCSV)
	if env.Code != 400 {
		t.Fatalf("部分坏 CSV 应整批拒绝: %+v", env)
	}
	var csvCount int64
	db.Model(&models.QuestionBank{}).Where("question_bank_type_id = ?", csvBank.ID).Count(&csvCount)
	var originalAfter models.QuestionBank
	db.First(&originalAfter, original.ID)
	if csvCount != 1 || originalAfter.QuestionText != "原始题目" {
		t.Fatalf("坏 CSV 不得 replace 旧库: count=%d text=%q", csvCount, originalAfter.QuestionText)
	}

	goodCSV := "题号,题库,题型编码,题型名称,题干,选项A,选项B,选项C,选项D,选项E,参考答案\n" +
		"1,CSV,single_choice,单选题,单选,甲,乙,,, ,A\n" +
		"2,CSV,multiple_choice,多选题,多选,甲,乙,丙,,,AC\n" +
		"3,CSV,true_false,判断题,判断,对,错,,, ,B\n" +
		"4,CSV,fill_blank,填空题,填空,,,,,, Tea\n" +
		"5,CSV,short_answer,简答题,简答,,,,,,\"参考第一行\n参考第二行\"\n"
	env = performCSVImport(t, router, adminToken, csvBank.ID, goodCSV)
	if env.Code != 0 {
		t.Fatalf("合法五题型 CSV 导入失败: %+v", env)
	}
	db.Model(&models.QuestionBank{}).Where("question_bank_type_id = ?", csvBank.ID).Count(&csvCount)
	if csvCount != 5 {
		t.Fatalf("合法 CSV 应导入 5 题，实际 %d", csvCount)
	}
	var shortCSV models.QuestionBank
	db.Where("question_bank_type_id = ? AND question_type_code = ?", csvBank.ID, models.QuestionTypeShortAnswer).First(&shortCSV)
	if shortCSV.CorrectAnswer != "参考第一行\n参考第二行" {
		t.Fatalf("CSV 多行参考答案未保留: %q", shortCSV.CorrectAnswer)
	}
}

func mustBank(t *testing.T, svc *services.AdminService, category, name string) models.QuestionBankType {
	t.Helper()
	bank, err := svc.CreateQuestionBank(services.QuestionBankTypeInput{CategoryCode: category, Name: name, Status: 1})
	if err != nil {
		t.Fatalf("创建题库 %s 失败: %v", name, err)
	}
	return *bank
}

func mustUser(t *testing.T, svc *services.AdminService, name string) services.AdminExamUser {
	t.Helper()
	user, err := svc.CreateExamUser(services.ExamUserInput{Name: name, Password: "pass1234"})
	if err != nil {
		t.Fatalf("创建用户 %s 失败: %v", name, err)
	}
	return *user
}

func grantBanks(t *testing.T, svc *services.AdminService, userID uint, banks ...models.QuestionBankType) {
	t.Helper()
	ids := make([]uint, 0, len(banks))
	for _, bank := range banks {
		ids = append(ids, bank.ID)
	}
	if err := svc.ReplaceUserQuestionBanks(userID, ids); err != nil {
		t.Fatalf("保存用户授权失败: %v", err)
	}
}

func assertNoAccess(t *testing.T, svc *services.AdminService, userID uint) {
	t.Helper()
	access, err := svc.GetUserQuestionBanks(userID)
	if err != nil {
		t.Fatalf("读取授权失败: %v", err)
	}
	if len(access.QuestionBankIDs) != 0 {
		t.Fatalf("用户 %d 不应自动获得授权: %+v", userID, access)
	}
}

func mustQuestion(t *testing.T, db *gorm.DB, bankID uint, no int, questionType, text, optionA, optionB, answer string) models.QuestionBank {
	t.Helper()
	question := models.QuestionBank{
		QuestionBankTypeID: &bankID,
		QuestionNo:         no,
		QuestionTypeCode:   questionType,
		QuestionTypeName:   services.QuestionTypeName(questionType),
		QuestionText:       text,
		OptionA:            optionA,
		OptionB:            optionB,
		CorrectAnswer:      answer,
	}
	if err := db.Create(&question).Error; err != nil {
		t.Fatalf("创建题目失败: %v", err)
	}
	return question
}

func seedFiveTypes(t *testing.T, db *gorm.DB, bankID uint) map[string]models.QuestionBank {
	t.Helper()
	return map[string]models.QuestionBank{
		models.QuestionTypeSingleChoice:   mustQuestion(t, db, bankID, 1, models.QuestionTypeSingleChoice, "单选", "甲", "乙", "A"),
		models.QuestionTypeMultipleChoice: mustQuestion(t, db, bankID, 2, models.QuestionTypeMultipleChoice, "多选", "甲", "乙", "AC"),
		models.QuestionTypeTrueFalse:      mustQuestion(t, db, bankID, 3, models.QuestionTypeTrueFalse, "判断", "对", "错", "B"),
		models.QuestionTypeFillBlank:      mustQuestion(t, db, bankID, 4, models.QuestionTypeFillBlank, "填空", "", "", "茶\nTea"),
		models.QuestionTypeShortAnswer:    mustQuestion(t, db, bankID, 5, models.QuestionTypeShortAnswer, "简答", "", "", "参考答案\n第二行"),
	}
}

func submitFiveTypes(t *testing.T, svc *services.ExamService, exam *models.ExamRecord, userID uint, questions map[string]models.QuestionBank) {
	t.Helper()
	order := []string{models.QuestionTypeSingleChoice, models.QuestionTypeMultipleChoice, models.QuestionTypeTrueFalse, models.QuestionTypeFillBlank, models.QuestionTypeShortAnswer}
	answers := map[string]string{
		models.QuestionTypeSingleChoice:   "A",
		models.QuestionTypeMultipleChoice: "CA",
		models.QuestionTypeTrueFalse:      "B",
		models.QuestionTypeFillBlank:      " 茶\r\nTea ",
		models.QuestionTypeShortAnswer:    "我的回答\n第二行",
	}
	for i, questionType := range order {
		_, _, completed, _, err := svc.SubmitAnswer(exam.ID, userID, questions[questionType].ID, answers[questionType])
		if err != nil {
			t.Fatalf("提交 %s 失败: %v", questionType, err)
		}
		if (i == len(order)-1) != completed {
			t.Fatalf("完成状态错误 type=%s completed=%v", questionType, completed)
		}
	}
}

func mustStart(t *testing.T, svc *services.ExamService, user services.AdminExamUser, bank models.QuestionBankType, entry string) *models.ExamRecord {
	t.Helper()
	exam, err := svc.StartExamForEntry(user.ID, user.Name, bank.ID, entry)
	if err != nil {
		t.Fatalf("开始考试失败 bank=%s: %v", bank.Name, err)
	}
	return exam
}

func mustResult(t *testing.T, svc *services.ExamService, examID, userID uint) *models.ExamRecord {
	t.Helper()
	result, err := svc.GetExamResultDetail(examID, userID)
	if err != nil {
		t.Fatalf("读取结果失败: %v", err)
	}
	return result
}

func assertResult(t *testing.T, result *models.ExamRecord, completed, correct, wrong, ungraded, score, graded int, accuracy float64) {
	t.Helper()
	if result.CompletedCount != completed || result.CorrectCount != correct || result.WrongCount != wrong ||
		result.UngradedCount != ungraded || result.TotalScore != score || result.GradedQuestions != graded ||
		result.AccuracyRate != accuracy || !result.HasGradedQuestions {
		t.Fatalf("结果口径错误: %+v", result)
	}
}

func countExamProgress(t *testing.T, db *gorm.DB, examID uint) int64 {
	t.Helper()
	var count int64
	if err := db.Model(&models.ExamProgress{}).Where("exam_record_id = ?", examID).Count(&count).Error; err != nil {
		t.Fatalf("统计进度失败: %v", err)
	}
	return count
}

func assertColumnType(t *testing.T, db *gorm.DB, table, column, want string) {
	t.Helper()
	var dataType string
	if err := db.Raw(`SELECT DATA_TYPE FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?`, table, column).Scan(&dataType).Error; err != nil {
		t.Fatalf("读取列类型失败: %v", err)
	}
	if !strings.Contains(strings.ToLower(dataType), want) {
		t.Fatalf("%s.%s 类型为 %s，期望包含 %s", table, column, dataType, want)
	}
}

func assertColumnCharset(t *testing.T, db *gorm.DB, table, column, want string) {
	t.Helper()
	var charset string
	if err := db.Raw(`SELECT CHARACTER_SET_NAME FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?`, table, column).Scan(&charset).Error; err != nil {
		t.Fatalf("读取列字符集失败: %v", err)
	}
	if !strings.EqualFold(charset, want) {
		t.Fatalf("%s.%s 字符集为 %s，期望 %s", table, column, charset, want)
	}
}

func mustToken(t *testing.T, id uint, name, role, secret string) string {
	t.Helper()
	token, err := middleware.GenerateToken(id, name, role, &config.JWTConfig{Secret: secret})
	if err != nil {
		t.Fatalf("生成 Token 失败: %v", err)
	}
	return token
}

func performJSON(t *testing.T, handler http.Handler, method, path string, body interface{}, token string) stage2Envelope {
	t.Helper()
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("序列化请求失败: %v", err)
		}
		reader = bytes.NewReader(data)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	var env stage2Envelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &env); err != nil {
		t.Fatalf("解析响应失败: %v body=%s", err, recorder.Body.String())
	}
	return env
}

func performCSVImport(t *testing.T, handler http.Handler, token string, bankID uint, content string) stage2Envelope {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("mode", "replace"); err != nil {
		t.Fatalf("写 mode 失败: %v", err)
	}
	if err := writer.WriteField("question_bank_id", strconv.Itoa(int(bankID))); err != nil {
		t.Fatalf("写题库 ID 失败: %v", err)
	}
	file, err := writer.CreateFormFile("file", "stage2.csv")
	if err != nil {
		t.Fatalf("创建文件字段失败: %v", err)
	}
	if _, err := file.Write([]byte(content)); err != nil {
		t.Fatalf("写 CSV 内容失败: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("关闭 multipart 失败: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/admin/questions/import", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	var env stage2Envelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &env); err != nil {
		t.Fatalf("解析导入响应失败: %v body=%s", err, recorder.Body.String())
	}
	return env
}

func dsnDatabaseName(dsn string) string {
	slash := strings.LastIndex(dsn, "/")
	if slash < 0 {
		return ""
	}
	name := dsn[slash+1:]
	if query := strings.Index(name, "?"); query >= 0 {
		name = name[:query]
	}
	return name
}

func withDSNParam(dsn, key, value string) string {
	if strings.Contains(dsn, key+"=") {
		return dsn
	}
	separator := "?"
	if strings.Contains(dsn, "?") {
		separator = "&"
	}
	return dsn + separator + key + "=" + value
}
