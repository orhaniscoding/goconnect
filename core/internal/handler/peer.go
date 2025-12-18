package handler

import (
"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/service"
)

// PeerHandler handles HTTP requests for peer management
type PeerHandler struct {
	peerService *service.PeerService
}

// NewPeerHandler creates a new peer handler
func NewPeerHandler(peerService *service.PeerService) *PeerHandler {
	return &PeerHandler{
		peerService: peerService,
	}
}

// CreatePeer creates a new peer
// @Summary Create a new peer
// @Description Creates a new WireGuard peer for a device in a network
// @Tags peers
// @Accept json
// @Produce json
// @Param request body domain.CreatePeerRequest true "Peer creation request"
// @Success 201 {object} domain.Peer
// @Failure 400 {object} domain.Error
// @Failure 409 {object} domain.Error
// @Failure 500 {object} domain.Error
// @Router /v1/peers [post]
func (h *PeerHandler) CreatePeer(c *gin.Context) {
	var req domain.CreatePeerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewError(domain.ErrValidation, err.Error(), nil))
		return
	}

	peer, err := h.peerService.CreatePeer(c.Request.Context(), &req)
	if err != nil {
		var domainErr *domain.Error; if errors.As(err, &domainErr) {
			c.JSON(domainErr.ToHTTPStatus(), domainErr)
			return
		}
		slog.Error("CreatePeer: Failed to create peer", "error", err)
		c.JSON(http.StatusInternalServerError, domain.NewError(domain.ErrInternalServer, "failed to create peer", nil))
		return
	}

	slog.Info("Peer created successfully", "peer_id", peer.ID)
	c.JSON(http.StatusCreated, peer)
}

// GetPeer retrieves a peer by ID
// @Summary Get peer by ID
// @Description Retrieves a peer by its ID
// @Tags peers
// @Produce json
// @Param id path string true "Peer ID"
// @Success 200 {object} domain.Peer
// @Failure 404 {object} domain.Error
// @Failure 500 {object} domain.Error
// @Router /v1/peers/{id} [get]
func (h *PeerHandler) GetPeer(c *gin.Context) {
	peerID := c.Param("id")

	peer, err := h.peerService.GetPeer(c.Request.Context(), peerID)
	if err != nil {
		var domainErr *domain.Error; if errors.As(err, &domainErr) {
			c.JSON(domainErr.ToHTTPStatus(), domainErr)
			return
		}
		c.JSON(http.StatusInternalServerError, domain.NewError(domain.ErrInternalServer, "failed to get peer", nil))
		return
	}

	c.JSON(http.StatusOK, peer)
}

// GetPeersByNetwork retrieves all peers in a network
// @Summary Get peers by network
// @Description Retrieves all peers in a network
// @Tags peers
// @Produce json
// @Param network_id path string true "Network ID"
// @Success 200 {array} domain.Peer
// @Failure 404 {object} domain.Error
// @Failure 500 {object} domain.Error
// @Router /v1/networks/{network_id}/peers [get]
func (h *PeerHandler) GetPeersByNetwork(c *gin.Context) {
	networkID := c.Param("network_id")

	peers, err := h.peerService.GetPeersByNetwork(c.Request.Context(), networkID)
	if err != nil {
		var domainErr *domain.Error; if errors.As(err, &domainErr) {
			c.JSON(domainErr.ToHTTPStatus(), domainErr)
			return
		}
		c.JSON(http.StatusInternalServerError, domain.NewError(domain.ErrInternalServer, "failed to get peers", nil))
		return
	}

	c.JSON(http.StatusOK, peers)
}

