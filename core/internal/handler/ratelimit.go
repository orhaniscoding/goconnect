package handler

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

type rateBucket struct {
	mu     sync.Mutex
	tokens int
	max    int
	refill time.Duration
	last   time.Time
}

func (b *rateBucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	if b.last.IsZero() {
		b.last = now
	}
	// Refill one token per interval
	elapsed := now.Sub(b.last)
	add := int(elapsed / b.refill)
	if add > 0 {
		b.tokens = min(b.max, b.tokens+add)
		b.last = b.last.Add(time.Duration(add) * b.refill)
	}
	if b.tokens > 0 {
		b.tokens--
		return true
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RateLimitMiddleware limits requests per key (user or IP).
func RateLimitMiddleware(capacity int, per time.Duration) gin.HandlerFunc {
	buckets := sync.Map{}
	return func(c *gin.Context) {
		// Key by user if available, else IP
		key := c.ClientIP()
		if uid, ok := c.Get("user_id"); ok {
			key = uid.(string)
		}
		v, _ := buckets.LoadOrStore(key, &rateBucket{tokens: capacity, max: capacity, refill: per, last: time.Now()})
		b := v.(*rateBucket)
		if !b.allow() {
			derr := domain.NewError(domain.ErrRateLimited, "Too many requests", nil)
			derr.RetryAfter = int(per.Seconds())
			c.Header("Retry-After", fmt.Sprintf("%d", derr.RetryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, derr)
			return
		}
		c.Next()
	}
}

// NewRateLimiterFromEnv creates a rate limiter using environment overrides.
// Env vars:
//   - SERVER_RL_CAPACITY: integer tokens per window (default: defCap)
//   - SERVER_RL_WINDOW_MS: integer window in milliseconds (default: defPer)
func NewRateLimiterFromEnv(defCap int, defPer time.Duration) gin.HandlerFunc {
	cap := defCap
	if v := os.Getenv("SERVER_RL_CAPACITY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cap = n
		}
	}
	per := defPer
	if v := os.Getenv("SERVER_RL_WINDOW_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			per = time.Duration(n) * time.Millisecond
		}
	}
	return RateLimitMiddleware(cap, per)
}

// EndpointRateLimits contains rate limits for specific endpoint types
type EndpointRateLimits struct {
	// Auth endpoints (login, register)
	AuthCapacity int
	AuthWindow   time.Duration
	// Join endpoints (network join requests)
	JoinCapacity int
	JoinWindow   time.Duration
	// Chat endpoints (message sending)
	ChatCapacity int
	ChatWindow   time.Duration
	// Invite endpoints (invite creation)
	InviteCapacity int
	InviteWindow   time.Duration
	// Default for other endpoints
	DefaultCapacity int
	DefaultWindow   time.Duration
}

// DefaultEndpointRateLimits returns sensible defaults as per TECH_SPEC
func DefaultEndpointRateLimits() EndpointRateLimits {
	return EndpointRateLimits{
		// Login: 5 attempts per minute per IP (brute force protection)
		AuthCapacity: 5,
		AuthWindow:   time.Minute,
		// Join: 10 requests per minute per user
		JoinCapacity: 10,
		JoinWindow:   time.Minute,
		// Chat: 30 messages per minute per user
		ChatCapacity: 30,
		ChatWindow:   time.Minute,
		// Invite: 20 creates per minute per user
		InviteCapacity: 20,
		InviteWindow:   time.Minute,
		// Default: 100 requests per minute
		DefaultCapacity: 100,
		DefaultWindow:   time.Minute,
	}
}

// LoadEndpointRateLimitsFromEnv loads rate limits from environment variables
func LoadEndpointRateLimitsFromEnv() EndpointRateLimits {
	limits := DefaultEndpointRateLimits()

	if v := os.Getenv("RL_AUTH_CAPACITY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limits.AuthCapacity = n
		}
	}
	if v := os.Getenv("RL_AUTH_WINDOW_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limits.AuthWindow = time.Duration(n) * time.Second
		}
	}
	if v := os.Getenv("RL_JOIN_CAPACITY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limits.JoinCapacity = n
		}
	}
	if v := os.Getenv("RL_JOIN_WINDOW_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limits.JoinWindow = time.Duration(n) * time.Second
		}
	}
	if v := os.Getenv("RL_CHAT_CAPACITY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limits.ChatCapacity = n
		}
	}
	if v := os.Getenv("RL_CHAT_WINDOW_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limits.ChatWindow = time.Duration(n) * time.Second
		}
	}
	if v := os.Getenv("RL_INVITE_CAPACITY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limits.InviteCapacity = n
		}
	}
	if v := os.Getenv("RL_INVITE_WINDOW_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limits.InviteWindow = time.Duration(n) * time.Second
		}
	}

	return limits
}

