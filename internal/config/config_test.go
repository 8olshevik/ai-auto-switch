package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear any env vars that might interfere
	envVars := []string{
		"PORT", "PROXY_PORT", "PROXY_LISTEN_ADDR",
		"DATABASE_PATH", "JWT_SECRET", "ADMIN_USERNAME",
		"ADMIN_PASSWORD", "CORS_ORIGINS", "LOG_LEVEL",
		"LOG_DIR", "ASSISTANT_MODEL",
	}
	for _, key := range envVars {
		t.Setenv(key, "")
		os.Unsetenv(key)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want 8080", cfg.Port)
	}
	if cfg.ProxyPort != 18100 {
		t.Errorf("ProxyPort = %d, want 18100", cfg.ProxyPort)
	}
	if cfg.ProxyListenAddr != "0.0.0.0" {
		t.Errorf("ProxyListenAddr = %q, want %q", cfg.ProxyListenAddr, "0.0.0.0")
	}
	if cfg.AdminUsername != "admin" {
		t.Errorf("AdminUsername = %q, want %q", cfg.AdminUsername, "admin")
	}
	if cfg.CORSOrigins != "*" {
		t.Errorf("CORSOrigins = %q, want %q", cfg.CORSOrigins, "*")
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "info")
	}
	if cfg.AssistantModel != "claude-sonnet-4-20250514" {
		t.Errorf("AssistantModel = %q, want %q", cfg.AssistantModel, "claude-sonnet-4-20250514")
	}
}

func TestLoad_FromEnv(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("PROXY_PORT", "19000")
	t.Setenv("PROXY_LISTEN_ADDR", "127.0.0.1")
	t.Setenv("DATABASE_PATH", "/tmp/test.db")
	t.Setenv("JWT_SECRET", "my-secret")
	t.Setenv("ADMIN_USERNAME", "superadmin")
	t.Setenv("ADMIN_PASSWORD", "pass123")
	t.Setenv("CORS_ORIGINS", "http://localhost:3000")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_DIR", "/tmp/logs")
	t.Setenv("ASSISTANT_MODEL", "gpt-4")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Port != 9090 {
		t.Errorf("Port = %d, want 9090", cfg.Port)
	}
	if cfg.ProxyPort != 19000 {
		t.Errorf("ProxyPort = %d, want 19000", cfg.ProxyPort)
	}
	if cfg.ProxyListenAddr != "127.0.0.1" {
		t.Errorf("ProxyListenAddr = %q, want %q", cfg.ProxyListenAddr, "127.0.0.1")
	}
	if cfg.DatabasePath != "/tmp/test.db" {
		t.Errorf("DatabasePath = %q, want %q", cfg.DatabasePath, "/tmp/test.db")
	}
	if cfg.JWTSecret != "my-secret" {
		t.Errorf("JWTSecret = %q, want %q", cfg.JWTSecret, "my-secret")
	}
	if cfg.AdminUsername != "superadmin" {
		t.Errorf("AdminUsername = %q, want %q", cfg.AdminUsername, "superadmin")
	}
	if cfg.AdminPassword != "pass123" {
		t.Errorf("AdminPassword = %q, want %q", cfg.AdminPassword, "pass123")
	}
	if cfg.CORSOrigins != "http://localhost:3000" {
		t.Errorf("CORSOrigins = %q, want %q", cfg.CORSOrigins, "http://localhost:3000")
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "debug")
	}
	if cfg.LogDir != "/tmp/logs" {
		t.Errorf("LogDir = %q, want %q", cfg.LogDir, "/tmp/logs")
	}
	if cfg.AssistantModel != "gpt-4" {
		t.Errorf("AssistantModel = %q, want %q", cfg.AssistantModel, "gpt-4")
	}
}

func TestLoad_ExpandsHomePath(t *testing.T) {
	// Clear env vars to use defaults with ~
	os.Unsetenv("DATABASE_PATH")
	os.Unsetenv("LOG_DIR")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() returned error: %v", err)
	}

	expectedDB := filepath.Join(home, ".code-switch/app.db")
	if cfg.DatabasePath != expectedDB {
		t.Errorf("DatabasePath = %q, want %q", cfg.DatabasePath, expectedDB)
	}

	expectedLogDir := filepath.Join(home, ".code-switch/logs")
	if cfg.LogDir != expectedLogDir {
		t.Errorf("LogDir = %q, want %q", cfg.LogDir, expectedLogDir)
	}
}

func TestLoad_ExpandsHomeTilde(t *testing.T) {
	t.Setenv("DATABASE_PATH", "~/custom/path/db.sqlite")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() returned error: %v", err)
	}

	expected := filepath.Join(home, "custom/path/db.sqlite")
	if cfg.DatabasePath != expected {
		t.Errorf("DatabasePath = %q, want %q", cfg.DatabasePath, expected)
	}
}

func TestValidate_MissingJWTSecret(t *testing.T) {
	cfg := &AppConfig{
		JWTSecret:     "",
		AdminPassword: "password123",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() should return error when JWT_SECRET is missing")
	}
	if got := err.Error(); got != "required environment variables not set: JWT_SECRET" {
		t.Errorf("error = %q, want mention of JWT_SECRET", got)
	}
}

func TestValidate_MissingAdminPassword(t *testing.T) {
	cfg := &AppConfig{
		JWTSecret:     "secret",
		AdminPassword: "",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() should return error when ADMIN_PASSWORD is missing")
	}
	if got := err.Error(); got != "required environment variables not set: ADMIN_PASSWORD" {
		t.Errorf("error = %q, want mention of ADMIN_PASSWORD", got)
	}
}

func TestValidate_MissingBoth(t *testing.T) {
	cfg := &AppConfig{
		JWTSecret:     "",
		AdminPassword: "",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() should return error when both required fields are missing")
	}
	if got := err.Error(); got != "required environment variables not set: JWT_SECRET, ADMIN_PASSWORD" {
		t.Errorf("error = %q, want mention of both JWT_SECRET and ADMIN_PASSWORD", got)
	}
}

func TestValidate_AllSet(t *testing.T) {
	cfg := &AppConfig{
		JWTSecret:     "my-secret",
		AdminPassword: "my-password",
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Validate() returned unexpected error: %v", err)
	}
}

func TestGetEnvInt_InvalidValue(t *testing.T) {
	t.Setenv("PORT", "not-a-number")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Should fall back to default
	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want 8080 (default) when env is invalid", cfg.Port)
	}
}
