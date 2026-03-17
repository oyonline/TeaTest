package models

import (
	"time"

	"gorm.io/gorm"
)

// AdminConfig 管理员配置表
type AdminConfig struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	AdminPassword string   `gorm:"size:100;not null" json:"admin_password"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
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

// QuestionBank 题库表
type QuestionBank struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	QuestionNo       int       `gorm:"not null;uniqueIndex" json:"question_no"`
	BankName         string    `gorm:"size:100" json:"bank_name"`
	QuestionTypeCode string    `gorm:"size:20" json:"question_type_code"`
	QuestionTypeName string    `gorm:"size:20" json:"question_type_name"`
	QuestionText     string    `gorm:"type:text;not null" json:"question_text"`
	OptionA          string    `gorm:"size:255" json:"option_a"`
	OptionB          string    `gorm:"size:255" json:"option_b"`
	OptionC          string    `gorm:"size:255" json:"option_c"`
	OptionD          string    `gorm:"size:255" json:"option_d"`
	OptionE          string    `gorm:"size:255" json:"option_e"`
	CorrectAnswer    string    `gorm:"size:10;not null" json:"correct_answer"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TableName 指定表名
func (QuestionBank) TableName() string {
	return "question_bank"
}

// ExamRecord 考试记录表
type ExamRecord struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	UserID          uint           `gorm:"not null;index" json:"user_id"`
	UserName        string         `gorm:"size:50" json:"user_name"`
	StartTime       time.Time      `json:"start_time"`
	EndTime         *time.Time     `json:"end_time"`
	DurationSeconds int            `json:"duration_seconds"`
	CompletedCount  int            `json:"completed_count"`
	CorrectCount    int            `json:"correct_count"`
	WrongCount      int            `json:"wrong_count"`
	TotalScore      int            `json:"total_score"`
	AccuracyRate    float64        `json:"accuracy_rate"`
	Status          string         `gorm:"size:20;default:'in_progress'" json:"status"` // in_progress, completed
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (ExamRecord) TableName() string {
	return "exam_records"
}

// ExamProgress 考试进度表
type ExamProgress struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	ExamRecordID uint      `gorm:"not null;index" json:"exam_record_id"`
	QuestionID  uint       `gorm:"not null;index" json:"question_id"`
	UserAnswer  string     `gorm:"size:10" json:"user_answer"`
	IsCorrect   bool       `json:"is_correct"`
	IsLocked    bool       `gorm:"default:false" json:"is_locked"`
	AnsweredAt  *time.Time `json:"answered_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
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
		&QuestionBank{},
		&ExamRecord{},
		&ExamProgress{},
	)
}
