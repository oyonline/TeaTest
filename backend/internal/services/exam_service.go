package services

import (
	"errors"
	"math"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"tea-exam/internal/models"
)

var (
	ErrExamNotFound  = errors.New("考试记录不存在")
	ErrExamForbidden = errors.New("无权访问此考试")
)

// QuestionBankOption 是用户端可选择的具体题库。
type QuestionBankOption struct {
	ID            uint       `json:"id"`
	CategoryCode  string     `json:"category_code"`
	CategoryName  string     `json:"category_name"`
	Name          string     `json:"name"`
	SortOrder     int        `json:"sort_order"`
	Status        int        `json:"status"`
	QuestionCount int64      `json:"question_count"`
	HasInProgress bool       `json:"has_in_progress"`
	CurrentExamID uint       `json:"current_exam_id,omitempty"`
	AnsweredCount int64      `json:"answered_count"`
	ExamStartTime *time.Time `json:"exam_start_time,omitempty"`
}

// ExamService 考试服务
type ExamService struct {
	db *gorm.DB
}

// NewExamService 创建考试服务
func NewExamService(db *gorm.DB) *ExamService {
	return &ExamService{db: db}
}

func QuestionBankCategoryName(code string) string {
	switch code {
	case models.QuestionBankCategoryCompetition:
		return "竞赛题库"
	case models.QuestionBankCategoryCertification:
		return "考级题库"
	default:
		return "历史未分类"
	}
}

func questionScope(db *gorm.DB, record *models.ExamRecord) *gorm.DB {
	query := db.Model(&models.QuestionBank{}).Where("question_text IS NOT NULL AND question_text != ?", "")
	if record.QuestionBankTypeID == nil {
		// 升级前创建的进行中考试继续跟随全部题目，避免管理员归类后无法续答。
		return query
	}
	return query.Where("question_bank_type_id = ?", *record.QuestionBankTypeID)
}

func (s *ExamService) getCurrentQuestionIDs(db *gorm.DB, record *models.ExamRecord) ([]uint, error) {
	var ids []uint
	err := questionScope(db, record).Order("question_no ASC, id ASC").Pluck("id", &ids).Error
	return ids, err
}

// GetQuestionBanks 获取用户可见题库；已停用但有进行中考试的题库仍返回，供用户继续答题。
func (s *ExamService) GetQuestionBanks(userID uint) ([]QuestionBankOption, error) {
	var banks []models.QuestionBankType
	if err := s.db.Order("CASE category_code WHEN 'competition' THEN 0 ELSE 1 END, sort_order ASC, id ASC").Find(&banks).Error; err != nil {
		return nil, err
	}

	result := make([]QuestionBankOption, 0, len(banks))
	for _, bank := range banks {
		var questionCount int64
		if err := s.db.Model(&models.QuestionBank{}).
			Where("question_bank_type_id = ? AND question_text IS NOT NULL AND question_text != ?", bank.ID, "").
			Count(&questionCount).Error; err != nil {
			return nil, err
		}

		record, err := s.GetInProgressExam(userID, &bank.ID)
		if err != nil {
			return nil, err
		}
		if bank.Status != 1 && record == nil {
			continue
		}

		option := QuestionBankOption{
			ID:            bank.ID,
			CategoryCode:  bank.CategoryCode,
			CategoryName:  QuestionBankCategoryName(bank.CategoryCode),
			Name:          bank.Name,
			SortOrder:     bank.SortOrder,
			Status:        bank.Status,
			QuestionCount: questionCount,
		}
		if record != nil {
			ids, err := s.getCurrentQuestionIDs(s.db, record)
			if err != nil {
				return nil, err
			}
			answeredCount, err := countAnsweredQuestions(s.db, record.ID, ids, false)
			if err != nil {
				return nil, err
			}
			option.HasInProgress = true
			option.CurrentExamID = record.ID
			option.AnsweredCount = answeredCount
			option.ExamStartTime = &record.StartTime
		}
		result = append(result, option)
	}
	return result, nil
}

