package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"codeswitch/internal/ws"
	"codeswitch/services"

	"github.com/daodao97/xgo/xdb"
	"github.com/gin-gonic/gin"
)

// AssistantHandler handles AI Assistant HTTP endpoints.
type AssistantHandler struct {
	relaySvc    *services.ProviderRelayService
	providerSvc *services.ProviderService
	settingsSvc *services.SettingsService
	appSettingsSvc *services.AppSettingsService
	logSvc      *services.LogService
	hub         *ws.Hub
	proxyAddr   string
}

// Conversation represents a single message in the conversation history.
type Conversation struct {
	ID         int64  `json:"id"`
	UserID     int64  `json:"user_id"`
	Role       string `json:"role"` // "user" or "assistant"
	Content    string `json:"content"`
	Model      string `json:"model,omitempty"`
	TokensUsed int    `json:"tokens_used,omitempty"`
	CreatedAt  string `json:"created_at"`
}

// ChatRequest represents the request body for POST /assistant/chat.
type ChatRequest struct {
	Message string `json:"message" binding:"required"`
	Model   string `json:"model"`
}

// ChatResponse represents the response for POST /assistant/chat.
type ChatResponse struct {
	Reply      string `json:"reply"`
	Model      string `json:"model"`
	TokensUsed int    `json:"tokens_used"`
}

// NewAssistantHandler creates a new AssistantHandler.
func NewAssistantHandler(relaySvc *services.ProviderRelayService, providerSvc *services.ProviderService, settingsSvc *services.SettingsService, appSettingsSvc *services.AppSettingsService, logSvc *services.LogService, hub *ws.Hub, proxyAddr string) *AssistantHandler {
	return &AssistantHandler{
		relaySvc:       relaySvc,
		providerSvc:    providerSvc,
		settingsSvc:    settingsSvc,
		appSettingsSvc: appSettingsSvc,
		logSvc:         logSvc,
		hub:            hub,
		proxyAddr:      proxyAddr,
	}
}

// =============================================================================
// Function Calling Tools Definition
// =============================================================================

// Tool represents an available function calling tool.
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  ToolParameters `json:"parameters"`
}

// ToolParameters defines the parameters for a tool.
type ToolParameters struct {
	Type       string                  `json:"type"`
	Properties map[string]ToolProperty `json:"properties"`
	Required   []string                `json:"required"`
}

// ToolProperty defines a single parameter property.
type ToolProperty struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

// GetTools returns the list of available tools for the AI assistant.
func (h *AssistantHandler) GetTools() []Tool {
	return []Tool{
		{
			Name:        "list_providers",
			Description: "列出指定平台的所有供应商配置。返回供应商名称、API URL、启用状态等信息。",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"kind": {
						Type:        "string",
						Description: "平台类型: claude, codex, gemini",
						Enum:        []string{"claude", "codex", "gemini"},
					},
				},
				Required: []string{"kind"},
			},
		},
		{
			Name:        "add_provider",
			Description: "添加新的供应商配置。敏感操作，需要用户确认。",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"kind": {
						Type:        "string",
						Description: "平台类型: claude, codex, gemini",
						Enum:        []string{"claude", "codex", "gemini"},
					},
					"name": {
						Type:        "string",
						Description: "供应商名称",
					},
					"apiUrl": {
						Type:        "string",
						Description: "API 端点 URL",
					},
					"apiKey": {
						Type:        "string",
						Description: "API 密钥（敏感信息）",
					},
					"level": {
						Type:        "number",
						Description: "优先级（1-10，数字越小优先级越高）",
					},
				},
				Required: []string{"kind", "name", "apiUrl", "apiKey"},
			},
		},
		{
			Name:        "toggle_provider",
			Description: "启用或禁用指定的供应商",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"kind": {
						Type:        "string",
						Description: "平台类型: claude, codex, gemini",
						Enum:        []string{"claude", "codex", "gemini"},
					},
					"id": {
						Type:        "number",
						Description: "供应商 ID",
					},
					"enabled": {
						Type:        "boolean",
						Description: "是否启用",
					},
				},
				Required: []string{"kind", "id", "enabled"},
			},
		},
		{
			Name:        "get_proxy_status",
			Description: "获取代理服务的运行状态和监听地址",
			Parameters: ToolParameters{
				Type:       "object",
				Properties: map[string]ToolProperty{},
				Required:   []string{},
			},
		},
		{
			Name:        "start_proxy",
			Description: "启动代理服务",
			Parameters: ToolParameters{
				Type:       "object",
				Properties: map[string]ToolProperty{},
				Required:   []string{},
			},
		},
		{
			Name:        "stop_proxy",
			Description: "停止代理服务",
			Parameters: ToolParameters{
				Type:       "object",
				Properties: map[string]ToolProperty{},
				Required:   []string{},
			},
		},
		{
			Name:        "get_stats",
			Description: "获取用量统计信息，包括请求数、Token 使用量、成本等",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"days": {
						Type:        "number",
						Description: "统计天数（默认7天）",
					},
				},
				Required: []string{},
			},
		},
		{
			Name:        "get_settings",
			Description: "获取系统设置信息，包括代理端口、日志级别等",
			Parameters: ToolParameters{
				Type:       "object",
				Properties: map[string]ToolProperty{},
				Required:   []string{},
			},
		},
		{
			Name:        "update_settings",
			Description: "更新系统设置。某些设置项为敏感操作，需要用户确认。",
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]ToolProperty{
					"key": {
						Type:        "string",
						Description: "设置项键名: proxyPort, proxyListenAddr, logLevel, databasePath, port",
						Enum:        []string{"proxyPort", "proxyListenAddr", "logLevel", "databasePath", "port"},
					},
					"value": {
						Type:        "any",
						Description: "设置值",
					},
				},
				Required: []string{"key", "value"},
			},
		},
	}
}

