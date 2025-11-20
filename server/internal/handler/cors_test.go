package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewCORSMiddleware_AllowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.CORSConfig{
		AllowedOrigins:   []string{"http://localhost:3000", "https://app.example.com"},
		AllowCredentials: true,
		MaxAge:           24 * time.Hour,
	}

	r := gin.New()
	r.Use(NewCORSMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Authorization, Content-Type, X-Request-ID", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "X-Request-ID", w.Header().Get("Access-Control-Expose-Headers"))
	assert.Contains(t, w.Header().Get("Access-Control-Max-Age"), "24h")
}

func TestNewCORSMiddleware_DisallowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.CORSConfig{
		AllowedOrigins: []string{"http://localhost:3000"},
	}

	r := gin.New()
	r.Use(NewCORSMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://evil.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)                                   // Still responds
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin")) // But no CORS headers
}

func TestNewCORSMiddleware_NoOriginHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.CORSConfig{
		AllowedOrigins: []string{"http://localhost:3000"},
	}

	r := gin.New()
	r.Use(NewCORSMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	// No Origin header set
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestNewCORSMiddleware_PreflightRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.CORSConfig{
		AllowedOrigins: []string{"http://localhost:3000"},
	}

	r := gin.New()
	r.Use(NewCORSMiddleware(cfg))
	r.OPTIONS("/test", func(c *gin.Context) {
		c.String(200, "should not reach here")
	})
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Body.String()) // Preflight should not have response body
}

func TestNewCORSMiddleware_WithoutCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.CORSConfig{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowCredentials: false, // Explicitly false
	}

	r := gin.New()
	r.Use(NewCORSMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestNewCORSMiddleware_ZeroMaxAge(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.CORSConfig{
		AllowedOrigins: []string{"http://localhost:3000"},
		MaxAge:         0, // Zero max age
	}

	r := gin.New()
	r.Use(NewCORSMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Max-Age"))
}

func TestNewCORSMiddleware_TrimWhitespace(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.CORSConfig{
		AllowedOrigins: []string{"  http://localhost:3000  ", " https://app.example.com "},
	}

	r := gin.New()
	r.Use(NewCORSMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	// Test first origin (with trimmed whitespace)
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))

	// Test second origin
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("Origin", "https://app.example.com")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	assert.Equal(t, 200, w2.Code)
	assert.Equal(t, "https://app.example.com", w2.Header().Get("Access-Control-Allow-Origin"))
}

func TestCheckOrigin_AllowedOrigin(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins: []string{"http://localhost:3000", "https://app.example.com"},
	}

	checker := CheckOrigin(cfg)

	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	assert.True(t, checker(req))
}

func TestCheckOrigin_DisallowedOrigin(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins: []string{"http://localhost:3000"},
	}

	checker := CheckOrigin(cfg)

	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Origin", "http://evil.com")

	assert.False(t, checker(req))
}

func TestCheckOrigin_NoOriginHeader(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins: []string{"http://localhost:3000"},
	}

	checker := CheckOrigin(cfg)

	req := httptest.NewRequest("GET", "/ws", nil)
	// No Origin header

	assert.True(t, checker(req)) // Should allow requests without Origin header
}

func TestCheckOrigin_WhitespaceTrimming(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins: []string{"  http://localhost:3000  "},
	}

	checker := CheckOrigin(cfg)

	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	assert.True(t, checker(req))
}

func TestCheckOrigin_MultipleOrigins(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"https://app.example.com",
			"https://staging.example.com",
		},
	}

	checker := CheckOrigin(cfg)

	tests := []struct {
		origin   string
		expected bool
	}{
		{"http://localhost:3000", true},
		{"https://app.example.com", true},
		{"https://staging.example.com", true},
		{"http://evil.com", false},
		{"https://phishing.com", false},
		{"", true}, // No origin
	}

	for _, tt := range tests {
		t.Run(tt.origin, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/ws", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			assert.Equal(t, tt.expected, checker(req))
		})
	}
}
