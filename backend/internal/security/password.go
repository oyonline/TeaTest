package security

import (
	"crypto/subtle"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用 bcrypt 保存新密码。
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword 同时兼容旧版明文密码和新版 bcrypt 密码。
func VerifyPassword(storedPassword, password string) bool {
	if strings.HasPrefix(storedPassword, "$2") {
		return bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password)) == nil
	}
	return subtle.ConstantTimeCompare([]byte(storedPassword), []byte(password)) == 1
}

// CanHashPassword 表示密码是否在当前 bcrypt 实现支持的长度内。
func CanHashPassword(password string) bool {
	return len([]byte(password)) <= 72
}

// NeedsUpgrade 判断旧版明文或低成本哈希是否需要升级。
func NeedsUpgrade(storedPassword string) bool {
	cost, err := bcrypt.Cost([]byte(storedPassword))
	return err != nil || cost < bcrypt.DefaultCost
}