// ToolCall represents a function call from the AI model.
type ToolCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ToolResult represents the result of a tool execution.
type ToolResult struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PendingConfirmation stores pending sensitive operations.
type PendingConfirmation struct {
	ToolCall  ToolCall `json:"tool_call"`
	UserID    int64    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

// pendingConfirmations stores pending confirmations in memory.
// In production, this should be stored in a database or Redis.
var pendingConfirmations = make(map[string]PendingConfirmation)

// SensitiveTools defines which tools require user confirmation.
var sensitiveTools = map[string]bool{
	"add_provider":    true,
	"update_settings": true,
}

// Chat handles POST /assistant/chat.
// Receives user message, forwards to AI model via Proxy, streams response via WebSocket.
func (h *AssistantHandler) Chat(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	userIDInt := userID.(int64)

	// Parse request
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: message is required"})
		return
	}

	if req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message cannot be empty"})
		return
	}

	// Default model if not specified
	model := req.Model
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	// Save user message to database
	if err := h.saveMessage(userIDInt, "user", req.Message, model, 0); err != nil {
		fmt.Printf("[assistant] save user message error: %v\n", err)
	}

	// Build the request to forward to the proxy
	// Use Anthropic-compatible format
	requestBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": req.Message,
			},
		},
		"stream": true,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build request"})
		return
	}

	// Forward request to proxy
	proxyURL := fmt.Sprintf("http://%s/v1/messages", h.proxyAddr)
	proxyReq, err := http.NewRequest("POST", proxyURL, bytes.NewReader(bodyBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create proxy request"})
		return
	}
	proxyReq.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{
		Timeout: 120 * time.Second,
	}
	resp, err := client.Do(proxyReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("failed to connect to AI service: %v", err)})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("AI service error: %s", string(body))})
		return
	}

	// Process streaming response (SSE format)
	assistantReply := ""
	tokensUsed := 0

	// Read SSE stream line by line
	scanner := &sseScanner{reader: resp.Body}
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and "data: [DONE]" messages
		if line == "" || line == "[DONE]" {
			continue
		}

		// Remove "data: " prefix if present
		if len(line) > 6 && line[:6] == "data: " {
			line = line[6:]
		}

		// Parse JSON from SSE data
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			continue
		}

		// Extract content from the response (Anthropic SSE format)
		if delta, ok := data["delta"]; ok {
			deltaMap, ok := delta.(map[string]interface{})
			if !ok {
				continue
			}
			if text, ok := deltaMap["text"]; ok {
				textStr, ok := text.(string)
				if !ok {
					continue
				}
				assistantReply += textStr
				// Push streaming response to WebSocket
				h.pushAssistantReply(userIDInt, textStr, model, false)
			}
		}

		// Check for usage data
		if usage, ok := data["usage"]; ok {
			usageMap, ok := usage.(map[string]interface{})
			if !ok {
				continue
			}
			if outputTokens, ok := usageMap["output_tokens"].(float64); ok {
				tokensUsed = int(outputTokens)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("[assistant] stream read error: %v\n", err)
	}

	// Push final message to WebSocket
	h.pushAssistantReply(userIDInt, assistantReply, model, true)

	// Save assistant message to database
	if err := h.saveMessage(userIDInt, "assistant", assistantReply, model, tokensUsed); err != nil {
		fmt.Printf("[assistant] save assistant message error: %v\n", err)
	}

	// Return final response
	c.JSON(http.StatusOK, ChatResponse{
		Reply:      assistantReply,
		Model:      model,
		TokensUsed: tokensUsed,
	})
}

