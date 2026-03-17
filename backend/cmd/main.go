package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"tea-exam/internal/config"
	"tea-exam/internal/handlers"
	"tea-exam/internal/middleware"
	"tea-exam/internal/services"
	"tea-exam/internal/utils"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 创建数据库（如果不存在）
	if err := utils.CreateDatabaseIfNotExists(&cfg.Database); err != nil {
		log.Printf("创建数据库失败或已存在: %v", err)
	}

	// 初始化数据库连接
	db, err := utils.InitDB(&cfg.Database)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 自动迁移表结构
	if err := utils.InitDatabase(db); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 初始化基础数据
	if err := utils.SeedData(db); err != nil {
		log.Printf("初始化基础数据失败: %v", err)
	}

	// 创建服务
	userService := services.NewUserService(db)
	examService := services.NewExamService(db)
	adminService := services.NewAdminService(db)

	// 创建处理器
	authHandler := handlers.NewAuthHandler(userService, adminService, &cfg.JWT)
	examHandler := handlers.NewExamHandler(examService)
	adminHandler := handlers.NewAdminHandler(adminService)

	// 创建 Gin 引擎
	r := gin.Default()

	// 跨域配置
	r.Use(middleware.CORSMiddleware())

	// API 路由组
	api := r.Group("/api")

	// 公开路由
	{
		// 登录相关
		api.POST("/auth/user/login", authHandler.UserLogin)
		api.POST("/auth/admin/login", authHandler.AdminLogin)
	}

	// 需要用户认证的路由
	userGroup := api.Group("/")
	userGroup.Use(middleware.JWTAuth(cfg.JWT.Secret))
	{
		// 用户信息
		userGroup.GET("/auth/me", authHandler.GetCurrentUser)

		// 考试相关
		userGroup.GET("/exam/in-progress", examHandler.GetInProgressExam)
		userGroup.POST("/exam/start", examHandler.StartExam)
		userGroup.GET("/exam/questions", examHandler.GetQuestions)
		userGroup.GET("/exam/all-status", examHandler.GetAllQuestionStatus)
		userGroup.POST("/exam/:exam_id/answer", examHandler.SubmitAnswer)
		userGroup.GET("/exam/:exam_id/unanswered", examHandler.GetUnansweredQuestions)
		userGroup.GET("/exam/:exam_id/result", examHandler.GetExamResult)
		userGroup.GET("/exam/stats", examHandler.GetExamStats)
	}

	// 需要管理员认证的路由
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.AdminAuth(cfg.JWT.Secret))
	{
		// 题库统计
		adminGroup.GET("/stats", adminHandler.GetBankStats)

		// 考试记录
		adminGroup.GET("/records", adminHandler.GetExamRecords)

		// 题库导入
		adminGroup.POST("/questions/import", adminHandler.ImportQuestions)
	}

	// 启动服务器
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	// 允许通过环境变量指定端口
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	log.Printf("服务器启动在端口 %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
