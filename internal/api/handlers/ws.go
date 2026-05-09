package handlers

import (
	"net/http"

	"codeswitch/internal/api/middleware"
	"codeswitch/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// upgrader configures the WebSocket upgrade with a permissive CheckOrigin
// (CORS is already handled by the middleware layer).
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WSHandler handles WebSocket connections for real-time event streaming.
type WSHandler struct {
	hub       *ws.Hub
	jwtSecret string
}

// NewWSHandler creates a new WebSocket handler.
func NewWSHandler(hub *ws.Hub, jwtSecret string) *WSHandler {
	return &WSHandler{
		hub:       hub,
		jwtSecret: jwtSecret,
	}
}

// HandleConnect upgrades the HTTP connection to WebSocket after validating
// the JWT token from the query parameter `?token=<jwt>`.
// WebSocket connections cannot easily set Authorization headers, so the token
// is passed as a query parameter instead.
func (h *WSHandler) HandleConnect(c *gin.Context) {
	// Validate JWT from query parameter
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token query parameter is required"})
		return
	}

	claims, err := middleware.ParseToken(token, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// Upgrade already wrote the error response
		return
	}

	// Register client with the hub (starts read/write pumps)
	_ = claims // claims available for future per-user filtering
	h.hub.RegisterClient(conn)
}
