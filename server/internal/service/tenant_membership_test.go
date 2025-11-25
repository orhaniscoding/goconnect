package service

import (
	"context"
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
	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, domain.ErrAlreadyMember, domainErr.Code)
}

func TestTenantMembershipService_JoinTenant_NonExistent(t *testing.T) {
	svc, _ := createTestTenantMembershipService()
	ctx := context.Background()

	_, err := svc.JoinTenant(ctx, "user-1", "non-existent", &domain.JoinTenantRequest{})

	require.Error(t, err)
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
	domainErr, ok := err.(*domain.Error)
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
	domainErr, ok := err.(*domain.Error)
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
	domainErr, ok := err.(*domain.Error)
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
	domainErr, ok := err.(*domain.Error)
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
	domainErr, ok := err.(*domain.Error)
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
	domainErr, ok := err.(*domain.Error)
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
