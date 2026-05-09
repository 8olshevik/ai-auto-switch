package handlers

import (
	"net/http"

	"codeswitch/services"

	"github.com/gin-gonic/gin"
)

// PromptsHandler handles prompt management endpoints.
type PromptsHandler struct {
	svc *services.PromptService
}

// NewPromptsHandler creates a new PromptsHandler.
func NewPromptsHandler(svc *services.PromptService) *PromptsHandler {
	return &PromptsHandler{svc: svc}
}

// GetPrompts handles GET /prompts/:platform.
func (h *PromptsHandler) GetPrompts(c *gin.Context) {
	platform := c.Param("platform")
	if platform == "" {
		platform = "claude"
	}

	prompts, err := h.svc.GetPrompts(platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, prompts)
}

// upsertPromptRequest represents the request body for creating/updating a prompt.
type upsertPromptRequest struct {
	ID      string          `json:"id" binding:"required"`
	Prompt  services.Prompt `json:"prompt" binding:"required"`
}

// CreatePrompt handles POST /prompts/:platform.
func (h *PromptsHandler) CreatePrompt(c *gin.Context) {
	platform := c.Param("platform")
	if platform == "" {
		platform = "claude"
	}

	var req upsertPromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id and prompt are required"})
		return
	}

	if req.Prompt.Name == "" || req.Prompt.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt name and content are required"})
		return
	}

	if err := h.svc.UpsertPrompt(platform, req.ID, req.Prompt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "prompt created"})
}

// UpdatePrompt handles PUT /prompts/:platform/:id.
func (h *PromptsHandler) UpdatePrompt(c *gin.Context) {
	platform := c.Param("platform")
	id := c.Param("id")
	if platform == "" || id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "platform and id are required"})
		return
	}

	var prompt services.Prompt
	if err := c.ShouldBindJSON(&prompt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.svc.UpsertPrompt(platform, id, prompt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "prompt updated"})
}

// DeletePrompt handles DELETE /prompts/:platform/:id.
func (h *PromptsHandler) DeletePrompt(c *gin.Context) {
	platform := c.Param("platform")
	id := c.Param("id")
	if platform == "" || id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "platform and id are required"})
		return
	}

	if err := h.svc.DeletePrompt(platform, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "prompt deleted"})
}

// EnablePrompt handles POST /prompts/:platform/:id/enable.
func (h *PromptsHandler) EnablePrompt(c *gin.Context) {
	platform := c.Param("platform")
	id := c.Param("id")
	if platform == "" || id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "platform and id are required"})
		return
	}

	if err := h.svc.EnablePrompt(platform, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "prompt enabled"})
}

// GetCurrentFileContent handles GET /prompts/:platform/file.
func (h *PromptsHandler) GetCurrentFileContent(c *gin.Context) {
	platform := c.Param("platform")
	if platform == "" {
		platform = "claude"
	}

	content, err := h.svc.GetCurrentFileContent(platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"content": content})
}
