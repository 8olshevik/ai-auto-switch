package handlers

import (
	"net/http"
	"strconv"

	"codeswitch/services"

	"github.com/daodao97/xgo/xdb"
	"github.com/gin-gonic/gin"
)

// LogsHandler handles log-related HTTP endpoints.
type LogsHandler struct {
	svc *services.LogService
}

// NewLogsHandler creates a new LogsHandler with the given LogService.
func NewLogsHandler(svc *services.LogService) *LogsHandler {
	return &LogsHandler{svc: svc}
}

// ListLogs handles GET /logs/.
// Supports query parameters: page, pageSize, keyword, startDate, endDate, platform, provider.
func (h *LogsHandler) ListLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))
	platform := c.Query("platform")
	provider := c.Query("provider")
	keyword := c.Query("keyword")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 1000 {
		pageSize = 1000
	}

	// Use a larger limit to support pagination with filtering
	limit := page * pageSize
	if limit > 1000 {
		limit = 1000
	}

	logs, err := h.svc.ListRequestLogs(platform, provider, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch logs"})
		return
	}

	// Apply keyword filter if provided
	if keyword != "" {
		filtered := make([]services.ReqeustLog, 0)
		for _, log := range logs {
			if containsKeyword(log, keyword) {
				filtered = append(filtered, log)
			}
		}
		logs = filtered
	}

	// Apply time range filter if provided
	if startDate != "" || endDate != "" {
		filtered := make([]services.ReqeustLog, 0)
		for _, log := range logs {
			if matchesTimeRange(log.CreatedAt, startDate, endDate) {
				filtered = append(filtered, log)
			}
		}
		logs = filtered
	}

	// Calculate pagination
	total := len(logs)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     logs[start:end],
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetStats handles GET /logs/stats.
// Returns log statistics: total requests, success rate, average response time, token usage, cost.
func (h *LogsHandler) GetStats(c *gin.Context) {
	platform := c.Query("platform")

	stats, err := h.svc.StatsSince(platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch stats"})
		return
	}

	providerStats, err := h.svc.ProviderDailyStats(platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch provider stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats":          stats,
		"providerStats":  providerStats,
	})
}

// GetHeatmap handles GET /logs/heatmap.
// Returns heatmap data for the specified number of days.
func (h *LogsHandler) GetHeatmap(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 {
		days = 30
	}

	heatmap, err := h.svc.HeatmapStats(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch heatmap data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": heatmap,
	})
}

// ClearLogs handles DELETE /logs/.
// Deletes all request log records.
func (h *LogsHandler) ClearLogs(c *gin.Context) {
	model := xdb.New("request_log")
	_, err := model.Delete()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logs cleared successfully"})
}

// containsKeyword checks if a log entry matches the given keyword.
func containsKeyword(log services.ReqeustLog, keyword string) bool {
	// Search across multiple fields
	fields := []string{
		log.Platform,
		log.Model,
		log.Provider,
		strconv.Itoa(log.HttpCode),
	}
	for _, field := range fields {
		if len(field) > 0 && contains(field, keyword) {
			return true
		}
	}
	return false
}

// matchesTimeRange checks if a timestamp falls within the given time range.
func matchesTimeRange(createdAt, startDate, endDate string) bool {
	if createdAt == "" {
		return false
	}
	if startDate != "" && createdAt < startDate {
		return false
	}
	if endDate != "" && createdAt > endDate {
		return false
	}
	return true
}

// contains performs a case-insensitive substring check.
func contains(s, substr string) bool {
	// Simple case-insensitive contains using lowercase comparison
	sLower := toLower(s)
	substrLower := toLower(substr)
	return len(substrLower) <= len(sLower) && indexSubstring(sLower, substrLower) >= 0
}

// toLower converts a string to lowercase (ASCII only for performance).
func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

// indexSubstring returns the index of substr in s, or -1 if not found.
func indexSubstring(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
