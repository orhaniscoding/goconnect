package handler

import (
	"context"
	"fmt"
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
func (h *WireGuardHandler) GetProfile(c *gin.Context) {
	networkID := c.Param("id")
	userID := c.GetString("user_id")
	deviceID := c.Query("device_id")     // Required query parameter
	privateKey := c.Query("private_key") // Client's WireGuard private key

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

	if privateKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_VALIDATION",
			"message": "private_key query parameter is required",
		})
		return
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

	// Get user for email
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		// Fallback if user service unavailable
		user = &domain.User{Email: "unknown@example.com"}
	}

	// Calculate prefix length from CIDR
	_, ipnet, err := net.ParseCIDR(network.CIDR)
	prefixLen := 24 // Default
	if err == nil {
		ones, _ := ipnet.Mask.Size()
		prefixLen = ones
	}

	// Generate WireGuard profile
	req := &wireguard.ProfileRequest{
		NetworkID:        network.ID,
		NetworkName:      network.Name,
		NetworkCIDR:      network.CIDR,
		DeviceID:         device.ID,
		DeviceName:       device.Name,
		DeviceIP:         deviceIP,
		DevicePrivateKey: privateKey, // Client provides their own private key
		PrefixLen:        prefixLen,
		UserID:           userID,
		UserEmail:        user.Email,
		IncludeHostScope: false, // Option for future enhancement
	}

	// Generate base config
	configTemplate, err := h.profileGenerator.GenerateClientConfig(c.Request.Context(), req)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			c.JSON(domainErr.ToHTTPStatus(), domainErr)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_INTERNAL_SERVER",
			"message": "Failed to generate WireGuard profile",
		})
		return
	}

	// Add mesh peer configuration (other active peers in the network)
	meshPeers, err := h.peerService.GetActivePeersConfig(c.Request.Context(), networkID)

	// Check for JSON request (used by Client Daemon for rich metadata)
	if c.GetHeader("Accept") == "application/json" {
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
		interfaceConfig := domain.InterfaceConfig{
			PrivateKey: privateKey,
			ListenPort: 51820, // Default, client can override
			Addresses:  []string{deviceIP + "/" + fmt.Sprintf("%d", prefixLen)},
			DNS:        dnsServers,
		}

		// Filter out self from peers
		peers := make([]domain.PeerConfig, 0)
		if err == nil {
			for _, p := range meshPeers {
				// We can't easily check DeviceID here because PeerConfig doesn't have it
				// But we can check PublicKey
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

		// Audit the profile generation
		h.auditor.Event(c.Request.Context(), tenantID, "PROFILE_RENDERED_JSON", userID, networkID, map[string]any{
			"device_id":   deviceID,
			"device_name": device.Name,
			"peers_count": len(peers),
		})
		return
	}

	if err == nil && len(meshPeers) > 0 {
		var meshConfig strings.Builder
		meshConfig.WriteString("\n# Mesh Peers (other devices in this network)\n")

		for _, meshPeer := range meshPeers {
			// Skip self (check by public key since we don't have DeviceID in PeerConfig)
			if meshPeer.PublicKey == device.PubKey {
				continue
			}

			// Add peer configuration
			meshConfig.WriteString("\n[Peer]\n")
			if meshPeer.Name != "" {
				meshConfig.WriteString(fmt.Sprintf("# Name: %s\n", meshPeer.Name))
			}
			if meshPeer.Hostname != "" {
				meshConfig.WriteString(fmt.Sprintf("# Hostname: %s\n", meshPeer.Hostname))
			}
			meshConfig.WriteString(fmt.Sprintf("PublicKey = %s\n", meshPeer.PublicKey))
			meshConfig.WriteString(fmt.Sprintf("AllowedIPs = %s\n", strings.Join(meshPeer.AllowedIPs, ", ")))

			// Add endpoint if known (for direct peer-to-peer connection)
			if meshPeer.Endpoint != "" {
				meshConfig.WriteString(fmt.Sprintf("Endpoint = %s\n", meshPeer.Endpoint))
			}

			// Add persistent keepalive for NAT traversal
			if meshPeer.PersistentKeepalive > 0 {
				meshConfig.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", meshPeer.PersistentKeepalive))
			}

			// Add preshared key if configured
			if meshPeer.PresharedKey != "" {
				meshConfig.WriteString(fmt.Sprintf("PresharedKey = %s\n", meshPeer.PresharedKey))
			}
		}

		configTemplate += meshConfig.String()
	}

	// Audit the profile generation
	h.auditor.Event(c.Request.Context(), tenantID, "PROFILE_RENDERED", userID, networkID, map[string]any{
		"device_id":   deviceID,
		"device_name": device.Name,
		"device_ip":   deviceIP,
		"network":     network.Name,
		"mesh_peers":  len(meshPeers) - 1, // Excluding self (approx)
	})

	// Return config as plain text with proper content type
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=goconnect-"+network.Name+"-"+device.Name+".conf")
	c.String(http.StatusOK, configTemplate)
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
