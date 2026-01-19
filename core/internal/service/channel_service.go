package service

import (
	"context"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

// ══════════════════════════════════════════════════════════════════════════════
// CHANNEL SERVICE
// ══════════════════════════════════════════════════════════════════════════════
// Business logic layer for channel management with RBAC integration

type ChannelService struct {
	channelRepo        repository.ChannelRepository
	sectionRepo        repository.SectionRepository
	permissionResolver *PermissionResolver
}

// NewChannelService creates a new channel service
func NewChannelService(
	channelRepo repository.ChannelRepository,
	sectionRepo repository.SectionRepository,
	permissionResolver *PermissionResolver,
) *ChannelService {
	return &ChannelService{
		channelRepo:        channelRepo,
		sectionRepo:        sectionRepo,
		permissionResolver: permissionResolver,
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// CREATE
// ══════════════════════════════════════════════════════════════════════════════

// CreateChannelInput represents input for creating a channel
type CreateChannelInput struct {
	UserID      string
	TenantID    *string // One of TenantID, SectionID, or NetworkID must be set
	SectionID   *string
	NetworkID   *string
	Name        string
	Description string
	Type        domain.ChannelType
	Bitrate     int
	UserLimit   int
	Slowmode    int
	NSFW        bool
}

// Create creates a new channel
// Permissions: Requires MANAGE_CHANNELS permission
func (s *ChannelService) Create(ctx context.Context, input CreateChannelInput) (*domain.Channel, error) {
	// Step 1: Validate parent (exactly one must be set)
	parentCount := 0
	var tenantID string

	if input.TenantID != nil {
		parentCount++
		tenantID = *input.TenantID
	}
	if input.SectionID != nil {
		parentCount++
		// Get section to extract tenant ID
		section, err := s.sectionRepo.GetByID(ctx, *input.SectionID)
		if err != nil {
			return nil, fmt.Errorf("failed to get section: %w", err)
		}
		tenantID = section.TenantID
	}
	if input.NetworkID != nil {
		parentCount++
		// Network channels require network context (to be implemented)
		return nil, domain.NewError(domain.ErrValidation, "network channels not yet supported", nil)
	}

	if parentCount != 1 {
		return nil, domain.NewError(domain.ErrValidation, "exactly one parent (tenant, section, or network) must be specified", map[string]any{
			"parent_count": parentCount,
		})
	}

	// Step 2: Check permissions
	permResult, err := s.permissionResolver.CheckPermission(
		ctx,
		input.UserID,
		tenantID,
		"",
		domain.PermissionManageChannels,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}

	if !permResult.Allowed {
		return nil, domain.NewError(domain.ErrForbidden, "insufficient permissions to create channel", map[string]any{
			"user_id":    input.UserID,
			"tenant_id":  tenantID,
			"permission": domain.PermissionManageChannels,
			"reason":     permResult.Reason,
		})
	}

	// Step 3: Validate input
	validatedName, err := domain.ValidateChannelName(input.Name)
	if err != nil {
		return nil, err
	}

	if len(input.Description) > 500 {
		return nil, domain.NewError(domain.ErrValidation, "description too long", map[string]any{
			"field": "description",
			"max":   500,
		})
	}

	// Step 4: Validate channel type
	switch input.Type {
	case domain.ChannelTypeText, domain.ChannelTypeVoice, domain.ChannelTypeAnnouncement:
		// Valid
	default:
		return nil, domain.NewError(domain.ErrValidation, "invalid channel type", map[string]any{
			"type": input.Type,
		})
	}

	// Step 5: Apply defaults and validate voice settings
	if input.Type == domain.ChannelTypeVoice {
		if input.Bitrate == 0 {
			input.Bitrate = 64000 // 64kbps default
		}
		if input.Bitrate < 8000 || input.Bitrate > 384000 {
			return nil, domain.NewError(domain.ErrValidation, "bitrate out of range", map[string]any{
				"bitrate": input.Bitrate,
				"min":     8000,
				"max":     384000,
			})
		}
		if input.UserLimit < 0 || input.UserLimit > 99 {
			return nil, domain.NewError(domain.ErrValidation, "user limit out of range", map[string]any{
				"user_limit": input.UserLimit,
				"min":        0,
				"max":        99,
			})
		}
	}

	// Step 6: Validate slowmode
	if input.Slowmode < 0 || input.Slowmode > 21600 {
		return nil, domain.NewError(domain.ErrValidation, "slowmode out of range", map[string]any{
			"slowmode": input.Slowmode,
			"min":      0,
			"max":      21600,
		})
	}

	// Step 7: Get existing channels to determine position
	filter := repository.ChannelFilter{
		TenantID:  input.TenantID,
		SectionID: input.SectionID,
		NetworkID: input.NetworkID,
		Limit:     100, // Get all channels in parent
	}

	existingChannels, _, err := s.channelRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing channels: %w", err)
	}

	// New channel goes to the end
	position := len(existingChannels)

	// Step 8: Create channel entity
	channel := &domain.Channel{
		ID:          domain.GenerateChannelID(),
		TenantID:    input.TenantID,
		SectionID:   input.SectionID,
		NetworkID:   input.NetworkID,
		Name:        validatedName,
		Description: input.Description,
		Type:        input.Type,
		Position:    position,
		Bitrate:     input.Bitrate,
		UserLimit:   input.UserLimit,
		Slowmode:    input.Slowmode,
		NSFW:        input.NSFW,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Step 9: Persist to database
	if err := s.channelRepo.Create(ctx, channel); err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	return channel, nil
}

// ══════════════════════════════════════════════════════════════════════════════
// READ
// ══════════════════════════════════════════════════════════════════════════════

// GetByID retrieves a channel by ID
// Permissions: Requires VIEW_CHANNELS permission (with channel-specific overrides)
func (s *ChannelService) GetByID(ctx context.Context, userID, channelID string) (*domain.Channel, error) {
	// Get channel
	channel, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return nil, err
	}

	// Determine tenant ID from channel's parent
	var tenantID string
	if channel.TenantID != nil {
		tenantID = *channel.TenantID
	} else if channel.SectionID != nil {
		section, err := s.sectionRepo.GetByID(ctx, *channel.SectionID)
		if err != nil {
			return nil, fmt.Errorf("failed to get section: %w", err)
		}
		tenantID = section.TenantID
	} else {
		// Network channel (not yet implemented)
		return nil, domain.NewError(domain.ErrValidation, "network channels not yet supported", nil)
	}

	// Check permission (with channel-specific overrides)
	permResult, err := s.permissionResolver.CheckPermission(
		ctx,
		userID,
		tenantID,
		channelID,
		domain.PermissionViewChannels,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}

	if !permResult.Allowed {
		return nil, domain.NewError(domain.ErrForbidden, "insufficient permissions to view channel", map[string]any{
			"user_id":    userID,
			"channel_id": channelID,
			"reason":     permResult.Reason,
		})
	}

	return channel, nil
}

// ListChannelsInput represents input for listing channels
type ListChannelsInput struct {
	UserID    string
	TenantID  *string
	SectionID *string
	NetworkID *string
	Type      *domain.ChannelType
	Limit     int
	Cursor    string
}

// List retrieves channels with filters
// Permissions: Requires VIEW_CHANNELS permission
func (s *ChannelService) List(ctx context.Context, input ListChannelsInput) ([]domain.Channel, string, error) {
	// Determine tenant ID
	var tenantID string
	if input.TenantID != nil {
		tenantID = *input.TenantID
	} else if input.SectionID != nil {
		section, err := s.sectionRepo.GetByID(ctx, *input.SectionID)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get section: %w", err)
		}
		tenantID = section.TenantID
	} else {
		return nil, "", domain.NewError(domain.ErrValidation, "tenant or section ID required", nil)
	}

	// Check permissions
	permResult, err := s.permissionResolver.CheckPermission(
		ctx,
		input.UserID,
		tenantID,
		"",
		domain.PermissionViewChannels,
	)
	if err != nil {
		return nil, "", fmt.Errorf("failed to check permissions: %w", err)
	}

	if !permResult.Allowed {
		return nil, "", domain.NewError(domain.ErrForbidden, "insufficient permissions to list channels", map[string]any{
			"user_id":   input.UserID,
			"tenant_id": tenantID,
		})
	}

	// Build filter
	filter := repository.ChannelFilter{
		TenantID:  input.TenantID,
		SectionID: input.SectionID,
		NetworkID: input.NetworkID,
		Type:      input.Type,
		Limit:     input.Limit,
		Cursor:    input.Cursor,
	}

	// Get channels
	channels, nextCursor, err := s.channelRepo.List(ctx, filter)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list channels: %w", err)
	}

	// Filter channels based on per-channel VIEW_CHANNELS permission overrides
	// Even if user has VIEW_CHANNELS at tenant level, individual channels may deny access
	filteredChannels := make([]domain.Channel, 0, len(channels))
	for _, channel := range channels {
		// Handle pointer TenantID
		tenantID := ""
		if channel.TenantID != nil {
			tenantID = *channel.TenantID
		}

		// Check per-channel permission
		channelPermResult, err := s.permissionResolver.CheckPermission(
			ctx,
			input.UserID,
			tenantID,
			channel.ID,
			domain.PermissionViewChannels,
		)
		if err != nil {
			// Log error but continue - skip channels where permission check fails
			continue
		}

		// Only include channels user can view
		if channelPermResult.Allowed {
			filteredChannels = append(filteredChannels, channel)
		}
	}

	// Adjust cursor if filtering reduced results below limit
	// If we filtered out all channels and there might be more, keep the original cursor
	if len(filteredChannels) == 0 && len(channels) > 0 && nextCursor != "" {
		// No visible channels in this page, but there might be more
		return filteredChannels, nextCursor, nil
	}

	return filteredChannels, nextCursor, nil
}

