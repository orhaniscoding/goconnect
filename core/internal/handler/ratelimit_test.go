package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestRateLimit_Join429(t *testing.T) {
	router, _ := setupTestRouter()

	// Create a network as admin (separate bucket)
	create := domain.CreateNetworkRequest{
		Name:       "RL Net",
		Visibility: domain.NetworkVisibilityPublic,
		JoinPolicy: domain.JoinPolicyApproval,
		CIDR:       "10.99.0.0/24",
	}
	payload, _ := json.Marshal(create)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/networks", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer admin-token")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 create network, got %d body=%s", w.Code, w.Body.String())
	}
	var cresp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &cresp); err != nil {
		t.Fatalf("unmarshal create: %v", err)
	}
	netID := cresp.Data.ID
	if netID == "" {
		t.Fatalf("network id empty in create response: %s", w.Body.String())
	}

	// Now send 6 join requests as user_dev within 1s
	var last *httptest.ResponseRecorder
	for i := 0; i < 6; i++ {
		last = httptest.NewRecorder()
		jreq, _ := http.NewRequest("POST", "/v1/networks/"+netID+"/join", nil)
		jreq.Header.Set("Authorization", "Bearer valid-token")
		jreq.Header.Set("Idempotency-Key", domain.GenerateIdempotencyKey())
		router.ServeHTTP(last, jreq)
	}
	if last.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 on 6th request, got %d body=%s", last.Code, last.Body.String())
	}
	var errResp domain.Error
	if err := json.Unmarshal(last.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("unmarshal rate limit error: %v", err)
	}
	if errResp.Code != domain.ErrRateLimited {
		t.Fatalf("expected code %s, got %s", domain.ErrRateLimited, errResp.Code)
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"a less than b", 3, 5, 3},
		{"a greater than b", 7, 4, 4},
		{"a equals b", 5, 5, 5},
		{"negative numbers", -3, -1, -3},
		{"zero and positive", 0, 5, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := min(tc.a, tc.b)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestNewRateLimitStore(t *testing.T) {
	limits := DefaultEndpointRateLimits()
	store := NewRateLimitStore(limits)

	assert.NotNil(t, store)
	assert.Equal(t, limits, store.limits)
	assert.NotNil(t, store.buckets)
	assert.Contains(t, store.buckets, "auth")
	assert.Contains(t, store.buckets, "join")
	assert.Contains(t, store.buckets, "chat")
	assert.Contains(t, store.buckets, "invite")
	assert.Contains(t, store.buckets, "default")
}

func TestRateLimitStore_Check(t *testing.T) {
	limits := EndpointRateLimits{
		AuthCapacity:    2,
		AuthWindow:      time.Minute,
		JoinCapacity:    2,
		JoinWindow:      time.Minute,
		ChatCapacity:    2,
		ChatWindow:      time.Minute,
		InviteCapacity:  2,
		InviteWindow:    time.Minute,
		DefaultCapacity: 2,
		DefaultWindow:   time.Minute,
	}
	store := NewRateLimitStore(limits)

	t.Run("auth rate limit works", func(t *testing.T) {
		allowed1, _ := store.check("auth", "ip1")
		assert.True(t, allowed1)

		allowed2, _ := store.check("auth", "ip1")
		assert.True(t, allowed2)

		allowed3, retryAfter := store.check("auth", "ip1")
		assert.False(t, allowed3)
		assert.Greater(t, retryAfter, 0)
	})

	t.Run("different keys have separate buckets", func(t *testing.T) {
		allowed1, _ := store.check("join", "user1")
		assert.True(t, allowed1)

		allowed2, _ := store.check("join", "user2")
		assert.True(t, allowed2)
	})

	t.Run("unknown limit type allows all", func(t *testing.T) {
		allowed, retryAfter := store.check("unknown", "key1")
		assert.True(t, allowed)
		assert.Equal(t, 0, retryAfter)
	})
}

func TestRateLimitStore_AuthRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limits := EndpointRateLimits{
		AuthCapacity: 1,
		AuthWindow:   time.Minute,
	}
	store := NewRateLimitStore(limits)

	r := gin.New()
	r.Use(store.AuthRateLimit())
	r.POST("/auth/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// First request should pass
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/auth/login", nil)
	req1.RemoteAddr = "192.168.1.100:12345"
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request should be rate limited
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/auth/login", nil)
	req2.RemoteAddr = "192.168.1.100:12345"
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestRateLimitStore_JoinRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limits := EndpointRateLimits{
		JoinCapacity: 1,
		JoinWindow:   time.Minute,
	}
	store := NewRateLimitStore(limits)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user")
		c.Next()
	})
	r.Use(store.JoinRateLimit())
	r.POST("/join", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// First request should pass
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/join", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request should be rate limited
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/join", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestRateLimitStore_ChatRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limits := EndpointRateLimits{
		ChatCapacity: 1,
		ChatWindow:   time.Minute,
	}
	store := NewRateLimitStore(limits)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "chat-user")
		c.Next()
	})
	r.Use(store.ChatRateLimit())
	r.POST("/chat", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// First request should pass
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/chat", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request should be rate limited
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/chat", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestRateLimitStore_InviteRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limits := EndpointRateLimits{
		InviteCapacity: 1,
		InviteWindow:   time.Minute,
	}
	store := NewRateLimitStore(limits)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "invite-user")
		c.Next()
	})
	r.Use(store.InviteRateLimit())
	r.POST("/invite", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// First request should pass
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/invite", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request should be rate limited
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/invite", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestDefaultEndpointRateLimits(t *testing.T) {
	limits := DefaultEndpointRateLimits()

	assert.Equal(t, 5, limits.AuthCapacity)
	assert.Equal(t, time.Minute, limits.AuthWindow)
	assert.Equal(t, 10, limits.JoinCapacity)
	assert.Equal(t, time.Minute, limits.JoinWindow)
	assert.Equal(t, 30, limits.ChatCapacity)
	assert.Equal(t, time.Minute, limits.ChatWindow)
	assert.Equal(t, 20, limits.InviteCapacity)
	assert.Equal(t, time.Minute, limits.InviteWindow)
	assert.Equal(t, 100, limits.DefaultCapacity)
	assert.Equal(t, time.Minute, limits.DefaultWindow)
}

