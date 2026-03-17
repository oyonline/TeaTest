package utils

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"tea-exam/internal/config"
	"tea-exam/internal/models"
)

// InitDB 初始化数据库连接
func InitDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.Charset)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	return db, nil
}

// InitDatabase 初始化数据库（创建表和基础数据）
func InitDatabase(db *gorm.DB) error {
	// 自动迁移表结构
	if err := models.AutoMigrate(db); err != nil {
		return fmt.Errorf("自动迁移失败: %v", err)
	}

	return nil
}

// CreateDatabaseIfNotExists 如果数据库不存在则创建
func CreateDatabaseIfNotExists(cfg *config.DatabaseConfig) error {
	// 先连接 MySQL（不指定数据库）
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=%s&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Charset)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("连接 MySQL 失败: %v", err)
	}

	// 创建数据库（如果不存在）
	createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET %s COLLATE utf8mb4_unicode_ci",
		cfg.DBName, cfg.Charset)
	if err := db.Exec(createDBSQL).Error; err != nil {
		return fmt.Errorf("创建数据库失败: %v", err)
	}

	return nil
}
