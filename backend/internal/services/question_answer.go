package services

import (
	"errors"
	"strings"
	"unicode/utf8"

	"tea-exam/internal/models"
)

const MaxTextAnswerRunes = 4000

var (
	ErrTextAnswerEmpty   = errors.New("答案不能为空")
	ErrTextAnswerTooLong = errors.New("答案不能超过4000个字符")
)

// IsTextQuestionType 判断题型是否使用整段文字作答。
func IsTextQuestionType(code string) bool {
	return code == models.QuestionTypeFillBlank || code == models.QuestionTypeShortAnswer
}

// IsAutoGradedQuestionType 判断题型是否由系统自动判分。
func IsAutoGradedQuestionType(code string) bool {
	return code == models.QuestionTypeSingleChoice ||
		code == models.QuestionTypeMultipleChoice ||
		code == models.QuestionTypeTrueFalse ||
		code == models.QuestionTypeFillBlank
}

// QuestionTypeName 返回题型的中文名称。
func QuestionTypeName(code string) string {
	switch code {
	case models.QuestionTypeSingleChoice:
		return "单选题"
	case models.QuestionTypeMultipleChoice:
		return "多选题"
	case models.QuestionTypeTrueFalse:
		return "判断题"
	case models.QuestionTypeFillBlank:
		return "填空题"
	case models.QuestionTypeShortAnswer:
		return "简答题"
	default:
		return "未知题型"
	}
}

// NormalizeTextAnswer 仅统一换行并去掉首尾空白；内部空格、大小写和标点保持不变。
func NormalizeTextAnswer(answer string) string {
	answer = strings.ReplaceAll(answer, "\r\n", "\n")
	answer = strings.ReplaceAll(answer, "\r", "\n")
	return strings.TrimSpace(answer)
}

// ValidateTextAnswer 校验文字答案长度（按 Unicode 码点）并返回用于保存的原文。
func ValidateTextAnswer(answer string) (string, error) {
	normalized := NormalizeTextAnswer(answer)
	if normalized == "" {
		return "", ErrTextAnswerEmpty
	}
	if utf8.RuneCountInString(normalized) > MaxTextAnswerRunes {
		return "", ErrTextAnswerTooLong
	}
	return normalized, nil
}

// GradeTextAnswer 返回是否自动评分以及是否答对；简答题永远不自动评分。
func GradeTextAnswer(questionType, correctAnswer, userAnswer string) (graded bool, correct bool) {
	if questionType == models.QuestionTypeShortAnswer {
		return false, false
	}
	if questionType == models.QuestionTypeFillBlank {
		return true, NormalizeTextAnswer(userAnswer) == NormalizeTextAnswer(correctAnswer)
	}
	return false, false
}
