package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512 KB

	// Send buffer size
	sendBufferSize = 256

	// Rate limit: 10 messages per second, burst of 20
	rateLimit = 10
	rateBurst = 20
)

// Client represents a WebSocket client connection
type Client struct {
	hub          *Hub
	conn         *websocket.Conn
	send         chan []byte
	userID       string
	tenantID     string
	isAdmin      bool            // Global admin flag from JWT
	isModerator  bool            // Content moderator flag from JWT
	rooms        map[string]bool // rooms this client is subscribed to
	lastActivity time.Time       // Last activity timestamp for presence tracking
	limiter      *rate.Limiter   // Rate limiter
	mu           sync.RWMutex
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, userID, tenantID string, isAdmin, isModerator bool) *Client {
	return &Client{
		hub:          hub,
		conn:         conn,
		send:         make(chan []byte, sendBufferSize),
		userID:       userID,
		tenantID:     tenantID,
		isAdmin:      isAdmin,
		isModerator:  isModerator,
		rooms:        make(map[string]bool),
		lastActivity: time.Now(),
		limiter:      rate.NewLimiter(rate.Limit(rateLimit), rateBurst),
	}
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("WebSocket error: %v\n", err)
			}
			break
		}

		// Rate limit check
		if !c.limiter.Allow() {
			c.sendError("", "ERR_RATE_LIMIT", "Too many messages", nil)
			continue
		}

		// Parse inbound message
		var inbound InboundMessage
		if err := json.Unmarshal(message, &inbound); err != nil {
			c.sendError("", "ERR_INVALID_MESSAGE", "Failed to parse message", nil)
			continue
		}

		// Handle message
		c.hub.handleInbound <- &InboundEvent{
			client:  c,
			message: &inbound,
		}
	}
}

// writePump pumps messages from the hub to the websocket connection
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
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
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

// Send sends a message to the client
func (c *Client) Send(msg *OutboundMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	select {
	case c.send <- data:
		return nil
	case <-time.After(writeWait):
		return fmt.Errorf("send timeout")
	}
}

// sendError sends an error message to the client
func (c *Client) sendError(opID, code, message string, details map[string]string) {
	msg := &OutboundMessage{
		Type: TypeError,
		OpID: opID,
		Error: &ErrorData{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
	c.Send(msg)
}

// sendAck sends an acknowledgment to the client
func (c *Client) sendAck(opID string, data interface{}) {
	msg := &OutboundMessage{
		Type: TypeAck,
		OpID: opID,
		Data: data,
	}
	c.Send(msg)
}

// JoinRoom adds the client to a room
func (c *Client) JoinRoom(room string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rooms[room] = true
}

// LeaveRoom removes the client from a room
func (c *Client) LeaveRoom(room string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.rooms, room)
}

// IsInRoom checks if client is in a room
func (c *Client) IsInRoom(room string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.rooms[room]
}

// GetRooms returns all rooms the client is in
func (c *Client) GetRooms() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	rooms := make([]string, 0, len(c.rooms))
	for room := range c.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

// UpdateActivity updates the last activity timestamp
func (c *Client) UpdateActivity() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastActivity = time.Now()
}

// GetLastActivity returns the last activity timestamp
func (c *Client) GetLastActivity() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastActivity
}

// sendMessage sends a message to the client (internal use)
func (c *Client) sendMessage(msg *OutboundMessage) {
	data := mustMarshal(msg)
	select {
	case c.send <- data:
	default:
		// Send buffer full, connection will be closed by hub
	}
}

// InboundEvent represents an inbound message event
type InboundEvent struct {
	client  *Client
	message *InboundMessage
}

// Run starts the client's read and write pumps
func (c *Client) Run(ctx context.Context) {
	go c.writePump()
	go c.readPump()
}
