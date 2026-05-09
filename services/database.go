package services

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/daodao97/xgo/xdb"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

// InitDatabase 初始化数据库连接（必须在所有服务构造之前调用）
// 【修复】解决数据库初始化时序问题：
// 1. 确保配置目录存在
// 2. 初始化 xdb 连接池
// 3. 显式设置 PRAGMA（WAL 模式 + busy_timeout）
// 4. 确保表结构存在
// 5. 预热连接池
func InitDatabase() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户目录失败: %w", err)
	}

	// 1. 确保配置目录存在（SQLite 不会自动创建父目录）
	configDir := filepath.Join(home, ".code-switch")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 2. 初始化 xdb 连接池
	// 【修复】移除 DSN 中的 PRAGMA 参数，modernc.org/sqlite 需要显式执行 PRAGMA
	dbPath := filepath.Join(configDir, "app.db?cache=shared&mode=rwc")
	if err := xdb.Inits([]xdb.Config{
		{
			Name:   "default",
			Driver: "sqlite",
			DSN:    dbPath,
		},
	}); err != nil {
		return fmt.Errorf("初始化数据库失败: %w", err)
	}

	// 3. 显式设置 PRAGMA（解决 SQLITE_BUSY 问题）
	db, err := xdb.DB("default")
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 3.1 设置 busy_timeout（30秒，确保高并发下有足够等待时间）
	if _, err := db.Exec("PRAGMA busy_timeout = 30000"); err != nil {
		return fmt.Errorf("设置 busy_timeout 失败: %w", err)
	}

	// 3.2 设置 WAL 模式（允许读写并发）
	var journalMode string
	if err := db.QueryRow("PRAGMA journal_mode = WAL").Scan(&journalMode); err != nil {
		return fmt.Errorf("设置 WAL 模式失败: %w", err)
	}
	fmt.Printf("✅ SQLite PRAGMA 已设置: journal_mode=%s, busy_timeout=30000ms\n", journalMode)

	// 4. 确保表结构存在
	if err := ensureRequestLogTable(); err != nil {
		return fmt.Errorf("初始化 request_log 表失败: %w", err)
	}
	if err := ensureBlacklistTables(); err != nil {
		return fmt.Errorf("初始化黑名单表失败: %w", err)
	}
	if err := ensureProviderAliasTable(); err != nil {
		return fmt.Errorf("初始化 provider_alias 表失败: %w", err)
	}
	if err := ensureUsersTable(); err != nil {
		return fmt.Errorf("初始化 users 表失败: %w", err)
	}
	if err := ensureGatewayKeysTable(); err != nil {
		return fmt.Errorf("初始化 gateway_keys 表失败: %w", err)
	}
	if err := ensureAssistantConversationsTable(); err != nil {
		return fmt.Errorf("初始化 assistant_conversations 表失败: %w", err)
	}

	// 5. 预热连接池：强制建立数据库连接，避免首次写入时失败
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM request_log").Scan(&count); err != nil {
		fmt.Printf("⚠️  连接池预热查询失败: %v\n", err)
	} else {
		fmt.Printf("✅ 数据库连接已预热（request_log 记录数: %d）\n", count)
	}

	return nil
}

// ensureBlacklistTables 确保黑名单相关表存在
func ensureBlacklistTables() error {
	db, err := xdb.DB("default")
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 1. 创建 app_settings 表
	const createAppSettingsSQL = `CREATE TABLE IF NOT EXISTS app_settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT UNIQUE NOT NULL,
		value TEXT
	)`
	if _, err := db.Exec(createAppSettingsSQL); err != nil {
		return fmt.Errorf("创建 app_settings 表失败: %w", err)
	}

	// 2. 创建 provider_blacklist 表
	const createBlacklistSQL = `CREATE TABLE IF NOT EXISTS provider_blacklist (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		platform TEXT NOT NULL,
		provider_name TEXT NOT NULL,
		failure_count INTEGER DEFAULT 0,
		blacklisted_at DATETIME,
		blacklisted_until DATETIME,
		last_failure_at DATETIME,
		blacklist_level INTEGER DEFAULT 0,
		last_recovered_at DATETIME,
		last_degrade_hour INTEGER DEFAULT 0,
		last_failure_window_start DATETIME,
		auto_recovered INTEGER DEFAULT 0,
		UNIQUE(platform, provider_name)
	)`
	if _, err := db.Exec(createBlacklistSQL); err != nil {
		return fmt.Errorf("创建 provider_blacklist 表失败: %w", err)
	}

	// 3. 确保 app_settings 中有默认的黑名单配置
	defaultSettings := []struct {
		key   string
		value string
	}{
		{"enable_blacklist", "false"},
		{"blacklist_failure_threshold", "3"},
		{"blacklist_duration_minutes", "30"},
	}

	for _, s := range defaultSettings {
		_, err := db.Exec(`
			INSERT OR IGNORE INTO app_settings (key, value) VALUES (?, ?)
		`, s.key, s.value)
		if err != nil {
			return fmt.Errorf("插入默认设置 %s 失败: %w", s.key, err)
		}
	}

	return nil
}