// ══════════════════════════════════════════════════════════════════════════════
// UPDATE
// ══════════════════════════════════════════════════════════════════════════════

// UpdateChannelInput represents input for updating a channel
type UpdateChannelInput struct {
	UserID      string
	ChannelID   string
	Name        *string
	Description *string
	Bitrate     *int
	UserLimit   *int
	Slowmode    *int
	NSFW        *bool
}

// Update updates a channel
// Permissions: Requires MANAGE_CHANNELS permission
func (s *ChannelService) Update(ctx context.Context, input UpdateChannelInput) (*domain.Channel, error) {
	// Step 1: Get existing channel
	channel, err := s.channelRepo.GetByID(ctx, input.ChannelID)
	if err != nil {
		return nil, err
	}

	// Step 2: Determine tenant ID
	var tenantID string
	if channel.TenantID != nil {
		tenantID = *channel.TenantID
	} else if channel.SectionID != nil {
		section, err := s.sectionRepo.GetByID(ctx, *channel.SectionID)
		if err != nil {
			return nil, fmt.Errorf("failed to get section: %w", err)
		}
		tenantID = section.TenantID
	}

	// Step 3: Check permissions
	permResult, err := s.permissionResolver.CheckPermission(
		ctx,
		input.UserID,
		tenantID,
		"",
		domain.PermissionManageChannels,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}

	if !permResult.Allowed {
		return nil, domain.NewError(domain.ErrForbidden, "insufficient permissions to update channel", map[string]any{
			"user_id":    input.UserID,
			"channel_id": input.ChannelID,
			"reason":     permResult.Reason,
		})
	}

	// Step 4: Apply updates
	if input.Name != nil {
		validatedName, err := domain.ValidateChannelName(*input.Name)
		if err != nil {
			return nil, err
		}
		channel.Name = validatedName
	}

	if input.Description != nil {
		if len(*input.Description) > 500 {
			return nil, domain.NewError(domain.ErrValidation, "description too long", map[string]any{
				"field": "description",
				"max":   500,
			})
		}
		channel.Description = *input.Description
	}

	if input.Bitrate != nil {
		if *input.Bitrate < 8000 || *input.Bitrate > 384000 {
			return nil, domain.NewError(domain.ErrValidation, "bitrate out of range", map[string]any{
				"bitrate": *input.Bitrate,
				"min":     8000,
				"max":     384000,
			})
		}
		channel.Bitrate = *input.Bitrate
	}

	if input.UserLimit != nil {
		if *input.UserLimit < 0 || *input.UserLimit > 99 {
			return nil, domain.NewError(domain.ErrValidation, "user limit out of range", map[string]any{
				"user_limit": *input.UserLimit,
				"min":        0,
				"max":        99,
			})
		}
		channel.UserLimit = *input.UserLimit
	}

	if input.Slowmode != nil {
		if *input.Slowmode < 0 || *input.Slowmode > 21600 {
			return nil, domain.NewError(domain.ErrValidation, "slowmode out of range", map[string]any{
				"slowmode": *input.Slowmode,
				"min":      0,
				"max":      21600,
			})
		}
		channel.Slowmode = *input.Slowmode
	}

	if input.NSFW != nil {
		channel.NSFW = *input.NSFW
	}

	channel.UpdatedAt = time.Now()

	// Step 5: Persist changes
	if err := s.channelRepo.Update(ctx, channel); err != nil {
		return nil, fmt.Errorf("failed to update channel: %w", err)
	}

	return channel, nil
}

