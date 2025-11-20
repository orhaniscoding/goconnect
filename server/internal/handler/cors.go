package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/config"
)

// NewCORSMiddleware creates a CORS middleware from configuration
func NewCORSMiddleware(cfg *config.CORSConfig) gin.HandlerFunc {
	allowedOriginsMap := make(map[string]bool)
	for _, origin := range cfg.AllowedOrigins {
		allowedOriginsMap[strings.TrimSpace(origin)] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		if origin != "" && allowedOriginsMap[origin] {
			c.Header("Access-Control-Allow-Origin", origin)

			if cfg.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}

			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
			c.Header("Access-Control-Expose-Headers", "X-Request-ID")

			if cfg.MaxAge > 0 {
				c.Header("Access-Control-Max-Age", cfg.MaxAge.String())
			}
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// CheckOrigin creates a WebSocket origin checker from configuration
func CheckOrigin(cfg *config.CORSConfig) func(*http.Request) bool {
	allowedOriginsMap := make(map[string]bool)
	for _, origin := range cfg.AllowedOrigins {
		allowedOriginsMap[strings.TrimSpace(origin)] = true
	}

	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true // Allow requests without Origin header (e.g., same-origin)
		}
		return allowedOriginsMap[origin]
	}
}
