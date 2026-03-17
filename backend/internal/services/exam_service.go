package services

import (
	"errors"
	"math"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
	"tea-exam/internal/models"
)

// ExamService 考试服务
type ExamService struct {
	db *gorm.DB
}

// NewExamService 创建考试服务
func NewExamService(db *gorm.DB) *ExamService {
	return &ExamService{db: db}
}

// GetInProgressExam 获取用户进行中的考试
func (s *ExamService) GetInProgressExam(userID uint) (*models.ExamRecord, error) {
	var record models.ExamRecord
	err := s.db.Where("user_id = ? AND status = 'in_progress'", userID).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

// StartExam 开始考试
func (s *ExamService) StartExam(userID uint, userName string) (*models.ExamRecord, error) {
	// 检查是否有进行中的考试
	existing, err := s.GetInProgressExam(userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	// 检查题库是否有题目（过滤空题干）
	var count int64
	if err := s.db.Model(&models.QuestionBank{}).Where("question_text IS NOT NULL AND question_text != ?", "").Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("题库为空，请联系管理员导入题库")
	}

	// 创建新考试记录
	record := models.ExamRecord{
		UserID:   userID,
		UserName: userName,
		StartTime: time.Now(),
		Status:   "in_progress",
	}

	if err := s.db.Create(&record).Error; err != nil {
		return nil, err
	}

	return &record, nil
}

// GetQuestions 获取分页题目
func (s *ExamService) GetQuestions(page, pageSize int) ([]models.QuestionBank, int64, error) {
	var questions []models.QuestionBank
	var total int64

	offset := (page - 1) * pageSize

	if err := s.db.Model(&models.QuestionBank{}).Where("question_text IS NOT NULL AND question_text != ?", "").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.db.Where("question_text IS NOT NULL AND question_text != ?", "").Order("question_no ASC").Limit(pageSize).Offset(offset).Find(&questions).Error; err != nil {
		return nil, 0, err
	}

	return questions, total, nil
}

// GetExamProgress 获取考试进度
func (s *ExamService) GetExamProgress(examRecordID uint) ([]models.ExamProgress, error) {
	var progress []models.ExamProgress
	err := s.db.Where("exam_record_id = ?", examRecordID).Find(&progress).Error
	return progress, err
}

// normalizeAnswer 将答案标准化为排序后的大写字母字符串
// 例如 "BAC" -> "ABC", "b,a,c" -> "ABC"
func normalizeAnswer(answer string) string {
	// 移除所有非字母字符，只保留 A-E
	var letters []string
	for _, ch := range strings.ToUpper(answer) {
		if ch >= 'A' && ch <= 'E' {
			letters = append(letters, string(ch))
		}
	}
	// 去重并排序
	seen := make(map[string]bool)
	unique := []string{}
	for _, letter := range letters {
		if !seen[letter] {
			seen[letter] = true
			unique = append(unique, letter)
		}
	}
	sort.Strings(unique)
	return strings.Join(unique, "")
}

// SubmitAnswer 提交答案
func (s *ExamService) SubmitAnswer(examRecordID uint, questionID uint, userAnswer string) (*models.ExamProgress, error) {
	// 开启事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取题目
	var question models.QuestionBank
	if err := tx.First(&question, questionID).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("题目不存在")
	}

	// 检查是否已锁定
	var existingProgress models.ExamProgress
	err := tx.Where("exam_record_id = ? AND question_id = ?", examRecordID, questionID).First(&existingProgress).Error
	if err == nil && existingProgress.IsLocked {
		tx.Rollback()
		return nil, errors.New("该题已锁定，无法修改答案")
	}

	// 标准化答案并判断是否正确
	normalizedUserAnswer := normalizeAnswer(userAnswer)
	normalizedCorrectAnswer := normalizeAnswer(question.CorrectAnswer)
	isCorrect := normalizedUserAnswer == normalizedCorrectAnswer

	now := time.Now()
	progress := models.ExamProgress{
		ExamRecordID: examRecordID,
		QuestionID:   questionID,
		UserAnswer:   normalizedUserAnswer,
		IsCorrect:    isCorrect,
		IsLocked:     true,
		AnsweredAt:   &now,
	}

	if err := tx.Create(&progress).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &progress, nil
}

