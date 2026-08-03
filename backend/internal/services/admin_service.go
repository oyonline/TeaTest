package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"tea-exam/internal/models"
)

type AdminQuestionBank struct {
	ID            uint      `json:"id"`
	CategoryCode  string    `json:"category_code"`
	CategoryName  string    `json:"category_name"`
	Name          string    `json:"name"`
	SortOrder     int       `json:"sort_order"`
	Status        int       `json:"status"`
	QuestionCount int64     `json:"question_count"`
	ExamCount     int64     `json:"exam_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type QuestionBankTypeInput struct {
	CategoryCode string `json:"category_code"`
	Name         string `json:"name"`
	SortOrder    int    `json:"sort_order"`
	Status       int    `json:"status"`
}

type AdminExamUser struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Status    int       `json:"status"`
	ExamCount int64     `json:"exam_count"`
	CreatedAt time.Time `json:"created_at"`
}

type ExamUserInput struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

// AdminService 管理员服务
type AdminService struct {
	db *gorm.DB
}

func NewAdminService(db *gorm.DB) *AdminService {
	return &AdminService{db: db}
}

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

func (s *AdminService) GetAdminConfig() (*models.AdminConfig, error) {
	var config models.AdminConfig
	if err := s.db.First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (s *AdminService) InitAdminConfig(password string) error {
	var count int64
	if err := s.db.Model(&models.AdminConfig{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return s.db.Create(&models.AdminConfig{AdminPassword: password}).Error
}

func validateExamUserInput(input *ExamUserInput) error {
	input.Name = strings.TrimSpace(input.Name)
	input.Password = strings.TrimSpace(input.Password)
	if input.Name == "" {
		return errors.New("答题人姓名不能为空")
	}
	if len([]rune(input.Name)) > 50 {
		return errors.New("答题人姓名不能超过50个字")
	}
	if input.Password == "" {
		return errors.New("登录密码不能为空")
	}
	if len([]rune(input.Password)) > 100 {
		return errors.New("登录密码不能超过100个字符")
	}
	return nil
}

func (s *AdminService) ListExamUsers(page, pageSize int, keyword string) ([]AdminExamUser, int64, error) {
	keyword = strings.TrimSpace(keyword)
	query := s.db.Model(&models.ExamUser{})
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []models.ExamUser
	if err := query.Order("created_at DESC, id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	type examCountRow struct {
		UserID uint
		Count  int64
	}
	examCounts := make(map[uint]int64, len(users))
	if len(users) > 0 {
		userIDs := make([]uint, 0, len(users))
		for _, user := range users {
			userIDs = append(userIDs, user.ID)
		}
		var countRows []examCountRow
		if err := s.db.Model(&models.ExamRecord{}).
			Select("user_id, COUNT(*) AS count").
			Where("user_id IN ?", userIDs).
			Group("user_id").
			Scan(&countRows).Error; err != nil {
			return nil, 0, err
		}
		for _, row := range countRows {
			examCounts[row.UserID] = row.Count
		}
	}

	result := make([]AdminExamUser, 0, len(users))
	for _, user := range users {
		result = append(result, AdminExamUser{
			ID:        user.ID,
			Name:      user.Name,
			Status:    user.Status,
			ExamCount: examCounts[user.ID],
			CreatedAt: user.CreatedAt,
		})
	}
	return result, total, nil
}

func (s *AdminService) CreateExamUser(input ExamUserInput) (*AdminExamUser, error) {
	if err := validateExamUserInput(&input); err != nil {
		return nil, err
	}

	user := models.ExamUser{Name: input.Name, Password: input.Password, Status: 1}
	if err := s.db.Create(&user).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return nil, errors.New("该答题人已存在")
		}
		return nil, err
	}
	return &AdminExamUser{
		ID:        user.ID,
		Name:      user.Name,
		Status:    user.Status,
		ExamCount: 0,
		CreatedAt: user.CreatedAt,
	}, nil
}

func validateQuestionBankInput(input *QuestionBankTypeInput) error {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return errors.New("题库名称不能为空")
	}
	if input.CategoryCode != models.QuestionBankCategoryCompetition && input.CategoryCode != models.QuestionBankCategoryCertification {
		return errors.New("一级分类只能选择竞赛题库或考级题库")
	}
	if input.Status != 0 && input.Status != 1 {
		return errors.New("题库状态无效")
	}
	return nil
}

func (s *AdminService) ListQuestionBanks() ([]AdminQuestionBank, error) {
	var banks []models.QuestionBankType
	if err := s.db.Order("CASE category_code WHEN 'competition' THEN 0 ELSE 1 END, sort_order ASC, id ASC").Find(&banks).Error; err != nil {
		return nil, err
	}

	result := make([]AdminQuestionBank, 0, len(banks))
	for _, bank := range banks {
		var questionCount int64
		if err := s.db.Model(&models.QuestionBank{}).Where("question_bank_type_id = ?", bank.ID).Count(&questionCount).Error; err != nil {
			return nil, err
		}
		var examCount int64
		if err := s.db.Model(&models.ExamRecord{}).Where("question_bank_type_id = ?", bank.ID).Count(&examCount).Error; err != nil {
			return nil, err
		}
		result = append(result, AdminQuestionBank{
			ID:            bank.ID,
			CategoryCode:  bank.CategoryCode,
			CategoryName:  QuestionBankCategoryName(bank.CategoryCode),
			Name:          bank.Name,
			SortOrder:     bank.SortOrder,
			Status:        bank.Status,
			QuestionCount: questionCount,
			ExamCount:     examCount,
			CreatedAt:     bank.CreatedAt,
			UpdatedAt:     bank.UpdatedAt,
		})
	}
	return result, nil
}

func (s *AdminService) CreateQuestionBank(input QuestionBankTypeInput) (*models.QuestionBankType, error) {
	if err := validateQuestionBankInput(&input); err != nil {
		return nil, err
	}
	bank := models.QuestionBankType{
		Code:         fmt.Sprintf("custom-%d", time.Now().UnixNano()),
		CategoryCode: input.CategoryCode,
		Name:         input.Name,
		SortOrder:    input.SortOrder,
		Status:       input.Status,
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&bank).Error; err != nil {
			return err
		}
		// GORM 会把带 default 标签的零值写成默认值；显式恢复管理员选择的停用状态。
		if input.Status == 0 {
			if err := tx.Model(&bank).Update("status", 0).Error; err != nil {
				return err
			}
			bank.Status = 0
		}
		return nil
	})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return nil, errors.New("同一分类下已存在同名题库")
		}
		return nil, err
	}
	return &bank, nil
}

func (s *AdminService) UpdateQuestionBank(id uint, input QuestionBankTypeInput) (*models.QuestionBankType, error) {
	if err := validateQuestionBankInput(&input); err != nil {
		return nil, err
	}
	var bank models.QuestionBankType
	if err := s.db.First(&bank, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("题库不存在")
		}
		return nil, err
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&bank).Updates(map[string]interface{}{
			"category_code": input.CategoryCode,
			"name":          input.Name,
			"sort_order":    input.SortOrder,
			"status":        input.Status,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&models.QuestionBank{}).Where("question_bank_type_id = ?", bank.ID).Update("bank_name", input.Name).Error
	})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return nil, errors.New("同一分类下已存在同名题库")
		}
		return nil, err
	}
	if err := s.db.First(&bank, id).Error; err != nil {
		return nil, err
	}
	return &bank, nil
}

// GetExamRecords 获取考试记录，bankFilter 可为具体题库 ID 或 unclassified。
func (s *AdminService) GetExamRecords(page, pageSize int, keyword, bankFilter string) ([]models.ExamRecord, int64, error) {
	var records []models.ExamRecord
	var total int64
	query := s.db.Model(&models.ExamRecord{})
	if keyword != "" {
		query = query.Where("user_name LIKE ?", "%"+keyword+"%")
	}
	if bankFilter == "unclassified" {
		query = query.Where("question_bank_type_id IS NULL")
	} else if bankFilter != "" {
		bankID, err := strconv.ParseUint(bankFilter, 10, 32)
		if err != nil {
			return nil, 0, errors.New("题库筛选条件无效")
		}
		query = query.Where("question_bank_type_id = ?", uint(bankID))
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

func (s *AdminService) GetBankStats() (map[string]interface{}, error) {
	var totalQuestions, unclassifiedQuestions, totalExams, completedExams, inProgressExams, activeBanks int64
	if err := s.db.Model(&models.QuestionBank{}).Count(&totalQuestions).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&models.QuestionBank{}).Where("question_bank_type_id IS NULL").Count(&unclassifiedQuestions).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&models.ExamRecord{}).Count(&totalExams).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&models.ExamRecord{}).Where("status = 'completed'").Count(&completedExams).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&models.ExamRecord{}).Where("status = 'in_progress'").Count(&inProgressExams).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&models.QuestionBankType{}).Where("status = ?", 1).Count(&activeBanks).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"total_questions":        totalQuestions,
		"unclassified_questions": unclassifiedQuestions,
		"active_banks":           activeBanks,
		"total_exams":            totalExams,
		"completed_exams":        completedExams,
		"in_progress_exams":      inProgressExams,
	}, nil
}

// ImportQuestions 导入题目；覆盖模式只清空选中的具体题库。
func (s *AdminService) ImportQuestions(questions []models.QuestionBank, mode string, bankID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var bank models.QuestionBankType
		if err := tx.First(&bank, bankID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("所选题库不存在")
			}
			return err
		}
		if mode == "replace" {
			if err := tx.Where("question_bank_type_id = ?", bankID).Delete(&models.QuestionBank{}).Error; err != nil {
				return err
			}
		}
		for i := range questions {
			questions[i].QuestionBankTypeID = &bank.ID
			questions[i].BankName = bank.Name
		}
		if len(questions) > 0 {
			if err := tx.CreateInBatches(questions, 100).Error; err != nil {
				if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
					return errors.New("所选题库中存在重复题号，请检查导入文件或改用覆盖模式")
				}
				return err
			}
		}
		return nil
	})
}

// ListQuestions 获取管理员题目列表。
func (s *AdminService) ListQuestions(page, pageSize int, keyword, bankFilter string) ([]models.QuestionBank, int64, error) {
	var questions []models.QuestionBank
	var total int64
	query := s.db.Model(&models.QuestionBank{}).Preload("QuestionBankType")
	if keyword != "" {
		query = query.Where("question_text LIKE ? OR CAST(question_no AS CHAR) = ?", "%"+keyword+"%", keyword)
	}
	if bankFilter == "unclassified" {
		query = query.Where("question_bank_type_id IS NULL")
	} else if bankFilter != "" {
		bankID, err := strconv.ParseUint(bankFilter, 10, 32)
		if err != nil {
			return nil, 0, errors.New("题库筛选条件无效")
		}
		query = query.Where("question_bank_type_id = ?", uint(bankID))
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("question_no ASC, id ASC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&questions).Error; err != nil {
		return nil, 0, err
	}
	return questions, total, nil
}

// ReclassifyQuestions 批量归类题目；allUnclassified 为 true 时归类全部未分类题目。
func (s *AdminService) ReclassifyQuestions(questionIDs []uint, bankID uint, allUnclassified bool) (int64, error) {
	if !allUnclassified && len(questionIDs) == 0 {
		return 0, errors.New("请选择要归类的题目")
	}

	var updatedCount int64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var bank models.QuestionBankType
		if err := tx.First(&bank, bankID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("目标题库不存在")
			}
			return err
		}

		query := tx.Model(&models.QuestionBank{})
		if allUnclassified {
			query = query.Where("question_bank_type_id IS NULL")
		} else {
			query = query.Where("id IN ?", questionIDs)
		}

		var sourceBankIDs []uint
		if err := query.Where("question_bank_type_id IS NOT NULL AND question_bank_type_id <> ?", bank.ID).
			Distinct().Pluck("question_bank_type_id", &sourceBankIDs).Error; err != nil {
			return err
		}

		result := query.Updates(map[string]interface{}{
			"question_bank_type_id": bank.ID,
			"bank_name":             bank.Name,
		})
		if result.Error != nil {
			return result.Error
		}
		updatedCount = result.RowsAffected

		for _, sourceBankID := range sourceBankIDs {
			var questionIDs []uint
			if err := tx.Model(&models.QuestionBank{}).
				Where("question_bank_type_id = ? AND question_text IS NOT NULL AND question_text != ?", sourceBankID, "").
				Order("question_no ASC, id ASC").Pluck("id", &questionIDs).Error; err != nil {
				return err
			}
			var records []models.ExamRecord
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("question_bank_type_id = ? AND status = 'in_progress'", sourceBankID).
				Find(&records).Error; err != nil {
				return err
			}
			for i := range records {
				if _, err := finalizeExamIfComplete(tx, &records[i], questionIDs); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return 0, errors.New("目标题库中存在相同题号，请调整后重试")
		}
		return 0, err
	}
	return updatedCount, nil
}
