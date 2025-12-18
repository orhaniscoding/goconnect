package service

import (
"errors"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestTenantMembershipService creates a service with in-memory repositories for testing
func createTestTenantMembershipService() (*TenantMembershipService, *repository.InMemoryTenantRepository) {
	memberRepo := repository.NewInMemoryTenantMemberRepository()
	inviteRepo := repository.NewInMemoryTenantInviteRepository()
	announcementRepo := repository.NewInMemoryTenantAnnouncementRepository()
	chatRepo := repository.NewInMemoryTenantChatRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	userRepo := repository.NewInMemoryUserRepository()

	svc := NewTenantMembershipService(memberRepo, inviteRepo, announcementRepo, chatRepo, tenantRepo, userRepo)
	return svc, tenantRepo
}

// Helper to setup a tenant with owner for tests
func setupTenantWithOwner(ctx context.Context, svc *TenantMembershipService, tenantRepo *repository.InMemoryTenantRepository) (string, string) {
	ownerID := "owner-" + time.Now().Format("150405")

	req := &domain.CreateTenantRequest{
		Name:        "Test Tenant",
		Description: "A test tenant",
		Visibility:  domain.TenantVisibilityPublic,
		AccessType:  domain.TenantAccessOpen,
		MaxMembers:  100,
	}

	tenant, _ := svc.CreateTenant(ctx, ownerID, req)
	return tenant.ID, ownerID
}

// ==================== CREATE TENANT TESTS ====================

func TestTenantMembershipService_CreateTenant_Success(t *testing.T) {
	svc, _ := createTestTenantMembershipService()
	ctx := context.Background()

	req := &domain.CreateTenantRequest{
		Name:        "My Server",
		Description: "A cool VPN server",
		Visibility:  domain.TenantVisibilityPublic,
		AccessType:  domain.TenantAccessOpen,
		MaxMembers:  50,
	}

	tenant, err := svc.CreateTenant(ctx, "user-1", req)

	require.NoError(t, err)
	assert.NotEmpty(t, tenant.ID)
	assert.Equal(t, "My Server", tenant.Name)
	assert.Equal(t, "user-1", tenant.OwnerID)
	assert.Equal(t, 1, tenant.MemberCount)
}

func TestTenantMembershipService_CreateTenant_OwnerIsMember(t *testing.T) {
	svc, _ := createTestTenantMembershipService()
	ctx := context.Background()

	req := &domain.CreateTenantRequest{Name: "Test Server"}
	tenant, _ := svc.CreateTenant(ctx, "user-1", req)

	// Verify owner is a member
	members, _ := svc.GetUserTenants(ctx, "user-1")
	assert.Equal(t, 1, len(members))
	assert.Equal(t, tenant.ID, members[0].TenantID)
	assert.Equal(t, domain.TenantRoleOwner, members[0].Role)
}

func TestTenantMembershipService_CreateTenant_Multiple(t *testing.T) {
	svc, _ := createTestTenantMembershipService()
	ctx := context.Background()

	// Create multiple tenants
	for i := 0; i < 3; i++ {
		req := &domain.CreateTenantRequest{Name: "Server " + string(rune('A'+i))}
		_, err := svc.CreateTenant(ctx, "user-1", req)
		require.NoError(t, err)
	}

	// User should be member of all 3
	members, _ := svc.GetUserTenants(ctx, "user-1")
	assert.Equal(t, 3, len(members))
}

func TestTenantMembershipService_CreateTenant_PasswordRequiresSecret(t *testing.T) {
	svc, _ := createTestTenantMembershipService()
	ctx := context.Background()

	req := &domain.CreateTenantRequest{
		Name:       "Secret Tenant",
		Visibility: domain.TenantVisibilityPrivate,
		AccessType: domain.TenantAccessPassword,
	}

	_, err := svc.CreateTenant(ctx, "owner", req)
	require.Error(t, err)
}

// ==================== GET TENANT TESTS ====================

func TestTenantMembershipService_GetTenant_Success(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, _ := setupTenantWithOwner(ctx, svc, tenantRepo)

	result, err := svc.GetTenant(ctx, tenantID)

	require.NoError(t, err)
	assert.Equal(t, tenantID, result.ID)
	assert.Equal(t, "Test Tenant", result.Name)
	assert.Equal(t, 1, result.MemberCount)
}

func TestTenantMembershipService_GetTenant_NotFound(t *testing.T) {
	svc, _ := createTestTenantMembershipService()
	ctx := context.Background()

	_, err := svc.GetTenant(ctx, "non-existent")

	require.Error(t, err)
}

// ==================== DISCOVERY TESTS ====================

func TestTenantMembershipService_ListPublicTenants_FiltersVisibility(t *testing.T) {
	svc, _ := createTestTenantMembershipService()
	ctx := context.Background()

	// Two public tenants and two that should be hidden
	_, _ = svc.CreateTenant(ctx, "owner-alpha", &domain.CreateTenantRequest{
		Name:       "Alpha Ops",
		Visibility: domain.TenantVisibilityPublic,
		AccessType: domain.TenantAccessOpen,
	})
	_, _ = svc.CreateTenant(ctx, "owner-beta", &domain.CreateTenantRequest{
		Name:       "Beta Labs",
		Visibility: domain.TenantVisibilityPrivate,
		AccessType: domain.TenantAccessInviteOnly,
	})
	_, _ = svc.CreateTenant(ctx, "owner-gamma", &domain.CreateTenantRequest{
		Name:       "Gamma Edge",
		Visibility: domain.TenantVisibilityPublic,
		AccessType: domain.TenantAccessOpen,
	})
	_, _ = svc.CreateTenant(ctx, "owner-delta", &domain.CreateTenantRequest{
		Name:       "Delta Hidden",
		Visibility: domain.TenantVisibilityUnlisted,
		AccessType: domain.TenantAccessOpen,
	})

	results, cursor, err := svc.ListPublicTenants(ctx, &domain.ListTenantsRequest{Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, "", cursor)
	require.Len(t, results, 2)
	nameSet := map[string]bool{}
	for _, tenant := range results {
		nameSet[tenant.Name] = true
		assert.Equal(t, 1, tenant.MemberCount)
	}
	assert.True(t, nameSet["Alpha Ops"])
	assert.True(t, nameSet["Gamma Edge"])
}

func TestTenantMembershipService_ListPublicTenants_SearchAndCursor(t *testing.T) {
	svc, _ := createTestTenantMembershipService()
	ctx := context.Background()

	// Create three public tenants with two matching the search term
	_, _ = svc.CreateTenant(ctx, "owner-alpha", &domain.CreateTenantRequest{
		Name:       "Alpha Core",
		Visibility: domain.TenantVisibilityPublic,
		AccessType: domain.TenantAccessOpen,
	})
	_, _ = svc.CreateTenant(ctx, "owner-beta", &domain.CreateTenantRequest{
		Name:       "Beta Ops",
		Visibility: domain.TenantVisibilityPublic,
		AccessType: domain.TenantAccessOpen,
	})
	_, _ = svc.CreateTenant(ctx, "owner-alpine", &domain.CreateTenantRequest{
		Name:       "Alpine Edge",
		Visibility: domain.TenantVisibilityPublic,
		AccessType: domain.TenantAccessOpen,
	})

	results, cursor, err := svc.ListPublicTenants(ctx, &domain.ListTenantsRequest{
		Limit:  1,
		Search: "al",
	})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.True(t, strings.Contains(strings.ToLower(results[0].Name), "al"))
	assert.NotEmpty(t, cursor)

	// Second page using cursor should also return a matching tenant
	secondPage, nextCursor, err := svc.ListPublicTenants(ctx, &domain.ListTenantsRequest{
		Limit:  1,
		Search: "al",
		Cursor: cursor,
	})
	require.NoError(t, err)
	require.Len(t, secondPage, 1)
	assert.True(t, strings.Contains(strings.ToLower(secondPage[0].Name), "al"))
	// No more matches afterwards
	assert.Equal(t, "", nextCursor)
}

func TestTenantMembershipService_ListPublicTenants_InvalidCursor(t *testing.T) {
	svc, _ := createTestTenantMembershipService()
	ctx := context.Background()

	_, _, err := svc.ListPublicTenants(ctx, &domain.ListTenantsRequest{Cursor: "not-a-number"})
	require.Error(t, err)
	var domainErr *domain.Error; ok := errors.As(err, &domainErr)
	require.True(t, ok)
	assert.Equal(t, domain.ErrInvalidRequest, domainErr.Code)
}

// ==================== JOIN TENANT TESTS ====================

func TestTenantMembershipService_JoinTenant_Success(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, _ := setupTenantWithOwner(ctx, svc, tenantRepo)

	member, err := svc.JoinTenant(ctx, "user-2", tenantID, &domain.JoinTenantRequest{})

	require.NoError(t, err)
	assert.Equal(t, tenantID, member.TenantID)
	assert.Equal(t, "user-2", member.UserID)
	assert.Equal(t, domain.TenantRoleMember, member.Role)
}

func TestTenantMembershipService_JoinTenant_AlreadyMember(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Owner tries to join again
	_, err := svc.JoinTenant(ctx, ownerID, tenantID, &domain.JoinTenantRequest{})

	require.Error(t, err)
	var domainErr *domain.Error; ok := errors.As(err, &domainErr)
	require.True(t, ok)
	assert.Equal(t, domain.ErrAlreadyMember, domainErr.Code)
}

func TestTenantMembershipService_JoinTenant_NonExistent(t *testing.T) {
	svc, _ := createTestTenantMembershipService()
	ctx := context.Background()

	_, err := svc.JoinTenant(ctx, "user-1", "non-existent", &domain.JoinTenantRequest{})

	require.Error(t, err)
}

func TestTenantMembershipService_JoinTenant_PasswordProtected(t *testing.T) {
	svc, _ := createTestTenantMembershipService()
	ctx := context.Background()

	tenant, err := svc.CreateTenant(ctx, "owner", &domain.CreateTenantRequest{
		Name:       "Secret",
		Visibility: domain.TenantVisibilityPrivate,
		AccessType: domain.TenantAccessPassword,
		Password:   "super-secret",
	})
	require.NoError(t, err)

	_, err = svc.JoinTenant(ctx, "user-1", tenant.ID, &domain.JoinTenantRequest{})
	require.Error(t, err)
	var domainErr *domain.Error; ok := errors.As(err, &domainErr)
	require.True(t, ok)
	assert.Equal(t, domain.ErrInvalidRequest, domainErr.Code)

	_, err = svc.JoinTenant(ctx, "user-1", tenant.ID, &domain.JoinTenantRequest{Password: "bad-pass"})
	require.Error(t, err)
	domainErr, ok = err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrInvalidCredentials, domainErr.Code)

	member, err := svc.JoinTenant(ctx, "user-1", tenant.ID, &domain.JoinTenantRequest{Password: "super-secret"})
	require.NoError(t, err)
	assert.Equal(t, tenant.ID, member.TenantID)
}

func TestTenantMembershipService_JoinTenant_InviteOnly(t *testing.T) {
	svc, _ := createTestTenantMembershipService()
	ctx := context.Background()

	tenant, err := svc.CreateTenant(ctx, "owner", &domain.CreateTenantRequest{
		Name:       "Private",
		Visibility: domain.TenantVisibilityPrivate,
		AccessType: domain.TenantAccessInviteOnly,
	})
	require.NoError(t, err)

	_, err = svc.JoinTenant(ctx, "user-1", tenant.ID, &domain.JoinTenantRequest{})
	require.Error(t, err)
	var domainErr *domain.Error; ok := errors.As(err, &domainErr)
	require.True(t, ok)
	assert.Equal(t, domain.ErrForbidden, domainErr.Code)
}

func TestTenantMembershipService_JoinTenant_MaxMembersReached(t *testing.T) {
	svc, _ := createTestTenantMembershipService()
	ctx := context.Background()

	tenant, err := svc.CreateTenant(ctx, "owner", &domain.CreateTenantRequest{
		Name:       "Small",
		Visibility: domain.TenantVisibilityPrivate,
		AccessType: domain.TenantAccessOpen,
		MaxMembers: 1,
	})
	require.NoError(t, err)

	_, err = svc.JoinTenant(ctx, "user-1", tenant.ID, &domain.JoinTenantRequest{})
	require.Error(t, err)
	var domainErr *domain.Error; ok := errors.As(err, &domainErr)
	require.True(t, ok)
	assert.Equal(t, domain.ErrForbidden, domainErr.Code)
}

func TestTenantMembershipService_JoinByCode_RespectsCapacity(t *testing.T) {
	svc, _ := createTestTenantMembershipService()
	ctx := context.Background()

	tenant, err := svc.CreateTenant(ctx, "owner", &domain.CreateTenantRequest{
		Name:       "InviteOnly",
		Visibility: domain.TenantVisibilityPrivate,
		AccessType: domain.TenantAccessInviteOnly,
		MaxMembers: 2,
	})
	require.NoError(t, err)

	invite, err := svc.CreateInvite(ctx, "owner", tenant.ID, &domain.CreateTenantInviteRequest{MaxUses: 5, ExpiresIn: 3600})
	require.NoError(t, err)

	member, err := svc.JoinByCode(ctx, "user-1", invite.Code)
	require.NoError(t, err)
	assert.Equal(t, tenant.ID, member.TenantID)

	_, err = svc.JoinByCode(ctx, "user-2", invite.Code)
	require.Error(t, err)
	var domainErr *domain.Error; ok := errors.As(err, &domainErr)
	require.True(t, ok)
	assert.Equal(t, domain.ErrForbidden, domainErr.Code)
}

// ==================== LEAVE TENANT TESTS ====================

func TestTenantMembershipService_LeaveTenant_Success(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, _ := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add a member
	_, _ = svc.JoinTenant(ctx, "user-2", tenantID, &domain.JoinTenantRequest{})

	// Member leaves
	err := svc.LeaveTenant(ctx, "user-2", tenantID)

	require.NoError(t, err)

	// Verify user is no longer a member
	members, _ := svc.GetUserTenants(ctx, "user-2")
	assert.Equal(t, 0, len(members))
}

func TestTenantMembershipService_LeaveTenant_OwnerCannot(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	err := svc.LeaveTenant(ctx, ownerID, tenantID)

	require.Error(t, err)
	var domainErr *domain.Error; ok := errors.As(err, &domainErr)
	require.True(t, ok)
	assert.Equal(t, domain.ErrForbidden, domainErr.Code)
}

func TestTenantMembershipService_LeaveTenant_NotMember(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, _ := setupTenantWithOwner(ctx, svc, tenantRepo)

	err := svc.LeaveTenant(ctx, "not-a-member", tenantID)

	require.Error(t, err)
}

// ==================== GET USER TENANTS TESTS ====================

func TestTenantMembershipService_GetUserTenants_Multiple(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	// Create 3 tenants
	tenant1, _ := setupTenantWithOwner(ctx, svc, tenantRepo)
	tenant2, _ := setupTenantWithOwner(ctx, svc, tenantRepo)
	tenant3, _ := setupTenantWithOwner(ctx, svc, tenantRepo)

	// User joins all 3
	_, _ = svc.JoinTenant(ctx, "user-joiner", tenant1, &domain.JoinTenantRequest{})
	_, _ = svc.JoinTenant(ctx, "user-joiner", tenant2, &domain.JoinTenantRequest{})
	_, _ = svc.JoinTenant(ctx, "user-joiner", tenant3, &domain.JoinTenantRequest{})

	members, err := svc.GetUserTenants(ctx, "user-joiner")

	require.NoError(t, err)
	assert.Equal(t, 3, len(members))
}

func TestTenantMembershipService_GetUserTenants_Empty(t *testing.T) {
	svc, _ := createTestTenantMembershipService()
	ctx := context.Background()

	members, err := svc.GetUserTenants(ctx, "user-with-no-tenants")

	require.NoError(t, err)
	assert.Equal(t, 0, len(members))
}

// ==================== GET TENANT MEMBERS TESTS ====================

func TestTenantMembershipService_GetTenantMembers_Success(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, _ := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add members
	for i := 0; i < 5; i++ {
		_, _ = svc.JoinTenant(ctx, "member-"+string(rune('a'+i)), tenantID, &domain.JoinTenantRequest{})
	}

	members, cursor, err := svc.GetTenantMembers(ctx, tenantID, &domain.ListTenantMembersRequest{Limit: 10})

	require.NoError(t, err)
	assert.Equal(t, 6, len(members)) // 5 members + 1 owner
	assert.Empty(t, cursor)
}

func TestTenantMembershipService_GetTenantMembers_Pagination(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, _ := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add 10 members
	for i := 0; i < 10; i++ {
		_, _ = svc.JoinTenant(ctx, "member-"+string(rune('a'+i)), tenantID, &domain.JoinTenantRequest{})
	}

	members, _, err := svc.GetTenantMembers(ctx, tenantID, &domain.ListTenantMembersRequest{Limit: 5})

	require.NoError(t, err)
	assert.Equal(t, 5, len(members))
	// Note: In-memory implementation doesn't use cursor
}

// ==================== UPDATE MEMBER ROLE TESTS ====================

func TestTenantMembershipService_UpdateMemberRole_OwnerPromotesToAdmin(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add a member
	member, _ := svc.JoinTenant(ctx, "user-2", tenantID, &domain.JoinTenantRequest{})

	// Owner promotes to admin
	err := svc.UpdateMemberRole(ctx, ownerID, tenantID, member.ID, domain.TenantRoleAdmin)

	require.NoError(t, err)
}

func TestTenantMembershipService_UpdateMemberRole_AdminCannotPromoteToAdmin(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add admin
	admin, _ := svc.JoinTenant(ctx, "admin-user", tenantID, &domain.JoinTenantRequest{})
	_ = svc.UpdateMemberRole(ctx, ownerID, tenantID, admin.ID, domain.TenantRoleAdmin)

	// Add member
	member, _ := svc.JoinTenant(ctx, "regular-user", tenantID, &domain.JoinTenantRequest{})

	// Admin tries to promote to admin - should fail
	err := svc.UpdateMemberRole(ctx, "admin-user", tenantID, member.ID, domain.TenantRoleAdmin)

	require.Error(t, err)
	var domainErr *domain.Error; ok := errors.As(err, &domainErr)
	require.True(t, ok)
	assert.Equal(t, domain.ErrForbidden, domainErr.Code)
}

func TestTenantMembershipService_UpdateMemberRole_CannotSetOwner(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	member, _ := svc.JoinTenant(ctx, "user-2", tenantID, &domain.JoinTenantRequest{})

	err := svc.UpdateMemberRole(ctx, ownerID, tenantID, member.ID, domain.TenantRoleOwner)

	require.Error(t, err)
	var domainErr *domain.Error; ok := errors.As(err, &domainErr)
	require.True(t, ok)
	assert.Equal(t, domain.ErrForbidden, domainErr.Code)
}

func TestTenantMembershipService_UpdateMemberRole_CannotChangeOwnRole(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add admin
	admin, _ := svc.JoinTenant(ctx, "admin-user", tenantID, &domain.JoinTenantRequest{})
	_ = svc.UpdateMemberRole(ctx, ownerID, tenantID, admin.ID, domain.TenantRoleAdmin)

	// Admin tries to change own role
	err := svc.UpdateMemberRole(ctx, "admin-user", tenantID, admin.ID, domain.TenantRoleModerator)

	require.Error(t, err)
	var domainErr *domain.Error; ok := errors.As(err, &domainErr)
	require.True(t, ok)
	assert.Equal(t, domain.ErrForbidden, domainErr.Code)
}

func TestTenantMembershipService_UpdateMemberRole_NotMember(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, _ := setupTenantWithOwner(ctx, svc, tenantRepo)

	member, _ := svc.JoinTenant(ctx, "user-2", tenantID, &domain.JoinTenantRequest{})

	// Non-member tries to change role
	err := svc.UpdateMemberRole(ctx, "outsider", tenantID, member.ID, domain.TenantRoleAdmin)

	require.Error(t, err)
}

// ==================== REMOVE MEMBER TESTS ====================

func TestTenantMembershipService_RemoveMember_ModeratorKicksMember(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add moderator
	mod, _ := svc.JoinTenant(ctx, "mod-user", tenantID, &domain.JoinTenantRequest{})
	_ = svc.UpdateMemberRole(ctx, ownerID, tenantID, mod.ID, domain.TenantRoleModerator)

	// Add member
	member, _ := svc.JoinTenant(ctx, "regular-user", tenantID, &domain.JoinTenantRequest{})

	// Moderator kicks member
	err := svc.RemoveMember(ctx, "mod-user", tenantID, member.ID)

	require.NoError(t, err)
}

func TestTenantMembershipService_RemoveMember_CannotKickOwner(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add admin
	admin, _ := svc.JoinTenant(ctx, "admin-user", tenantID, &domain.JoinTenantRequest{})
	_ = svc.UpdateMemberRole(ctx, ownerID, tenantID, admin.ID, domain.TenantRoleAdmin)

	// Get owner member ID
	ownerMembership, _ := svc.memberRepo.GetByUserAndTenant(ctx, ownerID, tenantID)

	// Admin tries to kick owner
	err := svc.RemoveMember(ctx, "admin-user", tenantID, ownerMembership.ID)

	require.Error(t, err)
	var domainErr *domain.Error; ok := errors.As(err, &domainErr)
	require.True(t, ok)
	assert.Equal(t, domain.ErrForbidden, domainErr.Code)
}

func TestTenantMembershipService_RemoveMember_CannotKickHigherRole(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add admin
	admin, _ := svc.JoinTenant(ctx, "admin-user", tenantID, &domain.JoinTenantRequest{})
	_ = svc.UpdateMemberRole(ctx, ownerID, tenantID, admin.ID, domain.TenantRoleAdmin)

	// Add moderator
	mod, _ := svc.JoinTenant(ctx, "mod-user", tenantID, &domain.JoinTenantRequest{})
	_ = svc.UpdateMemberRole(ctx, ownerID, tenantID, mod.ID, domain.TenantRoleModerator)

	// Moderator tries to kick admin
	err := svc.RemoveMember(ctx, "mod-user", tenantID, admin.ID)

	require.Error(t, err)
	var domainErr *domain.Error; ok := errors.As(err, &domainErr)
	require.True(t, ok)
	assert.Equal(t, domain.ErrForbidden, domainErr.Code)
}

func TestTenantMembershipService_RemoveMember_CannotKickSelf(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add moderator
	mod, _ := svc.JoinTenant(ctx, "mod-user", tenantID, &domain.JoinTenantRequest{})
	_ = svc.UpdateMemberRole(ctx, ownerID, tenantID, mod.ID, domain.TenantRoleModerator)

	// Moderator tries to kick self
	err := svc.RemoveMember(ctx, "mod-user", tenantID, mod.ID)

	require.Error(t, err)
}

func TestTenantMembershipService_RemoveMember_MemberCannotKick(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, _ := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add members
	member1, _ := svc.JoinTenant(ctx, "member-1", tenantID, &domain.JoinTenantRequest{})
	_, _ = svc.JoinTenant(ctx, "member-2", tenantID, &domain.JoinTenantRequest{})

	// Member tries to kick another member
	err := svc.RemoveMember(ctx, "member-2", tenantID, member1.ID)

	require.Error(t, err)
}

// ==================== UPDATE TENANT TESTS ====================

func TestTenantMembershipService_UpdateTenant_Success(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Owner updates tenant
	newName := "Updated Tenant Name"
	newDesc := "Updated description"
	req := &domain.UpdateTenantRequest{
		Name:        &newName,
		Description: &newDesc,
	}

	updated, err := svc.UpdateTenant(ctx, ownerID, tenantID, req)
	require.NoError(t, err)
	assert.Equal(t, newName, updated.Name)
	assert.Equal(t, newDesc, updated.Description)
}

func TestTenantMembershipService_UpdateTenant_MemberForbidden(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, _ := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add a regular member
	_, _ = svc.JoinTenant(ctx, "regular-user", tenantID, &domain.JoinTenantRequest{})

	// Member tries to update tenant
	newName := "Hacked Name"
	req := &domain.UpdateTenantRequest{
		Name: &newName,
	}

	_, err := svc.UpdateTenant(ctx, "regular-user", tenantID, req)
	require.Error(t, err)
	var domErr *domain.Error; if errors.As(err, &domErr) {
		assert.Equal(t, domain.ErrForbidden, domErr.Code)
	}
}

func TestTenantMembershipService_UpdateTenant_NotMember(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, _ := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Non-member tries to update tenant
	newName := "Hacked Name"
	req := &domain.UpdateTenantRequest{
		Name: &newName,
	}

	_, err := svc.UpdateTenant(ctx, "stranger", tenantID, req)
	require.Error(t, err)
}

// ==================== DELETE TENANT TESTS ====================

func TestTenantMembershipService_DeleteTenant_Success(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	err := svc.DeleteTenant(ctx, ownerID, tenantID)
	require.NoError(t, err)

	// Verify tenant is deleted
	_, err = tenantRepo.GetByID(ctx, tenantID)
	require.Error(t, err)
}

func TestTenantMembershipService_DeleteTenant_NonOwnerForbidden(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add admin
	_, _ = svc.JoinTenant(ctx, "admin-user", tenantID, &domain.JoinTenantRequest{})
	// Need to make them admin through UpdateMemberRole
	admin, _ := svc.JoinTenant(ctx, "admin2", tenantID, &domain.JoinTenantRequest{})
	_ = svc.UpdateMemberRole(ctx, ownerID, tenantID, admin.ID, domain.TenantRoleAdmin)

	// Admin tries to delete tenant (only owner can)
	err := svc.DeleteTenant(ctx, "admin2", tenantID)
	require.Error(t, err)
	var domErr *domain.Error; if errors.As(err, &domErr) {
		assert.Equal(t, domain.ErrForbidden, domErr.Code)
	}
}

// ==================== BAN/UNBAN MEMBER TESTS ====================

func TestTenantMembershipService_BanMember_Success(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add a member
	member, _ := svc.JoinTenant(ctx, "member-to-ban", tenantID, &domain.JoinTenantRequest{})

	// Owner bans member
	err := svc.BanMember(ctx, ownerID, tenantID, member.ID)
	require.NoError(t, err)
}

func TestTenantMembershipService_BanMember_NotAdmin(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, _ := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add members
	member1, _ := svc.JoinTenant(ctx, "member-1", tenantID, &domain.JoinTenantRequest{})
	_, _ = svc.JoinTenant(ctx, "member-2", tenantID, &domain.JoinTenantRequest{})

	// Regular member tries to ban another member
	err := svc.BanMember(ctx, "member-2", tenantID, member1.ID)
	require.Error(t, err)
}

func TestTenantMembershipService_UnbanMember_Success(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add and ban a member
	member, _ := svc.JoinTenant(ctx, "member-to-unban", tenantID, &domain.JoinTenantRequest{})
	_ = svc.BanMember(ctx, ownerID, tenantID, member.ID)

	// Owner unbans member
	err := svc.UnbanMember(ctx, ownerID, tenantID, member.ID)
	require.NoError(t, err)
}

func TestTenantMembershipService_UnbanMember_NotAdmin(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add and ban a member
	member, _ := svc.JoinTenant(ctx, "banned-member", tenantID, &domain.JoinTenantRequest{})
	_ = svc.BanMember(ctx, ownerID, tenantID, member.ID)

	// Add another regular member who tries to unban
	_, _ = svc.JoinTenant(ctx, "regular-member", tenantID, &domain.JoinTenantRequest{})

	err := svc.UnbanMember(ctx, "regular-member", tenantID, member.ID)
	require.Error(t, err)
}

func TestTenantMembershipService_ListBannedMembers_Success(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add and ban members
	member1, _ := svc.JoinTenant(ctx, "ban-list-1", tenantID, &domain.JoinTenantRequest{})
	member2, _ := svc.JoinTenant(ctx, "ban-list-2", tenantID, &domain.JoinTenantRequest{})
	_ = svc.BanMember(ctx, ownerID, tenantID, member1.ID)
	_ = svc.BanMember(ctx, ownerID, tenantID, member2.ID)

	// List banned members
	banned, err := svc.ListBannedMembers(ctx, ownerID, tenantID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(banned), 2)
}

func TestTenantMembershipService_ListBannedMembers_NotAdmin(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, _ := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add regular member
	_, _ = svc.JoinTenant(ctx, "regular", tenantID, &domain.JoinTenantRequest{})

	// Regular member tries to list banned members
	_, err := svc.ListBannedMembers(ctx, "regular", tenantID)
	require.Error(t, err)
}

func TestTenantMembershipService_UnbanMember_NotBanned(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add a member (not banned)
	member, _ := svc.JoinTenant(ctx, "not-banned-member", tenantID, &domain.JoinTenantRequest{})

	// Try to unban a member who is not banned
	err := svc.UnbanMember(ctx, ownerID, tenantID, member.ID)
	require.Error(t, err)
	var domainErr *domain.Error; ok := errors.As(err, &domainErr)
	require.True(t, ok)
	assert.Equal(t, domain.ErrValidation, domainErr.Code)
}

func TestTenantMembershipService_UnbanMember_MemberNotFound(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Try to unban a non-existent member
	err := svc.UnbanMember(ctx, ownerID, tenantID, "nonexistent-member-id")
	require.Error(t, err)
}

func TestTenantMembershipService_UnbanMember_WrongTenant(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	// Create two tenants
	tenantID1, ownerID1 := setupTenantWithOwner(ctx, svc, tenantRepo)

	tenant2 := &domain.Tenant{
		ID:         "tenant-2",
		Name:       "Second Tenant",
		Visibility: domain.TenantVisibilityPublic,
		AccessType: domain.TenantAccessOpen,
	}
	require.NoError(t, tenantRepo.Create(ctx, tenant2))
	svc.JoinTenant(ctx, "owner-2", tenant2.ID, &domain.JoinTenantRequest{})

	// Add and ban a member in tenant1
	member, _ := svc.JoinTenant(ctx, "member-t1", tenantID1, &domain.JoinTenantRequest{})
	_ = svc.BanMember(ctx, ownerID1, tenantID1, member.ID)

	// Try to unban from wrong tenant (owner2 tries to unban from tenant2 using member from tenant1)
	err := svc.UnbanMember(ctx, "owner-2", tenant2.ID, member.ID)
	require.Error(t, err)
}

// ==================== INVITE TESTS ====================

func TestTenantMembershipService_ListInvites_Success(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Create an invite first
	_, err := svc.CreateInvite(ctx, ownerID, tenantID, &domain.CreateTenantInviteRequest{
		MaxUses:   10,
		ExpiresIn: 86400, // 24 hours in seconds
	})
	require.NoError(t, err)

	// List invites as admin (owner)
	invites, err := svc.ListInvites(ctx, ownerID, tenantID)
	require.NoError(t, err)
	assert.Len(t, invites, 1)
}

func TestTenantMembershipService_ListInvites_NotAdmin(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, _ := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add regular member
	_, _ = svc.JoinTenant(ctx, "regular", tenantID, &domain.JoinTenantRequest{})

	// Regular member tries to list invites
	_, err := svc.ListInvites(ctx, "regular", tenantID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "admin")
}

func TestTenantMembershipService_RevokeInvite_Success(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Create an invite
	invite, err := svc.CreateInvite(ctx, ownerID, tenantID, &domain.CreateTenantInviteRequest{
		MaxUses:   10,
		ExpiresIn: 86400,
	})
	require.NoError(t, err)

	// Revoke the invite
	err = svc.RevokeInvite(ctx, ownerID, tenantID, invite.ID)
	require.NoError(t, err)
	// Success is indicated by no error returned
}

func TestTenantMembershipService_RevokeInvite_NotAdmin(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Create an invite
	invite, _ := svc.CreateInvite(ctx, ownerID, tenantID, &domain.CreateTenantInviteRequest{
		MaxUses:   10,
		ExpiresIn: 86400,
	})

	// Add regular member
	_, _ = svc.JoinTenant(ctx, "regular", tenantID, &domain.JoinTenantRequest{})

	// Regular member tries to revoke invite
	err := svc.RevokeInvite(ctx, "regular", tenantID, invite.ID)
	require.Error(t, err)
}

// ==================== ANNOUNCEMENT TESTS ====================

func TestTenantMembershipService_CreateAnnouncement_Success(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	announcement, err := svc.CreateAnnouncement(ctx, ownerID, tenantID, &domain.CreateAnnouncementRequest{
		Title:    "Important Update",
		Content:  "This is an important announcement.",
		IsPinned: true,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, announcement.ID)
	assert.Equal(t, "Important Update", announcement.Title)
}

func TestTenantMembershipService_CreateAnnouncement_NotModerator(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, _ := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add regular member
	_, _ = svc.JoinTenant(ctx, "regular", tenantID, &domain.JoinTenantRequest{})

	// Regular member tries to create announcement
	_, err := svc.CreateAnnouncement(ctx, "regular", tenantID, &domain.CreateAnnouncementRequest{
		Title:   "Test",
		Content: "Content",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "moderator")
}

func TestTenantMembershipService_GetAnnouncements_Success(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Create an announcement
	_, _ = svc.CreateAnnouncement(ctx, ownerID, tenantID, &domain.CreateAnnouncementRequest{
		Title:   "Announcement 1",
		Content: "Content 1",
	})

	// Get announcements
	announcements, _, err := svc.GetAnnouncements(ctx, ownerID, tenantID, &domain.ListAnnouncementsRequest{})
	require.NoError(t, err)
	assert.Len(t, announcements, 1)
}

func TestTenantMembershipService_GetAnnouncements_NotMember(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Create an announcement
	_, _ = svc.CreateAnnouncement(ctx, ownerID, tenantID, &domain.CreateAnnouncementRequest{
		Title:   "Test",
		Content: "Content",
	})

	// Non-member tries to get announcements
	_, _, err := svc.GetAnnouncements(ctx, "non-member", tenantID, &domain.ListAnnouncementsRequest{})
	require.Error(t, err)
}

func TestTenantMembershipService_UpdateAnnouncement_Success(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Create an announcement
	announcement, err := svc.CreateAnnouncement(ctx, ownerID, tenantID, &domain.CreateAnnouncementRequest{
		Title:   "Original Title",
		Content: "Original Content",
	})
	require.NoError(t, err)

	// Update the announcement
	newTitle := "Updated Title"
	newContent := "Updated Content"
	isPinned := true
	err = svc.UpdateAnnouncement(ctx, ownerID, tenantID, announcement.ID, &domain.UpdateAnnouncementRequest{
		Title:    &newTitle,
		Content:  &newContent,
		IsPinned: &isPinned,
	})
	require.NoError(t, err)
}

func TestTenantMembershipService_UpdateAnnouncement_NotModerator(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Create an announcement
	announcement, _ := svc.CreateAnnouncement(ctx, ownerID, tenantID, &domain.CreateAnnouncementRequest{
		Title:   "Test",
		Content: "Content",
	})

	// Add regular member
	_, _ = svc.JoinTenant(ctx, "regular", tenantID, &domain.JoinTenantRequest{})

	// Regular member tries to update
	newTitle := "Hacked"
	err := svc.UpdateAnnouncement(ctx, "regular", tenantID, announcement.ID, &domain.UpdateAnnouncementRequest{
		Title: &newTitle,
	})
	require.Error(t, err)
}

func TestTenantMembershipService_DeleteAnnouncement_Success(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Create an announcement
	announcement, _ := svc.CreateAnnouncement(ctx, ownerID, tenantID, &domain.CreateAnnouncementRequest{
		Title:   "To Delete",
		Content: "Content",
	})

	// Delete announcement
	err := svc.DeleteAnnouncement(ctx, ownerID, tenantID, announcement.ID)
	require.NoError(t, err)
}

func TestTenantMembershipService_DeleteAnnouncement_NotModerator(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Create an announcement
	announcement, _ := svc.CreateAnnouncement(ctx, ownerID, tenantID, &domain.CreateAnnouncementRequest{
		Title:   "Test",
		Content: "Content",
	})

	// Add regular member
	_, _ = svc.JoinTenant(ctx, "regular", tenantID, &domain.JoinTenantRequest{})

	// Regular member tries to delete
	err := svc.DeleteAnnouncement(ctx, "regular", tenantID, announcement.ID)
	require.Error(t, err)
}

// ==================== CHAT TESTS ====================

func TestTenantMembershipService_SendChatMessage_Success(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Send a chat message
	msg, err := svc.SendChatMessage(ctx, ownerID, tenantID, &domain.SendChatMessageRequest{
		Content: "Hello, world!",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, msg.ID)
	assert.Equal(t, "Hello, world!", msg.Content)
}

func TestTenantMembershipService_SendChatMessage_NotMember(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, _ := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Non-member tries to send message
	_, err := svc.SendChatMessage(ctx, "non-member", tenantID, &domain.SendChatMessageRequest{
		Content: "Hello",
	})
	require.Error(t, err)
}

func TestTenantMembershipService_GetChatHistory_Success(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Send a few messages
	_, _ = svc.SendChatMessage(ctx, ownerID, tenantID, &domain.SendChatMessageRequest{Content: "Msg 1"})
	_, _ = svc.SendChatMessage(ctx, ownerID, tenantID, &domain.SendChatMessageRequest{Content: "Msg 2"})

	// Get chat history
	messages, err := svc.GetChatHistory(ctx, ownerID, tenantID, &domain.ListChatMessagesRequest{})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(messages), 2)
}

func TestTenantMembershipService_GetChatHistory_NotMember(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, _ := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Non-member tries to get chat history
	_, err := svc.GetChatHistory(ctx, "non-member", tenantID, &domain.ListChatMessagesRequest{})
	require.Error(t, err)
}

func TestTenantMembershipService_DeleteChatMessage_Success(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Send a message
	msg, _ := svc.SendChatMessage(ctx, ownerID, tenantID, &domain.SendChatMessageRequest{Content: "To Delete"})

	// Delete message (owner can delete)
	err := svc.DeleteChatMessage(ctx, ownerID, tenantID, msg.ID)
	require.NoError(t, err)
}

func TestTenantMembershipService_DeleteChatMessage_NotAuthorized(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Owner sends a message
	msg, _ := svc.SendChatMessage(ctx, ownerID, tenantID, &domain.SendChatMessageRequest{Content: "Owner's message"})

	// Add regular member
	_, _ = svc.JoinTenant(ctx, "regular", tenantID, &domain.JoinTenantRequest{})

	// Regular member tries to delete owner's message
	err := svc.DeleteChatMessage(ctx, "regular", tenantID, msg.ID)
	require.Error(t, err)
}

// ==================== PERMISSION CHECK TESTS ====================

func TestTenantMembershipService_CheckTenantPermission_Success(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Check owner has owner permission
	err := svc.CheckTenantPermission(ctx, ownerID, tenantID, domain.TenantRoleOwner)
	require.NoError(t, err)

	// Check owner has admin permission (owner >= admin)
	err = svc.CheckTenantPermission(ctx, ownerID, tenantID, domain.TenantRoleAdmin)
	require.NoError(t, err)
}

func TestTenantMembershipService_CheckTenantPermission_Forbidden(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, _ := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add regular member
	_, _ = svc.JoinTenant(ctx, "regular", tenantID, &domain.JoinTenantRequest{})

	// Regular member doesn't have admin permission
	err := svc.CheckTenantPermission(ctx, "regular", tenantID, domain.TenantRoleAdmin)
	require.Error(t, err)
}

// ==================== JOIN BY CODE ADDITIONAL TESTS ====================

func TestTenantMembershipService_JoinByCode_ExpiredInvite(t *testing.T) {
	svc, _ := createTestTenantMembershipService()
	ctx := context.Background()

	// Create tenant with invite-only access
	tenant, err := svc.CreateTenant(ctx, "owner-expired", &domain.CreateTenantRequest{
		Name:       "Expired Invite Tenant",
		Visibility: domain.TenantVisibilityPrivate,
		AccessType: domain.TenantAccessInviteOnly,
		MaxMembers: 100,
	})
	require.NoError(t, err)

	// Create invite with very short expiry (1 second)
	invite, err := svc.CreateInvite(ctx, "owner-expired", tenant.ID, &domain.CreateTenantInviteRequest{
		MaxUses:   10,
		ExpiresIn: 1, // 1 second
	})
	require.NoError(t, err)

	// Wait for invite to expire
	time.Sleep(2 * time.Second)

	// Try to use expired invite
	_, err = svc.JoinByCode(ctx, "user-late", invite.Code)
	require.Error(t, err)
	var domainErr *domain.Error; ok := errors.As(err, &domainErr)
	require.True(t, ok)
	assert.Equal(t, domain.ErrInviteTokenExpired, domainErr.Code)
}

func TestTenantMembershipService_JoinByCode_MaxUsesReached(t *testing.T) {
	svc, _ := createTestTenantMembershipService()
	ctx := context.Background()

	// Create tenant
	tenant, err := svc.CreateTenant(ctx, "owner-max-uses", &domain.CreateTenantRequest{
		Name:       "Max Uses Tenant",
		Visibility: domain.TenantVisibilityPrivate,
		AccessType: domain.TenantAccessInviteOnly,
		MaxMembers: 100,
	})
	require.NoError(t, err)

	// Create invite with max uses = 1
	invite, err := svc.CreateInvite(ctx, "owner-max-uses", tenant.ID, &domain.CreateTenantInviteRequest{
		MaxUses:   1,
		ExpiresIn: 3600,
	})
	require.NoError(t, err)

	// First use should succeed
	_, err = svc.JoinByCode(ctx, "user-first", invite.Code)
	require.NoError(t, err)

	// Second use should fail
	_, err = svc.JoinByCode(ctx, "user-second", invite.Code)
	require.Error(t, err)
	var domainErr *domain.Error; ok := errors.As(err, &domainErr)
	require.True(t, ok)
	assert.Equal(t, domain.ErrInviteTokenExpired, domainErr.Code)
}

func TestTenantMembershipService_JoinByCode_InvalidCode(t *testing.T) {
	svc, _ := createTestTenantMembershipService()
	ctx := context.Background()

	// Try to join with invalid code
	_, err := svc.JoinByCode(ctx, "user-invalid", "invalid-code-12345")
	require.Error(t, err)
}

func TestTenantMembershipService_JoinByCode_AlreadyMember(t *testing.T) {
	svc, _ := createTestTenantMembershipService()
	ctx := context.Background()

	// Create tenant
	tenant, err := svc.CreateTenant(ctx, "owner-already", &domain.CreateTenantRequest{
		Name:       "Already Member Tenant",
		Visibility: domain.TenantVisibilityPrivate,
		AccessType: domain.TenantAccessInviteOnly,
		MaxMembers: 100,
	})
	require.NoError(t, err)

	// Create invite
	invite, err := svc.CreateInvite(ctx, "owner-already", tenant.ID, &domain.CreateTenantInviteRequest{
		MaxUses:   10,
		ExpiresIn: 3600,
	})
	require.NoError(t, err)

	// First join should succeed
	_, err = svc.JoinByCode(ctx, "user-join-twice", invite.Code)
	require.NoError(t, err)

	// Second join should fail (already a member)
	_, err = svc.JoinByCode(ctx, "user-join-twice", invite.Code)
	require.Error(t, err)
	var domainErr *domain.Error; ok := errors.As(err, &domainErr)
	require.True(t, ok)
	assert.Equal(t, domain.ErrAlreadyMember, domainErr.Code)
}

// ==================== UPDATE TENANT ADDITIONAL TESTS ====================

func TestTenantMembershipService_UpdateTenant_ChangeVisibility(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Change visibility from public to private
	visibility := domain.TenantVisibilityPrivate
	req := &domain.UpdateTenantRequest{
		Visibility: &visibility,
	}

	updated, err := svc.UpdateTenant(ctx, ownerID, tenantID, req)
	require.NoError(t, err)
	assert.Equal(t, domain.TenantVisibilityPrivate, updated.Visibility)
}

func TestTenantMembershipService_UpdateTenant_ChangeAccessType(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Change access type to invite only
	accessType := domain.TenantAccessInviteOnly
	req := &domain.UpdateTenantRequest{
		AccessType: &accessType,
	}

	updated, err := svc.UpdateTenant(ctx, ownerID, tenantID, req)
	require.NoError(t, err)
	assert.Equal(t, domain.TenantAccessInviteOnly, updated.AccessType)
}

func TestTenantMembershipService_UpdateTenant_AdminCanUpdate(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add admin
	adminMember, _ := svc.JoinTenant(ctx, "admin-updater", tenantID, &domain.JoinTenantRequest{})
	_ = svc.UpdateMemberRole(ctx, ownerID, tenantID, adminMember.ID, domain.TenantRoleAdmin)

	// Admin updates tenant
	newDesc := "Admin updated description"
	req := &domain.UpdateTenantRequest{
		Description: &newDesc,
	}

	updated, err := svc.UpdateTenant(ctx, "admin-updater", tenantID, req)
	require.NoError(t, err)
	assert.Equal(t, newDesc, updated.Description)
}

// ==================== BAN MEMBER ADDITIONAL TESTS ====================

func TestTenantMembershipService_BanMember_CannotBanOwner(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add admin
	adminMember, _ := svc.JoinTenant(ctx, "admin-ban-owner", tenantID, &domain.JoinTenantRequest{})
	_ = svc.UpdateMemberRole(ctx, ownerID, tenantID, adminMember.ID, domain.TenantRoleAdmin)

	// Get owner's member ID
	ownerMembership, _ := svc.memberRepo.GetByUserAndTenant(ctx, ownerID, tenantID)

	// Admin tries to ban owner
	err := svc.BanMember(ctx, "admin-ban-owner", tenantID, ownerMembership.ID)
	require.Error(t, err)
	var domainErr *domain.Error; ok := errors.As(err, &domainErr)
	require.True(t, ok)
	assert.Equal(t, domain.ErrForbidden, domainErr.Code)
}

func TestTenantMembershipService_BanMember_CannotBanHigherRole(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add admin
	adminMember, _ := svc.JoinTenant(ctx, "admin-to-ban", tenantID, &domain.JoinTenantRequest{})
	_ = svc.UpdateMemberRole(ctx, ownerID, tenantID, adminMember.ID, domain.TenantRoleAdmin)

	// Add moderator
	modMember, _ := svc.JoinTenant(ctx, "mod-banner", tenantID, &domain.JoinTenantRequest{})
	_ = svc.UpdateMemberRole(ctx, ownerID, tenantID, modMember.ID, domain.TenantRoleModerator)

	// Moderator tries to ban admin
	err := svc.BanMember(ctx, "mod-banner", tenantID, adminMember.ID)
	require.Error(t, err)
	var domainErr *domain.Error; ok := errors.As(err, &domainErr)
	require.True(t, ok)
	assert.Equal(t, domain.ErrForbidden, domainErr.Code)
}

func TestTenantMembershipService_BanMember_BannedUserCannotRejoin(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, ownerID := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Add and ban a member
	member, _ := svc.JoinTenant(ctx, "banned-user", tenantID, &domain.JoinTenantRequest{})
	_ = svc.BanMember(ctx, ownerID, tenantID, member.ID)

	// Banned user tries to rejoin
	_, err := svc.JoinTenant(ctx, "banned-user", tenantID, &domain.JoinTenantRequest{})
	require.Error(t, err)
	var domainErr *domain.Error; ok := errors.As(err, &domainErr)
	require.True(t, ok)
	assert.Equal(t, domain.ErrForbidden, domainErr.Code)
}

// ==================== DELETE TENANT EDGE CASE TESTS ====================

func TestTenantMembershipService_DeleteTenant_NotMember(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	tenantID, _ := setupTenantWithOwner(ctx, svc, tenantRepo)

	// Non-member tries to delete tenant
	err := svc.DeleteTenant(ctx, "random-user", tenantID)
	require.Error(t, err)
	var domainErr *domain.Error; ok := errors.As(err, &domainErr)
	require.True(t, ok)
	assert.Equal(t, domain.ErrForbidden, domainErr.Code)
}

func TestTenantMembershipService_DeleteTenant_TenantNotFound(t *testing.T) {
	svc, tenantRepo := createTestTenantMembershipService()
	ctx := context.Background()

	// Create a tenant and owner first
	tenant := &domain.Tenant{
		ID:         "delete-test-tenant",
		Name:       "Test Tenant",
		Visibility: domain.TenantVisibilityPublic,
		AccessType: domain.TenantAccessOpen,
	}
	require.NoError(t, tenantRepo.Create(ctx, tenant))

	// Create membership
	_, _ = svc.JoinTenant(ctx, "owner-user", tenant.ID, &domain.JoinTenantRequest{})

	// Delete tenant from repo directly to simulate inconsistent state
	_ = tenantRepo.Delete(ctx, tenant.ID)

	// Now try to delete (member exists but tenant doesn't)
	err := svc.DeleteTenant(ctx, "owner-user", tenant.ID)
	// Should fail because member lookup will fail or tenant not found
	require.Error(t, err)
}
