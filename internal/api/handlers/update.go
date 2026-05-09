package handlers

import (
	"net/http"

	"codeswitch/services"

	"github.com/gin-gonic/gin"
)

// AppVersion is the current application version
var AppVersion = "v2.6.30"

// UpdateHandler handles update check and install endpoints.
type UpdateHandler struct {
	svc *services.UpdateService
}

// NewUpdateHandler creates a new UpdateHandler.
func NewUpdateHandler(svc *services.UpdateService) *UpdateHandler {
	return &UpdateHandler{svc: svc}
}

// GetVersion handles GET /update/version — returns current app version.
func (h *UpdateHandler) GetVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": AppVersion})
}

// CheckUpdate handles GET /update/check — checks for available updates.
func (h *UpdateHandler) CheckUpdate(c *gin.Context) {
	info, err := h.svc.CheckUpdate()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if info == nil {
		c.JSON(http.StatusOK, gin.H{"available": false, "state": h.svc.GetState()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"available": true, "info": info, "state": h.svc.GetState()})
}

// GetState handles GET /update/state — returns current update state.
func (h *UpdateHandler) GetState(c *gin.Context) {
	state := h.svc.GetState()
	c.JSON(http.StatusOK, state)
}

// DownloadUpdate handles POST /update/download — starts downloading the update.
func (h *UpdateHandler) DownloadUpdate(c *gin.Context) {
	if err := h.svc.DownloadUpdate(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "download started"})
}

// InstallUpdate handles POST /update/install — requests restart to apply update.
func (h *UpdateHandler) InstallUpdate(c *gin.Context) {
	if err := h.svc.RequestRestart(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "restart requested"})
}

// dismissRequest represents the request body for dismissing an update.
type dismissRequest struct {
	Version string `json:"version" binding:"required"`
}

// DismissUpdate handles POST /update/dismiss — dismisses a specific version.
func (h *UpdateHandler) DismissUpdate(c *gin.Context) {
	var req dismissRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "version is required"})
		return
	}

	if err := h.svc.DismissUpdate(req.Version); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "update dismissed"})
}

// CancelDownload handles POST /update/cancel — cancels an ongoing download.
func (h *UpdateHandler) CancelDownload(c *gin.Context) {
	if err := h.svc.CancelDownload(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "download cancelled"})
}
