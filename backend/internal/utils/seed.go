package utils

import (
	"tea-exam/internal/models"
	"tea-exam/internal/security"
	"tea-exam/internal/services"

	"gorm.io/gorm"
)

// SeedData 初始化基础数据
func SeedData(db *gorm.DB, adminInitialPassword string, seedDemoUsers bool) error {
	// 初始化管理员配置
	adminService := services.NewAdminService(db)
	if err := adminService.InitAdminConfig(adminInitialPassword); err != nil {
		return err
	}

	// 初始化示例答题用户（可选，可以通过数据库直接维护）
	if seedDemoUsers {
		if err := seedExamUsers(db); err != nil {
			return err
		}
	}

	if err := seedQuestionBankTypes(db); err != nil {
		return err
	}

	return nil
}

func seedQuestionBankTypes(db *gorm.DB) error {
	banks := []models.QuestionBankType{
		{Code: "competition-sensory-review", CategoryCode: models.QuestionBankCategoryCompetition, Name: "茶叶感官审评", SortOrder: 10, Status: 1},
		{Code: "competition-processing", CategoryCode: models.QuestionBankCategoryCompetition, Name: "茶叶加工", SortOrder: 20, Status: 1},
		{Code: "certification-tea-appraiser-34", CategoryCode: models.QuestionBankCategoryCertification, Name: "评茶员-3级/4级", SortOrder: 10, Status: 1},
		{Code: "certification-tea-processing-worker-34", CategoryCode: models.QuestionBankCategoryCertification, Name: "茶叶加工工-3级/4级", SortOrder: 20, Status: 1},
		{Code: "certification-tea-appraiser-technician-2", CategoryCode: models.QuestionBankCategoryCertification, Name: "评茶技师-2级", SortOrder: 30, Status: 1},
		{Code: "certification-tea-processing-technician-2", CategoryCode: models.QuestionBankCategoryCertification, Name: "茶叶加工技师-2级", SortOrder: 40, Status: 1},
	}

	for _, bank := range banks {
		if err := db.Where("code = ?", bank.Code).FirstOrCreate(&bank).Error; err != nil {
			return err
		}
	}

	return nil
}

// seedExamUsers 初始化示例答题用户
func seedExamUsers(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.ExamUser{}).Count(&count).Error; err != nil {
		return err
	}

	// 如果没有用户，创建几个示例用户
	if count == 0 {
		password, err := security.HashPassword("123456")
		if err != nil {
			return err
		}
		users := []models.ExamUser{
			{Name: "张三", Password: password, Status: 1},
			{Name: "李四", Password: password, Status: 1},
			{Name: "王五", Password: password, Status: 1},
		}

		for _, user := range users {
			if err := db.Create(&user).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