// ensureProviderAliasTable 创建 provider_alias 表,用于 rename 后 48h 内承接旧名 in-flight 写入。
func ensureProviderAliasTable() error {
	db, err := xdb.DB("default")
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	const createSQL = `CREATE TABLE IF NOT EXISTS provider_alias (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		platform TEXT NOT NULL,
		provider_id INTEGER NOT NULL,
		alias_name TEXT NOT NULL COLLATE NOCASE,
		canonical_name TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME NOT NULL,
		UNIQUE(platform, alias_name)
	)`
	if _, err := db.Exec(createSQL); err != nil {
		return fmt.Errorf("创建 provider_alias 表失败: %w", err)
	}

	const createIndexSQL = `
		CREATE INDEX IF NOT EXISTS idx_provider_alias_pid ON provider_alias(platform, provider_id);
		CREATE INDEX IF NOT EXISTS idx_provider_alias_expires ON provider_alias(expires_at);
	`
	if _, err := db.Exec(createIndexSQL); err != nil {
		return fmt.Errorf("创建 provider_alias 索引失败: %w", err)
	}

	return nil
}

// ensureUsersTable 创建 users 表，用于用户认证
func ensureUsersTable() error {
	db, err := xdb.DB("default")
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	const createSQL = `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT DEFAULT 'admin',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`
	if _, err := db.Exec(createSQL); err != nil {
		return fmt.Errorf("创建 users 表失败: %w", err)
	}

	return nil
}

// ensureGatewayKeysTable 创建 gateway_keys 表，用于 API 网关密钥管理
func ensureGatewayKeysTable() error {
	db, err := xdb.DB("default")
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	const createSQL = `CREATE TABLE IF NOT EXISTS gateway_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		key_hash TEXT UNIQUE NOT NULL,
		key_prefix TEXT NOT NULL,
		rate_limit INTEGER DEFAULT 60,
		enabled BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_used_at DATETIME
	)`
	if _, err := db.Exec(createSQL); err != nil {
		return fmt.Errorf("创建 gateway_keys 表失败: %w", err)
	}

	return nil
}

// ensureAssistantConversationsTable 创建 assistant_conversations 表，用于 AI 助手对话历史
func ensureAssistantConversationsTable() error {
	db, err := xdb.DB("default")
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	const createSQL = `CREATE TABLE IF NOT EXISTS assistant_conversations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		model TEXT,
		tokens_used INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	)`
	if _, err := db.Exec(createSQL); err != nil {
		return fmt.Errorf("创建 assistant_conversations 表失败: %w", err)
	}

	return nil
}

// CreateDefaultAdmin 首次启动时创建默认管理员账户。
// 如果 users 表中已有用户，则跳过创建。
// username 和 password 从环境变量（ADMIN_USERNAME、ADMIN_PASSWORD）读取。
func CreateDefaultAdmin(username, password string) error {
	db, err := xdb.DB("default")
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 检查是否已有用户
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return fmt.Errorf("查询用户数失败: %w", err)
	}
	if count > 0 {
		return nil // 已有用户，跳过
	}

	// 使用 bcrypt 哈希密码
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码哈希失败: %w", err)
	}

	// 插入默认管理员
	_, err = db.Exec(
		"INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)",
		username, string(hash), "admin",
	)
	if err != nil {
		return fmt.Errorf("创建默认管理员失败: %w", err)
	}

	fmt.Printf("✅ 默认管理员账户已创建 (用户名: %s)\n", username)
	return nil
}
