package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config 应用配置
type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	JWT      JWTConfig
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	Charset  string
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string
	Mode string
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret string
}

// LoadConfig 加载配置
func LoadConfig() *Config {
	// 尝试加载 .env 文件（如果存在）
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "tea_exam"),
			Charset:  "utf8mb4",
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "tea-exam-secret-key-2024"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
