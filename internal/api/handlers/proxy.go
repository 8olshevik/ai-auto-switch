package handlers

import (
	"net/http"

	"codeswitch/services"

	"github.com/gin-gonic/gin"
)

// ProxyHandler handles proxy management HTTP endpoints.
type ProxyHandler struct {
	svc *services.ProviderRelayService
}

// NewProxyHandler creates a new ProxyHandler with the given ProviderRelayService.
func NewProxyHandler(svc *services.ProviderRelayService) *ProxyHandler {
	return &ProxyHandler{svc: svc}
}

// proxyStatusResponse represents the JSON response for GET /proxy/status.
type proxyStatusResponse struct {
	Running bool   `json:"running"`
	Addr    string `json:"addr"`
}

// GetStatus handles GET /proxy/status.
// Returns the current running state and listen address of the proxy service.
func (h *ProxyHandler) GetStatus(c *gin.Context) {
	c.JSON(http.StatusOK, proxyStatusResponse{
		Running: h.svc.IsRunning(),
		Addr:    h.svc.Addr(),
	})
}

// Start handles POST /proxy/start.
// Starts the proxy relay service.
func (h *ProxyHandler) Start(c *gin.Context) {
	if h.svc.IsRunning() {
		c.JSON(http.StatusOK, gin.H{
			"message": "proxy is already running",
			"addr":    h.svc.Addr(),
		})
		return
	}

	if err := h.svc.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "proxy started",
		"addr":    h.svc.Addr(),
	})
}

// Stop handles POST /proxy/stop.
// Stops the proxy relay service.
func (h *ProxyHandler) Stop(c *gin.Context) {
	if !h.svc.IsRunning() {
		c.JSON(http.StatusOK, gin.H{"message": "proxy is not running"})
		return
	}

	if err := h.svc.Stop(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "proxy stopped"})
}

// GetLastUsed handles GET /proxy/last-used.
// Returns the last used provider for each platform.
func (h *ProxyHandler) GetLastUsed(c *gin.Context) {
	lastUsed := h.svc.GetAllLastUsedProviders()
	c.JSON(http.StatusOK, lastUsed)
}
