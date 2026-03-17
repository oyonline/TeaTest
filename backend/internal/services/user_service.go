package services

import (
	"errors"

	"gorm.io/gorm"
	"tea-exam/internal/models"
)

// UserService 用户服务
type UserService struct {
	db *gorm.DB
}

// NewUserService 创建用户服务
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// Login 用户登录
func (s *UserService) Login(name, password string) (*models.ExamUser, error) {
	var user models.ExamUser
	err := s.db.Where("name = ? AND password = ? AND status = 1", name, password).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByID 根据 ID 获取用户
func (s *UserService) GetUserByID(id uint) (*models.ExamUser, error) {
	var user models.ExamUser
	err := s.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