// RateLimitStore manages multiple rate limit buckets for different endpoint types
type RateLimitStore struct {
	limits  EndpointRateLimits
	buckets map[string]*sync.Map // Type -> (key -> bucket)
	mu      sync.RWMutex
}

// NewRateLimitStore creates a new rate limit store
func NewRateLimitStore(limits EndpointRateLimits) *RateLimitStore {
	return &RateLimitStore{
		limits: limits,
		buckets: map[string]*sync.Map{
			"auth":    {},
			"join":    {},
			"chat":    {},
			"invite":  {},
			"default": {},
		},
	}
}

// check performs rate limit check for a specific type and key
func (s *RateLimitStore) check(limitType, key string) (bool, int) {
	s.mu.RLock()
	bucketsMap, exists := s.buckets[limitType]
	s.mu.RUnlock()

	if !exists {
		return true, 0 // No limit defined
	}

	var capacity int
	var window time.Duration

	switch limitType {
	case "auth":
		capacity = s.limits.AuthCapacity
		window = s.limits.AuthWindow
	case "join":
		capacity = s.limits.JoinCapacity
		window = s.limits.JoinWindow
	case "chat":
		capacity = s.limits.ChatCapacity
		window = s.limits.ChatWindow
	case "invite":
		capacity = s.limits.InviteCapacity
		window = s.limits.InviteWindow
	default:
		capacity = s.limits.DefaultCapacity
		window = s.limits.DefaultWindow
	}

	v, _ := bucketsMap.LoadOrStore(key, &rateBucket{
		tokens: capacity,
		max:    capacity,
		refill: window / time.Duration(capacity),
		last:   time.Now(),
	})
	bucket := v.(*rateBucket)

	if bucket.allow() {
		return true, 0
	}

	return false, int(window.Seconds())
}

// AuthRateLimit returns middleware for auth endpoint rate limiting
func (s *RateLimitStore) AuthRateLimit() gin.HandlerFunc {
	return s.rateLimitMiddleware("auth", true) // Use IP for auth (not logged in yet)
}

// JoinRateLimit returns middleware for join endpoint rate limiting
func (s *RateLimitStore) JoinRateLimit() gin.HandlerFunc {
	return s.rateLimitMiddleware("join", false) // Use user_id
}

// ChatRateLimit returns middleware for chat endpoint rate limiting
func (s *RateLimitStore) ChatRateLimit() gin.HandlerFunc {
	return s.rateLimitMiddleware("chat", false) // Use user_id
}

// InviteRateLimit returns middleware for invite endpoint rate limiting
func (s *RateLimitStore) InviteRateLimit() gin.HandlerFunc {
	return s.rateLimitMiddleware("invite", false) // Use user_id
}

// rateLimitMiddleware creates a rate limit middleware for a specific type
func (s *RateLimitStore) rateLimitMiddleware(limitType string, useIPOnly bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if !useIPOnly {
			if uid, ok := c.Get("user_id"); ok {
				key = uid.(string)
			}
		}

		allowed, retryAfter := s.check(limitType, key)
		if !allowed {
			derr := domain.NewError(domain.ErrRateLimited, "Too many requests", nil)
			derr.RetryAfter = retryAfter
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, derr)
			return
		}
		c.Next()
	}
}