// GetInProgressExam 获取用户在指定题库中的进行中考试。bankID 为 nil 时查询历史未分类考试。
func (s *ExamService) GetInProgressExam(userID uint, bankID *uint) (*models.ExamRecord, error) {
	var record models.ExamRecord
	query := s.db.Where("user_id = ? AND status = 'in_progress'", userID)
	if bankID == nil {
		query = query.Where("question_bank_type_id IS NULL")
	} else {
		query = query.Where("question_bank_type_id = ?", *bankID)
	}
	err := query.Order("created_at ASC").First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// StartExam 开始指定题库考试。同一用户每个题库最多一场进行中考试。
func (s *ExamService) StartExam(userID uint, userName string, bankID uint) (*models.ExamRecord, error) {
	if bankID == 0 {
		return nil, errors.New("请选择题库")
	}

	var result *models.ExamRecord
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var user models.ExamUser
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
			return err
		}

		var bank models.QuestionBankType
		if err := tx.First(&bank, bankID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("所选题库不存在")
			}
			return err
		}
		if bank.Status != 1 {
			return errors.New("所选题库已停用，暂时不能开始新考试")
		}

		var existing models.ExamRecord
		err := tx.Where("user_id = ? AND question_bank_type_id = ? AND status = 'in_progress'", userID, bankID).First(&existing).Error
		if err == nil {
			result = &existing
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		var count int64
		if err := tx.Model(&models.QuestionBank{}).
			Where("question_bank_type_id = ? AND question_text IS NOT NULL AND question_text != ?", bankID, "").
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return errors.New("所选题库为空，请联系管理员导入题目")
		}

		record := models.ExamRecord{
			UserID:               userID,
			UserName:             userName,
			QuestionBankTypeID:   &bank.ID,
			QuestionBankName:     bank.Name,
			QuestionBankCategory: bank.CategoryCode,
			StartTime:            time.Now(),
			Status:               "in_progress",
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		result = &record
		return nil
	})
	return result, err
}

func (s *ExamService) getExamForUser(examRecordID, userID uint) (*models.ExamRecord, error) {
	var record models.ExamRecord
	if err := s.db.First(&record, examRecordID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrExamNotFound
		}
		return nil, err
	}
	if record.UserID != userID {
		return nil, ErrExamForbidden
	}
	return &record, nil
}

// GetExamInfo 获取考试基本信息和当前题数。
func (s *ExamService) GetExamInfo(examRecordID, userID uint) (*models.ExamRecord, int64, error) {
	record, err := s.getExamForUser(examRecordID, userID)
	if err != nil {
		return nil, 0, err
	}
	var total int64
	if err := questionScope(s.db, record).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	return record, total, nil
}

// normalizeAnswer 将答案标准化为排序后的大写字母字符串。
func normalizeAnswer(answer string) string {
	letters := make([]string, 0, len(answer))
	for _, ch := range strings.ToUpper(answer) {
		if ch >= 'A' && ch <= 'E' {
			letters = append(letters, string(ch))
		}
	}
	seen := make(map[string]bool)
	unique := make([]string, 0, len(letters))
	for _, letter := range letters {
		if !seen[letter] {
			seen[letter] = true
			unique = append(unique, letter)
		}
	}
	sort.Strings(unique)
	return strings.Join(unique, "")
}

func countAnsweredQuestions(db *gorm.DB, examRecordID uint, questionIDs []uint, correctOnly bool) (int64, error) {
	if len(questionIDs) == 0 {
		return 0, nil
	}
	query := db.Model(&models.ExamProgress{}).
		Where("exam_record_id = ? AND question_id IN ? AND is_locked = ?", examRecordID, questionIDs, true)
	if correctOnly {
		query = query.Where("is_correct = ?", true)
	}
	var count int64
	err := query.Count(&count).Error
	return count, err
}

// finalizeExamIfComplete 按题库当前内容结算考试。题库被管理员清空时也会正常结束，
// 避免动态题库变化后留下无法继续作答的进行中记录。
func finalizeExamIfComplete(db *gorm.DB, record *models.ExamRecord, questionIDs []uint) (bool, error) {
	completedCount, err := countAnsweredQuestions(db, record.ID, questionIDs, false)
	if err != nil {
		return false, err
	}
	if len(questionIDs) > 0 && completedCount < int64(len(questionIDs)) {
		return false, nil
	}

	correctCount, err := countAnsweredQuestions(db, record.ID, questionIDs, true)
	if err != nil {
		return false, err
	}
	accuracyRate := float64(0)
	if len(questionIDs) > 0 {
		accuracyRate = math.Round((float64(correctCount)/float64(len(questionIDs))*100)*100) / 100
	}
	endTime := time.Now()
	updates := map[string]interface{}{
		"end_time":         endTime,
		"duration_seconds": int(endTime.Sub(record.StartTime).Seconds()),
		"completed_count":  int(completedCount),
		"correct_count":    int(correctCount),
		"wrong_count":      int(completedCount - correctCount),
		"total_score":      int(correctCount),
		"accuracy_rate":    accuracyRate,
		"status":           "completed",
	}
	if err := db.Model(record).Updates(updates).Error; err != nil {
		return false, err
	}
	if err := db.First(record, record.ID).Error; err != nil {
		return false, err
	}
	return true, nil
}