// ══════════════════════════════════════════════════════════════════════════════
// DELETE
// ══════════════════════════════════════════════════════════════════════════════

// Delete deletes a channel
// Permissions: Requires MANAGE_CHANNELS permission
func (s *ChannelService) Delete(ctx context.Context, userID, channelID string) error {
	// Step 1: Get existing channel
	channel, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return err
	}

	// Step 2: Determine tenant ID
	var tenantID string
	if channel.TenantID != nil {
		tenantID = *channel.TenantID
	} else if channel.SectionID != nil {
		section, err := s.sectionRepo.GetByID(ctx, *channel.SectionID)
		if err != nil {
			return fmt.Errorf("failed to get section: %w", err)
		}
		tenantID = section.TenantID
	}

	// Step 3: Check permissions
	permResult, err := s.permissionResolver.CheckPermission(
		ctx,
		userID,
		tenantID,
		"",
		domain.PermissionManageChannels,
	)
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}

	if !permResult.Allowed {
		return domain.NewError(domain.ErrForbidden, "insufficient permissions to delete channel", map[string]any{
			"user_id":    userID,
			"channel_id": channelID,
			"reason":     permResult.Reason,
		})
	}

	// Step 4: Delete channel (soft delete)
	if err := s.channelRepo.Delete(ctx, channelID); err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}

	return nil
}

