package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

// GDPRExportData represents the complete user data export
type GDPRExportData struct {
	ExportedAt  time.Time         `json:"exported_at"`
	User        *GDPRUserData     `json:"user"`
	Devices     []GDPRDeviceData  `json:"devices"`
	Memberships []GDPRMemberData  `json:"memberships"`
	Networks    []GDPRNetworkData `json:"networks_owned"`
}

// GDPRUserData represents exported user profile
type GDPRUserData struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Locale       string    `json:"locale,omitempty"`
	TwoFAEnabled bool      `json:"two_fa_enabled"`
	AuthProvider string    `json:"auth_provider"`
	IsAdmin      bool      `json:"is_admin"`
	IsModerator  bool      `json:"is_moderator"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GDPRDeviceData represents exported device info
type GDPRDeviceData struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Platform  string    `json:"platform"`
	PublicKey string    `json:"public_key"`
	IPAddress string    `json:"ip_address,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	LastSeen  time.Time `json:"last_seen"`
	Online    bool      `json:"online"`
}

// GDPRMemberData represents exported membership info
type GDPRMemberData struct {
	NetworkID   string     `json:"network_id"`
	NetworkName string     `json:"network_name"`
	Role        string     `json:"role"`
	Status      string     `json:"status"`
	JoinedAt    *time.Time `json:"joined_at,omitempty"`
}

// GDPRNetworkData represents exported owned network info
type GDPRNetworkData struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	CIDR       string    `json:"cidr"`
	Visibility string    `json:"visibility"`
	JoinPolicy string    `json:"join_policy"`
	CreatedAt  time.Time `json:"created_at"`
}

// GDPRDeleteRequest represents a user deletion request
type GDPRDeleteRequest struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Status      string    `json:"status"` // pending, processing, completed, failed
	RequestedAt time.Time `json:"requested_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	Error       string    `json:"error,omitempty"`
}

// GDPRService handles GDPR/DSR operations
type GDPRService struct {
	userRepo       repository.UserRepository
	deviceRepo     repository.DeviceRepository
	networkRepo    repository.NetworkRepository
	membershipRepo repository.MembershipRepository
}

// NewGDPRService creates a new GDPR service
func NewGDPRService(
	userRepo repository.UserRepository,
	deviceRepo repository.DeviceRepository,
	networkRepo repository.NetworkRepository,
	membershipRepo repository.MembershipRepository,
) *GDPRService {
	return &GDPRService{
		userRepo:       userRepo,
		deviceRepo:     deviceRepo,
		networkRepo:    networkRepo,
		membershipRepo: membershipRepo,
	}
}

// ExportUserData exports all user data for GDPR compliance
func (s *GDPRService) ExportUserData(ctx context.Context, userID, tenantID string) (*GDPRExportData, error) {
	// Get user profile
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, domain.NewError(domain.ErrUserNotFound, "User not found", nil)
	}

	export := &GDPRExportData{
		ExportedAt: time.Now().UTC(),
		User: &GDPRUserData{
			ID:           user.ID,
			Email:        user.Email,
			Locale:       user.Locale,
			TwoFAEnabled: user.TwoFAEnabled,
			AuthProvider: user.AuthProvider,
			IsAdmin:      user.IsAdmin,
			IsModerator:  user.IsModerator,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
		},
		Devices:     []GDPRDeviceData{},
		Memberships: []GDPRMemberData{},
		Networks:    []GDPRNetworkData{},
	}

	// Get devices using filter
	deviceFilter := domain.DeviceFilter{
		UserID: userID,
		Limit:  1000,
	}
	devices, _, err := s.deviceRepo.List(ctx, deviceFilter)
	if err == nil {
		for _, d := range devices {
			export.Devices = append(export.Devices, GDPRDeviceData{
				ID:        d.ID,
				Name:      d.Name,
				Platform:  d.Platform,
				PublicKey: d.PubKey,
				IPAddress: d.IPAddress,
				CreatedAt: d.CreatedAt,
				LastSeen:  d.LastSeen,
				Online:    d.Active,
			})
		}
	}

	// Get all networks to find memberships and owned networks
	// Use IsAdmin=true to bypass visibility filter for GDPR export
	networkFilter := repository.NetworkFilter{
		TenantID:   tenantID,
		IsAdmin:    true,
		Visibility: "all",
		Limit:      1000,
	}
	networks, _, err := s.networkRepo.List(ctx, networkFilter)
	if err == nil {
		for _, n := range networks {
			// Check if user is owner
			if n.CreatedBy == userID {
				export.Networks = append(export.Networks, GDPRNetworkData{
					ID:         n.ID,
					Name:       n.Name,
					CIDR:       n.CIDR,
					Visibility: string(n.Visibility),
					JoinPolicy: string(n.JoinPolicy),
					CreatedAt:  n.CreatedAt,
				})
			}

			// Check membership
			membership, err := s.membershipRepo.Get(ctx, n.ID, userID)
			if err == nil && membership != nil {
				export.Memberships = append(export.Memberships, GDPRMemberData{
					NetworkID:   n.ID,
					NetworkName: n.Name,
					Role:        string(membership.Role),
					Status:      string(membership.Status),
					JoinedAt:    membership.JoinedAt,
				})
			}
		}
	}

	return export, nil
}

// ExportUserDataJSON exports user data as JSON bytes
func (s *GDPRService) ExportUserDataJSON(ctx context.Context, userID, tenantID string) ([]byte, error) {
	data, err := s.ExportUserData(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(data, "", "  ")
}

// RequestDeletion initiates an async user deletion request
func (s *GDPRService) RequestDeletion(ctx context.Context, userID string) (*GDPRDeleteRequest, error) {
	// Verify user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, domain.NewError(domain.ErrUserNotFound, "User not found", nil)
	}

	// Create deletion request
	// In production, this would be stored in a queue table for async processing
	req := &GDPRDeleteRequest{
		ID:          fmt.Sprintf("del_%s_%d", userID, time.Now().Unix()),
		UserID:      userID,
		Status:      "pending",
		RequestedAt: time.Now().UTC(),
	}

	// TODO: Store in deletion_requests table for async worker processing
	// The actual deletion will be handled by a background worker

	return req, nil
}

// ProcessDeletion performs the actual user data deletion
// This should be called by a background worker, not directly by the API
func (s *GDPRService) ProcessDeletion(ctx context.Context, userID, tenantID string) error {
	// 1. Delete all user devices
	deviceFilter := domain.DeviceFilter{UserID: userID, Limit: 1000}
	devices, _, _ := s.deviceRepo.List(ctx, deviceFilter)
	for _, d := range devices {
		if err := s.deviceRepo.Delete(ctx, d.ID); err != nil {
			return fmt.Errorf("failed to delete device %s: %w", d.ID, err)
		}
	}

	// 2. Remove all memberships
	networkFilter := repository.NetworkFilter{TenantID: tenantID, Limit: 1000}
	networks, _, _ := s.networkRepo.List(ctx, networkFilter)
	for _, n := range networks {
		if _, err := s.membershipRepo.Get(ctx, n.ID, userID); err == nil {
			if err := s.membershipRepo.Remove(ctx, n.ID, userID); err != nil {
				return fmt.Errorf("failed to remove membership: %w", err)
			}
		}
	}

	// 3. Delete user record
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
