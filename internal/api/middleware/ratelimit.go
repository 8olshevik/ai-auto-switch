package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/daodao97/xgo/xdb"
	"github.com/gin-gonic/gin"
)

// bucket represents a token bucket for rate limiting a specific API key.
type bucket struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

// bucketStore stores token buckets for each API key.
var bucketStore = &sync.Map{} // map[string]*bucket

// getBucket retrieves or creates a token bucket for the given API key hash.
func getBucket(keyHash string, rateLimit int) *bucket {
	// rateLimit is requests per minute, convert to tokens per second
	refillRate := float64(rateLimit) / 60.0

	if existing, ok := bucketStore.Load(keyHash); ok {
		b := existing.(*bucket)
		b.mu.Lock()
		b.refillRate = refillRate
		b.mu.Unlock()
		return b
	}

	now := time.Now()
	b := &bucket{
		tokens:     float64(rateLimit),
		maxTokens:  float64(rateLimit),
		refillRate: refillRate,
		lastRefill: now,
	}
	bucketStore.Store(keyHash, b)
	return b
}

// refill adds tokens based on time elapsed since last refill.
func (b *bucket) refill() {
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	added := elapsed * b.refillRate
	b.tokens = min(b.maxTokens, b.tokens+added)
	b.lastRefill = now
}

// consume attempts to consume one token from the bucket.
// Returns true if successful, false if rate limited.
func (b *bucket) consume() bool {
	b.refill()
	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// remainingTokens returns the number of tokens remaining in the bucket.
func (b *bucket) remainingTokens() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.refill()
	return b.tokens
}

// lookupKey looks up an API key in the database by its hash.
// Returns the rate_limit if found and enabled, or -1 if not found or disabled.
func lookupKey(keyHash string) (rateLimit int, err error) {
	db, err := xdb.DB("default")
	if err != nil {
		return -1, fmt.Errorf("database connection failed: %w", err)
	}

	var rateLimitVal int
	var enabled bool
	err = db.QueryRow(`
		SELECT rate_limit, enabled 
		FROM gateway_keys 
		WHERE key_hash = ?
	`, keyHash).Scan(&rateLimitVal, &enabled)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return -1, nil // Key not found
		}
		return -1, fmt.Errorf("database query failed: %w", err)
	}

	if !enabled {
		return -1, nil // Key is disabled
	}

	return rateLimitVal, nil
}

// extractAPIKey extracts the API key from the request.
// Supports both "Authorization: Bearer <key>" and "X-API-Key: <key>" headers.
func extractAPIKey(c *gin.Context) string {
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

// hashKey hashes the API key using SHA256.
func hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// RateLimitMiddleware returns a Gin middleware that implements token bucket rate limiting.
// It extracts the API key from the request, looks up the rate limit configuration from the
// gateway_keys table, and allows/rejects requests based on token availability.
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract API key from request
		apiKey := extractAPIKey(c)
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "API key is required (use X-API-Key header or Authorization: Bearer <key>)",
			})
			return
		}

		// Hash the key for lookup
		keyHash := hashKey(apiKey)

		// Look up the key in the database to get rate limit config
		rateLimit, err := lookupKey(keyHash)
		if err != nil {
			// Log the error but don't expose internal details to client
			fmt.Printf("[RATELIMIT] Error looking up key: %v\n", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
			return
		}

		// If key not found or disabled, reject the request
		if rateLimit < 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or disabled API key",
			})
			return
		}

		// Get or create token bucket for this key
		bucket := getBucket(keyHash, rateLimit)

		// Try to consume a token
		if !bucket.consume() {
			// Rate limited - return 429
			remaining := int(bucket.remainingTokens())
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rateLimit))
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			c.Header("X-RateLimit-Reset", "60") // Reset after 60 seconds
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":            "rate limit exceeded",
				"rateLimit":        rateLimit,
				"rateLimitResetIn": 60,
			})
			return
		}

		// Request allowed - set rate limit headers
		remaining := int(bucket.remainingTokens())
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rateLimit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", "60")

		// Store the key hash in context for downstream handlers
		c.Set("apiKeyHash", keyHash)

		// Update last_used_at in database (async to not block the request)
		go func() {
			db, err := xdb.DB("default")
			if err != nil {
				fmt.Printf("[RATELIMIT] Failed to get database connection: %v\n", err)
				return
			}
			_, err = db.Exec("UPDATE gateway_keys SET last_used_at = datetime('now') WHERE key_hash = ?", keyHash)
			if err != nil {
				fmt.Printf("[RATELIMIT] Failed to update last_used_at: %v\n", err)
			}
		}()

		c.Next()
	}
}