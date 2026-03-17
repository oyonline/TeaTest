package utils

import (
	"tea-exam/internal/models"
	"tea-exam/internal/services"

	"gorm.io/gorm"
)

// SeedData 初始化基础数据
func SeedData(db *gorm.DB) error {
	// 初始化管理员配置
	adminService := services.NewAdminService(db)
	if err := adminService.InitAdminConfig("123456"); err != nil {
		return err
	}

	// 初始化示例答题用户（可选，可以通过数据库直接维护）
	if err := seedExamUsers(db); err != nil {
		return err
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
		users := []models.ExamUser{
			{Name: "张三", Password: "123456", Status: 1},
			{Name: "李四", Password: "123456", Status: 1},
			{Name: "王五", Password: "123456", Status: 1},
		}

		for _, user := range users {
			if err := db.Create(&user).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
