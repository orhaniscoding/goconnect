package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Hub maintains active clients and broadcasts messages to them
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Rooms map room name to clients in that room
	rooms map[string]map[*Client]bool

	// Inbound messages from clients
	handleInbound chan *InboundEvent

	// Outbound broadcasts
	broadcast chan *BroadcastMessage

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Message handler
	handler MessageHandler

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// BroadcastMessage represents a message to be broadcast to a room
type BroadcastMessage struct {
	Room    string           // Room to broadcast to (empty = all clients)
	Message *OutboundMessage // Message to send
	Exclude *Client          // Optional client to exclude from broadcast
}

// MessageHandler handles inbound WebSocket messages
type MessageHandler interface {
	HandleMessage(ctx context.Context, client *Client, msg *InboundMessage) error
}

// NewHub creates a new WebSocket hub
func NewHub(handler MessageHandler) *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		rooms:         make(map[string]map[*Client]bool),
		handleInbound: make(chan *InboundEvent, 256),
		broadcast:     make(chan *BroadcastMessage, 256),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		handler:       handler,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// Graceful shutdown
			h.mu.Lock()
			for client := range h.clients {
				close(client.send)
			}
			h.mu.Unlock()
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			fmt.Printf("Client registered: %s (total: %d)\n", client.userID, len(h.clients))
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				// Remove from all rooms
				for room := range client.rooms {
					if clients, ok := h.rooms[room]; ok {
						delete(clients, client)
						if len(clients) == 0 {
							delete(h.rooms, room)
						} else {
							// Broadcast offline status to remaining clients in the room
							// Use goroutine to avoid deadlock/blocking if broadcast channel is full
							go func(r string, uid string) {
								h.broadcast <- &BroadcastMessage{
									Room: r,
									Message: &OutboundMessage{
										Type: TypePresenceUpdate,
										Data: &PresenceUpdateData{
											UserID: uid,
											Status: "offline",
											Since:  time.Now().Format(time.RFC3339),
										},
									},
								}
							}(room, client.userID)
						}
					}
				}
				close(client.send)
				fmt.Printf("Client unregistered: %s (total: %d)\n", client.userID, len(h.clients))
			}
			h.mu.Unlock()

		case event := <-h.handleInbound:
			// Handle message in goroutine to avoid blocking
			go func(e *InboundEvent) {
				if err := h.handler.HandleMessage(ctx, e.client, e.message); err != nil {
					e.client.sendError(e.message.OpID, "ERR_HANDLER_FAILED", err.Error(), nil)
				}
			}(event)

		case msg := <-h.broadcast:
			h.mu.RLock()
			if msg.Room == "" {
				// Broadcast to all clients
				for client := range h.clients {
					if msg.Exclude != nil && client == msg.Exclude {
						continue
					}
					select {
					case client.send <- mustMarshal(msg.Message):
					default:
						// Client's send buffer is full, close connection
						go func(c *Client) {
							h.unregister <- c
						}(client)
					}
				}
			} else {
				// Broadcast to room
				if clients, ok := h.rooms[msg.Room]; ok {
					for client := range clients {
						if msg.Exclude != nil && client == msg.Exclude {
							continue
						}
						select {
						case client.send <- mustMarshal(msg.Message):
						default:
							go func(c *Client) {
								h.unregister <- c
							}(client)
						}
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Register registers a client
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister unregisters a client
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Broadcast sends a message to a room or all clients
func (h *Hub) Broadcast(room string, msg *OutboundMessage, exclude *Client) {
	h.broadcast <- &BroadcastMessage{
		Room:    room,
		Message: msg,
		Exclude: exclude,
	}
}

// JoinRoom adds a client to a room
func (h *Hub) JoinRoom(client *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.rooms[room]; !ok {
		h.rooms[room] = make(map[*Client]bool)
	}
	h.rooms[room][client] = true
	client.JoinRoom(room)
	fmt.Printf("Client %s joined room %s\n", client.userID, room)
}

// LeaveRoom removes a client from a room
func (h *Hub) LeaveRoom(client *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.rooms[room]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.rooms, room)
		}
	}
	client.LeaveRoom(room)
	fmt.Printf("Client %s left room %s\n", client.userID, room)
}

// GetRoomClients returns all clients in a room
func (h *Hub) GetRoomClients(room string) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := make([]*Client, 0)
	if roomClients, ok := h.rooms[room]; ok {
		for client := range roomClients {
			clients = append(clients, client)
		}
	}
	return clients
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetRoomCount returns the number of active rooms
func (h *Hub) GetRoomCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms)
}

// GetActiveConnectionCount returns the number of currently connected clients
func (h *Hub) GetActiveConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// mustMarshal marshals a message or panics (should never fail for valid types)
func mustMarshal(msg *OutboundMessage) []byte {
	data, err := json.Marshal(msg)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal message: %v", err))
	}
	return data
}
