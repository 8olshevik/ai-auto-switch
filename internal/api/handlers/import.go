package handlers

import (
	"net/http"

	"codeswitch/services"

	"github.com/gin-gonic/gin"
)

// ImportHandler handles import/export configuration endpoints.
type ImportHandler struct {
	svc *services.ImportService
}

// NewImportHandler creates a new ImportHandler.
func NewImportHandler(svc *services.ImportService) *ImportHandler {
	return &ImportHandler{svc: svc}
}

// importConfigRequest represents the request body for importing config.
type importConfigRequest struct {
	Path string `json:"path"`
}

// ImportConfig handles POST /import/config — imports configuration from a path or default location.
func (h *ImportHandler) ImportConfig(c *gin.Context) {
	var req importConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// If no body, import from default path
		req.Path = ""
	}

	var result services.ConfigImportResult
	var err error

	if req.Path != "" {
		result, err = h.svc.ImportFromPath(req.Path)
	} else {
		result, err = h.svc.ImportAll()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ExportConfig handles GET /import/export — exports current configuration status.
func (h *ImportHandler) ExportConfig(c *gin.Context) {
	status, err := h.svc.GetStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}

// GetStatus handles GET /import/status — returns import status without performing import.
func (h *ImportHandler) GetStatus(c *gin.Context) {
	status, err := h.svc.GetStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}

// parseMCPRequest represents the request body for parsing MCP JSON.
type parseMCPRequest struct {
	JSON string `json:"json" binding:"required"`
}

// ParseMCPJSON handles POST /import/mcp/parse — parses MCP JSON for preview.
func (h *ImportHandler) ParseMCPJSON(c *gin.Context) {
	var req parseMCPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "json field is required"})
		return
	}

	result, err := h.svc.ParseMCPJSON(req.JSON)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// importMCPServersRequest represents the request body for importing MCP servers.
type importMCPServersRequest struct {
	Servers  []services.MCPServer `json:"servers" binding:"required"`
	Strategy string               `json:"strategy"`
}

// ImportMCPServers handles POST /import/mcp — imports MCP servers.
func (h *ImportHandler) ImportMCPServers(c *gin.Context) {
	var req importMCPServersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "servers array is required"})
		return
	}

	strategy := req.Strategy
	if strategy == "" {
		strategy = "skip"
	}

	count, err := h.svc.ImportMCPServers(req.Servers, strategy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"imported": count})
}
