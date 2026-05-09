package handlers

import (
	"net/http"

	"codeswitch/services"

	"github.com/gin-gonic/gin"
)

// SpeedTestHandler handles speed test endpoints.
type SpeedTestHandler struct {
	svc *services.SpeedTestService
}

// NewSpeedTestHandler creates a new SpeedTestHandler.
func NewSpeedTestHandler(svc *services.SpeedTestService) *SpeedTestHandler {
	return &SpeedTestHandler{svc: svc}
}

// runSpeedTestRequest represents the request body for running a speed test.
type runSpeedTestRequest struct {
	URLs       []string `json:"urls" binding:"required"`
	TimeoutSec *int     `json:"timeout_sec"`
}

// Run handles POST /speedtest/run — executes endpoint latency tests.
func (h *SpeedTestHandler) Run(c *gin.Context) {
	var req runSpeedTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "urls array is required"})
		return
	}

	if len(req.URLs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "urls array cannot be empty"})
		return
	}

	results := h.svc.TestEndpoints(req.URLs, req.TimeoutSec)
	c.JSON(http.StatusOK, results)
}
