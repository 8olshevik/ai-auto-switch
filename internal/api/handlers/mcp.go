package handlers

import (
	"net/http"

	"codeswitch/services"

	"github.com/gin-gonic/gin"
)

// McpHandler handles MCP server management HTTP endpoints.
type McpHandler struct {
	svc *services.MCPService
}

// NewMcpHandler creates a new McpHandler with the given MCPService.
func NewMcpHandler(svc *services.MCPService) *McpHandler {
	return &McpHandler{svc: svc}
}

// ListServers handles GET /mcp/.
// Returns the list of all configured MCP servers.
func (h *McpHandler) List(c *gin.Context) {
	servers, err := h.svc.ListServers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch MCP servers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": servers})
}

// AddServer handles POST /mcp/.
// Adds a new MCP server to the list and syncs to config files.
func (h *McpHandler) Create(c *gin.Context) {
	var input services.MCPServer
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if input.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "server name is required"})
		return
	}
	if input.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "server type is required"})
		return
	}

	// Load existing servers
	servers, err := h.svc.ListServers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load MCP servers"})
		return
	}

	// Check for duplicate name
	for _, s := range servers {
		if s.Name == input.Name {
			c.JSON(http.StatusConflict, gin.H{"error": "server with this name already exists"})
			return
		}
	}

	// Append new server and save all
	servers = append(servers, input)
	if err := h.svc.SaveServers(servers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "MCP server added successfully"})
}

// UpdateServer handles PUT /mcp/:id.
// Updates an existing MCP server by name (id = server name) and syncs to config files.
func (h *McpHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "server id is required"})
		return
	}

	var input services.MCPServer
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Load existing servers
	servers, err := h.svc.ListServers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load MCP servers"})
		return
	}

	// Find and update the target server
	found := false
	for i, s := range servers {
		if s.Name == id {
			// If name is being changed, use the new name; otherwise keep the original
			if input.Name == "" {
				input.Name = id
			}
			servers[i] = input
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "MCP server not found"})
		return
	}

	if err := h.svc.SaveServers(servers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "MCP server updated successfully"})
}

// DeleteServer handles DELETE /mcp/:id.
// Removes an MCP server by name (id = server name) and syncs to config files.
func (h *McpHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "server id is required"})
		return
	}

	// Load existing servers
	servers, err := h.svc.ListServers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load MCP servers"})
		return
	}

	// Find and remove the target server
	found := false
	filtered := make([]services.MCPServer, 0, len(servers))
	for _, s := range servers {
		if s.Name == id {
			found = true
			continue
		}
		filtered = append(filtered, s)
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "MCP server not found"})
		return
	}

	if err := h.svc.SaveServers(filtered); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "MCP server deleted successfully"})
}