// sseScanner reads SSE (Server-Sent Events) lines from a reader.
type sseScanner struct {
	reader io.Reader
	buf    []byte
}

func (s *sseScanner) Scan() bool {
	// Read until we get a line (newline)
	s.buf = s.buf[:0]
	for {
		buf := make([]byte, 1)
		n, err := s.reader.Read(buf)
		if err != nil {
			return len(s.buf) > 0
		}
		if n == 0 {
			return len(s.buf) > 0
		}

		if buf[0] == '\n' {
			// Remove carriage return if present
			if len(s.buf) > 0 && s.buf[len(s.buf)-1] == '\r' {
				s.buf = s.buf[:len(s.buf)-1]
			}
			return true
		}
		if buf[0] != '\r' {
			s.buf = append(s.buf, buf[0])
		}
	}
}

func (s *sseScanner) Text() string {
	return string(s.buf)
}

func (s *sseScanner) Err() error {
	return nil
}

// pushAssistantReply pushes an assistant reply event to the WebSocket.
func (h *AssistantHandler) pushAssistantReply(userID int64, content string, model string, isFinal bool) {
	if h.hub == nil {
		return
	}

	h.hub.BroadcastEvent(ws.WSEventAssistantReply, map[string]interface{}{
		"user_id":  userID,
		"content":  content,
		"model":    model,
		"is_final": isFinal,
	})
}

// History handles GET /assistant/history.
// Retrieves conversation history for the current user.
func (h *AssistantHandler) History(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	userIDInt := userID.(int64)

	// Query conversation history
	history, err := h.getHistory(userIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get history: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
	})
}

// ClearHistory handles DELETE /assistant/history.
// Clears all conversation history for the current user.
func (h *AssistantHandler) ClearHistory(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	userIDInt := userID.(int64)

	// Clear conversation history
	if err := h.clearHistory(userIDInt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to clear history: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "conversation history cleared",
	})
}

// saveMessage saves a conversation message to the database.
func (h *AssistantHandler) saveMessage(userID int64, role, content, model string, tokensUsed int) error {
	db, err := xdb.DB("default")
	if err != nil {
		return fmt.Errorf("get database: %w", err)
	}

	_, err = db.Exec(`
		INSERT INTO assistant_conversations (user_id, role, content, model, tokens_used, created_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'))
	`, userID, role, content, model, tokensUsed)

	if err != nil {
		return fmt.Errorf("insert conversation: %w", err)
	}

	return nil
}

// getHistory retrieves conversation history for a user.
func (h *AssistantHandler) getHistory(userID int64) ([]Conversation, error) {
	db, err := xdb.DB("default")
	if err != nil {
		return nil, fmt.Errorf("get database: %w", err)
	}

	rows, err := db.Query(`
		SELECT id, user_id, role, content, model, tokens_used, created_at
		FROM assistant_conversations
		WHERE user_id = ?
		ORDER BY created_at ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}
	defer rows.Close()

	var history []Conversation
	for rows.Next() {
		var conv Conversation
		if err := rows.Scan(&conv.ID, &conv.UserID, &conv.Role, &conv.Content, &conv.Model, &conv.TokensUsed, &conv.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		history = append(history, conv)
	}

	if history == nil {
		history = []Conversation{}
	}

	return history, nil
}

// clearHistory clears all conversation history for a user.
func (h *AssistantHandler) clearHistory(userID int64) error {
	db, err := xdb.DB("default")
	if err != nil {
		return fmt.Errorf("get database: %w", err)
	}

	_, err = db.Exec(`DELETE FROM assistant_conversations WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("delete history: %w", err)
	}

	return nil
}
// =============================================================================
// Tool Execution Engine
// =============================================================================

// executeTool executes a tool call and returns the result.
func (h *AssistantHandler) executeTool(toolCall ToolCall) ToolResult {
	switch toolCall.Name {
	case "list_providers":
		return h.executeListProviders(toolCall.Arguments)
	case "add_provider":
		return h.executeAddProvider(toolCall.Arguments)
	case "toggle_provider":
		return h.executeToggleProvider(toolCall.Arguments)
	case "get_proxy_status":
		return h.executeGetProxyStatus()
	case "start_proxy":
		return h.executeStartProxy()
	case "stop_proxy":
		return h.executeStopProxy()
	case "get_stats":
		return h.executeGetStats(toolCall.Arguments)
	case "get_settings":
		return h.executeGetSettings()
	case "update_settings":
		return h.executeUpdateSettings(toolCall.Arguments)
	default:
		return ToolResult{
			Success: false,
			Error:   fmt.Sprintf("unknown tool: %s", toolCall.Name),
		}
	}
}

