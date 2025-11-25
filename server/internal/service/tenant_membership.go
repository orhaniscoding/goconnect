package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"golang.org/x/crypto/argon2"
)

// TenantMembershipService handles tenant membership operations
type TenantMembershipService struct {
	memberRepo       repository.TenantMemberRepository
	inviteRepo       repository.TenantInviteRepository
	announcementRepo repository.TenantAnnouncementRepository
	chatRepo         repository.TenantChatRepository
	tenantRepo       repository.TenantRepository
	userRepo         repository.UserRepository
}

// NewTenantMembershipService creates a new tenant membership service
func NewTenantMembershipService(
	memberRepo repository.TenantMemberRepository,
	inviteRepo repository.TenantInviteRepository,
	announcementRepo repository.TenantAnnouncementRepository,
	chatRepo repository.TenantChatRepository,
	tenantRepo repository.TenantRepository,
	userRepo repository.UserRepository,
) *TenantMembershipService {
	return &TenantMembershipService{
		memberRepo:       memberRepo,
		inviteRepo:       inviteRepo,
		announcementRepo: announcementRepo,
		chatRepo:         chatRepo,
		tenantRepo:       tenantRepo,
		userRepo:         userRepo,
	}
}

// ==================== TENANT OPERATIONS ====================

// CreateTenant creates a new tenant with the user as owner
func (s *TenantMembershipService) CreateTenant(ctx context.Context, userID string, req *domain.CreateTenantRequest) (*domain.TenantExtended, error) {
	// Generate ID
	tenantID := uuid.New().String()
	visibility := req.Visibility
	if visibility == "" {
		visibility = domain.TenantVisibilityPrivate
	}
	accessType := req.AccessType
	if accessType == "" {
		accessType = domain.TenantAccessInviteOnly
	}
	var passwordHash string
	if accessType == domain.TenantAccessPassword {
		if strings.TrimSpace(req.Password) == "" {
			return nil, domain.NewError(domain.ErrValidation, "Password required for password-protected tenants", map[string]string{"field": "password"})
		}
		passwordHash = HashTenantPassword(req.Password)
	}
	now := time.Now()
	tenant := &domain.Tenant{
		ID:           tenantID,
		Name:         req.Name,
		Description:  req.Description,
		Visibility:   visibility,
		AccessType:   accessType,
		PasswordHash: passwordHash,
		MaxMembers:   req.MaxMembers,
		OwnerID:      userID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Create tenant in repository
	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Add creator as owner
	member := &domain.TenantMember{
		TenantID: tenantID,
		UserID:   userID,
		Role:     domain.TenantRoleOwner,
	}
	if err := s.memberRepo.Create(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to add owner as member: %w", err)
	}

	return &domain.TenantExtended{
		Tenant:      *tenant,
		MemberCount: 1,
	}, nil
}

// GetTenant gets a tenant by ID with member count
func (s *TenantMembershipService) GetTenant(ctx context.Context, tenantID string) (*domain.TenantExtended, error) {
	basicTenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	memberCount, err := s.memberRepo.CountByTenant(ctx, tenantID)
	if err != nil {
		memberCount = 0 // Non-critical, continue
	}

	return &domain.TenantExtended{
		Tenant:      *basicTenant,
		MemberCount: memberCount,
	}, nil
}

// UpdateTenant updates tenant settings (owner/admin only)
func (s *TenantMembershipService) UpdateTenant(ctx context.Context, userID, tenantID string, req *domain.UpdateTenantRequest) (*domain.TenantExtended, error) {
	// Check if user has permission (owner or admin)
	member, err := s.memberRepo.GetByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return nil, domain.NewError(domain.ErrForbidden, "You are not a member of this tenant", nil)
	}

	if member.Role != domain.TenantRoleOwner && member.Role != domain.TenantRoleAdmin {
		return nil, domain.NewError(domain.ErrForbidden, "Only owner or admin can update tenant settings", nil)
	}

	// Get current tenant
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
		tenant.Name = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		tenant.Description = strings.TrimSpace(*req.Description)
	}
	if req.Visibility != nil {
		tenant.Visibility = *req.Visibility
	}
	if req.AccessType != nil {
		tenant.AccessType = *req.AccessType
		// If changing to password access, require password
		if *req.AccessType == domain.TenantAccessPassword {
			if req.Password == nil || strings.TrimSpace(*req.Password) == "" {
				return nil, domain.NewError(domain.ErrValidation, "Password required for password-protected access", map[string]string{"field": "password"})
			}
		}
	}
	if req.Password != nil && strings.TrimSpace(*req.Password) != "" {
		tenant.PasswordHash = HashTenantPassword(*req.Password)
	}
	if req.MaxMembers != nil {
		tenant.MaxMembers = *req.MaxMembers
	}

	tenant.UpdatedAt = time.Now()

	// Save updates
	if err := s.tenantRepo.Update(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	// Get member count for response
	memberCount, _ := s.memberRepo.CountByTenant(ctx, tenantID)

	return &domain.TenantExtended{
		Tenant:      *tenant,
		MemberCount: memberCount,
	}, nil
}

