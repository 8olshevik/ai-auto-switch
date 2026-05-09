package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// AppConfig holds all application configuration loaded from environment variables.
type AppConfig struct {
	// Server
	Port            int    // env: PORT, default: 8080
	ProxyPort       int    // env: PROXY_PORT, default: 18100
	ProxyListenAddr string // env: PROXY_LISTEN_ADDR, default: "0.0.0.0"

	// Database
	DatabasePath string // env: DATABASE_PATH, default: "~/.code-switch/app.db"

	// Authentication
	JWTSecret     string // env: JWT_SECRET, required
	AdminUsername string // env: ADMIN_USERNAME, default: "admin"
	AdminPassword string // env: ADMIN_PASSWORD, required

	// CORS
	CORSOrigins string // env: CORS_ORIGINS, default: "*"

	// Logging
	LogLevel string // env: LOG_LEVEL, default: "info"
	LogDir   string // env: LOG_DIR, default: "~/.code-switch/logs"

	// AI Assistant
	AssistantModel string // env: ASSISTANT_MODEL, default: "claude-sonnet-4-20250514"
}

// Load reads configuration from environment variables, applying defaults for unset values.
// It expands ~ in path fields to the user's home directory.
func Load() (*AppConfig, error) {
	cfg := &AppConfig{
		Port:            getEnvInt("PORT", 8080),
		ProxyPort:       getEnvInt("PROXY_PORT", 18100),
		ProxyListenAddr: getEnv("PROXY_LISTEN_ADDR", "0.0.0.0"),
		DatabasePath:    getEnv("DATABASE_PATH", "~/.code-switch/app.db"),
		JWTSecret:       getEnv("JWT_SECRET", ""),
		AdminUsername:   getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:   getEnv("ADMIN_PASSWORD", ""),
		CORSOrigins:     getEnv("CORS_ORIGINS", "*"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		LogDir:          getEnv("LOG_DIR", "~/.code-switch/logs"),
		AssistantModel:  getEnv("ASSISTANT_MODEL", "claude-sonnet-4-20250514"),
	}

	// Expand ~ in paths
	var err error
	cfg.DatabasePath, err = expandHome(cfg.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("expanding DatabasePath: %w", err)
	}
	cfg.LogDir, err = expandHome(cfg.LogDir)
	if err != nil {
		return nil, fmt.Errorf("expanding LogDir: %w", err)
	}

	return cfg, nil
}

// Validate checks that all required fields are set and returns an error if any are missing.
func (c *AppConfig) Validate() error {
	var missing []string

	if c.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if c.AdminPassword == "" {
		missing = append(missing, "ADMIN_PASSWORD")
	}

	if len(missing) > 0 {
		return fmt.Errorf("required environment variables not set: %s", strings.Join(missing, ", "))
	}

	return nil
}

// getEnv returns the value of the environment variable named by key,
// or defaultVal if the variable is not set or is empty.
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// getEnvInt returns the integer value of the environment variable named by key,
// or defaultVal if the variable is not set, empty, or not a valid integer.
func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return n
}

// expandHome replaces a leading ~ with the user's home directory.
func expandHome(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}

	return filepath.Join(home, path[1:]), nil
}
