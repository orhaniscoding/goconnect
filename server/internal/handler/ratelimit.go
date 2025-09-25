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
