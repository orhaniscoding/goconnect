package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/service"
	"github.com/orhaniscoding/goconnect/server/internal/wireguard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestKey() string {
	kp, _ := wireguard.GenerateKeyPair()
	return kp.PublicKey
}

func setupPeerTest() (*gin.Engine, *PeerHandler, *service.PeerService, repository.NetworkRepository, repository.DeviceRepository) {
	gin.SetMode(gin.TestMode)

	// Setup repositories
	peerRepo := repository.NewInMemoryPeerRepository()
	deviceRepo := repository.NewInMemoryDeviceRepository()
	networkRepo := repository.NewInMemoryNetworkRepository()

	// Create test network
	ctx := context.Background()
	networkRepo.Create(ctx, &domain.Network{
		ID:       "network-123",
		TenantID: "tenant-1",
		Name:     "Test Network",
		CIDR:     "10.0.0.0/24",
	})

	// Create test device
	deviceRepo.Create(ctx, &domain.Device{
		ID:       "device-123",
		UserID:   "user-123",
		TenantID: "tenant-1",
		Name:     "Test Device",
		Platform: "linux",
		PubKey:   "devicePublicKeyxxxxxxxxxxxxxxxxxxxxxxxxxx=",
	})

	// Setup service
	peerService := service.NewPeerService(peerRepo, deviceRepo, networkRepo)

	// Setup handler
	handler := NewPeerHandler(peerService)

	// Setup router
	r := gin.New()

	return r, handler, peerService, networkRepo, deviceRepo
}

func TestPeerHandler_CreatePeer(t *testing.T) {
	r, handler, _, _, _ := setupPeerTest()

	r.POST("/v1/peers", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("tenant_id", "tenant-1")
		handler.CreatePeer(c)
	})

	t.Run("Success - create new peer", func(t *testing.T) {
		body := map[string]interface{}{
			"network_id":  "network-123",
			"device_id":   "device-123",
			"public_key":  generateTestKey(),
			"allowed_ips": []string{"10.0.0.1/32"},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/peers", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response["id"])
		assert.Equal(t, "network-123", response["network_id"])
		assert.Equal(t, "device-123", response["device_id"])
	})

	t.Run("Bad Request - missing network_id", func(t *testing.T) {
		body := map[string]interface{}{
			"device_id":   "device-123",
			"public_key":  "testPublicKeyxxxxxxxxxxxxxxxxxxxxxxxxxxx=",
			"allowed_ips": []string{"10.0.0.1/32"},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/peers", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Bad Request - invalid public key", func(t *testing.T) {
		body := map[string]interface{}{
			"network_id":  "network-123",
			"device_id":   "device-123",
			"public_key":  "short",
			"allowed_ips": []string{"10.0.0.1/32"},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/peers", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestPeerHandler_GetPeer(t *testing.T) {
	r, handler, peerService, _, _ := setupPeerTest()

	// Create a peer first
	ctx := context.Background()
	peer, err := peerService.CreatePeer(ctx, &domain.CreatePeerRequest{
		NetworkID:  "network-123",
		DeviceID:   "device-123",
		PublicKey:  generateTestKey(),
		AllowedIPs: []string{"10.0.0.2/32"},
	})
	require.NoError(t, err)

	r.GET("/v1/peers/:id", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("tenant_id", "tenant-1")
		handler.GetPeer(c)
	})

	t.Run("Success - get existing peer", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/v1/peers/%s", peer.ID), nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, peer.ID, response["id"])
	})

	t.Run("Not Found - peer does not exist", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/peers/non-existent-peer", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestPeerHandler_GetPeersByNetwork(t *testing.T) {
	r, handler, peerService, _, _ := setupPeerTest()

	// Create peers
	ctx := context.Background()
	_, err := peerService.CreatePeer(ctx, &domain.CreatePeerRequest{
		NetworkID:  "network-123",
		DeviceID:   "device-123",
		PublicKey:  generateTestKey(),
		AllowedIPs: []string{"10.0.0.3/32"},
	})
	require.NoError(t, err)

	r.GET("/v1/networks/:network_id/peers", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("tenant_id", "tenant-1")
		handler.GetPeersByNetwork(c)
	})

	t.Run("Success - get peers by network", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/networks/network-123/peers", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(response), 1)
	})

	t.Run("Success - empty list for non-existent network", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/networks/non-existent/peers", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		// May return 200 with empty array or 404 depending on implementation
		assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, w.Code)
	})
}

func TestPeerHandler_DeletePeer(t *testing.T) {
	r, handler, peerService, _, _ := setupPeerTest()

	// Create a peer
	ctx := context.Background()
	peer, err := peerService.CreatePeer(ctx, &domain.CreatePeerRequest{
		NetworkID:  "network-123",
		DeviceID:   "device-123",
		PublicKey:  generateTestKey(),
		AllowedIPs: []string{"10.0.0.4/32"},
	})
	require.NoError(t, err)

	r.DELETE("/v1/peers/:id", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("tenant_id", "tenant-1")
		handler.DeletePeer(c)
	})

	t.Run("Success - delete peer", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", fmt.Sprintf("/v1/peers/%s", peer.ID), nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("Not Found - peer already deleted", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", fmt.Sprintf("/v1/peers/%s", peer.ID), nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		// Delete is idempotent - may return 204 or 404
		assert.Contains(t, []int{http.StatusNoContent, http.StatusNotFound}, w.Code)
	})
}

func TestPeerHandler_UpdatePeer(t *testing.T) {
	r, handler, peerService, _, _ := setupPeerTest()

	// Create a peer
	ctx := context.Background()
	peer, err := peerService.CreatePeer(ctx, &domain.CreatePeerRequest{
		NetworkID:  "network-123",
		DeviceID:   "device-123",
		PublicKey:  generateTestKey(),
		AllowedIPs: []string{"10.0.0.5/32"},
	})
	require.NoError(t, err)

	r.PATCH("/v1/peers/:id", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("tenant_id", "tenant-1")
		handler.UpdatePeer(c)
	})

	t.Run("Success - update peer", func(t *testing.T) {
		newEndpoint := "192.168.1.100:51820"
		body := map[string]interface{}{
			"endpoint": newEndpoint,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("PATCH", fmt.Sprintf("/v1/peers/%s", peer.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, newEndpoint, response["endpoint"])
	})

	t.Run("Not Found - peer does not exist", func(t *testing.T) {
		body := map[string]interface{}{
			"endpoint": "192.168.1.200:51820",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("PATCH", "/v1/peers/non-existent", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
