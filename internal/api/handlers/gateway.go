package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/daodao97/xgo/xdb"
	"github.com/gin-gonic/gin"
)

// GatewayHandler handles API Gateway endpoints for key management and usage stats.
type GatewayHandler struct{}

// NewGatewayHandler creates a new GatewayHandler.
func NewGatewayHandler() *GatewayHandler {
	return &GatewayHandler{}
}

// GatewayKey represents an API key in the database.
type GatewayKey struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	KeyHash     string     `json:"-"`
	KeyPrefix   string     `json:"keyPrefix"`
	RateLimit   int        `json:"rateLimit"`
	Enabled     bool       `json:"enabled"`
	CreatedAt   time.Time  `json:"createdAt"`
	LastUsedAt  *time.Time `json:"lastUsedAt,omitempty"`
}

// KeyResponse represents the response when creating a new API key.
type KeyResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Key       string    `json:"key"`       // The plain key (only returned once on creation)
	KeyPrefix string    `json:"keyPrefix"`
	RateLimit int       `json:"rateLimit"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
}

// ListKeyResponse represents an API key in the list response.
type ListKeyResponse struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	KeyPrefix string     `json:"keyPrefix"`
	RateLimit int        `json:"rateLimit"`
	Enabled   bool       `json:"enabled"`
	CreatedAt time.Time  `json:"createdAt"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
}

// createKeyRequest represents the request body for creating an API key.
type createKeyRequest struct {
	Name      string `json:"name" binding:"required"`
	RateLimit *int   `json:"rateLimit"`
}

// updateRateLimitRequest represents the request body for setting rate limit.
type updateRateLimitRequest struct {
	KeyID     int `json:"keyId" binding:"required"`
	RateLimit int `json:"rateLimit" binding:"required,min=1"`
}

// UsageStats represents gateway usage statistics.
type UsageStats struct {
	TotalRequests int            `json:"totalRequests"`
	BySourceApp   map[string]int `json:"bySourceApp"`
	Period        string         `json:"period"`
}

// generateKey generates a random API key (32 bytes = 64 hex characters).
func generateKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// hashKey hashes the API key using SHA256.
func hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// getKeyPrefix returns the first 8 characters of the key.
func getKeyPrefix(key string) string {
	if len(key) > 8 {
		return key[:8]
	}
	return key
}

// CreateKey handles POST /gateway/keys — creates a new API key.
func (h *GatewayHandler) CreateKey(c *gin.Context) {
	var req createKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	// Generate a random API key
	plainKey, err := generateKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Hash the key for storage
	keyHash := hashKey(plainKey)
	keyPrefix := getKeyPrefix(plainKey)

	// Set default rate limit if not provided
	rateLimit := 60
	if req.RateLimit != nil && *req.RateLimit > 0 {
		rateLimit = *req.RateLimit
	}

	// Insert into database
	db, err := xdb.DB("default")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
		return
	}

	result, err := db.Exec(`
		INSERT INTO gateway_keys (name, key_hash, key_prefix, rate_limit, enabled, created_at)
		VALUES (?, ?, ?, ?, 1, datetime('now'))
	`, req.Name, keyHash, keyPrefix, rateLimit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create key: %v", err)})
		return
	}

	id, _ := result.LastInsertId()

	// Return the plain key (only available once)
	c.JSON(http.StatusCreated, KeyResponse{
		ID:        int(id),
		Name:      req.Name,
		Key:       plainKey,
		KeyPrefix: keyPrefix,
		RateLimit: rateLimit,
		Enabled:   true,
		CreatedAt: time.Now(),
	})
}

