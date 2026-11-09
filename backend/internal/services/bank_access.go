package services

import (
	"errors"
	"strings"

	"gorm.io/gorm"
	"tea-exam/internal/models"
)

var (
	ErrBankAccessDenied      = errors.New("无权访问此题库")
	ErrLegacyExamUnsupported = errors.New("历史未分类考试已保留，请选择具体题库开始")
	ErrInvalidEntry          = errors.New("入口参数无效")
)

const (
	EntryTea      = "tea"
	EntryPractice = "practice"
)

// ParseEntry 把客户端入口参数收窄为固定白名单；空值默认旧茶入口。
func ParseEntry(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", EntryTea:
		return EntryTea, nil
	case EntryPractice:
		return EntryPractice, nil
	default:
		return "", ErrInvalidEntry
	}
}

// EntryCategoryCodes 返回入口允许展示的一级分类。入口只能收窄，不会越权。
func EntryCategoryCodes(entry string) []string {
	if entry == EntryPractice {
		return []string{models.QuestionBankCategoryGeneral, models.QuestionBankCategorySpecial}
	}
	return []string{models.QuestionBankCategoryCompetition, models.QuestionBankCategoryCertification}
}

// EntryAllowsCategory 判断题库分类是否属于该入口。
func EntryAllowsCategory(entry, categoryCode string) bool {
	for _, code := range EntryCategoryCodes(entry) {
		if code == categoryCode {
			return true
		}
	}
	return false
}

// userHasBankAccess 权限始终以数据库关系为准，不依赖 JWT 或客户端入口参数。
func (s *ExamService) userHasBankAccess(db *gorm.DB, userID, bankID uint) (bool, error) {
	if userID == 0 || bankID == 0 {
		return false, nil
	}
	var count int64
	err := db.Model(&models.ExamUserQuestionBank{}).
		Where("user_id = ? AND question_bank_type_id = ?", userID, bankID).
		Count(&count).Error
	return count > 0, err
}