// SubmitAnswer 提交答案，并在当前题库全部答完时结算。
func (s *ExamService) SubmitAnswer(examRecordID, userID, questionID uint, userAnswer string) (*models.ExamProgress, *models.ExamRecord, bool, string, error) {
	normalizedUserAnswer := normalizeAnswer(userAnswer)
	if normalizedUserAnswer == "" {
		return nil, nil, false, "", errors.New("答案格式无效")
	}

	var progress models.ExamProgress
	var record models.ExamRecord
	var correctAnswer string
	completed := false
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&record, examRecordID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrExamNotFound
			}
			return err
		}
		if record.UserID != userID {
			return ErrExamForbidden
		}
		if record.Status != "in_progress" {
			return errors.New("考试已完成，不能继续答题")
		}

		var question models.QuestionBank
		if err := questionScope(tx, &record).Where("id = ?", questionID).First(&question).Error; err != nil {
			return errors.New("题目不属于本次考试")
		}
		correctAnswer = normalizeAnswer(question.CorrectAnswer)

		var existing models.ExamProgress
		err := tx.Where("exam_record_id = ? AND question_id = ?", examRecordID, questionID).First(&existing).Error
		if err == nil && existing.IsLocked {
			return errors.New("该题已锁定，无法修改答案")
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		now := time.Now()
		progress = models.ExamProgress{
			ExamRecordID: examRecordID,
			QuestionID:   questionID,
			UserAnswer:   normalizedUserAnswer,
			IsCorrect:    normalizedUserAnswer == correctAnswer,
			IsLocked:     true,
			AnsweredAt:   &now,
		}
		if existing.ID > 0 {
			progress.ID = existing.ID
			if err := tx.Model(&existing).Updates(&progress).Error; err != nil {
				return err
			}
		} else if err := tx.Create(&progress).Error; err != nil {
			return err
		}

		questionIDs, err := s.getCurrentQuestionIDs(tx, &record)
		if err != nil {
			return err
		}
		completed, err = finalizeExamIfComplete(tx, &record, questionIDs)
		return err
	})
	if err != nil {
		return nil, nil, false, "", err
	}
	return &progress, &record, completed, correctAnswer, nil
}

// GetUnansweredQuestions 获取当前题库中的未答题 ID。
func (s *ExamService) GetUnansweredQuestions(examRecordID, userID uint) ([]uint, error) {
	record, err := s.getExamForUser(examRecordID, userID)
	if err != nil {
		return nil, err
	}
	allIDs, err := s.getCurrentQuestionIDs(s.db, record)
	if err != nil {
		return nil, err
	}
	var answeredIDs []uint
	if len(allIDs) > 0 {
		if err := s.db.Model(&models.ExamProgress{}).
			Where("exam_record_id = ? AND question_id IN ? AND is_locked = ?", examRecordID, allIDs, true).
			Pluck("question_id", &answeredIDs).Error; err != nil {
			return nil, err
		}
	}
	answeredMap := make(map[uint]bool, len(answeredIDs))
	for _, id := range answeredIDs {
		answeredMap[id] = true
	}
	unanswered := make([]uint, 0)
	for _, id := range allIDs {
		if !answeredMap[id] {
			unanswered = append(unanswered, id)
		}
	}
	return unanswered, nil
}

// GetExamResult 获取考试结果。
func (s *ExamService) GetExamResult(examRecordID, userID uint) (*models.ExamRecord, error) {
	return s.getExamForUser(examRecordID, userID)
}