// executeListProviders lists all providers for a given platform.
func (h *AssistantHandler) executeListProviders(args json.RawMessage) ToolResult {
	var params struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("invalid arguments: %v", err)}
	}

	if params.Kind == "" {
		return ToolResult{Success: false, Error: "kind is required"}
	}

	providers, err := h.providerSvc.LoadProviders(params.Kind)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to load providers: %v", err)}
	}

	// Mask API keys in response
	for i := range providers {
		if len(providers[i].APIKey) > 4 {
			providers[i].APIKey = "****" + providers[i].APIKey[len(providers[i].APIKey)-4:]
		} else {
			providers[i].APIKey = "****"
		}
	}

	return ToolResult{Success: true, Result: providers}
}

// executeAddProvider adds a new provider (sensitive operation).
func (h *AssistantHandler) executeAddProvider(args json.RawMessage) ToolResult {
	var params struct {
		Kind   string `json:"kind"`
		Name   string `json:"name"`
		APIURL string `json:"apiUrl"`
		APIKey string `json:"apiKey"`
		Level  int    `json:"level"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("invalid arguments: %v", err)}
	}

	if params.Kind == "" || params.Name == "" || params.APIURL == "" || params.APIKey == "" {
		return ToolResult{Success: false, Error: "kind, name, apiUrl, and apiKey are required"}
	}

	// Load existing providers
	providers, err := h.providerSvc.LoadProviders(params.Kind)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to load providers: %v", err)}
	}

	// Find max ID
	maxID := int64(0)
	for _, p := range providers {
		if p.ID > maxID {
			maxID = p.ID
		}
	}

	// Create new provider
	newProvider := services.Provider{
		ID:     maxID + 1,
		Name:   params.Name,
		APIURL: params.APIURL,
		APIKey: params.APIKey,
		Level:  params.Level,
	}
	if newProvider.Level == 0 {
		newProvider.Level = 1
	}

	providers = append(providers, newProvider)

	// Save providers
	if err := h.providerSvc.SaveProviders(params.Kind, providers); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to save provider: %v", err)}
	}

	return ToolResult{
		Success: true,
		Result:  gin.H{"message": "provider added successfully", "id": newProvider.ID},
	}
}

// executeToggleProvider enables or disables a provider.
func (h *AssistantHandler) executeToggleProvider(args json.RawMessage) ToolResult {
	var params struct {
		Kind    string `json:"kind"`
		ID      int64  `json:"id"`
		Enabled bool   `json:"enabled"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("invalid arguments: %v", err)}
	}

	if params.Kind == "" || params.ID == 0 {
		return ToolResult{Success: false, Error: "kind and id are required"}
	}

	// Load providers
	providers, err := h.providerSvc.LoadProviders(params.Kind)
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to load providers: %v", err)}
	}

	// Find and update provider
	found := false
	for i, p := range providers {
		if p.ID == params.ID {
			providers[i].Enabled = params.Enabled
			found = true
			break
		}
	}

	if !found {
		return ToolResult{Success: false, Error: "provider not found"}
	}

	// Save
	if err := h.providerSvc.SaveProviders(params.Kind, providers); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to save providers: %v", err)}
	}

	return ToolResult{
		Success: true,
		Result:  gin.H{"message": fmt.Sprintf("provider %s", map[bool]string{true: "enabled", false: "disabled"}[params.Enabled])},
	}
}

// executeGetProxyStatus returns the proxy service status.
func (h *AssistantHandler) executeGetProxyStatus() ToolResult {
	if h.relaySvc == nil {
		return ToolResult{Success: false, Error: "relay service not available"}
	}

	return ToolResult{
		Success: true,
		Result: gin.H{
			"running": h.relaySvc.IsRunning(),
			"addr":    h.relaySvc.Addr(),
		},
	}
}

// executeStartProxy starts the proxy service.
func (h *AssistantHandler) executeStartProxy() ToolResult {
	if h.relaySvc == nil {
		return ToolResult{Success: false, Error: "relay service not available"}
	}

	if h.relaySvc.IsRunning() {
		return ToolResult{
			Success: true,
			Result:  gin.H{"message": "proxy is already running", "addr": h.relaySvc.Addr()},
		}
	}

	if err := h.relaySvc.Start(); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to start proxy: %v", err)}
	}

	return ToolResult{
		Success: true,
		Result:  gin.H{"message": "proxy started", "addr": h.relaySvc.Addr()},
	}
}

