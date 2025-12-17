package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/service"
	"github.com/orhaniscoding/goconnect/server/internal/wireguard"
)

// WireGuardHandler handles WireGuard profile generation requests
type WireGuardHandler struct {
	networkRepo      NetworkRepository
	membershipRepo   MembershipRepository
	deviceService    *service.DeviceService
	peerService      *service.PeerService
	userRepo         UserRepository
	profileGenerator *wireguard.ProfileGenerator
	auditor          service.Auditor
}

// NetworkRepository interface for network operations
type NetworkRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Network, error)
}

// MembershipRepository interface for membership operations
type MembershipRepository interface {
	Get(ctx context.Context, networkID, userID string) (*domain.Membership, error)
}

// UserRepository interface for user operations
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
}

// NewWireGuardHandler creates a new WireGuard handler
func NewWireGuardHandler(
	networkRepo NetworkRepository,
	membershipRepo MembershipRepository,
	deviceService *service.DeviceService,
	peerService *service.PeerService,
	userRepo UserRepository,
	profileGenerator *wireguard.ProfileGenerator,
	auditor service.Auditor,
) *WireGuardHandler {
	return &WireGuardHandler{
		networkRepo:      networkRepo,
		membershipRepo:   membershipRepo,
		deviceService:    deviceService,
		peerService:      peerService,
		userRepo:         userRepo,
		profileGenerator: profileGenerator,
		auditor:          auditor,
	}
}

// GetProfile handles GET /v1/networks/:id/wg/profile
// SECURITY: This endpoint no longer accepts private_key as a query parameter.
// The client must inject their locally-stored private key into the config.
// Only JSON responses are supported (Accept: application/json required).
func (h *WireGuardHandler) GetProfile(c *gin.Context) {
	networkID := c.Param("id")
	userID := c.GetString("user_id")
	deviceID := c.Query("device_id") // Required query parameter

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "ERR_UNAUTHORIZED",
			"message": "Authentication required",
		})
		return
	}

	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_VALIDATION",
			"message": "device_id query parameter is required",
		})
		return
	}

	// SECURITY: Only JSON API is supported - private keys should NEVER be sent to server
	acceptHeader := c.GetHeader("Accept")
	// Allow typical Accept variants like "application/json, */*"
	if !strings.Contains(acceptHeader, "application/json") {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_VALIDATION",
			"message": "Only JSON API is supported (Accept: application/json). Private keys should never leave the device.",
		})
		return
	}

	// Log if someone tries to send private_key (deprecated, ignored)
	if c.Query("private_key") != "" {
		slog.Warn("SECURITY: private_key in URL is ignored for security. Client must inject key locally.",
			"user_id", userID, "device_id", deviceID, "network_id", networkID)
	}

	// Get network
	network, err := h.networkRepo.GetByID(c.Request.Context(), networkID)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			c.JSON(domainErr.ToHTTPStatus(), domainErr)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "ERR_NOT_FOUND",
			"message": "Network not found",
		})
		return
	}

	// Enforce tenant isolation
	tenantID := c.GetString("tenant_id")
	if network.TenantID != tenantID {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "ERR_NOT_FOUND",
			"message": "Network not found",
		})
		return
	}

	// Check membership (user must be approved member)
	membership, err := h.membershipRepo.Get(c.Request.Context(), networkID, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    "ERR_FORBIDDEN",
			"message": "You are not a member of this network",
		})
		return
	}

	if membership.Status != domain.StatusApproved {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    "ERR_FORBIDDEN",
			"message": "Your membership is not approved yet",
			"details": gin.H{"status": membership.Status},
		})
		return
	}

	// Get device (must belong to user)
	isAdmin := c.GetBool("is_admin")
	device, err := h.deviceService.GetDevice(c.Request.Context(), deviceID, userID, tenantID, isAdmin)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			c.JSON(domainErr.ToHTTPStatus(), domainErr)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL_SERVER",
			"message": "Failed to get device",
		})
		return
	}

	// Get peer for this device in this network
	peer, err := h.peerService.GetPeerByNetworkAndDevice(c.Request.Context(), networkID, deviceID)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			c.JSON(domainErr.ToHTTPStatus(), domainErr)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "ERR_NOT_FOUND",
			"message": "No peer configuration found for this device in this network. Please ensure the device is registered and you are a member of the network.",
		})
		return
	}

	// Extract device IP from peer's AllowedIPs
	var deviceIP string
	if len(peer.AllowedIPs) > 0 {
		// AllowedIPs format: "10.0.0.5/32"
		// Extract IP without CIDR notation
		ip := peer.AllowedIPs[0]
		if idx := len(ip) - 3; idx > 0 && ip[idx:] == "/32" {
			deviceIP = ip[:idx]
		} else {
			// Fallback: use as-is
			deviceIP = ip
		}
	}

	if deviceIP == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL_SERVER",
			"message": "Device has no IP allocation in this network",
		})
		return
	}

	// Calculate prefix length from CIDR
	_, ipnet, err := net.ParseCIDR(network.CIDR)
	prefixLen := 24 // Default
	if err == nil {
		ones, _ := ipnet.Mask.Size()
		prefixLen = ones
	}

	// JSON-only response: return device config WITHOUT private key (client must inject locally).
	meshPeers, meshErr := h.peerService.GetActivePeersConfig(c.Request.Context(), networkID)

		// Build DNS list from network config
		dnsServers := []string{}
		if network.DNS != nil && *network.DNS != "" {
			for _, s := range strings.Split(*network.DNS, ",") {
				if trimmed := strings.TrimSpace(s); trimmed != "" {
					dnsServers = append(dnsServers, trimmed)
				}
			}
		}

		// Construct Interface Config
		// NOTE: PrivateKey is intentionally empty - client MUST inject locally
		interfaceConfig := domain.InterfaceConfig{
			PrivateKey: "", // SECURITY: Client must inject their locally-stored private key
		ListenPort: 51820,
			Addresses:  []string{deviceIP + "/" + fmt.Sprintf("%d", prefixLen)},
			DNS:        dnsServers,
		}

		// Filter out self from peers
		peers := make([]domain.PeerConfig, 0)
	if meshErr == nil {
			for _, p := range meshPeers {
			// Skip self (check by public key since PeerConfig doesn't have DeviceID)
				if p.PublicKey == device.PubKey {
					continue
				}
				peers = append(peers, p)
			}
		}

		c.JSON(http.StatusOK, domain.DeviceConfig{
			Interface: interfaceConfig,
			Peers:     peers,
		})

		h.auditor.Event(c.Request.Context(), tenantID, "PROFILE_RENDERED_JSON", userID, networkID, map[string]any{
			"device_id":   deviceID,
			"device_name": device.Name,
			"peers_count": len(peers),
		})
}

// RegisterWireGuardRoutes registers WireGuard routes
func RegisterWireGuardRoutes(
	r *gin.Engine,
	handler *WireGuardHandler,
	authMiddleware gin.HandlerFunc,
) {
	wg := r.Group("/v1/networks/:id/wg")
	wg.Use(authMiddleware) // Requires authentication

	wg.GET("/profile", handler.GetProfile) // Generate WireGuard profile for device
}
