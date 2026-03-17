package services

import (
	"errors"

	"gorm.io/gorm"
	"tea-exam/internal/models"
)

// AdminService 管理员服务
type AdminService struct {
	db *gorm.DB
}

// NewAdminService 创建管理员服务
func NewAdminService(db *gorm.DB) *AdminService {
	return &AdminService{db: db}
}

// Login 管理员登录
func (s *AdminService) Login(password string) error {
	var config models.AdminConfig
	err := s.db.First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("管理员配置不存在")
		}
		return err
	}

	if config.AdminPassword != password {
		return errors.New("密码错误")
	}

	return nil
}

// GetAdminConfig 获取管理员配置
func (s *AdminService) GetAdminConfig() (*models.AdminConfig, error) {
	var config models.AdminConfig
	err := s.db.First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// InitAdminConfig 初始化管理员配置
func (s *AdminService) InitAdminConfig(password string) error {
	var count int64
	if err := s.db.Model(&models.AdminConfig{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil // 已存在，不重复初始化
	}

	config := models.AdminConfig{
		AdminPassword: password,
	}
	return s.db.Create(&config).Error
}

// GetExamRecords 获取考试记录列表
func (s *AdminService) GetExamRecords(page, pageSize int, keyword string) ([]models.ExamRecord, int64, error) {
	var records []models.ExamRecord
	var total int64

	offset := (page - 1) * pageSize

	query := s.db.Model(&models.ExamRecord{})
	if keyword != "" {
		query = query.Where("user_name LIKE ?", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// GetBankStats 获取题库统计信息
func (s *AdminService) GetBankStats() (map[string]interface{}, error) {
	var totalQuestions int64
	if err := s.db.Model(&models.QuestionBank{}).Count(&totalQuestions).Error; err != nil {
		return nil, err
	}

	// 获取考试记录统计
	var totalExams int64
	var completedExams int64
	var inProgressExams int64

	s.db.Model(&models.ExamRecord{}).Count(&totalExams)
	s.db.Model(&models.ExamRecord{}).Where("status = 'completed'").Count(&completedExams)
	s.db.Model(&models.ExamRecord{}).Where("status = 'in_progress'").Count(&inProgressExams)

	return map[string]interface{}{
		"total_questions":  totalQuestions,
		"total_exams":      totalExams,
		"completed_exams":  completedExams,
		"in_progress_exams": inProgressExams,
	}, nil
}

// ImportQuestions 导入题目（覆盖模式）
func (s *AdminService) ImportQuestions(questions []models.QuestionBank, mode string) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if mode == "replace" {
		// 覆盖模式：清空旧题库
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.QuestionBank{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 批量插入新题目
	if len(questions) > 0 {
		if err := tx.CreateInBatches(questions, 100).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}
