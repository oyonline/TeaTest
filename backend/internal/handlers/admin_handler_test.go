package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"tea-exam/internal/config"
	"tea-exam/internal/middleware"
	"tea-exam/internal/models"
	"tea-exam/internal/services"
)

func resetPasswordTestRouter(t *testing.T, db *gorm.DB) (*gin.Engine, string, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	cfg := &config.JWTConfig{Secret: "reset-password-test-secret-at-least-32-characters"}
	adminToken, err := middleware.GenerateToken(1, "admin", "admin", cfg)
	if err != nil {
		t.Fatal(err)
	}
	userToken, err := middleware.GenerateToken(1, "test-user", "user", cfg)
	if err != nil {
		t.Fatal(err)
	}
	adminService := services.NewAdminService(db)
	handler := NewAdminHandler(adminService)
	router := gin.New()
	admin := router.Group("/api/admin", middleware.AdminAuth(cfg.Secret))
	admin.POST("/users/:id/reset-password", handler.ResetExamUserPassword)
	auth := NewAuthHandler(services.NewUserService(db), adminService, cfg)
	router.POST("/api/auth/user/login", auth.UserLogin)
	return router, adminToken, userToken
}

func postPasswordTest(t *testing.T, router http.Handler, path, body, token string) int {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	var result struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON response, HTTP status %d", recorder.Code)
	}
	return result.Code
}

func TestResetExamUserPasswordAuthorization(t *testing.T) {
	router, _, userToken := resetPasswordTestRouter(t, nil)
	body := `{"password":"new-password","confirm_password":"new-password"}`
	for _, test := range []struct {
		name  string
		token string
		code  int
	}{
		{"anonymous", "", 401},
		{"invalid token", "invalid-token", 401},
		{"answering user", userToken, 403},
	} {
		t.Run(test.name, func(t *testing.T) {
			if code := postPasswordTest(t, router, "/api/admin/users/1/reset-password", body, test.token); code != test.code {
				t.Fatalf("response code = %d, want %d", code, test.code)
			}
		})
	}
}

func TestResetExamUserPasswordMySQL(t *testing.T) {
	dsn := os.Getenv("TEA_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set TEA_TEST_MYSQL_DSN to run the MySQL integration test")
	}
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal("cannot initialize MySQL connection")
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	t.Cleanup(func() { sqlDB.Close() })
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("cannot connect to MySQL")
	}
	// These tables shadow existing names only on this connection and disappear on close.
	for _, statement := range []string{
		`CREATE TEMPORARY TABLE exam_users (id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT, name VARCHAR(50) NOT NULL UNIQUE, password VARCHAR(100) NOT NULL, status INT NOT NULL DEFAULT 1, created_at DATETIME(3), updated_at DATETIME(3))`,
		`CREATE TEMPORARY TABLE exam_records (id BIGINT PRIMARY KEY, user_id BIGINT, total_score INT)`,
		`CREATE TEMPORARY TABLE exam_progress (id BIGINT PRIMARY KEY, exam_record_id BIGINT, user_answer VARCHAR(10), is_locked BOOL)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("cannot create connection-local fixture: %v", err)
		}
	}
	oldPassword := "old-password@@"
	oldHash, err := bcrypt.GenerateFromPassword([]byte(oldPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	createdAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	users := []models.ExamUser{
		{ID: 1, Name: "reset-target", Password: string(oldHash), Status: 1, CreatedAt: createdAt},
		{ID: 2, Name: "other-user", Password: string(oldHash), Status: 1, CreatedAt: createdAt},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO exam_records VALUES (501, 1, 85)").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO exam_progress VALUES (601, 501, 'AC', true)").Error; err != nil {
		t.Fatal(err)
	}
	router, adminToken, _ := resetPasswordTestRouter(t, db)
	validBody := `{"password":"new-password@@","confirm_password":"new-password@@"}`
	for _, test := range []struct {
		name string
		id   string
		body string
	}{
		{"zero id", "0", validBody},
		{"negative id", "-1", validBody},
		{"nonnumeric id", "abc", validBody},
		{"overflow id", "4294967296", validBody},
		{"missing user", "999", validBody},
		{"malformed JSON", "1", "{"},
		{"empty password", "1", `{"password":"","confirm_password":""}`},
		{"missing confirmation", "1", `{"password":"new-password@@"}`},
		{"mismatch", "1", `{"password":"new-password@@","confirm_password":"different"}`},
		{"too long", "1", `{"password":"` + strings.Repeat("a", 73) + `","confirm_password":"` + strings.Repeat("a", 73) + `"}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			if code := postPasswordTest(t, router, "/api/admin/users/"+test.id+"/reset-password", test.body, adminToken); code != 400 {
				t.Fatalf("response code = %d, want 400", code)
			}
			var user models.ExamUser
			if err := db.First(&user, 1).Error; err != nil {
				t.Fatal(err)
			}
			if user.Password != string(oldHash) {
				t.Fatal("rejected reset changed the password")
			}
		})
	}
	newPassword := " new-password@@ "
	body, err := json.Marshal(map[string]string{"password": newPassword, "confirm_password": newPassword})
	if err != nil {
		t.Fatal(err)
	}
	if code := postPasswordTest(t, router, "/api/admin/users/1/reset-password", string(body), adminToken); code != 0 {
		t.Fatalf("reset response code = %d, want 0", code)
	}
	var target, other models.ExamUser
	if err := db.First(&target, 1).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.First(&other, 2).Error; err != nil {
		t.Fatal(err)
	}
	if bcrypt.CompareHashAndPassword([]byte(target.Password), []byte(newPassword)) != nil {
		t.Fatal("new password is not stored as a valid bcrypt hash")
	}
	if target.Name != "reset-target" || target.Status != 1 || !target.CreatedAt.Equal(createdAt) {
		t.Fatal("reset changed account identity, status or creation time")
	}
	if other.Password != string(oldHash) {
		t.Fatal("reset changed another user's password")
	}
	for _, test := range []struct {
		name     string
		password string
		code     int
	}{
		{"new password logs in", newPassword, 0},
		{"old password rejected", oldPassword, 401},
	} {
		t.Run(test.name, func(t *testing.T) {
			login, err := json.Marshal(map[string]string{"name": "reset-target", "password": test.password})
			if err != nil {
				t.Fatal(err)
			}
			if code := postPasswordTest(t, router, "/api/auth/user/login", string(login), ""); code != test.code {
				t.Fatalf("login response code = %d, want %d", code, test.code)
			}
		})
	}
	var history struct {
		ID         uint
		UserID     uint
		TotalScore int
	}
	if err := db.Raw("SELECT * FROM exam_records WHERE id=501").Scan(&history).Error; err != nil {
		t.Fatal(err)
	}
	if history.ID != 501 || history.UserID != 1 || history.TotalScore != 85 {
		t.Fatal("exam history changed")
	}
	var progress struct {
		ID           uint
		ExamRecordID uint
		UserAnswer   string
		IsLocked     bool
	}
	if err := db.Raw("SELECT * FROM exam_progress WHERE id=601").Scan(&progress).Error; err != nil {
		t.Fatal(err)
	}
	if progress.ID != 601 || progress.ExamRecordID != 501 || progress.UserAnswer != "AC" || !progress.IsLocked {
		t.Fatal("exam progress changed")
	}
}
