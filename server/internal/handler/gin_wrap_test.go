package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ==================== GIN WRAP TESTS ====================

func TestGinWrap_HandlerCalled(t *testing.T) {
	t.Run("Wrapped Handler Is Called", func(t *testing.T) {
		called := false
		originalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		r := gin.New()
		r.GET("/test", GinWrap(originalHandler))

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.True(t, called, "Original handler should be called")
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "OK", w.Body.String())
	})
}

func TestGinWrap_PathParams(t *testing.T) {
	t.Run("Path Parameters Are Passed", func(t *testing.T) {
		var receivedID string
		originalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedID = r.PathValue("id")
			w.WriteHeader(http.StatusOK)
		})

		r := gin.New()
		r.GET("/users/:id", GinWrap(originalHandler))

		req := httptest.NewRequest("GET", "/users/user-123", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, "user-123", receivedID)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Multiple Path Parameters", func(t *testing.T) {
		var receivedTenantID, receivedNetworkID string
		originalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedTenantID = r.PathValue("tenant_id")
			receivedNetworkID = r.PathValue("network_id")
			w.WriteHeader(http.StatusOK)
		})

		r := gin.New()
		r.GET("/tenants/:tenant_id/networks/:network_id", GinWrap(originalHandler))

		req := httptest.NewRequest("GET", "/tenants/t-123/networks/n-456", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, "t-123", receivedTenantID)
		assert.Equal(t, "n-456", receivedNetworkID)
	})
}

func TestGinWrap_ContextHeaders(t *testing.T) {
	t.Run("Tenant ID Header Is Set", func(t *testing.T) {
		var receivedTenantID string
		originalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedTenantID = r.Header.Get("X-Tenant-ID")
			w.WriteHeader(http.StatusOK)
		})

		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("tenant_id", "tenant-abc")
			c.Next()
		})
		r.GET("/test", GinWrap(originalHandler))

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, "tenant-abc", receivedTenantID)
	})

	t.Run("User ID Header Is Set", func(t *testing.T) {
		var receivedUserID string
		originalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedUserID = r.Header.Get("X-User-ID")
			w.WriteHeader(http.StatusOK)
		})

		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("user_id", "user-xyz")
			c.Next()
		})
		r.GET("/test", GinWrap(originalHandler))

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, "user-xyz", receivedUserID)
	})

	t.Run("Both Headers Are Set", func(t *testing.T) {
		var receivedTenantID, receivedUserID string
		originalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedTenantID = r.Header.Get("X-Tenant-ID")
			receivedUserID = r.Header.Get("X-User-ID")
			w.WriteHeader(http.StatusOK)
		})

		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("tenant_id", "tenant-123")
			c.Set("user_id", "user-456")
			c.Next()
		})
		r.GET("/test", GinWrap(originalHandler))

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, "tenant-123", receivedTenantID)
		assert.Equal(t, "user-456", receivedUserID)
	})

	t.Run("No Context Values - No Headers", func(t *testing.T) {
		var receivedTenantID, receivedUserID string
		originalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedTenantID = r.Header.Get("X-Tenant-ID")
			receivedUserID = r.Header.Get("X-User-ID")
			w.WriteHeader(http.StatusOK)
		})

		r := gin.New()
		r.GET("/test", GinWrap(originalHandler))

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Empty(t, receivedTenantID)
		assert.Empty(t, receivedUserID)
	})
}

func TestGinWrap_RequestPreserved(t *testing.T) {
	t.Run("Query Parameters Preserved", func(t *testing.T) {
		var receivedQuery string
		originalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedQuery = r.URL.Query().Get("filter")
			w.WriteHeader(http.StatusOK)
		})

		r := gin.New()
		r.GET("/search", GinWrap(originalHandler))

		req := httptest.NewRequest("GET", "/search?filter=active", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, "active", receivedQuery)
	})

	t.Run("Request Headers Preserved", func(t *testing.T) {
		var receivedAuth string
		originalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedAuth = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusOK)
		})

		r := gin.New()
		r.GET("/api", GinWrap(originalHandler))

		req := httptest.NewRequest("GET", "/api", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, "Bearer test-token", receivedAuth)
	})
}
