package models

import (
	"time"

	"gorm.io/gorm"
)

// AdminConfig 管理员配置表
type AdminConfig struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	AdminPassword string    `gorm:"size:100;not null" json:"admin_password"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName 指定表名
func (AdminConfig) TableName() string {
	return "admin_config"
}

// ExamUser 答题用户表
type ExamUser struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:50;not null;uniqueIndex" json:"name"`
	Password  string    `gorm:"size:100;not null" json:"password"`
	Status    int       `gorm:"default:1" json:"status"` // 1: 启用, 0: 禁用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ExamUser) TableName() string {
	return "exam_users"
}

const (
	QuestionBankCategoryCompetition   = "competition"
	QuestionBankCategoryCertification = "certification"
)

// QuestionBankType 具体题库。一级分类由 CategoryCode 固定为竞赛题库或考级题库。
type QuestionBankType struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Code         string    `gorm:"size:64;not null;uniqueIndex" json:"code"`
	CategoryCode string    `gorm:"size:32;not null;uniqueIndex:uk_question_bank_category_name" json:"category_code"`
	Name         string    `gorm:"size:100;not null;uniqueIndex:uk_question_bank_category_name" json:"name"`
	SortOrder    int       `gorm:"default:0" json:"sort_order"`
	Status       int       `gorm:"default:1;index" json:"status"` // 1: 启用, 0: 停用
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (QuestionBankType) TableName() string {
	return "question_bank_types"
}

// QuestionBank 题库表
type QuestionBank struct {
	ID                 uint              `gorm:"primaryKey" json:"id"`
	QuestionBankTypeID *uint             `gorm:"index;uniqueIndex:uk_question_bank_type_no" json:"question_bank_id"`
	QuestionBankType   *QuestionBankType `gorm:"foreignKey:QuestionBankTypeID" json:"question_bank,omitempty"`
	QuestionNo         int               `gorm:"not null;uniqueIndex:uk_question_bank_type_no" json:"question_no"`
	BankName           string            `gorm:"size:100" json:"bank_name"`
	QuestionTypeCode   string            `gorm:"size:20" json:"question_type_code"`
	QuestionTypeName   string            `gorm:"size:20" json:"question_type_name"`
	QuestionText       string            `gorm:"type:text;not null" json:"question_text"`
	OptionA            string            `gorm:"size:255" json:"option_a"`
	OptionB            string            `gorm:"size:255" json:"option_b"`
	OptionC            string            `gorm:"size:255" json:"option_c"`
	OptionD            string            `gorm:"size:255" json:"option_d"`
	OptionE            string            `gorm:"size:255" json:"option_e"`
	CorrectAnswer      string            `gorm:"size:10;not null" json:"correct_answer"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
}

// TableName 指定表名
func (QuestionBank) TableName() string {
	return "question_bank"
}

// ExamRecord 考试记录表
type ExamRecord struct {
	ID                   uint              `gorm:"primaryKey" json:"id"`
	UserID               uint              `gorm:"not null;index" json:"user_id"`
	UserName             string            `gorm:"size:50" json:"user_name"`
	QuestionBankTypeID   *uint             `gorm:"index" json:"question_bank_id"`
	QuestionBankType     *QuestionBankType `gorm:"foreignKey:QuestionBankTypeID" json:"question_bank,omitempty"`
	QuestionBankName     string            `gorm:"size:100" json:"question_bank_name"`
	QuestionBankCategory string            `gorm:"size:32" json:"question_bank_category"`
	StartTime            time.Time         `json:"start_time"`
	EndTime              *time.Time        `json:"end_time"`
	DurationSeconds      int               `json:"duration_seconds"`
	CompletedCount       int               `json:"completed_count"`
	CorrectCount         int               `json:"correct_count"`
	WrongCount           int               `json:"wrong_count"`
	TotalScore           int               `json:"total_score"`
	AccuracyRate         float64           `json:"accuracy_rate"`
	Status               string            `gorm:"size:20;default:'in_progress';index" json:"status"` // in_progress, completed
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
	DeletedAt            gorm.DeletedAt    `gorm:"index" json:"-"`
}

// TableName 指定表名
func (ExamRecord) TableName() string {
	return "exam_records"
}

// ExamProgress 考试进度表
type ExamProgress struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	ExamRecordID uint       `gorm:"not null;index;uniqueIndex:uk_exam_progress" json:"exam_record_id"`
	QuestionID   uint       `gorm:"not null;index;uniqueIndex:uk_exam_progress" json:"question_id"`
	UserAnswer   string     `gorm:"size:10" json:"user_answer"`
	IsCorrect    bool       `json:"is_correct"`
	IsLocked     bool       `gorm:"default:false" json:"is_locked"`
	AnsweredAt   *time.Time `json:"answered_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// TableName 指定表名
func (ExamProgress) TableName() string {
	return "exam_progress"
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&AdminConfig{},
		&ExamUser{},
		&QuestionBankType{},
		&QuestionBank{},
		&ExamRecord{},
		&ExamProgress{},
	)
}