// ListPublicTenants returns discoverable tenants for unauthenticated discovery/search flows
func (s *TenantMembershipService) ListPublicTenants(ctx context.Context, req *domain.ListTenantsRequest) ([]*domain.TenantExtended, string, error) {
	if req == nil {
		req = &domain.ListTenantsRequest{}
	}

	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset := 0
	if cursor := strings.TrimSpace(req.Cursor); cursor != "" {
		parsed, err := strconv.Atoi(cursor)
		if err != nil || parsed < 0 {
			return nil, "", domain.NewError(domain.ErrInvalidRequest, "Invalid cursor value", map[string]string{"cursor": req.Cursor})
		}
		offset = parsed
	}

	// Fetch tenants using existing repository pagination (order by created_at DESC)
	tenants, total, err := s.tenantRepo.ListAll(ctx, limit, offset, req.Search)
	if err != nil {
		return nil, "", err
	}

	var result []*domain.TenantExtended
	for _, tenant := range tenants {
		if tenant.Visibility != domain.TenantVisibilityPublic {
			continue
		}

		memberCount, err := s.memberRepo.CountByTenant(ctx, tenant.ID)
		if err != nil {
			return nil, "", err
		}

		result = append(result, &domain.TenantExtended{
			Tenant:      *tenant,
			MemberCount: memberCount,
		})
	}

	nextCursor := ""
	if offset+len(tenants) < total {
		nextCursor = strconv.Itoa(offset + len(tenants))
	}

	return result, nextCursor, nil
}

// ==================== MEMBERSHIP OPERATIONS ====================

// JoinTenant allows a user to join a tenant
func (s *TenantMembershipService) JoinTenant(ctx context.Context, userID, tenantID string, req *domain.JoinTenantRequest) (*domain.TenantMember, error) {
	// Check if already a member
	existing, _ := s.memberRepo.GetByUserAndTenant(ctx, userID, tenantID)
	if existing != nil {
		return nil, domain.NewError(domain.ErrAlreadyMember, "You are already a member of this tenant", nil)
	}

	// Get tenant to check access type
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	if tenant.AccessType == domain.TenantAccessInviteOnly {
		return nil, domain.NewError(domain.ErrForbidden, "This tenant requires an invite code to join", nil)
	}

	if tenant.AccessType == domain.TenantAccessPassword {
		password := ""
		if req != nil {
			password = req.Password
		}
		if strings.TrimSpace(password) == "" {
			return nil, domain.NewError(domain.ErrInvalidRequest, "Password required to join this tenant", map[string]string{"field": "password"})
		}
		if tenant.PasswordHash == "" || HashTenantPassword(password) != tenant.PasswordHash {
			return nil, domain.NewError(domain.ErrInvalidCredentials, "Invalid tenant password", nil)
		}
	}

	if err := s.ensureTenantCapacity(ctx, tenant); err != nil {
		return nil, err
	}

	// Create membership
	member := &domain.TenantMember{
		TenantID: tenantID,
		UserID:   userID,
		Role:     domain.TenantRoleMember,
	}

	if err := s.memberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	return member, nil
}

