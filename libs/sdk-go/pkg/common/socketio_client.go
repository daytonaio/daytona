// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Engine.IO v4 packet types
const (
	eioOpen    = '0'
	eioClose   = '1'
	eioPing    = '2'
	eioPong    = '3'
	eioMessage = '4'
	eioUpgrade = '5'
	eioNoop    = '6'
)

// Socket.IO v4 packet types (inside Engine.IO messages)
const (
	sioConnect      = '0'
	sioDisconnect   = '1'
	sioEvent        = '2'
	sioAck          = '3'
	sioConnectError = '4'
)

// socketIOOpenPayload represents the Engine.IO OPEN packet payload.
type socketIOOpenPayload struct {
	SID          string   `json:"sid"`
	Upgrades     []string `json:"upgrades"`
	PingInterval int      `json:"pingInterval"`
	PingTimeout  int      `json:"pingTimeout"`
}

// socketIOConnectPayload is sent to authenticate with Socket.IO.
type socketIOConnectPayload struct {
	Token string `json:"token"`
}

// socketIOConnectError is the Socket.IO connect error response.
type socketIOConnectError struct {
	Message string `json:"message"`
}

// socketIOEventHandler handles a Socket.IO event.
type socketIOEventHandler func(eventName string, data json.RawMessage)

// socketIOClient is a minimal Engine.IO/Socket.IO v4 client for connecting
// to a Socket.IO server over WebSocket.
//
// Engine.IO v4 heartbeat protocol (WebSocket transport):
//   - Server sends PING (type 2) every pingInterval ms
//   - Client must respond with PONG (type 3) within pingTimeout ms
//   - Client monitors for missing server activity to detect dead connections
type socketIOClient struct {
	mu sync.RWMutex

	conn               *websocket.Conn
	connected          bool
	sid                string
	pingInterval       time.Duration
	pingTimeout        time.Duration
	lastServerActivity time.Time // tracks ANY incoming server data
	eventHandler       socketIOEventHandler
	onDisconnect       func()
	closeCh            chan struct{}
	healthStopCh       chan struct{}
	writemu            sync.Mutex // serializes WebSocket writes
}

// socketIOClientConfig holds configuration for creating a socketIOClient.
type socketIOClientConfig struct {
	// APIURL is the full API URL (e.g. "https://app.daytona.io/api")
	APIURL string
	// Token is the auth token (API key or JWT)
	Token string
	// OrganizationID is optional and passed as a query parameter
	OrganizationID string
	// ConnectTimeout is the maximum time to wait for the connection
	ConnectTimeout time.Duration
	// EventHandler is called for each Socket.IO event received
	EventHandler socketIOEventHandler
	// OnDisconnect is called when the connection is lost
	OnDisconnect func()
}

// newSocketIOClient creates and connects a new Socket.IO client.
func newSocketIOClient(cfg socketIOClientConfig) (*socketIOClient, error) {
	if cfg.ConnectTimeout == 0 {
		cfg.ConnectTimeout = 5 * time.Second
	}

	c := &socketIOClient{
		eventHandler: cfg.EventHandler,
		onDisconnect: cfg.OnDisconnect,
		closeCh:      make(chan struct{}),
		healthStopCh: make(chan struct{}),
	}

	wsURL, err := buildWebSocketURL(cfg.APIURL, cfg.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to build WebSocket URL: %w", err)
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: cfg.ConnectTimeout,
	}

	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("WebSocket dial failed: %w", err)
	}
	c.conn = conn

	// Read Engine.IO OPEN packet
	if err := c.readOpenPacket(cfg.ConnectTimeout); err != nil {
		conn.Close()
		return nil, err
	}

	// Send Socket.IO CONNECT with auth
	if err := c.sendConnect(cfg.Token); err != nil {
		conn.Close()
		return nil, err
	}

	// Read Socket.IO CONNECT response
	if err := c.readConnectResponse(cfg.ConnectTimeout); err != nil {
		conn.Close()
		return nil, err
	}

	c.mu.Lock()
	c.connected = true
	c.lastServerActivity = time.Now()
	c.mu.Unlock()

	// Start background goroutines
	go c.readLoop()
	go c.healthMonitorLoop()

	return c, nil
}