// GetUnansweredQuestions 获取未答题的题目
func (s *ExamService) GetUnansweredQuestions(examRecordID uint) ([]uint, error) {
	var unanswered []uint

	// 获取所有题目ID（过滤空题干）
	var allQuestionIDs []uint
	if err := s.db.Model(&models.QuestionBank{}).Where("question_text IS NOT NULL AND question_text != ?", "").Pluck("id", &allQuestionIDs).Error; err != nil {
		return nil, err
	}

	// 获取已答题的题目ID
	var answeredQuestionIDs []uint
	if err := s.db.Model(&models.ExamProgress{}).
		Where("exam_record_id = ?", examRecordID).
		Pluck("question_id", &answeredQuestionIDs).Error; err != nil {
		return nil, err
	}

	// 找出未答题的题目
	answeredMap := make(map[uint]bool)
	for _, id := range answeredQuestionIDs {
		answeredMap[id] = true
	}

	for _, id := range allQuestionIDs {
		if !answeredMap[id] {
			unanswered = append(unanswered, id)
		}
	}

	return unanswered, nil
}

// CheckAndCompleteExam 检查并自动完成考试
func (s *ExamService) CheckAndCompleteExam(examRecordID uint) (*models.ExamRecord, bool, error) {
	// 获取考试记录
	var record models.ExamRecord
	if err := s.db.First(&record, examRecordID).Error; err != nil {
		return nil, false, err
	}

	if record.Status == "completed" {
		return &record, false, nil
	}

	// 获取总题数（过滤空题干）
	var totalQuestions int64
	if err := s.db.Model(&models.QuestionBank{}).Where("question_text IS NOT NULL AND question_text != ?", "").Count(&totalQuestions).Error; err != nil {
		return nil, false, err
	}

	// 获取已完成题数
	var completedCount int64
	if err := s.db.Model(&models.ExamProgress{}).
		Where("exam_record_id = ?", examRecordID).
		Count(&completedCount).Error; err != nil {
		return nil, false, err
	}

	// 检查是否全部完成
	if int(completedCount) < int(totalQuestions) {
		return &record, false, nil
	}

	// 计算统计信息
	var correctCount int64
	if err := s.db.Model(&models.ExamProgress{}).
		Where("exam_record_id = ? AND is_correct = ?", examRecordID, true).
		Count(&correctCount).Error; err != nil {
		return nil, false, err
	}

	wrongCount := int(completedCount) - int(correctCount)
	accuracyRate := float64(correctCount) / float64(totalQuestions) * 100

	now := time.Now()
	duration := int(now.Sub(record.StartTime).Seconds())

	// 更新考试记录
	updates := map[string]interface{}{
		"end_time":        now,
		"duration_seconds": duration,
		"completed_count": int(completedCount),
		"correct_count":   int(correctCount),
		"wrong_count":     wrongCount,
		"total_score":     int(correctCount),
		"accuracy_rate":   math.Round(accuracyRate*100) / 100,
		"status":          "completed",
	}

	if err := s.db.Model(&record).Updates(updates).Error; err != nil {
		return nil, false, err
	}

	// 刷新记录
	s.db.First(&record, examRecordID)

	return &record, true, nil
}