// JoinByCode allows a user to join a tenant via invite code
func (s *TenantMembershipService) JoinByCode(ctx context.Context, userID string, code string) (*domain.TenantMember, error) {
	// Find invite by code
	invite, err := s.inviteRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// Validate invite
	if !invite.IsValid() {
		if invite.RevokedAt != nil {
			return nil, domain.NewError(domain.ErrInviteTokenRevoked, "This invite code has been revoked", nil)
		}
		return nil, domain.NewError(domain.ErrInviteTokenExpired, "This invite code has expired", nil)
	}

	// Check if already a member
	existing, _ := s.memberRepo.GetByUserAndTenant(ctx, userID, invite.TenantID)
	if existing != nil {
		return nil, domain.NewError(domain.ErrAlreadyMember, "You are already a member of this tenant", nil)
	}

	tenant, err := s.tenantRepo.GetByID(ctx, invite.TenantID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureTenantCapacity(ctx, tenant); err != nil {
		return nil, err
	}

	// Increment use count
	if err := s.inviteRepo.IncrementUseCount(ctx, invite.ID); err != nil {
		return nil, err
	}

	// Create membership
	member := &domain.TenantMember{
		TenantID: invite.TenantID,
		UserID:   userID,
		Role:     domain.TenantRoleMember,
	}

	if err := s.memberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	return member, nil
}

// LeaveTenant allows a user to leave a tenant
func (s *TenantMembershipService) LeaveTenant(ctx context.Context, userID, tenantID string) error {
	member, err := s.memberRepo.GetByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return err
	}

	// Owner cannot leave (must transfer ownership first)
	if member.Role == domain.TenantRoleOwner {
		return domain.NewError(domain.ErrForbidden, "Owner cannot leave tenant. Transfer ownership first.", nil)
	}

	return s.memberRepo.Delete(ctx, member.ID)
}

// GetUserTenants returns all tenants a user is a member of
func (s *TenantMembershipService) GetUserTenants(ctx context.Context, userID string) ([]*domain.TenantMember, error) {
	return s.memberRepo.ListByUser(ctx, userID)
}

// GetTenantMembers returns all members of a tenant
func (s *TenantMembershipService) GetTenantMembers(ctx context.Context, tenantID string, req *domain.ListTenantMembersRequest) ([]*domain.TenantMember, string, error) {
	return s.memberRepo.ListByTenant(ctx, tenantID, req.Role, req.Limit, req.Cursor)
}

// UpdateMemberRole updates a member's role (requires admin+ permission)
func (s *TenantMembershipService) UpdateMemberRole(ctx context.Context, actorID, tenantID, targetMemberID string, newRole domain.TenantRole) error {
	// Get actor's role
	actorRole, err := s.memberRepo.GetUserRole(ctx, actorID, tenantID)
	if err != nil {
		return domain.NewError(domain.ErrForbidden, "You are not a member of this tenant", nil)
	}

	// Get target member
	targetMember, err := s.memberRepo.GetByID(ctx, targetMemberID)
	if err != nil {
		return err
	}

	// Verify target is in the same tenant
	if targetMember.TenantID != tenantID {
		return domain.NewError(domain.ErrNotFound, "Member not found in this tenant", nil)
	}

	// Permission checks
	// 1. Cannot change your own role (except owner -> admin for transfer)
	if targetMember.UserID == actorID && actorRole != domain.TenantRoleOwner {
		return domain.NewError(domain.ErrForbidden, "You cannot change your own role", nil)
	}

	// 2. Only owner can set admin role
	if newRole == domain.TenantRoleAdmin && actorRole != domain.TenantRoleOwner {
		return domain.NewError(domain.ErrForbidden, "Only owner can promote to admin", nil)
	}

	// 3. Cannot demote someone with higher/equal role
	if !actorRole.HasPermission(targetMember.Role) {
		return domain.NewError(domain.ErrForbidden, "Cannot modify member with higher or equal role", nil)
	}

	// 4. Cannot set role higher than your own (except owner)
	if actorRole != domain.TenantRoleOwner && !actorRole.HasPermission(newRole) {
		return domain.NewError(domain.ErrForbidden, "Cannot set role higher than your own", nil)
	}

	// 5. Cannot set owner role via this method
	if newRole == domain.TenantRoleOwner {
		return domain.NewError(domain.ErrForbidden, "Use transfer ownership to change owner", nil)
	}

	return s.memberRepo.UpdateRole(ctx, targetMemberID, newRole)
}

