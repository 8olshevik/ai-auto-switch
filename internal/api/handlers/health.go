package handlers

import (
	"net/http"
	"strconv"

	"codeswitch/services"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check endpoints.
type HealthHandler struct {
	svc *services.HealthCheckService
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(svc *services.HealthCheckService) *HealthHandler {
	return &HealthHandler{svc: svc}
}

// GetLatestResults handles GET /health/ — returns latest health check results.
func (h *HealthHandler) GetLatestResults(c *gin.Context) {
	results, err := h.svc.GetLatestResults()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

// runCheckRequest represents the request body for running a health check.
type runCheckRequest struct {
	Platform   string `json:"platform" binding:"required"`
	ProviderID int64  `json:"provider_id"`
}

// RunCheck handles POST /health/check — executes a health check.
func (h *HealthHandler) RunCheck(c *gin.Context) {
	var req runCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "platform is required"})
		return
	}

	if req.ProviderID > 0 {
		// Run single provider check
		result, err := h.svc.RunSingleCheck(req.Platform, req.ProviderID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
		return
	}

	// Run all checks
	results, err := h.svc.RunAllChecks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

// GetHistory handles GET /health/history — returns health check history for a provider.
func (h *HealthHandler) GetHistory(c *gin.Context) {
	platform := c.Query("platform")
	providerName := c.Query("provider")
	limitStr := c.DefaultQuery("limit", "50")

	if platform == "" || providerName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "platform and provider query params are required"})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 50
	}

	history, err := h.svc.GetHistory(platform, providerName, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, history)
}
