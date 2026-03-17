package handlers

import (
	"github.com/gin-gonic/gin"
	"tea-exam/internal/config"
	"tea-exam/internal/middleware"
	"tea-exam/internal/services"
	"tea-exam/pkg/response"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	userService  *services.UserService
	adminService *services.AdminService
	jwtConfig    *config.JWTConfig
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(userService *services.UserService, adminService *services.AdminService, jwtConfig *config.JWTConfig) *AuthHandler {
	return &AuthHandler{
		userService:  userService,
		adminService: adminService,
		jwtConfig:    jwtConfig,
	}
}

// UserLoginRequest 用户登录请求
type UserLoginRequest struct {
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AdminLoginRequest 管理员登录请求
type AdminLoginRequest struct {
	Password string `json:"password" binding:"required"`
}

// UserLogin 答题用户登录
func (h *AuthHandler) UserLogin(c *gin.Context) {
	var req UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入姓名和密码")
		return
	}

	user, err := h.userService.Login(req.Name, req.Password)
	if err != nil {
		response.Error(c, 401, err.Error())
		return
	}

	token, err := middleware.GenerateToken(user.ID, user.Name, "user", h.jwtConfig)
	if err != nil {
		response.ServerError(c, "生成 Token 失败")
		return
	}

	response.Success(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":   user.ID,
			"name": user.Name,
		},
	})
}

// AdminLogin 管理员登录
func (h *AuthHandler) AdminLogin(c *gin.Context) {
	var req AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入密码")
		return
	}

	if err := h.adminService.Login(req.Password); err != nil {
		response.Error(c, 401, err.Error())
		return
	}

	// 管理员使用固定 ID 1
	token, err := middleware.GenerateToken(1, "admin", "admin", h.jwtConfig)
	if err != nil {
		response.ServerError(c, "生成 Token 失败")
		return
	}

	response.Success(c, gin.H{
		"token": token,
		"role":  "admin",
	})
}

// GetCurrentUser 获取当前登录用户信息
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	userName, _ := c.Get("userName")
	role, _ := c.Get("role")

	response.Success(c, gin.H{
		"id":       userID,
		"name":     userName,
		"role":     role,
	})
}
