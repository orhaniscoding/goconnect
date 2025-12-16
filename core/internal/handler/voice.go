package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/logger"
	"github.com/redis/go-redis/v9"
)

// VoiceSignal represents a WebRTC signaling message
type VoiceSignal struct {
	Type      string          `json:"type"`                // "offer", "answer", "candidate"
	SDP       json.RawMessage `json:"sdp,omitempty"`       // For offer/answer
	Candidate json.RawMessage `json:"candidate,omitempty"` // For candidate
	TargetID  string          `json:"target_id"`           // Target user ID
	SenderID  string          `json:"sender_id,omitempty"` // Sender user ID
	NetworkID string          `json:"network_id"`          // Network scope
}

// VoiceHandler handles real-time voice signaling
type VoiceHandler struct {
	redis *redis.Client
}

// NewVoiceHandler creates a new voice handler
func NewVoiceHandler(redisClient *redis.Client) *VoiceHandler {
	return &VoiceHandler{
		redis: redisClient,
	}
}

// Signal handles POST /v1/voice/signal
func (h *VoiceHandler) Signal(c *gin.Context) {
	if h.redis == nil {
		errorResponse(c, domain.NewError(domain.ErrNotImplemented, "Voice chat requires Redis", nil))
		return
	}

	var sig VoiceSignal
	if err := c.ShouldBindJSON(&sig); err != nil {
		errorResponse(c, domain.NewError(domain.ErrValidation, "Invalid signal format", nil))
		return
	}

	senderID := c.GetString("user_id")
	if senderID == "" {
		errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Unauthorized", nil))
		return
	}
	sig.SenderID = senderID

	// Store signal in a list for the target user (Queue)
	// Key: "voice:queue:{target_id}"
	key := "voice:queue:" + sig.TargetID
	
	val, err := json.Marshal(sig)
	if err != nil {
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Failed to encode signal", nil))
		return
	}

	ctx := c.Request.Context()
	pipe := h.redis.Pipeline()
	pipe.RPush(ctx, key, val)
	pipe.Expire(ctx, key, 30*time.Second) // Signals are ephemeral
	// Optional: Limit list size
	pipe.LTrim(ctx, key, -50, -1) // Keep last 50
	
	// ... (Redis logic)
	
	if _, err := pipe.Exec(ctx); err != nil {
		logger.Error("Failed to send voice signal", 
			"sender_id", senderID, 
			"target_id", sig.TargetID, 
			"error", err,
		)
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Failed to send signal", nil))
		return
	}

	logger.Debug("Voice signal sent", 
		"type", sig.Type, 
		"sender_id", senderID, 
		"target_id", sig.TargetID,
	)

	c.Status(http.StatusAccepted)
}

// GetSignals handles GET /v1/voice/signals
func (h *VoiceHandler) GetSignals(c *gin.Context) {
	if h.redis == nil {
		errorResponse(c, domain.NewError(domain.ErrNotImplemented, "Voice chat requires Redis", nil))
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Unauthorized", nil))
		return
	}

	key := "voice:queue:" + userID
	
	// Atomic pop all elements (LRange + Del in transaction, or just LPop count)
	// Simple approach: Read all, delete key (acceptable race for voice signals? maybe lost signals if crash)
	// Better: RPop count? Redis doesn't support popping all atomically easily without Lua.
	// We will use a script to get simple atomicity: GET all + DISCARD
	
	script := redis.NewScript(`
		var msgs = redis.call("LRANGE", KEYS[1], 0, -1)
		redis.call("DEL", KEYS[1])
		return msgs
	`)

	result, err := script.Run(c.Request.Context(), h.redis, []string{key}).Result()
	if err != nil && err != redis.Nil {
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Failed to retrieve signals", nil))
		return
	}

	var signals []VoiceSignal = []VoiceSignal{}
	
	if items, ok := result.([]interface{}); ok {
		for _, item := range items {
			if s, ok := item.(string); ok {
				var sig VoiceSignal
				if json.Unmarshal([]byte(s), &sig) == nil {
					signals = append(signals, sig)
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": signals})
}
