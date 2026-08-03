package config

import "testing"

func TestValidateReleaseConfig(t *testing.T) {
	valid := &Config{
		Database: DatabaseConfig{Password: "database-password"},
		Server:   ServerConfig{Mode: "release"},
		JWT:      JWTConfig{Secret: "0123456789abcdef0123456789abcdef"},
		Security: SecurityConfig{AdminInitialPassword: "strong-admin-password"},
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid release config returned error: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{"default jwt", func(c *Config) { c.JWT.Secret = "tea-exam-secret-key-2024" }},
		{"short admin password", func(c *Config) { c.Security.AdminInitialPassword = "123456" }},
		{"empty database password", func(c *Config) { c.Database.Password = "" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := *valid
			test.mutate(&cfg)
			if err := cfg.Validate(); err == nil {
				t.Fatal("invalid release config should return an error")
			}
		})
	}
}

func TestValidateAllowsDevelopmentDefaults(t *testing.T) {
	cfg := &Config{Server: ServerConfig{Mode: "debug"}}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("development config returned error: %v", err)
	}
}

func TestGetEnvBoolRejectsInvalidValue(t *testing.T) {
	t.Setenv("TEST_BOOL", "flase")
	if _, err := getEnvBool("TEST_BOOL", true); err == nil {
		t.Fatal("malformed boolean should return an error")
	}
}

func TestLoadConfigReleaseDisablesDemoUsersByDefault(t *testing.T) {
	t.Setenv("GIN_MODE", "release")
	t.Setenv("SEED_DEMO_USERS", "")
	cfg := LoadConfig()
	if cfg.Security.SeedDemoUsers {
		t.Fatal("release mode should not seed demo users by default")
	}
}
