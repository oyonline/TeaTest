package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config 应用配置
type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	JWT      JWTConfig
	Security SecurityConfig
	loadErr  error
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

// SecurityConfig 安全与初始化配置。
type SecurityConfig struct {
	AdminInitialPassword string
	SeedDemoUsers        bool
	AllowedOrigins       []string
}

// LoadConfig 加载配置
func LoadConfig() *Config {
	// 尝试加载 .env 文件（如果存在）
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")

	mode := getEnv("GIN_MODE", "debug")
	defaultOrigins := "http://localhost:3000,http://127.0.0.1:3000"
	if mode == "release" {
		defaultOrigins = ""
	}
	seedDemoUsers, seedDemoUsersErr := getEnvBool("SEED_DEMO_USERS", mode != "release")

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
			Mode: mode,
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "tea-exam-secret-key-2024"),
		},
		Security: SecurityConfig{
			AdminInitialPassword: getEnv("ADMIN_INITIAL_PASSWORD", "123456"),
			SeedDemoUsers:        seedDemoUsers,
			AllowedOrigins:       getEnvList("CORS_ALLOWED_ORIGINS", defaultOrigins),
		},
		loadErr: seedDemoUsersErr,
	}
}

// Validate 校验生产环境必须显式设置的安全配置。
func (c *Config) Validate() error {
	if c.loadErr != nil {
		return c.loadErr
	}
	if c.Server.Mode != "release" {
		return nil
	}
	if len(c.JWT.Secret) < 32 || c.JWT.Secret == "tea-exam-secret-key-2024" {
		return errors.New("release 模式下 JWT_SECRET 必须设置为至少 32 个字符的随机值")
	}
	if len(c.Security.AdminInitialPassword) < 12 || c.Security.AdminInitialPassword == "123456" {
		return errors.New("release 模式下 ADMIN_INITIAL_PASSWORD 必须设置为至少 12 个字符的强密码")
	}
	if len([]byte(c.Security.AdminInitialPassword)) > 72 {
		return errors.New("ADMIN_INITIAL_PASSWORD 不能超过 72 个字节")
	}
	if c.Database.Password == "" {
		return errors.New("release 模式下 DB_PASSWORD 不能为空")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) (bool, error) {
	value, exists := os.LookupEnv(key)
	if !exists || strings.TrimSpace(value) == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s 必须是 true 或 false", key)
	}
	return parsed, nil
}

func getEnvList(key, defaultValue string) []string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	if strings.TrimSpace(value) == "" {
		return nil
	}

	items := strings.Split(value, ",")
	result := make([]string, 0, len(items))
	for _, item := range items {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