// ListKeys handles GET /gateway/keys — returns all API keys.
func (h *GatewayHandler) ListKeys(c *gin.Context) {
	db, err := xdb.DB("default")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
		return
	}

	rows, err := db.Query(`
		SELECT id, name, key_prefix, rate_limit, enabled, created_at, last_used_at
		FROM gateway_keys
		ORDER BY created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to list keys: %v", err)})
		return
	}
	defer rows.Close()

	var keys []ListKeyResponse
	for rows.Next() {
		var key ListKeyResponse
		var createdAt, lastUsedAt *time.Time

		err := rows.Scan(&key.ID, &key.Name, &key.KeyPrefix, &key.RateLimit, &key.Enabled, &createdAt, &lastUsedAt)
		if err != nil {
			continue
		}

		if createdAt != nil {
			key.CreatedAt = *createdAt
		}
		key.LastUsedAt = lastUsedAt

		keys = append(keys, key)
	}

	if keys == nil {
		keys = []ListKeyResponse{}
	}

	c.JSON(http.StatusOK, keys)
}

// DeleteKey handles DELETE /gateway/keys/:id — deletes an API key.
func (h *GatewayHandler) DeleteKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key id is required"})
		return
	}

	db, err := xdb.DB("default")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
		return
	}

	result, err := db.Exec("DELETE FROM gateway_keys WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to delete key: %v", err)})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "key deleted"})
}

// GetStats handles GET /gateway/stats — returns gateway usage statistics.
func (h *GatewayHandler) GetStats(c *gin.Context) {
	sourceApp := c.Query("sourceApp")
	period := c.DefaultQuery("period", "7d") // 7d, 24h, 30d

	// Calculate the time range based on period
	var startTime time.Time
	now := time.Now()
	switch period {
	case "24h":
		startTime = now.Add(-24 * time.Hour)
	case "7d":
		startTime = now.Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = now.Add(-30 * 24 * time.Hour)
	default:
		startTime = now.Add(-7 * 24 * time.Hour)
	}

	db, err := xdb.DB("default")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
		return
	}

	// Build the query based on whether sourceApp filter is provided
	var query string
	var args []interface{}

	if sourceApp != "" {
		query = `
			SELECT COUNT(*) as total, source_app
			FROM request_log
			WHERE source_app = ? AND created_at > ?
			GROUP BY source_app
		`
		args = []interface{}{sourceApp, startTime.Format(time.RFC3339)}
	} else {
		query = `
			SELECT COALESCE(SUM(request_count), 0), source_app
			FROM (
				SELECT COUNT(*) as request_count, source_app
				FROM request_log
				WHERE created_at > ?
				GROUP BY source_app
			)
			GROUP BY source_app
		`
		args = []interface{}{startTime.Format(time.RFC3339)}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get stats: %v", err)})
		return
	}
	defer rows.Close()

	bySourceApp := make(map[string]int)
	var totalRequests int

	for rows.Next() {
		var count int
		var app string

		if sourceApp != "" {
			err := rows.Scan(&count)
			if err != nil {
				continue
			}
			app = sourceApp
		} else {
			err := rows.Scan(&count, &app)
			if err != nil {
				continue
			}
		}

		totalRequests += count
		bySourceApp[app] = count
	}

	c.JSON(http.StatusOK, UsageStats{
		TotalRequests: totalRequests,
		BySourceApp:   bySourceApp,
		Period:        period,
	})
}

// UpdateRateLimit handles PUT /gateway/rate-limit — sets rate limit for a key.
func (h *GatewayHandler) UpdateRateLimit(c *gin.Context) {
	var req updateRateLimitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "keyId and rateLimit (min 1) are required"})
		return
	}

	db, err := xdb.DB("default")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
		return
	}

	result, err := db.Exec("UPDATE gateway_keys SET rate_limit = ? WHERE id = ?", req.RateLimit, req.KeyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to update rate limit: %v", err)})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"keyId":     req.KeyID,
		"rateLimit": req.RateLimit,
		"message":   "rate limit updated",
	})
}

// ToggleKey handles POST /gateway/keys/:id/toggle — enables or disables an API key.
func (h *GatewayHandler) ToggleKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key id is required"})
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "enabled is required"})
		return
	}

	db, err := xdb.DB("default")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
		return
	}

	result, err := db.Exec("UPDATE gateway_keys SET enabled = ? WHERE id = ?", req.Enabled, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to toggle key: %v", err)})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"enabled": req.Enabled,
		"message": "key status updated",
	})
}