// GetExamStats 获取用户总体统计。
func (s *ExamService) GetExamStats(userID uint) (map[string]interface{}, error) {
	var totalQuestions int64
	if err := s.db.Model(&models.QuestionBank{}).
		Joins("JOIN question_bank_types ON question_bank_types.id = question_bank.question_bank_type_id AND question_bank_types.status = 1").
		Where("question_bank.question_text IS NOT NULL AND question_bank.question_text != ?", "").Count(&totalQuestions).Error; err != nil {
		return nil, err
	}
	var completedExams int64
	if err := s.db.Model(&models.ExamRecord{}).Where("user_id = ? AND status = 'completed'", userID).Count(&completedExams).Error; err != nil {
		return nil, err
	}
	var inProgressCount int64
	if err := s.db.Model(&models.ExamRecord{}).Where("user_id = ? AND status = 'in_progress'", userID).Count(&inProgressCount).Error; err != nil {
		return nil, err
	}
	stats := map[string]interface{}{
		"total_questions":   totalQuestions,
		"completed_exams":   completedExams,
		"in_progress_count": inProgressCount,
		"has_in_progress":   inProgressCount > 0,
	}
	legacy, err := s.GetInProgressExam(userID, nil)
	if err != nil {
		return nil, err
	}
	if legacy != nil {
		ids, err := s.getCurrentQuestionIDs(s.db, legacy)
		if err != nil {
			return nil, err
		}
		answeredCount, err := countAnsweredQuestions(s.db, legacy.ID, ids, false)
		if err != nil {
			return nil, err
		}
		stats["legacy_in_progress"] = map[string]interface{}{
			"exam_id":         legacy.ID,
			"answered_count":  answeredCount,
			"total_questions": len(ids),
			"start_time":      legacy.StartTime,
		}
	}
	return stats, nil
}

func (s *ExamService) getProgressMap(examRecordID uint) (map[uint]*models.ExamProgress, error) {
	var progressList []models.ExamProgress
	if err := s.db.Where("exam_record_id = ? AND is_locked = ?", examRecordID, true).Find(&progressList).Error; err != nil {
		return nil, err
	}
	result := make(map[uint]*models.ExamProgress, len(progressList))
	for i := range progressList {
		result[progressList[i].QuestionID] = &progressList[i]
	}
	return result, nil
}

// GetQuestionsWithProgress 获取当前题库及本次考试进度。
func (s *ExamService) GetQuestionsWithProgress(page, pageSize int, examRecordID, userID uint) (map[string]interface{}, error) {
	record, err := s.getExamForUser(examRecordID, userID)
	if err != nil {
		return nil, err
	}
	var total int64
	if err := questionScope(s.db, record).Count(&total).Error; err != nil {
		return nil, err
	}
	var questions []models.QuestionBank
	if err := questionScope(s.db, record).Order("question_no ASC, id ASC").
		Limit(pageSize).Offset((page - 1) * pageSize).Find(&questions).Error; err != nil {
		return nil, err
	}
	progressMap, err := s.getProgressMap(examRecordID)
	if err != nil {
		return nil, err
	}

	questionList := make([]map[string]interface{}, 0, len(questions))
	for i, q := range questions {
		questionType := "single_choice"
		if q.QuestionTypeCode == "multiple_choice" || q.QuestionTypeCode == "true_false" {
			questionType = q.QuestionTypeCode
		}
		qData := map[string]interface{}{
			"id":             q.ID,
			"position":       (page-1)*pageSize + i + 1,
			"question_no":    q.QuestionNo,
			"question_type":  questionType,
			"question_text":  q.QuestionText,
			"option_a":       q.OptionA,
			"option_b":       q.OptionB,
			"option_c":       q.OptionC,
			"option_d":       q.OptionD,
			"option_e":       q.OptionE,
			"correct_answer": "",
			"has_answered":   false,
			"user_answer":    "",
			"is_correct":     false,
			"is_locked":      false,
		}
		if progress, ok := progressMap[q.ID]; ok {
			qData["correct_answer"] = normalizeAnswer(q.CorrectAnswer)
			qData["has_answered"] = true
			qData["user_answer"] = progress.UserAnswer
			qData["is_correct"] = progress.IsCorrect
			qData["is_locked"] = progress.IsLocked
		}
		questionList = append(questionList, qData)
	}

	return map[string]interface{}{
		"list":        questionList,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": int(math.Ceil(float64(total) / float64(pageSize))),
	}, nil
}

// GetAllQuestionNosWithStatus 获取当前题库全部题号及答题状态。
func (s *ExamService) GetAllQuestionNosWithStatus(examRecordID, userID uint) ([]map[string]interface{}, error) {
	record, err := s.getExamForUser(examRecordID, userID)
	if err != nil {
		return nil, err
	}
	var questions []models.QuestionBank
	if err := questionScope(s.db, record).Order("question_no ASC, id ASC").Find(&questions).Error; err != nil {
		return nil, err
	}
	progressMap, err := s.getProgressMap(examRecordID)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(questions))
	for i, q := range questions {
		status := "unanswered"
		if progress, ok := progressMap[q.ID]; ok {
			if progress.IsCorrect {
				status = "correct"
			} else {
				status = "wrong"
			}
		}
		result = append(result, map[string]interface{}{
			"question_id": q.ID,
			"question_no": q.QuestionNo,
			"position":    i + 1,
			"status":      status,
		})
	}
	return result, nil
}
