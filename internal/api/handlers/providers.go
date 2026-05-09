package handlers

import (
	"net/http"
	"strconv"

	"codeswitch/services"

	"github.com/gin-gonic/gin"
)

// ProviderHandler handles provider management HTTP endpoints.
type ProviderHandler struct {
	svc *services.ProviderService
}

// NewProviderHandler creates a new ProviderHandler with the given ProviderService.
func NewProviderHandler(svc *services.ProviderService) *ProviderHandler {
	return &ProviderHandler{svc: svc}
}

// GetProviders handles GET /providers/:kind.
// Returns the list of providers for the given platform kind.
func (h *ProviderHandler) GetProviders(c *gin.Context) {
	kind := c.Param("kind")
	if kind == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kind is required"})
		return
	}

	providers, err := h.svc.LoadProviders(kind)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if providers == nil {
		providers = []services.Provider{}
	}

	c.JSON(http.StatusOK, providers)
}

// SaveProviders handles POST /providers/:kind.
// Saves the full list of providers for the given platform kind.
// The API Key is encrypted by the service layer.
func (h *ProviderHandler) SaveProviders(c *gin.Context) {
	kind := c.Param("kind")
	if kind == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kind is required"})
		return
	}

	var providers []services.Provider
	if err := c.ShouldBindJSON(&providers); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	if err := h.svc.SaveProviders(kind, providers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "providers saved successfully"})
}

// DuplicateProvider handles POST /providers/:kind/duplicate/:id.
// Creates a copy of the specified provider.
func (h *ProviderHandler) DuplicateProvider(c *gin.Context) {
	kind := c.Param("kind")
	if kind == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kind is required"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider id"})
		return
	}

	provider, err := h.svc.DuplicateProvider(kind, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, provider)
}

// renameRequest represents the JSON body for the rename endpoint.
type renameRequest struct {
	NewName string `json:"newName" binding:"required"`
}

// RenameProvider handles PUT /providers/:kind/:id/rename.
// Renames the specified provider.
func (h *ProviderHandler) RenameProvider(c *gin.Context) {
	kind := c.Param("kind")
	if kind == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kind is required"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider id"})
		return
	}

	var req renameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "newName is required"})
		return
	}

	if err := h.svc.RenameProvider(kind, id, req.NewName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "provider renamed successfully"})
}

// reorderRequest represents the JSON body for the reorder endpoint.
type reorderRequest struct {
	IDs []int64 `json:"ids" binding:"required"`
}

// ReorderProviders handles POST /providers/:kind/reorder.
// Reorders providers according to the given list of IDs.
func (h *ProviderHandler) ReorderProviders(c *gin.Context) {
	kind := c.Param("kind")
	if kind == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kind is required"})
		return
	}

	var req reorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids is required"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids must not be empty"})
		return
	}

	// Load current providers
	providers, err := h.svc.LoadProviders(kind)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build a map of ID -> Provider for quick lookup
	providerMap := make(map[int64]services.Provider, len(providers))
	for _, p := range providers {
		providerMap[p.ID] = p
	}

	// Reorder according to the given IDs
	reordered := make([]services.Provider, 0, len(providers))
	for _, id := range req.IDs {
		if p, ok := providerMap[id]; ok {
			reordered = append(reordered, p)
			delete(providerMap, id)
		}
	}

	// Append any providers not in the IDs list (preserve them at the end)
	for _, p := range providers {
		if _, ok := providerMap[p.ID]; ok {
			reordered = append(reordered, p)
		}
	}

	// Save the reordered list
	if err := h.svc.SaveProviders(kind, reordered); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "providers reordered successfully"})
}