// GetExamResult 获取考试结果
func (s *ExamService) GetExamResult(examRecordID uint) (*models.ExamRecord, error) {
	var record models.ExamRecord
	if err := s.db.First(&record, examRecordID).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// GetExamStats 获取考试统计（用于欢迎页显示）
func (s *ExamService) GetExamStats(userID uint) (map[string]interface{}, error) {
	// 获取总题数（过滤空题干）
	var totalQuestions int64
	if err := s.db.Model(&models.QuestionBank{}).Where("question_text IS NOT NULL AND question_text != ?", "").Count(&totalQuestions).Error; err != nil {
		return nil, err
	}

	// 获取已完成考试数
	var completedExams int64
	if err := s.db.Model(&models.ExamRecord{}).
		Where("user_id = ? AND status = 'completed'", userID).
		Count(&completedExams).Error; err != nil {
		return nil, err
	}

	// 获取进行中的考试
	var inProgressExam models.ExamRecord
	err := s.db.Where("user_id = ? AND status = 'in_progress'", userID).First(&inProgressExam).Error
	hasInProgress := err == nil

	stats := map[string]interface{}{
		"total_questions":  totalQuestions,
		"completed_exams":  completedExams,
		"has_in_progress":  hasInProgress,
	}

	if hasInProgress {
		// 获取已答题数
		var answeredCount int64
		s.db.Model(&models.ExamProgress{}).
			Where("exam_record_id = ?", inProgressExam.ID).
			Count(&answeredCount)

		stats["current_exam_id"] = inProgressExam.ID
		stats["answered_count"] = answeredCount
		stats["start_time"] = inProgressExam.StartTime
	}

	return stats, nil
}

// GetProgressWithQuestions 获取带题目详情的进度
func (s *ExamService) GetProgressWithQuestions(examRecordID uint) (map[uint]*models.ExamProgress, error) {
	var progressList []models.ExamProgress
	if err := s.db.Where("exam_record_id = ?", examRecordID).Find(&progressList).Error; err != nil {
		return nil, err
	}

	progressMap := make(map[uint]*models.ExamProgress)
	for i := range progressList {
		progressMap[progressList[i].QuestionID] = &progressList[i]
	}

	return progressMap, nil
}

// GetQuestionsWithProgress 获取题目及答题进度
func (s *ExamService) GetQuestionsWithProgress(page, pageSize int, examRecordID uint) (map[string]interface{}, error) {
	// 获取题目
	questions, total, err := s.GetQuestions(page, pageSize)
	if err != nil {
		return nil, err
	}

	// 获取进度
	progressMap, err := s.GetProgressWithQuestions(examRecordID)
	if err != nil {
		return nil, err
	}

	// 组合数据
	var questionList []map[string]interface{}
	for _, q := range questions {
		// 确定题型，默认为单选题
		questionType := "single_choice"
		if q.QuestionTypeCode == "multiple_choice" {
			questionType = "multiple_choice"
		} else if q.QuestionTypeCode == "true_false" {
			questionType = "true_false"
		}

		qData := map[string]interface{}{
			"id":                 q.ID,
			"question_no":        q.QuestionNo,
			"question_type":      questionType,
			"question_text":      q.QuestionText,
			"option_a":           q.OptionA,
			"option_b":           q.OptionB,
			"option_c":           q.OptionC,
			"option_d":           q.OptionD,
			"option_e":           q.OptionE,
			"correct_answer":     q.CorrectAnswer,
			"has_answered":       false,
			"user_answer":        "",
			"is_correct":         false,
			"is_locked":          false,
		}

		if progress, ok := progressMap[q.ID]; ok {
			qData["has_answered"] = true
			qData["user_answer"] = progress.UserAnswer
			qData["is_correct"] = progress.IsCorrect
			qData["is_locked"] = progress.IsLocked
		}

		questionList = append(questionList, qData)
	}

	result := map[string]interface{}{
		"list":        questionList,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": int(math.Ceil(float64(total) / float64(pageSize))),
	}

	return result, nil
}

// GetAllQuestionNosWithStatus 获取所有题号及答题状态（用于进度网格）
func (s *ExamService) GetAllQuestionNosWithStatus(examRecordID uint) ([]map[string]interface{}, error) {
	// 获取所有题目（过滤空题干）
	var questions []models.QuestionBank
	if err := s.db.Where("question_text IS NOT NULL AND question_text != ?", "").Order("question_no ASC").Find(&questions).Error; err != nil {
		return nil, err
	}

	// 获取进度
	progressMap, err := s.GetProgressWithQuestions(examRecordID)
	if err != nil {
		return nil, err
	}

	// 组装状态
	var result []map[string]interface{}
	for _, q := range questions {
		status := "unanswered"
		if progress, ok := progressMap[q.ID]; ok {
			if progress.IsCorrect {
				status = "correct"
			} else {
				status = "wrong"
			}
		}

		result = append(result, map[string]interface{}{
			"question_id":   q.ID,
			"question_no":   q.QuestionNo,
			"status":        status,
		})
	}

	return result, nil
}