// buildWebSocketURL constructs the WebSocket URL for Socket.IO connection.
func buildWebSocketURL(apiURL string, organizationID string) (string, error) {
	parsed, err := url.Parse(apiURL)
	if err != nil {
		return "", err
	}

	// Determine WebSocket scheme
	wsScheme := "wss"
	if parsed.Scheme == "http" {
		wsScheme = "ws"
	}

	// Build the WebSocket URL
	query := url.Values{}
	query.Set("EIO", "4")
	query.Set("transport", "websocket")
	if organizationID != "" {
		query.Set("organizationId", organizationID)
	}

	wsURL := fmt.Sprintf("%s://%s/api/socket.io/?%s", wsScheme, parsed.Host, query.Encode())
	return wsURL, nil
}

// readOpenPacket reads and parses the Engine.IO OPEN packet.
func (c *socketIOClient) readOpenPacket(timeout time.Duration) error {
	_ = c.conn.SetReadDeadline(time.Now().Add(timeout))
	defer func() { _ = c.conn.SetReadDeadline(time.Time{}) }()

	_, msg, err := c.conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read Engine.IO OPEN: %w", err)
	}

	if len(msg) == 0 || msg[0] != byte(eioOpen) {
		return fmt.Errorf("expected Engine.IO OPEN packet, got: %s", string(msg))
	}

	var payload socketIOOpenPayload
	if err := json.Unmarshal(msg[1:], &payload); err != nil {
		return fmt.Errorf("failed to parse Engine.IO OPEN payload: %w", err)
	}

	c.sid = payload.SID
	c.pingInterval = time.Duration(payload.PingInterval) * time.Millisecond
	c.pingTimeout = time.Duration(payload.PingTimeout) * time.Millisecond

	if c.pingInterval == 0 {
		c.pingInterval = 25 * time.Second
	}
	if c.pingTimeout == 0 {
		c.pingTimeout = 20 * time.Second
	}

	return nil
}

// sendConnect sends the Socket.IO CONNECT packet with auth token.
func (c *socketIOClient) sendConnect(token string) error {
	authPayload := socketIOConnectPayload{Token: token}
	authJSON, err := json.Marshal(authPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal auth payload: %w", err)
	}

	// Socket.IO CONNECT for default namespace: "40" + JSON auth
	packet := fmt.Sprintf("%c%c%s", eioMessage, sioConnect, string(authJSON))
	return c.writeMessage(packet)
}

// readConnectResponse reads the Socket.IO CONNECT acknowledgement or error.
func (c *socketIOClient) readConnectResponse(timeout time.Duration) error {
	_ = c.conn.SetReadDeadline(time.Now().Add(timeout))
	defer func() { _ = c.conn.SetReadDeadline(time.Time{}) }()

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("failed to read Socket.IO CONNECT response: %w", err)
		}

		if len(msg) < 2 {
			continue
		}

		// Check for Engine.IO message containing Socket.IO packet
		if msg[0] == byte(eioMessage) {
			switch msg[1] {
			case byte(sioConnect):
				// Success: "40{\"sid\":\"...\"}"
				return nil
			case byte(sioConnectError):
				// Error: "44{\"message\":\"...\"}"
				var errPayload socketIOConnectError
				if parseErr := json.Unmarshal(msg[2:], &errPayload); parseErr == nil {
					return fmt.Errorf("Socket.IO connection rejected: %s", errPayload.Message)
				}
				return fmt.Errorf("Socket.IO connection rejected: %s", string(msg[2:]))
			}
		}

		// Handle Engine.IO ping during handshake
		if msg[0] == byte(eioPing) {
			_ = c.writeMessage(string([]byte{byte(eioPong)}))
		}
	}
}

// readLoop continuously reads messages from the WebSocket and dispatches them.
// A read deadline is set on each iteration to detect half-open TCP connections
// that stop delivering data without closing the socket.
func (c *socketIOClient) readLoop() {
	defer func() {
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()

		if c.onDisconnect != nil {
			c.onDisconnect()
		}
	}()

	for {
		select {
		case <-c.closeCh:
			return
		default:
		}

		// Set a read deadline so ReadMessage doesn't block forever on dead connections.
		// Allow pingInterval + pingTimeout + buffer for the server to send something.
		c.mu.RLock()
		deadline := c.pingInterval + c.pingTimeout + 5*time.Second
		c.mu.RUnlock()
		_ = c.conn.SetReadDeadline(time.Now().Add(deadline))

		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			// Check if this is an intentional close
			select {
			case <-c.closeCh:
				return
			default:
			}
			// Connection error (timeout, network drop, clean close, etc.)
			return
		}

		if len(msg) == 0 {
			continue
		}

		c.handleMessage(msg)
	}
}

