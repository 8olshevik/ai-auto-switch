package services

// WSBroadcaster is an interface for broadcasting WebSocket events.
// It decouples services from the concrete ws.Hub implementation,
// allowing services to broadcast events without importing the ws package.
type WSBroadcaster interface {
	BroadcastEvent(eventType string, payload interface{})
}

// WebSocket event type constants (mirrored from internal/ws for use in services).
const (
	WSEventProxyStatus = "proxy:status"
	WSEventRequestLog  = "request:log"
	WSEventHealthCheck = "health:result"
)