// RemoveMember removes a member from a tenant (kick)
func (s *TenantMembershipService) RemoveMember(ctx context.Context, actorID, tenantID, targetMemberID string) error {
	// Get actor's role
	actorRole, err := s.memberRepo.GetUserRole(ctx, actorID, tenantID)
	if err != nil {
		return domain.NewError(domain.ErrForbidden, "You are not a member of this tenant", nil)
	}

	// Need at least moderator to kick
	if !actorRole.HasPermission(domain.TenantRoleModerator) {
		return domain.NewError(domain.ErrForbidden, "You need moderator permission to remove members", nil)
	}

	// Get target member
	targetMember, err := s.memberRepo.GetByID(ctx, targetMemberID)
	if err != nil {
		return err
	}

	// Verify same tenant
	if targetMember.TenantID != tenantID {
		return domain.NewError(domain.ErrNotFound, "Member not found in this tenant", nil)
	}

	// Cannot kick yourself
	if targetMember.UserID == actorID {
		return domain.NewError(domain.ErrForbidden, "Use leave instead of removing yourself", nil)
	}

	// Cannot kick owner
	if targetMember.Role == domain.TenantRoleOwner {
		return domain.NewError(domain.ErrForbidden, "Cannot remove tenant owner", nil)
	}

	// Cannot kick someone with higher/equal role
	if !actorRole.HasPermission(targetMember.Role) || actorRole == targetMember.Role {
		return domain.NewError(domain.ErrForbidden, "Cannot remove member with higher or equal role", nil)
	}

	return s.memberRepo.Delete(ctx, targetMemberID)
}

func (s *TenantMembershipService) ensureTenantCapacity(ctx context.Context, tenant *domain.Tenant) error {
	if tenant == nil || tenant.MaxMembers <= 0 {
		return nil
	}

	count, err := s.memberRepo.CountByTenant(ctx, tenant.ID)
	if err != nil {
		return err
	}
	if count >= tenant.MaxMembers {
		return domain.NewError(domain.ErrForbidden, fmt.Sprintf("Tenant has reached the maximum of %d members", tenant.MaxMembers), map[string]int{"max_members": tenant.MaxMembers})
	}
	return nil
}

// ==================== INVITE OPERATIONS ====================

// CreateInvite creates a new tenant invite code
func (s *TenantMembershipService) CreateInvite(ctx context.Context, userID, tenantID string, req *domain.CreateTenantInviteRequest) (*domain.TenantInvite, error) {
	// Check permission (need admin+)
	hasRole, err := s.memberRepo.HasRole(ctx, userID, tenantID, domain.TenantRoleAdmin)
	if err != nil {
		return nil, err
	}
	if !hasRole {
		return nil, domain.NewError(domain.ErrForbidden, "You need admin permission to create invites", nil)
	}

	// Generate code
	code, err := domain.GenerateTenantInviteCode()
	if err != nil {
		return nil, err
	}

	invite := &domain.TenantInvite{
		TenantID:  tenantID,
		Code:      code,
		MaxUses:   req.MaxUses,
		CreatedBy: userID,
	}

	if req.ExpiresIn > 0 {
		expiresAt := time.Now().Add(time.Duration(req.ExpiresIn) * time.Second)
		invite.ExpiresAt = &expiresAt
	}

	if err := s.inviteRepo.Create(ctx, invite); err != nil {
		return nil, err
	}

	return invite, nil
}