// executeStopProxy stops the proxy service.
func (h *AssistantHandler) executeStopProxy() ToolResult {
	if h.relaySvc == nil {
		return ToolResult{Success: false, Error: "relay service not available"}
	}

	if !h.relaySvc.IsRunning() {
		return ToolResult{Success: true, Result: gin.H{"message": "proxy is not running"}}
	}

	if err := h.relaySvc.Stop(); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to stop proxy: %v", err)}
	}

	return ToolResult{Success: true, Result: gin.H{"message": "proxy stopped"}}
}

// executeGetStats returns usage statistics.
func (h *AssistantHandler) executeGetStats(args json.RawMessage) ToolResult {
	var params struct {
		Days int `json:"days"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		params.Days = 7 // default
	}
	if params.Days == 0 {
		params.Days = 7
	}

	if h.logSvc == nil {
		return ToolResult{Success: false, Error: "log service not available"}
	}

	stats, err := h.logSvc.StatsSince("")
	if err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("failed to get stats: %v", err)}
	}

	return ToolResult{Success: true, Result: stats}
}

// executeGetSettings returns system settings.
func (h *AssistantHandler) executeGetSettings() ToolResult {
	// This is a simple implementation - in production, you'd get from config
	return ToolResult{
		Success: true,
		Result: gin.H{
			"proxyPort":        18100,
			"proxyListenAddr":  "0.0.0.0",
			"logLevel":         "info",
			"databasePath":     "~/.code-switch/app.db",
			"port":             8080,
		},
	}
}

// executeUpdateSettings updates system settings (sensitive operation).
func (h *AssistantHandler) executeUpdateSettings(args json.RawMessage) ToolResult {
	var params struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return ToolResult{Success: false, Error: fmt.Sprintf("invalid arguments: %v", err)}
	}

	if params.Key == "" {
		return ToolResult{Success: false, Error: "key is required"}
	}

	// For now, we just validate the key and return success
	// In production, this would update the actual config
	validKeys := map[string]bool{
		"proxyPort":        true,
		"proxyListenAddr": true,
		"logLevel":        true,
		"databasePath":    true,
		"port":            true,
	}

	if !validKeys[params.Key] {
		return ToolResult{Success: false, Error: fmt.Sprintf("invalid key: %s", params.Key)}
	}

	return ToolResult{
		Success: true,
		Result:  gin.H{"message": fmt.Sprintf("setting %s updated", params.Key)},
	}
}

// =============================================================================
// Sensitive Operation Confirmation
// =============================================================================

// isSensitiveTool checks if a tool requires user confirmation.
func isSensitiveTool(toolName string) bool {
	return sensitiveTools[toolName]
}

// generateConfirmationID generates a unique ID for a pending confirmation.
func generateConfirmationID(userID int64, toolCall ToolCall) string {
	return fmt.Sprintf("%d-%s-%d", userID, toolCall.Name, time.Now().UnixNano())
}

// GetPendingConfirmation retrieves a pending confirmation by ID.
func (h *AssistantHandler) GetPendingConfirmation(confirmationID string) (PendingConfirmation, bool) {
	conf, ok := pendingConfirmations[confirmationID]
	return conf, ok
}

// ClearPendingConfirmation removes a pending confirmation.
func (h *AssistantHandler) ClearPendingConfirmation(confirmationID string) {
	delete(pendingConfirmations, confirmationID)
}

// Execute handles POST /assistant/execute.
// Executes a confirmed operation after user confirmation.
func (h *AssistantHandler) Execute(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	userIDInt := userID.(int64)

	// Parse request
	var req struct {
		ConfirmationID string `json:"confirmationId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "confirmationId is required"})
		return
	}

	// Find pending confirmation
	conf, ok := pendingConfirmations[req.ConfirmationID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "confirmation not found or expired"})
		return
	}

	// Verify it's for this user
	if conf.UserID != userIDInt {
		c.JSON(http.StatusForbidden, gin.H{"error": "confirmation does not match current user"})
		return
	}

	// Check if confirmation has expired (10 minutes)
	if time.Since(conf.Timestamp) > 10*time.Minute {
		delete(pendingConfirmations, req.ConfirmationID)
		c.JSON(http.StatusGone, gin.H{"error": "confirmation expired"})
		return
	}

	// Execute the tool
	result := h.executeTool(conf.ToolCall)

	// Clear the confirmation
	delete(pendingConfirmations, req.ConfirmationID)

	// Return result
	if result.Success {
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusBadRequest, result)
	}
}