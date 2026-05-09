package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-key-for-unit-tests"

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken(1, "admin", "admin", testSecret)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("GenerateToken returned empty token")
	}

	// Verify the token can be parsed back
	claims, err := ParseToken(token, testSecret)
	if err != nil {
		t.Fatalf("ParseToken failed for generated token: %v", err)
	}
	if claims.UserID != 1 {
		t.Errorf("expected UserID=1, got %d", claims.UserID)
	}
	if claims.Username != "admin" {
		t.Errorf("expected Username=admin, got %s", claims.Username)
	}
	if claims.Role != "admin" {
		t.Errorf("expected Role=admin, got %s", claims.Role)
	}
	if claims.Issuer != "codeswitch" {
		t.Errorf("expected Issuer=codeswitch, got %s", claims.Issuer)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	token, err := GenerateRefreshToken(2, "user1", "user", testSecret)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("GenerateRefreshToken returned empty token")
	}

	claims, err := ParseToken(token, testSecret)
	if err != nil {
		t.Fatalf("ParseToken failed for refresh token: %v", err)
	}
	if claims.UserID != 2 {
		t.Errorf("expected UserID=2, got %d", claims.UserID)
	}

	// Verify expiry is approximately 7 days from now
	expectedExpiry := time.Now().Add(RefreshTokenExpiry)
	diff := claims.ExpiresAt.Time.Sub(expectedExpiry)
	if diff > time.Minute || diff < -time.Minute {
		t.Errorf("refresh token expiry not within expected range: got %v, expected ~%v", claims.ExpiresAt.Time, expectedExpiry)
	}
}

func TestParseToken_InvalidSecret(t *testing.T) {
	token, _ := GenerateToken(1, "admin", "admin", testSecret)

	_, err := ParseToken(token, "wrong-secret")
	if err == nil {
		t.Fatal("ParseToken should fail with wrong secret")
	}
}

func TestParseToken_ExpiredToken(t *testing.T) {
	// Create a token that's already expired
	now := time.Now()
	claims := Claims{
		UserID:   1,
		Username: "admin",
		Role:     "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			Issuer:    "codeswitch",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(testSecret))

	_, err := ParseToken(tokenString, testSecret)
	if err == nil {
		t.Fatal("ParseToken should fail for expired token")
	}
}

func TestParseToken_MalformedToken(t *testing.T) {
	_, err := ParseToken("not-a-valid-token", testSecret)
	if err == nil {
		t.Fatal("ParseToken should fail for malformed token")
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	token, _ := GenerateToken(1, "admin", "admin", testSecret)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	r.Use(AuthMiddleware(testSecret))
	r.GET("/test", func(c *gin.Context) {
		userID, _ := c.Get("userID")
		username, _ := c.Get("username")
		role, _ := c.Get("role")
		c.JSON(http.StatusOK, gin.H{
			"userID":   userID,
			"username": username,
			"role":     role,
		})
	})

	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(AuthMiddleware(testSecret))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(AuthMiddleware(testSecret))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat token123")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(AuthMiddleware(testSecret))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-string")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	// Create an expired token
	now := time.Now()
	claims := Claims{
		UserID:   1,
		Username: "admin",
		Role:     "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			Issuer:    "codeswitch",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(testSecret))

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(AuthMiddleware(testSecret))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}