// ListInvites returns all invites for a tenant
func (s *TenantMembershipService) ListInvites(ctx context.Context, userID, tenantID string) ([]*domain.TenantInvite, error) {
	// Check permission (need admin+)
	hasRole, err := s.memberRepo.HasRole(ctx, userID, tenantID, domain.TenantRoleAdmin)
	if err != nil {
		return nil, err
	}
	if !hasRole {
		return nil, domain.NewError(domain.ErrForbidden, "You need admin permission to view invites", nil)
	}

	return s.inviteRepo.ListByTenant(ctx, tenantID)
}

// RevokeInvite revokes an invite code
func (s *TenantMembershipService) RevokeInvite(ctx context.Context, userID, tenantID, inviteID string) error {
	// Check permission
	hasRole, err := s.memberRepo.HasRole(ctx, userID, tenantID, domain.TenantRoleAdmin)
	if err != nil {
		return err
	}
	if !hasRole {
		return domain.NewError(domain.ErrForbidden, "You need admin permission to revoke invites", nil)
	}

	// Verify invite belongs to tenant
	invite, err := s.inviteRepo.GetByID(ctx, inviteID)
	if err != nil {
		return err
	}
	if invite.TenantID != tenantID {
		return domain.NewError(domain.ErrNotFound, "Invite not found in this tenant", nil)
	}

	return s.inviteRepo.Revoke(ctx, inviteID)
}

// ==================== ANNOUNCEMENT OPERATIONS ====================

// CreateAnnouncement creates a new announcement
func (s *TenantMembershipService) CreateAnnouncement(ctx context.Context, userID, tenantID string, req *domain.CreateAnnouncementRequest) (*domain.TenantAnnouncement, error) {
	// Check permission (need moderator+)
	hasRole, err := s.memberRepo.HasRole(ctx, userID, tenantID, domain.TenantRoleModerator)
	if err != nil {
		return nil, err
	}
	if !hasRole {
		return nil, domain.NewError(domain.ErrForbidden, "You need moderator permission to create announcements", nil)
	}

	announcement := &domain.TenantAnnouncement{
		TenantID: tenantID,
		Title:    req.Title,
		Content:  req.Content,
		AuthorID: userID,
		IsPinned: req.IsPinned,
	}

	if err := s.announcementRepo.Create(ctx, announcement); err != nil {
		return nil, err
	}

	return announcement, nil
}

// GetAnnouncements returns announcements for a tenant
func (s *TenantMembershipService) GetAnnouncements(ctx context.Context, userID, tenantID string, req *domain.ListAnnouncementsRequest) ([]*domain.TenantAnnouncement, string, error) {
	// Check membership (any member can view)
	_, err := s.memberRepo.GetByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return nil, "", domain.NewError(domain.ErrForbidden, "You are not a member of this tenant", nil)
	}

	pinnedOnly := req.Pinned != nil && *req.Pinned
	return s.announcementRepo.ListByTenant(ctx, tenantID, pinnedOnly, req.Limit, req.Cursor)
}

// UpdateAnnouncement updates an announcement
func (s *TenantMembershipService) UpdateAnnouncement(ctx context.Context, userID, tenantID, announcementID string, req *domain.UpdateAnnouncementRequest) error {
	// Check permission
	hasRole, err := s.memberRepo.HasRole(ctx, userID, tenantID, domain.TenantRoleModerator)
	if err != nil {
		return err
	}
	if !hasRole {
		return domain.NewError(domain.ErrForbidden, "You need moderator permission to update announcements", nil)
	}

	ann, err := s.announcementRepo.GetByID(ctx, announcementID)
	if err != nil {
		return err
	}
	if ann.TenantID != tenantID {
		return domain.NewError(domain.ErrNotFound, "Announcement not found in this tenant", nil)
	}

	// Apply updates
	if req.Title != nil {
		ann.Title = *req.Title
	}
	if req.Content != nil {
		ann.Content = *req.Content
	}
	if req.IsPinned != nil {
		ann.IsPinned = *req.IsPinned
	}

	return s.announcementRepo.Update(ctx, ann)
}

