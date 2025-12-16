package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// Mock Redis client isn't trivial with go-redis without an interface or miniredis.
// For this unit test, we will focus on input validation and Authorization checks which don't hit Redis if they fail early.
// For full integration, we would need a running Redis container or miniredis.

func TestVoiceHandler_Signal_Validation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Non-nil redis client so the handler reaches JSON parsing/auth checks.
	// Address is intentionally unreachable; these tests should return before Redis is used.
	rc := redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})
	h := NewVoiceHandler(rc)
	router := gin.New()
	
	// Mock auth middleware behavior by setting context user_id
	router.POST("/signal", func(c *gin.Context) {
		c.Set("user_id", "user_123")
		h.Signal(c)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/signal", bytes.NewBufferString("invalid-json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		// Assuming errorResponse returns JSON with error field
	})

	t.Run("Missing Redis (Internal Error)", func(t *testing.T) {
		// Valid JSON but nil Redis
		hNil := NewVoiceHandler(nil)
		routerNil := gin.New()
		routerNil.POST("/signal", func(c *gin.Context) {
			c.Set("user_id", "user_123")
			hNil.Signal(c)
		})

		payload := VoiceSignal{
			Type:      "offer",
			TargetID:  "user_456",
			NetworkID: "net_123",
		}
		jsonBytes, _ := json.Marshal(payload)
		
		req, _ := http.NewRequest("POST", "/signal", bytes.NewBuffer(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		routerNil.ServeHTTP(w, req)

		// Expect 501 Not Implemented as per our handler logic when redis is nil
		assert.Equal(t, http.StatusNotImplemented, w.Code)
	})
}

func TestVoiceHandler_Signal_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rc := redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})
	h := NewVoiceHandler(rc)
	router := gin.New()
	
	// No user_id in context
	router.POST("/signal", func(c *gin.Context) {
		h.Signal(c)
	})

	payload := VoiceSignal{
		Type:      "offer",
		TargetID:  "user_456",
		NetworkID: "net_123",
	}
	jsonBytes, _ := json.Marshal(payload)
	
	req, _ := http.NewRequest("POST", "/signal", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
