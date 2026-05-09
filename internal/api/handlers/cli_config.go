package handlers

import (
	"net/http"

	"codeswitch/services"

	"github.com/gin-gonic/gin"
)

// CLIConfigHandler handles CLI configuration endpoints.
type CLIConfigHandler struct {
	svc *services.CliConfigService
}

// NewCLIConfigHandler creates a new CLIConfigHandler.
func NewCLIConfigHandler(svc *services.CliConfigService) *CLIConfigHandler {
	return &CLIConfigHandler{svc: svc}
}

// GetConfig handles GET /cli-config/:platform.
func (h *CLIConfigHandler) GetConfig(c *gin.Context) {
	platform := c.Param("platform")
	if platform == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "platform is required"})
		return
	}

	config, err := h.svc.GetConfig(platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// saveConfigRequest represents the request body for saving CLI config.
type saveConfigRequest struct {
	Editable map[string]interface{} `json:"editable"`
}

// SaveConfig handles PUT /cli-config/:platform.
func (h *CLIConfigHandler) SaveConfig(c *gin.Context) {
	platform := c.Param("platform")
	if platform == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "platform is required"})
		return
	}

	var req saveConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.svc.SaveConfig(platform, req.Editable); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "config saved"})
}

// getSnapshotsRequest represents query params for config snapshots.
type getSnapshotsRequest struct {
	APIUrl      string `form:"apiUrl"`
	APIKey      string `form:"apiKey"`
	PreviewMode string `form:"previewMode"`
}

// GetConfigSnapshots handles GET /cli-config/:platform/snapshots.
func (h *CLIConfigHandler) GetConfigSnapshots(c *gin.Context) {
	platform := c.Param("platform")
	if platform == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "platform is required"})
		return
	}

	var req getSnapshotsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		return
	}

	snapshots, err := h.svc.GetConfigSnapshots(platform, req.APIUrl, req.APIKey, req.PreviewMode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, snapshots)
}

// saveConfigFileContentRequest represents the request body for saving raw config file content.
type saveConfigFileContentRequest struct {
	FilePath string `json:"filePath" binding:"required"`
	Content  string `json:"content" binding:"required"`
}

// SaveConfigFileContent handles PUT /cli-config/:platform/file.
func (h *CLIConfigHandler) SaveConfigFileContent(c *gin.Context) {
	platform := c.Param("platform")
	if platform == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "platform is required"})
		return
	}

	var req saveConfigFileContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filePath and content are required"})
		return
	}

	if err := h.svc.SaveConfigFileContent(platform, req.FilePath, req.Content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "file saved"})
}

// GetTemplate handles GET /cli-config/:platform/template.
func (h *CLIConfigHandler) GetTemplate(c *gin.Context) {
	platform := c.Param("platform")
	if platform == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "platform is required"})
		return
	}

	template, err := h.svc.GetTemplate(platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

// setTemplateRequest represents the request body for setting a template.
type setTemplateRequest struct {
	Template        map[string]interface{} `json:"template" binding:"required"`
	IsGlobalDefault bool                   `json:"isGlobalDefault"`
}

// SetTemplate handles PUT /cli-config/:platform/template.
func (h *CLIConfigHandler) SetTemplate(c *gin.Context) {
	platform := c.Param("platform")
	if platform == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "platform is required"})
		return
	}

	var req setTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "template is required"})
		return
	}

	if err := h.svc.SetTemplate(platform, req.Template, req.IsGlobalDefault); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "template saved"})
}

// RestoreDefault handles POST /cli-config/:platform/restore.
func (h *CLIConfigHandler) RestoreDefault(c *gin.Context) {
	platform := c.Param("platform")
	if platform == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "platform is required"})
		return
	}

	if err := h.svc.RestoreDefault(platform); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "default restored"})
}