// DeleteAnnouncement deletes an announcement
func (s *TenantMembershipService) DeleteAnnouncement(ctx context.Context, userID, tenantID, announcementID string) error {
	// Check permission
	hasRole, err := s.memberRepo.HasRole(ctx, userID, tenantID, domain.TenantRoleModerator)
	if err != nil {
		return err
	}
	if !hasRole {
		return domain.NewError(domain.ErrForbidden, "You need moderator permission to delete announcements", nil)
	}

	// Verify announcement belongs to tenant
	ann, err := s.announcementRepo.GetByID(ctx, announcementID)
	if err != nil {
		return err
	}
	if ann.TenantID != tenantID {
		return domain.NewError(domain.ErrNotFound, "Announcement not found in this tenant", nil)
	}

	return s.announcementRepo.Delete(ctx, announcementID)
}

// ==================== CHAT OPERATIONS ====================

// SendChatMessage sends a message to tenant chat
func (s *TenantMembershipService) SendChatMessage(ctx context.Context, userID, tenantID string, req *domain.SendChatMessageRequest) (*domain.TenantChatMessage, error) {
	// Check membership
	_, err := s.memberRepo.GetByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return nil, domain.NewError(domain.ErrForbidden, "You are not a member of this tenant", nil)
	}

	message := &domain.TenantChatMessage{
		TenantID: tenantID,
		UserID:   userID,
		Content:  req.Content,
	}

	if err := s.chatRepo.Create(ctx, message); err != nil {
		return nil, err
	}

	return message, nil
}

// GetChatHistory returns chat history for a tenant
func (s *TenantMembershipService) GetChatHistory(ctx context.Context, userID, tenantID string, req *domain.ListChatMessagesRequest) ([]*domain.TenantChatMessage, error) {
	// Check membership
	_, err := s.memberRepo.GetByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return nil, domain.NewError(domain.ErrForbidden, "You are not a member of this tenant", nil)
	}

	return s.chatRepo.ListByTenant(ctx, tenantID, req.Before, req.Limit)
}

// DeleteChatMessage deletes a chat message
func (s *TenantMembershipService) DeleteChatMessage(ctx context.Context, userID, tenantID, messageID string) error {
	// Get message
	msg, err := s.chatRepo.GetByID(ctx, messageID)
	if err != nil {
		return err
	}
	if msg.TenantID != tenantID {
		return domain.NewError(domain.ErrNotFound, "Message not found in this tenant", nil)
	}

	// Check permission: author can delete own message, moderator+ can delete any
	if msg.UserID != userID {
		hasRole, err := s.memberRepo.HasRole(ctx, userID, tenantID, domain.TenantRoleModerator)
		if err != nil {
			return err
		}
		if !hasRole {
			return domain.NewError(domain.ErrForbidden, "You can only delete your own messages", nil)
		}
	}

	return s.chatRepo.SoftDelete(ctx, messageID)
}

// ==================== PERMISSION HELPERS ====================

// CheckTenantPermission checks if user has required role in tenant
func (s *TenantMembershipService) CheckTenantPermission(ctx context.Context, userID, tenantID string, requiredRole domain.TenantRole) error {
	hasRole, err := s.memberRepo.HasRole(ctx, userID, tenantID, requiredRole)
	if err != nil {
		return err
	}
	if !hasRole {
		return domain.NewError(domain.ErrForbidden, fmt.Sprintf("You need %s permission for this action", requiredRole), nil)
	}
	return nil
}

// HashTenantPassword hashes password for tenant access
func HashTenantPassword(password string) string {
	salt := []byte("tenant-password-salt") // In production, use random salt per tenant
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return fmt.Sprintf("%x", hash)
}