func TestLoadEndpointRateLimitsFromEnv(t *testing.T) {
	// Test with default values (no env vars set)
	limits := LoadEndpointRateLimitsFromEnv()

	// Should return default values
	assert.Equal(t, 5, limits.AuthCapacity)
	assert.Equal(t, time.Minute, limits.AuthWindow)
}

func TestLoadEndpointRateLimitsFromEnv_WithEnvVars(t *testing.T) {
	// Set environment variables
	t.Setenv("RL_AUTH_CAPACITY", "10")
	t.Setenv("RL_AUTH_WINDOW_SEC", "30")
	t.Setenv("RL_JOIN_CAPACITY", "20")
	t.Setenv("RL_JOIN_WINDOW_SEC", "120")
	t.Setenv("RL_CHAT_CAPACITY", "50")
	t.Setenv("RL_CHAT_WINDOW_SEC", "60")
	t.Setenv("RL_INVITE_CAPACITY", "25")
	t.Setenv("RL_INVITE_WINDOW_SEC", "90")

	limits := LoadEndpointRateLimitsFromEnv()

	assert.Equal(t, 10, limits.AuthCapacity)
	assert.Equal(t, 30*time.Second, limits.AuthWindow)
	assert.Equal(t, 20, limits.JoinCapacity)
	assert.Equal(t, 120*time.Second, limits.JoinWindow)
	assert.Equal(t, 50, limits.ChatCapacity)
	assert.Equal(t, 60*time.Second, limits.ChatWindow)
	assert.Equal(t, 25, limits.InviteCapacity)
	assert.Equal(t, 90*time.Second, limits.InviteWindow)
	// Default doesn't have env var support, stays at default
	assert.Equal(t, 100, limits.DefaultCapacity)
	assert.Equal(t, time.Minute, limits.DefaultWindow)
}

func TestLoadEndpointRateLimitsFromEnv_InvalidValues(t *testing.T) {
	// Set invalid environment variables
	t.Setenv("RL_AUTH_CAPACITY", "invalid")
	t.Setenv("RL_AUTH_WINDOW_SEC", "not-a-number")
	t.Setenv("RL_JOIN_CAPACITY", "-5") // negative - should use default

	limits := LoadEndpointRateLimitsFromEnv()

	// Should use default values when env vars are invalid
	assert.Equal(t, 5, limits.AuthCapacity)
	assert.Equal(t, time.Minute, limits.AuthWindow)
	assert.Equal(t, 10, limits.JoinCapacity) // -5 is invalid, use default
}

func TestNewRateLimiterFromEnv(t *testing.T) {
	t.Run("Uses Defaults When No Env Vars", func(t *testing.T) {
		middleware := NewRateLimiterFromEnv(100, time.Minute)
		assert.NotNil(t, middleware)
	})

	t.Run("Uses Env Vars When Set", func(t *testing.T) {
		t.Setenv("SERVER_RL_CAPACITY", "50")
		t.Setenv("SERVER_RL_WINDOW_MS", "30000")

		middleware := NewRateLimiterFromEnv(100, time.Minute)
		assert.NotNil(t, middleware)
	})

	t.Run("Uses Defaults On Invalid Env Vars", func(t *testing.T) {
		t.Setenv("SERVER_RL_CAPACITY", "invalid")
		t.Setenv("SERVER_RL_WINDOW_MS", "not-a-number")

		middleware := NewRateLimiterFromEnv(100, time.Minute)
		assert.NotNil(t, middleware)
	})

	t.Run("Uses Defaults On Zero Or Negative Values", func(t *testing.T) {
		t.Setenv("SERVER_RL_CAPACITY", "0")
		t.Setenv("SERVER_RL_WINDOW_MS", "-1000")

		middleware := NewRateLimiterFromEnv(100, time.Minute)
		assert.NotNil(t, middleware)
	})
}

