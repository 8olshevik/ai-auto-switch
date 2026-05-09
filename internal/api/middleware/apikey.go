package middleware

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/daodao97/xgo/xdb"
	"github.com/gin-gonic/gin"
)

// APIKeyInfo holds the validated API key information
type APIKeyInfo struct {
	KeyHash   string
	Name      string
	RateLimit int
	SourceApp string
}

// ContextKey type for context values
type ContextKey string

const (
	// APIKeyInfoKey is the context key for API key info
	APIKeyInfoKey ContextKey = "apiKeyInfo"
	// SourceAppKey is the context key for source application
	SourceAppKey ContextKey = "sourceApp"
)

// lookupAPIKey looks up an API key in the database by its hash.
// Returns the key info if found and enabled, or an error if not found or disabled.
func lookupAPIKey(keyHash string) (*APIKeyInfo, error) {
	db, err := xdb.DB("default")
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	var keyInfo APIKeyInfo
	var id int
	var enabled bool

	err = db.QueryRow(`
		SELECT id, name, key_hash, rate_limit, enabled
		FROM gateway_keys
		WHERE key_hash = ?
	`, keyHash).Scan(&id, &keyInfo.Name, &keyInfo.KeyHash, &keyInfo.RateLimit, &enabled)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Key not found
		}
		return nil, fmt.Errorf("database query failed: %w", err)
	}

	if !enabled {
		return nil, nil // Key is disabled
	}

	return &keyInfo, nil
}

// extractAPIKey extracts the API key from the request.
// Supports both "Authorization: Bearer <key>" and "X-API-Key: <key>" headers.
func extractProxyAPIKey(c *gin.Context) string {
	// Check X-API-Key header first
	if apiKey := c.GetHeader("X-API-Key"); apiKey != "" {
		return strings.TrimSpace(apiKey)
	}

	// Check Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Support "Bearer <key>" format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}

	// Also support "Authorization: <key>" (direct key)
	return strings.TrimSpace(authHeader)
}

// hashAPIKey hashes the API key using SHA256.
func hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// ProxyAPIKeyAuthMiddleware returns a Gin middleware that validates API keys for the proxy service.
// It extracts the API key from the request, validates it against the gateway_keys table,
// and stores the API key info and source app in the context for downstream handlers.
func ProxyAPIKeyAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract API key from request
		apiKey := extractProxyAPIKey(c)
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "API key is required (use X-API-Key header or Authorization: Bearer <key>)",
			})
			return
		}

		// Hash the key for lookup
		keyHash := hashAPIKey(apiKey)

		// Look up the key in the database
		keyInfo, err := lookupAPIKey(keyHash)
		if err != nil {
			// Log the error but don't expose internal details to client
			fmt.Printf("[PROXY_API_KEY] Error looking up key: %v\n", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
			return
		}

		// If key not found or disabled, reject the request
		if keyInfo == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or disabled API key",
			})
			return
		}

		// Extract source app from header or use default "unknown"
		sourceApp := c.GetHeader("X-Source-App")
		if sourceApp == "" {
			sourceApp = "unknown"
		}

		// Store API key info and source app in context
		keyInfo.SourceApp = sourceApp
		c.Set(string(APIKeyInfoKey), keyInfo)
		c.Set(string(SourceAppKey), sourceApp)

		// Update last_used_at in database
		go func() {
			db, err := xdb.DB("default")
			if err != nil {
				fmt.Printf("[PROXY_API_KEY] Failed to get database connection: %v\n", err)
				return
			}
			_, err = db.Exec("UPDATE gateway_keys SET last_used_at = datetime('now') WHERE key_hash = ?", keyHash)
			if err != nil {
				fmt.Printf("[PROXY_API_KEY] Failed to update last_used_at: %v\n", err)
			}
		}()

		fmt.Printf("[PROXY_API_KEY] Authenticated request from source_app=%s, key_name=%s\n", sourceApp, keyInfo.Name)

		c.Next()
	}
}

// GetAPIKeyInfo retrieves the API key info from the Gin context
func GetAPIKeyInfo(c *gin.Context) (*APIKeyInfo, bool) {
	value, exists := c.Get(string(APIKeyInfoKey))
	if !exists {
		return nil, false
	}
	if info, ok := value.(*APIKeyInfo); ok {
		return info, true
	}
	return nil, false
}

// GetSourceApp retrieves the source app from the Gin context
func GetSourceApp(c *gin.Context) string {
	if sourceApp, exists := c.Get(string(SourceAppKey)); exists {
		if app, ok := sourceApp.(string); ok {
			return app
		}
	}
	return "unknown"
}