// handleMessage processes an Engine.IO/Socket.IO message.
func (c *socketIOClient) handleMessage(msg []byte) {
	// Track all server activity for health monitoring.
	// If the server stops sending ANY data (pings, events, etc.)
	// the health monitor will detect the dead connection.
	c.mu.Lock()
	c.lastServerActivity = time.Now()
	c.mu.Unlock()

	switch msg[0] {
	case byte(eioPing):
		// Server heartbeat — respond immediately with PONG
		_ = c.writeMessage(string([]byte{byte(eioPong)}))

	case byte(eioPong):
		// Unexpected in EIO v4 (server doesn't respond to client pings),
		// but handle gracefully — activity already tracked above.

	case byte(eioClose):
		// Server-initiated close
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()

	case byte(eioMessage):
		// Socket.IO packet inside Engine.IO message
		if len(msg) < 2 {
			return
		}
		c.handleSocketIOPacket(msg[1:])

	case byte(eioNoop):
		// No-op, ignore

	default:
		// Unknown packet type, ignore
	}
}

// handleSocketIOPacket processes a Socket.IO packet.
func (c *socketIOClient) handleSocketIOPacket(data []byte) {
	if len(data) == 0 {
		return
	}

	switch data[0] {
	case byte(sioEvent):
		// Event: "2[\"event_name\",{...}]"
		c.handleEvent(data[1:])

	case byte(sioDisconnect):
		// Server-initiated disconnect
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()

	default:
		// Other packet types (ACK, etc.) - ignore
	}
}

// handleEvent parses and dispatches a Socket.IO EVENT packet.
func (c *socketIOClient) handleEvent(data []byte) {
	if c.eventHandler == nil {
		return
	}

	// Parse the namespace prefix if present (e.g., "/namespace,")
	// For default namespace, there's no prefix
	jsonData := data

	// Skip namespace prefix (e.g., "/ns," prefix)
	if len(jsonData) > 0 && jsonData[0] == '/' {
		commaIdx := strings.IndexByte(string(jsonData), ',')
		if commaIdx >= 0 {
			jsonData = jsonData[commaIdx+1:]
		}
	}

	// Parse JSON array: ["event_name", data]
	var eventArray []json.RawMessage
	if err := json.Unmarshal(jsonData, &eventArray); err != nil {
		return
	}

	if len(eventArray) < 1 {
		return
	}

	var eventName string
	if err := json.Unmarshal(eventArray[0], &eventName); err != nil {
		return
	}

	var eventData json.RawMessage
	if len(eventArray) > 1 {
		eventData = eventArray[1]
	}

	c.eventHandler(eventName, eventData)
}

// healthCheckInterval is the frequency of connection health checks.
const healthCheckInterval = 5 * time.Second

// healthMonitorLoop monitors connection health by checking for server activity.
// In Engine.IO v4, the server sends PING every pingInterval ms.
// If no server activity (pings, events, any data) is seen within
// pingInterval + pingTimeout, the connection is considered dead.
func (c *socketIOClient) healthMonitorLoop() {
	ticker := time.NewTicker(healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.closeCh:
			return
		case <-c.healthStopCh:
			return
		case <-ticker.C:
			c.mu.RLock()
			connected := c.connected
			lastActivity := c.lastServerActivity
			threshold := c.pingInterval + c.pingTimeout
			c.mu.RUnlock()

			if !connected {
				return
			}

			// No server activity within expected window — connection is dead
			if time.Since(lastActivity) > threshold {
				c.mu.Lock()
				c.connected = false
				c.mu.Unlock()
				if c.onDisconnect != nil {
					c.onDisconnect()
				}
				return
			}
		}
	}
}

// writeMessage sends a text message on the WebSocket, serializing writes.
func (c *socketIOClient) writeMessage(msg string) error {
	c.writemu.Lock()
	defer c.writemu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("connection is closed")
	}

	return c.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

// IsConnected returns whether the client is currently connected.
func (c *socketIOClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// Close gracefully closes the Socket.IO and WebSocket connection.
func (c *socketIOClient) Close() {
	c.mu.Lock()
	if !c.connected && c.conn == nil {
		c.mu.Unlock()
		return
	}
	c.connected = false
	c.mu.Unlock()

	// Signal goroutines to stop
	select {
	case <-c.closeCh:
	default:
		close(c.closeCh)
	}

	select {
	case <-c.healthStopCh:
	default:
		close(c.healthStopCh)
	}

	// Send Engine.IO CLOSE
	_ = c.writeMessage(string([]byte{byte(eioClose)}))

	// Close WebSocket with proper close frame
	if c.conn != nil {
		_ = c.conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		)
		c.conn.Close()
	}
}