// GetPeersByDevice retrieves all peers for a device
// @Summary Get peers by device
// @Description Retrieves all peers for a device
// @Tags peers
// @Produce json
// @Param device_id path string true "Device ID"
// @Success 200 {array} domain.Peer
// @Failure 404 {object} domain.Error
// @Failure 500 {object} domain.Error
// @Router /v1/devices/{device_id}/peers [get]
func (h *PeerHandler) GetPeersByDevice(c *gin.Context) {
	deviceID := c.Param("device_id")

	peers, err := h.peerService.GetPeersByDevice(c.Request.Context(), deviceID)
	if err != nil {
		var domainErr *domain.Error; if errors.As(err, &domainErr) {
			c.JSON(domainErr.ToHTTPStatus(), domainErr)
			return
		}
		c.JSON(http.StatusInternalServerError, domain.NewError(domain.ErrInternalServer, "failed to get peers", nil))
		return
	}

	c.JSON(http.StatusOK, peers)
}

// GetActivePeers retrieves all active peers in a network
// @Summary Get active peers
// @Description Retrieves all active peers in a network
// @Tags peers
// @Produce json
// @Param network_id path string true "Network ID"
// @Success 200 {array} domain.Peer
// @Failure 404 {object} domain.Error
// @Failure 500 {object} domain.Error
// @Router /v1/networks/{network_id}/peers/active [get]
func (h *PeerHandler) GetActivePeers(c *gin.Context) {
	networkID := c.Param("network_id")

	peers, err := h.peerService.GetActivePeers(c.Request.Context(), networkID)
	if err != nil {
		var domainErr *domain.Error; if errors.As(err, &domainErr) {
			c.JSON(domainErr.ToHTTPStatus(), domainErr)
			return
		}
		c.JSON(http.StatusInternalServerError, domain.NewError(domain.ErrInternalServer, "failed to get active peers", nil))
		return
	}

	c.JSON(http.StatusOK, peers)
}

// UpdatePeer updates a peer
// @Summary Update peer
// @Description Updates a peer's configuration
// @Tags peers
// @Accept json
// @Produce json
// @Param id path string true "Peer ID"
// @Param request body domain.UpdatePeerRequest true "Peer update request"
// @Success 200 {object} domain.Peer
// @Failure 400 {object} domain.Error
// @Failure 404 {object} domain.Error
// @Failure 500 {object} domain.Error
// @Router /v1/peers/{id} [patch]
func (h *PeerHandler) UpdatePeer(c *gin.Context) {
	peerID := c.Param("id")

	var req domain.UpdatePeerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewError(domain.ErrValidation, err.Error(), nil))
		return
	}

	peer, err := h.peerService.UpdatePeer(c.Request.Context(), peerID, &req)
	if err != nil {
		var domainErr *domain.Error; if errors.As(err, &domainErr) {
			c.JSON(domainErr.ToHTTPStatus(), domainErr)
			return
		}
		c.JSON(http.StatusInternalServerError, domain.NewError(domain.ErrInternalServer, "failed to update peer", nil))
		return
	}

	c.JSON(http.StatusOK, peer)
}

// UpdatePeerStats updates peer statistics
// @Summary Update peer statistics
// @Description Updates peer statistics from WireGuard
// @Tags peers
// @Accept json
// @Produce json
// @Param id path string true "Peer ID"
// @Param request body domain.UpdatePeerStatsRequest true "Peer stats update request"
// @Success 204
// @Failure 400 {object} domain.Error
// @Failure 404 {object} domain.Error
// @Failure 500 {object} domain.Error
// @Router /v1/peers/{id}/stats [post]
func (h *PeerHandler) UpdatePeerStats(c *gin.Context) {
	peerID := c.Param("id")

	var req domain.UpdatePeerStatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewError(domain.ErrValidation, err.Error(), nil))
		return
	}

	if err := h.peerService.UpdatePeerStats(c.Request.Context(), peerID, &req); err != nil {
		var domainErr *domain.Error; if errors.As(err, &domainErr) {
			c.JSON(domainErr.ToHTTPStatus(), domainErr)
			return
		}
		c.JSON(http.StatusInternalServerError, domain.NewError(domain.ErrInternalServer, "failed to update peer stats", nil))
		return
	}

	c.Status(http.StatusNoContent)
}

