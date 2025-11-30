package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPeerRoutes(t *testing.T, authMiddleware gin.HandlerFunc) (*gin.Engine, *repository.InMemoryPeerRepository) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	peerRepo := repository.NewInMemoryPeerRepository()
	deviceRepo := repository.NewInMemoryDeviceRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()
	peerService := service.NewPeerService(peerRepo, deviceRepo, networkRepo)
	handler := NewPeerHandler(peerService)

	r := gin.New()
	RegisterPeerRoutes(r, handler, authMiddleware)

	return r, peerRepo
}

func TestRegisterPeerRoutes(t *testing.T) {
	auth := func(c *gin.Context) { c.Next() }
	r, _ := setupPeerRoutes(t, auth)

	routes := r.Routes()

	expected := map[string]bool{
		"POST /v1/peers":                            false,
		"GET /v1/peers/:id":                         false,
		"PATCH /v1/peers/:id":                       false,
		"DELETE /v1/peers/:id":                      false,
		"GET /v1/peers/:id/stats":                   false,
		"POST /v1/peers/:id/stats":                  false,
		"POST /v1/peers/:id/rotate-keys":            false,
		"GET /v1/networks/:network_id/peers":        false,
		"GET /v1/networks/:network_id/peers/active": false,
		"GET /v1/networks/:network_id/peers/stats":  false,
		"GET /v1/devices/:device_id/peers":          false,
	}

	for _, rt := range routes {
		key := rt.Method + " " + rt.Path
		if _, ok := expected[key]; ok {
			expected[key] = true
		}
	}

	for route, seen := range expected {
		assert.Truef(t, seen, "route not registered: %s", route)
	}
}

func TestPeerRoutes_AuthMiddleware(t *testing.T) {
	t.Run("blocks unauthorized requests", func(t *testing.T) {
		var middlewareCalled bool
		auth := func(c *gin.Context) {
			middlewareCalled = true
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		r, _ := setupPeerRoutes(t, auth)

		req := httptest.NewRequest("GET", "/v1/peers/any", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.True(t, middlewareCalled)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("passes through when authorized", func(t *testing.T) {
		auth := func(c *gin.Context) {
			c.Set("user_id", "user-1")
			c.Set("tenant_id", "tenant-1")
			c.Next()
		}

		r, peerRepo := setupPeerRoutes(t, auth)

		peer := &domain.Peer{
			NetworkID:  "net-1",
			DeviceID:   "dev-1",
			TenantID:   "tenant-1",
			PublicKey:  "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
			AllowedIPs: []string{"10.0.0.2/32"},
		}
		require.NoError(t, peerRepo.Create(context.Background(), peer))

		req := httptest.NewRequest("GET", "/v1/peers/"+peer.ID, nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestPeerRoutes_MethodNotAllowed(t *testing.T) {
	auth := func(c *gin.Context) { c.Next() }

	r, _ := setupPeerRoutes(t, auth)
	r.HandleMethodNotAllowed = true

	req := httptest.NewRequest("PUT", "/v1/peers/any", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
