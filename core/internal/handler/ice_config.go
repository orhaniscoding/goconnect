package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/service"
)

// ICEConfigHandler handles ICE configuration endpoints for NAT traversal
type ICEConfigHandler struct {
	turnService  *service.TURNService
	memberRepo   repository.MembershipRepository
}

// NewICEConfigHandler creates a new ICE config handler
func NewICEConfigHandler(turnService *service.TURNService, memberRepo repository.MembershipRepository) *ICEConfigHandler {
	return &ICEConfigHandler{
		turnService:  turnService,
		memberRepo:   memberRepo,
	}
}

// GetICEConfig handles GET /v1/networks/:id/ice-config
// Returns STUN/TURN configuration for P2P connections
// @Summary Get ICE configuration for a network
// @Description Returns STUN servers and time-limited TURN credentials for NAT traversal
// @Tags networks
// @Accept json
// @Produce json
// @Param id path string true "Network ID"
// @Success 200 {object} service.ICEConfig
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Not a member"
// @Failure 500 {object} map[string]string "Internal error"
// @Router /v1/networks/{id}/ice-config [get]
func (h *ICEConfigHandler) GetICEConfig(c *gin.Context) {
	networkID := c.Param("id")
	userID := c.GetString("user_id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Verify user is a member of the network
	membership, err := h.memberRepo.Get(c.Request.Context(), networkID, userID)
	if err != nil || membership == nil || membership.Status != "approved" {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member of this network"})
		return
	}

	// Generate ICE config with time-limited TURN credentials
	iceConfig := h.turnService.GetICEConfig(networkID)

	c.JSON(http.StatusOK, iceConfig)
}

// RegisterICEConfigRoutes registers ICE configuration routes
func RegisterICEConfigRoutes(r *gin.Engine, handler *ICEConfigHandler, authService TokenValidator) {
	networks := r.Group("/v1/networks")
	networks.Use(AuthMiddleware(authService))
	{
		networks.GET("/:id/ice-config", handler.GetICEConfig)
	}
}
