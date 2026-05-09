package ws

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// upgrader for test server
var testUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func TestNewHub(t *testing.T) {
	hub := NewHub()
	if hub == nil {
		t.Fatal("NewHub() returned nil")
	}
	if hub.clients == nil {
		t.Error("clients map not initialized")
	}
	if hub.register == nil {
		t.Error("register channel not initialized")
	}
	if hub.unregister == nil {
		t.Error("unregister channel not initialized")
	}
	if hub.broadcast == nil {
		t.Error("broadcast channel not initialized")
	}
}

func TestWSMessage_JSON(t *testing.T) {
	msg := WSMessage{
		Type:    WSEventProxyStatus,
		Payload: map[string]string{"status": "running"},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal WSMessage: %v", err)
	}

	var decoded WSMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal WSMessage: %v", err)
	}

	if decoded.Type != WSEventProxyStatus {
		t.Errorf("expected type %q, got %q", WSEventProxyStatus, decoded.Type)
	}
}

func TestClient_Subscribe(t *testing.T) {
	hub := NewHub()
	client := hub.NewClient(nil)

	// Initially no subscriptions — should receive everything
	if !client.isSubscribed(WSEventProxyStatus) {
		t.Error("client with no subscriptions should receive all events")
	}

	// Subscribe to specific events
	client.Subscribe(WSEventProxyStatus, WSEventRequestLog)

	if !client.isSubscribed(WSEventProxyStatus) {
		t.Error("client should be subscribed to proxy:status")
	}
	if !client.isSubscribed(WSEventRequestLog) {
		t.Error("client should be subscribed to request:log")
	}
	if client.isSubscribed(WSEventHealthCheck) {
		t.Error("client should NOT be subscribed to health:result")
	}
}

func TestClient_Unsubscribe(t *testing.T) {
	hub := NewHub()
	client := hub.NewClient(nil)

	client.Subscribe(WSEventProxyStatus, WSEventRequestLog)
	client.Unsubscribe(WSEventProxyStatus)

	if client.isSubscribed(WSEventProxyStatus) {
		t.Error("client should no longer be subscribed to proxy:status")
	}
	if !client.isSubscribed(WSEventRequestLog) {
		t.Error("client should still be subscribed to request:log")
	}
}

func TestHub_BroadcastEvent(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create a test WebSocket server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade failed: %v", err)
		}
		hub.RegisterClient(conn)
	}))
	defer server.Close()

	// Connect a client
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	// Give time for registration
	time.Sleep(50 * time.Millisecond)

	// Broadcast an event
	hub.BroadcastEvent(WSEventProxyStatus, map[string]string{"status": "running"})

	// Read the message from the client side
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read message failed: %v", err)
	}

	var msg WSMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if msg.Type != WSEventProxyStatus {
		t.Errorf("expected type %q, got %q", WSEventProxyStatus, msg.Type)
	}
}

func TestHub_SubscriptionFiltering(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create a test WebSocket server
	var registeredClient *Client
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade failed: %v", err)
		}
		registeredClient = hub.RegisterClient(conn)
	}))
	defer server.Close()

	// Connect a client
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	// Give time for registration
	time.Sleep(50 * time.Millisecond)

	// Subscribe only to health:result
	registeredClient.Subscribe(WSEventHealthCheck)

	// Broadcast a proxy:status event — client should NOT receive it
	hub.BroadcastEvent(WSEventProxyStatus, map[string]string{"status": "running"})

	// Broadcast a health:result event — client SHOULD receive it
	hub.BroadcastEvent(WSEventHealthCheck, map[string]bool{"healthy": true})

	// Read the message — should be the health check one
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read message failed: %v", err)
	}

	var msg WSMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if msg.Type != WSEventHealthCheck {
		t.Errorf("expected type %q, got %q", WSEventHealthCheck, msg.Type)
	}
}

func TestEventConstants(t *testing.T) {
	// Verify event constants are defined correctly
	if WSEventProxyStatus != "proxy:status" {
		t.Errorf("WSEventProxyStatus = %q, want %q", WSEventProxyStatus, "proxy:status")
	}
	if WSEventRequestLog != "request:log" {
		t.Errorf("WSEventRequestLog = %q, want %q", WSEventRequestLog, "request:log")
	}
	if WSEventHealthCheck != "health:result" {
		t.Errorf("WSEventHealthCheck = %q, want %q", WSEventHealthCheck, "health:result")
	}
	if WSEventAssistantReply != "assistant:reply" {
		t.Errorf("WSEventAssistantReply = %q, want %q", WSEventAssistantReply, "assistant:reply")
	}
}
