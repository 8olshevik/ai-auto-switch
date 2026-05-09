package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware returns a Gin middleware that handles Cross-Origin Resource Sharing.
// allowedOrigins is a comma-separated list of allowed origins (e.g., "http://localhost:3000,http://example.com").
// If allowedOrigins is "*", all origins are allowed.
func CORSMiddleware(allowedOrigins string) gin.HandlerFunc {
	origins := parseOrigins(allowedOrigins)
	allowAll := len(origins) == 1 && origins[0] == "*"

	return func(c *gin.Context) {
		requestOrigin := c.GetHeader("Origin")

		if allowAll {
			setCORSHeaders(c, requestOrigin)
		} else if isOriginAllowed(requestOrigin, origins) {
			setCORSHeaders(c, requestOrigin)
		} else if requestOrigin != "" {
			// Origin is present but not in the allowed list — reject
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// Handle preflight OPTIONS requests
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// setCORSHeaders sets the standard CORS response headers.
func setCORSHeaders(c *gin.Context, origin string) {
	c.Header("Access-Control-Allow-Origin", origin)
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
	c.Header("Access-Control-Allow-Credentials", "true")
	c.Header("Access-Control-Max-Age", "86400")
}

// parseOrigins splits a comma-separated origins string into a trimmed slice.
func parseOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	if len(origins) == 0 {
		return []string{"*"}
	}
	return origins
}

// isOriginAllowed checks whether the given origin is in the allowed list.
func isOriginAllowed(origin string, allowed []string) bool {
	for _, a := range allowed {
		if a == origin {
			return true
		}
	}
	return false
}
