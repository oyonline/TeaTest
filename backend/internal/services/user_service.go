package services

import (
	"errors"
	"strings"

	"gorm.io/gorm"
	"tea-exam/internal/models"
	"tea-exam/internal/security"
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
	err := s.db.Where("name = ? AND status = 1", strings.TrimSpace(name)).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}
	if !security.VerifyPassword(user.Password, password) {
		return nil, errors.New("用户名或密码错误")
	}
	if security.NeedsUpgrade(user.Password) && security.CanHashPassword(password) {
		hash, err := security.HashPassword(password)
		if err != nil {
			return nil, err
		}
		if err := s.db.Model(&user).Update("password", hash).Error; err != nil {
			return nil, err
		}
		user.Password = hash
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