// ══════════════════════════════════════════════════════════════════════════════
// POSITION MANAGEMENT
// ══════════════════════════════════════════════════════════════════════════════

// UpdatePositions updates positions for multiple channels within a parent
// Permissions: Requires MANAGE_CHANNELS permission
func (s *ChannelService) UpdatePositions(ctx context.Context, userID, parentID string, positions map[string]int) error {
	// TODO: Determine tenant ID from parent ID
	// For now, assume parentID is tenantID (simplified)

	// Check permissions
	permResult, err := s.permissionResolver.CheckPermission(
		ctx,
		userID,
		parentID,
		"",
		domain.PermissionManageChannels,
	)
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}

	if !permResult.Allowed {
		return domain.NewError(domain.ErrForbidden, "insufficient permissions to reorder channels", map[string]any{
			"user_id":   userID,
			"parent_id": parentID,
			"reason":    permResult.Reason,
		})
	}

	// Validate positions (must be non-negative)
	for channelID, position := range positions {
		if position < 0 {
			return domain.NewError(domain.ErrValidation, "invalid position", map[string]any{
				"channel_id": channelID,
				"position":   position,
			})
		}
	}

	// Update positions atomically
	if err := s.channelRepo.UpdatePositions(ctx, parentID, positions); err != nil {
		return fmt.Errorf("failed to update positions: %w", err)
	}

	return nil
}
