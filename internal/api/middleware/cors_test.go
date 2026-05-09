package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupRouter(allowedOrigins string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORSMiddleware(allowedOrigins))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	r.POST("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	return r
}

func TestCORSMiddleware_AllowAll(t *testing.T) {
	router := setupRouter("*")

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://example.com" {
		t.Errorf("expected Access-Control-Allow-Origin 'http://example.com', got '%s'", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("expected Access-Control-Allow-Credentials 'true', got '%s'", got)
	}
}

func TestCORSMiddleware_AllowedOrigin(t *testing.T) {
	router := setupRouter("http://localhost:3000,http://example.com")

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("expected Access-Control-Allow-Origin 'http://localhost:3000', got '%s'", got)
	}
}

func TestCORSMiddleware_DisallowedOrigin(t *testing.T) {
	router := setupRouter("http://localhost:3000,http://example.com")

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://evil.com")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

func TestCORSMiddleware_PreflightOptions(t *testing.T) {
	router := setupRouter("http://localhost:3000")

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("expected Access-Control-Allow-Methods header to be set")
	}
	if got := w.Header().Get("Access-Control-Max-Age"); got != "86400" {
		t.Errorf("expected Access-Control-Max-Age '86400', got '%s'", got)
	}
}

func TestCORSMiddleware_PreflightDisallowedOrigin(t *testing.T) {
	router := setupRouter("http://localhost:3000")

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://evil.com")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

func TestCORSMiddleware_NoOriginHeader(t *testing.T) {
	router := setupRouter("http://localhost:3000")

	// Request without Origin header should pass through (same-origin or non-browser)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestCORSMiddleware_HeaderValues(t *testing.T) {
	router := setupRouter("*")

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://any-origin.com")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":      "http://any-origin.com",
		"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, OPTIONS, PATCH",
		"Access-Control-Allow-Headers":     "Origin, Content-Type, Accept, Authorization, X-Requested-With",
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Max-Age":           "86400",
	}

	for header, expected := range expectedHeaders {
		if got := w.Header().Get(header); got != expected {
			t.Errorf("header %s: expected '%s', got '%s'", header, expected, got)
		}
	}
}

func TestParseOrigins(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"*", []string{"*"}},
		{"http://localhost:3000", []string{"http://localhost:3000"}},
		{"http://localhost:3000,http://example.com", []string{"http://localhost:3000", "http://example.com"}},
		{" http://localhost:3000 , http://example.com ", []string{"http://localhost:3000", "http://example.com"}},
		{"", []string{"*"}},
	}

	for _, tt := range tests {
		result := parseOrigins(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("parseOrigins(%q): expected %v, got %v", tt.input, tt.expected, result)
			continue
		}
		for i, v := range result {
			if v != tt.expected[i] {
				t.Errorf("parseOrigins(%q)[%d]: expected %q, got %q", tt.input, i, tt.expected[i], v)
			}
		}
	}
}
