package handlers

import (
	"database/sql"
	"net/http"

	"codeswitch/internal/api/middleware"

	"github.com/daodao97/xgo/xdb"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication-related HTTP endpoints.
type AuthHandler struct {
	jwtSecret string
}

// NewAuthHandler creates a new AuthHandler with the given JWT secret.
func NewAuthHandler(jwtSecret string) *AuthHandler {
	return &AuthHandler{jwtSecret: jwtSecret}
}

// loginRequest represents the JSON body for the login endpoint.
type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// loginResponse represents the JSON response for a successful login.
type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

// userInfoResponse represents the JSON response for the /auth/me endpoint.
type userInfoResponse struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// Login handles POST /auth/login.
// It validates the username and password against the database, and returns
// JWT access and refresh tokens on success.
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}

	// Query the user from the database
	db, err := xdb.DB("default")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection error"})
		return
	}

	var userID int64
	var username, passwordHash, role string
	err = db.QueryRow(
		"SELECT id, username, password_hash, role FROM users WHERE username = ?",
		req.Username,
	).Scan(&userID, &username, &passwordHash, &role)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database query error"})
		return
	}

	// Verify password with bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	// Generate tokens
	accessToken, err := middleware.GenerateToken(userID, username, role, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate access token"})
		return
	}

	refreshToken, err := middleware.GenerateRefreshToken(userID, username, role, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate refresh token"})
		return
	}

	c.JSON(http.StatusOK, loginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	})
}

// Logout handles POST /auth/logout.
// Since JWT is stateless, this endpoint simply acknowledges the logout request.
// The client is responsible for discarding the token.
func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// GetMe handles GET /auth/me.
// It reads user information from the gin context (set by the auth middleware)
// and returns it to the client.
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, _ := c.Get("userID")
	username, _ := c.Get("username")
	role, _ := c.Get("role")

	c.JSON(http.StatusOK, userInfoResponse{
		UserID:   userID.(int64),
		Username: username.(string),
		Role:     role.(string),
	})
}
