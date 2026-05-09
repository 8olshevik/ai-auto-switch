package ws

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket event type constants used across the application.
const (
	WSEventProxyStatus    = "proxy:status"
	WSEventRequestLog     = "request:log"
	WSEventHealthCheck    = "health:result"
	WSEventAssistantReply = "assistant:reply"
)

// Heartbeat and buffer configuration.
const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512

	// Send channel buffer size per client.
	sendBufferSize = 256
)

// WSMessage is the standard message format for WebSocket communication.
type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Client represents a single WebSocket connection.
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte

	// subscriptions is the set of event types this client is interested in.
	// An empty set means the client receives all events (no filtering).
	subscriptions map[string]struct{}
	mu            sync.RWMutex
}

// Subscribe adds event types to this client's subscription set.
// If no subscriptions are set, the client receives all broadcast messages.
func (c *Client) Subscribe(eventTypes ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.subscriptions == nil {
		c.subscriptions = make(map[string]struct{})
	}
	for _, et := range eventTypes {
		c.subscriptions[et] = struct{}{}
	}
}

// Unsubscribe removes event types from this client's subscription set.
func (c *Client) Unsubscribe(eventTypes ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, et := range eventTypes {
		delete(c.subscriptions, et)
	}
}

// isSubscribed checks whether the client should receive a message of the given type.
func (c *Client) isSubscribed(eventType string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	// No subscriptions means receive everything.
	if len(c.subscriptions) == 0 {
		return true
	}
	_, ok := c.subscriptions[eventType]
	return ok
}

// Hub manages all active WebSocket clients and handles broadcasting.
type Hub struct {
	// Registered clients.
	clients map[*Client]struct{}

	// Channel for registering new clients.
	register chan *Client

	// Channel for unregistering clients.
	unregister chan *Client

	// Channel for broadcasting messages to all (subscribed) clients.
	broadcast chan WSMessage

	mu sync.RWMutex
}

// NewHub creates a new Hub instance ready to be started with Run().
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]struct{}),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan WSMessage, 256),
	}
}

// Run starts the hub's main event loop. It should be launched as a goroutine.
// It processes client registration, unregistration, and message broadcasting.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = struct{}{}
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			data, err := json.Marshal(msg)
			if err != nil {
				log.Printf("[ws] failed to marshal broadcast message: %v", err)
				continue
			}

			h.mu.RLock()
			for client := range h.clients {
				if !client.isSubscribed(msg.Type) {
					continue
				}
				select {
				case client.send <- data:
				default:
					// Client send buffer is full — disconnect it.
					go h.removeClient(client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// removeClient unregisters a client from the hub.
func (h *Hub) removeClient(client *Client) {
	h.unregister <- client
}

// Broadcast sends a WSMessage to all connected clients that are subscribed
// to the message's event type.
func (h *Hub) Broadcast(msg WSMessage) {
	h.broadcast <- msg
}

// BroadcastEvent is a convenience method that constructs a WSMessage and broadcasts it.
func (h *Hub) BroadcastEvent(eventType string, payload interface{}) {
	h.Broadcast(WSMessage{
		Type:    eventType,
		Payload: payload,
	})
}

// NewClient creates a new Client associated with this hub and the given connection.
// After creation, the caller should start the client's read and write pumps.
func (h *Hub) NewClient(conn *websocket.Conn) *Client {
	return &Client{
		hub:           h,
		conn:          conn,
		send:          make(chan []byte, sendBufferSize),
		subscriptions: make(map[string]struct{}),
	}
}

// RegisterClient registers a client with the hub and starts its read/write pumps.
func (h *Hub) RegisterClient(conn *websocket.Conn) *Client {
	client := h.NewClient(conn)
	h.register <- client
	go client.writePump()
	go client.readPump()
	return client
}

// readPump pumps messages from the WebSocket connection to the hub.
// It handles ping/pong heartbeat and processes incoming subscribe/unsubscribe commands.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("[ws] unexpected close error: %v", err)
			}
			break
		}

		// Process client commands (subscribe/unsubscribe).
		c.handleMessage(message)
	}
}

// handleMessage processes incoming messages from the client.
// Supported commands:
//
//	{"type": "subscribe", "payload": ["proxy:status", "request:log"]}
//	{"type": "unsubscribe", "payload": ["proxy:status"]}
func (c *Client) handleMessage(message []byte) {
	var msg struct {
		Type    string   `json:"type"`
		Payload []string `json:"payload"`
	}
	if err := json.Unmarshal(message, &msg); err != nil {
		return
	}

	switch msg.Type {
	case "subscribe":
		c.Subscribe(msg.Payload...)
	case "unsubscribe":
		c.Unsubscribe(msg.Payload...)
	}
}

// writePump pumps messages from the hub to the WebSocket connection.
// It handles the ping/pong heartbeat mechanism to detect dead connections.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel — send a close frame.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Drain queued messages into the current write to reduce syscalls.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