// DeletePeer deletes a peer
// @Summary Delete peer
// @Description Soft-deletes a peer
// @Tags peers
// @Produce json
// @Param id path string true "Peer ID"
// @Success 204
// @Failure 404 {object} domain.Error
// @Failure 500 {object} domain.Error
// @Router /v1/peers/{id} [delete]
func (h *PeerHandler) DeletePeer(c *gin.Context) {
	peerID := c.Param("id")

	if err := h.peerService.DeletePeer(c.Request.Context(), peerID); err != nil {
		var domainErr *domain.Error; if errors.As(err, &domainErr) {
			c.JSON(domainErr.ToHTTPStatus(), domainErr)
			return
		}
		c.JSON(http.StatusInternalServerError, domain.NewError(domain.ErrInternalServer, "failed to delete peer", nil))
		return
	}

	c.Status(http.StatusNoContent)
}

// GetPeerStats retrieves peer statistics
// @Summary Get peer statistics
// @Description Retrieves real-time statistics for a peer
// @Tags peers
// @Produce json
// @Param id path string true "Peer ID"
// @Success 200 {object} domain.PeerStats
// @Failure 404 {object} domain.Error
// @Failure 500 {object} domain.Error
// @Router /v1/peers/{id}/stats [get]
func (h *PeerHandler) GetPeerStats(c *gin.Context) {
	peerID := c.Param("id")

	stats, err := h.peerService.GetPeerStats(c.Request.Context(), peerID)
	if err != nil {
		var domainErr *domain.Error; if errors.As(err, &domainErr) {
			c.JSON(domainErr.ToHTTPStatus(), domainErr)
			return
		}
		c.JSON(http.StatusInternalServerError, domain.NewError(domain.ErrInternalServer, "failed to get peer stats", nil))
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetNetworkPeerStats retrieves statistics for all peers in a network
// @Summary Get network peer statistics
// @Description Retrieves statistics for all peers in a network
// @Tags peers
// @Produce json
// @Param network_id path string true "Network ID"
// @Success 200 {array} domain.PeerStats
// @Failure 404 {object} domain.Error
// @Failure 500 {object} domain.Error
// @Router /v1/networks/{network_id}/peers/stats [get]
func (h *PeerHandler) GetNetworkPeerStats(c *gin.Context) {
	networkID := c.Param("network_id")

	stats, err := h.peerService.GetNetworkPeerStats(c.Request.Context(), networkID)
	if err != nil {
		var domainErr *domain.Error; if errors.As(err, &domainErr) {
			c.JSON(domainErr.ToHTTPStatus(), domainErr)
			return
		}
		c.JSON(http.StatusInternalServerError, domain.NewError(domain.ErrInternalServer, "failed to get network peer stats", nil))
		return
	}

	c.JSON(http.StatusOK, stats)
}

// RotatePeerKeys rotates the WireGuard keys for a peer
// @Summary Rotate peer keys
// @Description Rotates the WireGuard keys for a peer
// @Tags peers
// @Produce json
// @Param id path string true "Peer ID"
// @Success 200 {object} domain.Peer
// @Failure 404 {object} domain.Error
// @Failure 500 {object} domain.Error
// @Router /v1/peers/{id}/rotate-keys [post]
func (h *PeerHandler) RotatePeerKeys(c *gin.Context) {
	peerID := c.Param("id")

	peer, err := h.peerService.RotatePeerKeys(c.Request.Context(), peerID)
	if err != nil {
		var domainErr *domain.Error; if errors.As(err, &domainErr) {
			c.JSON(domainErr.ToHTTPStatus(), domainErr)
			return
		}
		c.JSON(http.StatusInternalServerError, domain.NewError(domain.ErrInternalServer, "failed to rotate keys", nil))
		return
	}

	c.JSON(http.StatusOK, peer)
}
