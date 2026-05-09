package handlers

import (
	"net/http"

	"codeswitch/services"

	"github.com/gin-gonic/gin"
)

// BlacklistHandler handles blacklist management endpoints.
type BlacklistHandler struct {
	svc *services.BlacklistService
}

// NewBlacklistHandler creates a new BlacklistHandler.
func NewBlacklistHandler(svc *services.BlacklistService) *BlacklistHandler {
	return &BlacklistHandler{svc: svc}
}

// GetStatus handles GET /blacklist/ — returns blacklist status for a platform.
func (h *BlacklistHandler) GetStatus(c *gin.Context) {
	platform := c.DefaultQuery("platform", "claude")

	statuses, err := h.svc.GetBlacklistStatus(platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, statuses)
}

// recoverRequest represents the request body for manual recovery.
type recoverRequest struct {
	Platform     string `json:"platform" binding:"required"`
	ProviderName string `json:"provider_name" binding:"required"`
}

// ManualRecover handles POST /blacklist/recover — manually recovers a blacklisted provider.
func (h *BlacklistHandler) ManualRecover(c *gin.Context) {
	var req recoverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "platform and provider_name are required"})
		return
	}

	if err := h.svc.ManualUnblockAndReset(req.Platform, req.ProviderName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "provider recovered"})
}

// GetSettings handles GET /blacklist/settings — returns blacklist configuration.
func (h *BlacklistHandler) GetSettings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"enabled":       h.svc.IsBlacklistEnabled(),
		"level_enabled": h.svc.IsLevelBlacklistEnabled(),
		"fixed_mode":    h.svc.ShouldUseFixedMode(),
		"retry_config":  h.svc.GetRetryConfig(),
	})
}
