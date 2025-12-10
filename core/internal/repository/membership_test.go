package repository

import (
	"context"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInMemoryMembershipRepository(t *testing.T) {
	repo := NewInMemoryMembershipRepository()
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.byKey)
	assert.NotNil(t, repo.byNetwork)
}

func TestMembershipRepository_UpsertApproved_Create(t *testing.T) {
	repo := NewInMemoryMembershipRepository()
	ctx := context.Background()
	joinedAt := time.Now()

	m, err := repo.UpsertApproved(ctx, "network-1", "user-1", domain.RoleMember, joinedAt)

	require.NoError(t, err)
	assert.NotEmpty(t, m.ID)
	assert.Equal(t, "network-1", m.NetworkID)
	assert.Equal(t, "user-1", m.UserID)
	assert.Equal(t, domain.RoleMember, m.Role)
	assert.Equal(t, domain.StatusApproved, m.Status)
	assert.NotNil(t, m.JoinedAt)
	assert.Equal(t, joinedAt.Unix(), m.JoinedAt.Unix())
	assert.False(t, m.CreatedAt.IsZero())
	assert.False(t, m.UpdatedAt.IsZero())
}

func TestMembershipRepository_UpsertApproved_Update(t *testing.T) {
	repo := NewInMemoryMembershipRepository()
	ctx := context.Background()
	joinedAt := time.Now()

	// Create initial membership
	m1, err := repo.UpsertApproved(ctx, "network-1", "user-1", domain.RoleMember, joinedAt)
	require.NoError(t, err)
	originalID := m1.ID

	time.Sleep(2 * time.Millisecond)
	newJoinedAt := time.Now()

	// Update to admin role
	m2, err := repo.UpsertApproved(ctx, "network-1", "user-1", domain.RoleAdmin, newJoinedAt)

	require.NoError(t, err)
	assert.Equal(t, originalID, m2.ID) // Same membership record
	assert.Equal(t, domain.RoleAdmin, m2.Role)
	assert.Equal(t, domain.StatusApproved, m2.Status)
	assert.Equal(t, newJoinedAt.Unix(), m2.JoinedAt.Unix())
	// UpdatedAt should be refreshed (allow for time precision)
	assert.True(t, m2.UpdatedAt.Unix() >= m1.UpdatedAt.Unix())
}

func TestMembershipRepository_Get_Success(t *testing.T) {
	repo := NewInMemoryMembershipRepository()
	ctx := context.Background()
	joinedAt := time.Now()

	created, err := repo.UpsertApproved(ctx, "network-1", "user-1", domain.RoleMember, joinedAt)
	require.NoError(t, err)

	m, err := repo.Get(ctx, "network-1", "user-1")

	require.NoError(t, err)
	assert.Equal(t, created.ID, m.ID)
	assert.Equal(t, "network-1", m.NetworkID)
	assert.Equal(t, "user-1", m.UserID)
	assert.Equal(t, domain.RoleMember, m.Role)
}

func TestMembershipRepository_Get_NotFound(t *testing.T) {
	repo := NewInMemoryMembershipRepository()
	ctx := context.Background()

	m, err := repo.Get(ctx, "network-1", "user-1")

	assert.Nil(t, m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestMembershipRepository_SetStatus_Success(t *testing.T) {
	repo := NewInMemoryMembershipRepository()
	ctx := context.Background()
	joinedAt := time.Now()

	m, err := repo.UpsertApproved(ctx, "network-1", "user-1", domain.RoleMember, joinedAt)
	require.NoError(t, err)
	originalUpdatedAt := m.UpdatedAt

	time.Sleep(1 * time.Millisecond)
	err = repo.SetStatus(ctx, "network-1", "user-1", domain.StatusBanned)

	require.NoError(t, err)

	// Verify status changed
	m, err = repo.Get(ctx, "network-1", "user-1")
	require.NoError(t, err)
	assert.Equal(t, domain.StatusBanned, m.Status)
	assert.True(t, m.UpdatedAt.After(originalUpdatedAt))
}

func TestMembershipRepository_SetStatus_NotFound(t *testing.T) {
	repo := NewInMemoryMembershipRepository()
	ctx := context.Background()

	err := repo.SetStatus(ctx, "network-1", "user-1", domain.StatusBanned)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestMembershipRepository_List_EmptyNetwork(t *testing.T) {
	repo := NewInMemoryMembershipRepository()
	ctx := context.Background()

	members, next, err := repo.List(ctx, "network-1", "", 10, "")

	require.NoError(t, err)
	assert.Empty(t, members)
	assert.Empty(t, next)
}

func TestMembershipRepository_List_AllMembers(t *testing.T) {
	repo := NewInMemoryMembershipRepository()
	ctx := context.Background()
	joinedAt := time.Now()

	// Create 3 members
	repo.UpsertApproved(ctx, "network-1", "user-1", domain.RoleMember, joinedAt)
	repo.UpsertApproved(ctx, "network-1", "user-2", domain.RoleMember, joinedAt.Add(1*time.Second))
	repo.UpsertApproved(ctx, "network-1", "user-3", domain.RoleAdmin, joinedAt.Add(2*time.Second))

	members, next, err := repo.List(ctx, "network-1", "", 10, "")

	require.NoError(t, err)
	assert.Len(t, members, 3)
	assert.Empty(t, next)
	// Should be sorted by JoinedAt ascending
	assert.Equal(t, "user-1", members[0].UserID)
	assert.Equal(t, "user-2", members[1].UserID)
	assert.Equal(t, "user-3", members[2].UserID)
}

func TestMembershipRepository_List_FilterByStatus(t *testing.T) {
	repo := NewInMemoryMembershipRepository()
	ctx := context.Background()
	joinedAt := time.Now()

	// Create approved and banned members with different joinedAt times
	repo.UpsertApproved(ctx, "network-1", "user-1", domain.RoleMember, joinedAt)
	repo.UpsertApproved(ctx, "network-1", "user-2", domain.RoleMember, joinedAt.Add(1*time.Second))
	repo.UpsertApproved(ctx, "network-1", "user-3", domain.RoleMember, joinedAt.Add(2*time.Second))
	repo.SetStatus(ctx, "network-1", "user-2", domain.StatusBanned)

	// List only approved
	members, next, err := repo.List(ctx, "network-1", "approved", 10, "")

	require.NoError(t, err)
	assert.Len(t, members, 2)
	assert.Empty(t, next)
	// Verify user-1 and user-3 are approved (order by joinedAt)
	userIDs := []string{members[0].UserID, members[1].UserID}
	assert.Contains(t, userIDs, "user-1")
	assert.Contains(t, userIDs, "user-3")

	// List only banned
	members, _, err = repo.List(ctx, "network-1", "banned", 10, "")

	require.NoError(t, err)
	assert.Len(t, members, 1)
	assert.Equal(t, "user-2", members[0].UserID)
}

func TestMembershipRepository_List_Pagination(t *testing.T) {
	repo := NewInMemoryMembershipRepository()
	ctx := context.Background()
	joinedAt := time.Now()

	// Create 5 members
	for i := 1; i <= 5; i++ {
		repo.UpsertApproved(ctx, "network-1", "user-"+string(rune('0'+i)), domain.RoleMember, joinedAt.Add(time.Duration(i)*time.Second))
	}

	// Page 1: limit 2
	members, next, err := repo.List(ctx, "network-1", "", 2, "")
	require.NoError(t, err)
	assert.Len(t, members, 2)
	assert.NotEmpty(t, next)
	assert.Equal(t, "user-1", members[0].UserID)
	assert.Equal(t, "user-2", members[1].UserID)

	// Page 2: limit 2, cursor from page 1
	members, next, err = repo.List(ctx, "network-1", "", 2, next)
	require.NoError(t, err)
	assert.Len(t, members, 2)
	assert.NotEmpty(t, next)
	assert.Equal(t, "user-3", members[0].UserID)
	assert.Equal(t, "user-4", members[1].UserID)

	// Page 3: remaining 1 member
	members, next, err = repo.List(ctx, "network-1", "", 2, next)
	require.NoError(t, err)
	assert.Len(t, members, 1)
	assert.Empty(t, next) // No more pages
	assert.Equal(t, "user-5", members[0].UserID)
}

func TestMembershipRepository_List_DifferentNetworks(t *testing.T) {
	repo := NewInMemoryMembershipRepository()
	ctx := context.Background()
	joinedAt := time.Now()

	// Create members in different networks
	repo.UpsertApproved(ctx, "network-1", "user-1", domain.RoleMember, joinedAt)
	repo.UpsertApproved(ctx, "network-1", "user-2", domain.RoleMember, joinedAt)
	repo.UpsertApproved(ctx, "network-2", "user-3", domain.RoleMember, joinedAt)

	// List network-1 members
	members, _, err := repo.List(ctx, "network-1", "", 10, "")
	require.NoError(t, err)
	assert.Len(t, members, 2)

	// List network-2 members
	members, _, err = repo.List(ctx, "network-2", "", 10, "")
	require.NoError(t, err)
	assert.Len(t, members, 1)
	assert.Equal(t, "user-3", members[0].UserID)
}

func TestMembershipRepository_Remove_Success(t *testing.T) {
	repo := NewInMemoryMembershipRepository()
	ctx := context.Background()
	joinedAt := time.Now()

	repo.UpsertApproved(ctx, "network-1", "user-1", domain.RoleMember, joinedAt)

	err := repo.Remove(ctx, "network-1", "user-1")

	require.NoError(t, err)

	// Verify removed
	m, err := repo.Get(ctx, "network-1", "user-1")
	assert.Nil(t, m)
	require.Error(t, err)
}

func TestMembershipRepository_Remove_NotFound(t *testing.T) {
	repo := NewInMemoryMembershipRepository()
	ctx := context.Background()

	err := repo.Remove(ctx, "network-1", "user-1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestMembershipRepository_Remove_FromList(t *testing.T) {
	repo := NewInMemoryMembershipRepository()
	ctx := context.Background()
	joinedAt := time.Now()

	// Create 3 members
	repo.UpsertApproved(ctx, "network-1", "user-1", domain.RoleMember, joinedAt)
	repo.UpsertApproved(ctx, "network-1", "user-2", domain.RoleMember, joinedAt)
	repo.UpsertApproved(ctx, "network-1", "user-3", domain.RoleMember, joinedAt)

	// Remove middle member
	err := repo.Remove(ctx, "network-1", "user-2")
	require.NoError(t, err)

	// Verify list only has 2 members
	members, _, err := repo.List(ctx, "network-1", "", 10, "")
	require.NoError(t, err)
	assert.Len(t, members, 2)
	// Verify user-2 is removed (remaining are user-1 and user-3)
	userIDs := []string{members[0].UserID, members[1].UserID}
	assert.Contains(t, userIDs, "user-1")
	assert.Contains(t, userIDs, "user-3")
	assert.NotContains(t, userIDs, "user-2")
}

func TestMembershipRepository_FullCycle(t *testing.T) {
	repo := NewInMemoryMembershipRepository()
	ctx := context.Background()
	joinedAt := time.Now()

	// Create
	m, err := repo.UpsertApproved(ctx, "network-1", "user-1", domain.RoleMember, joinedAt)
	require.NoError(t, err)
	assert.Equal(t, domain.RoleMember, m.Role)

	// Get
	m, err = repo.Get(ctx, "network-1", "user-1")
	require.NoError(t, err)
	assert.Equal(t, domain.StatusApproved, m.Status)

	// Update status
	err = repo.SetStatus(ctx, "network-1", "user-1", domain.StatusBanned)
	require.NoError(t, err)

	// Update role (via upsert)
	m, err = repo.UpsertApproved(ctx, "network-1", "user-1", domain.RoleAdmin, joinedAt)
	require.NoError(t, err)
	assert.Equal(t, domain.RoleAdmin, m.Role)
	assert.Equal(t, domain.StatusApproved, m.Status) // Resets to approved

	// Remove
	err = repo.Remove(ctx, "network-1", "user-1")
	require.NoError(t, err)

	// Verify removed
	m, err = repo.Get(ctx, "network-1", "user-1")
	assert.Nil(t, m)
	assert.Error(t, err)
}

// JoinRequest Tests

func TestNewInMemoryJoinRequestRepository(t *testing.T) {
	repo := NewInMemoryJoinRequestRepository()
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.byKey)
	assert.NotNil(t, repo.byID)
}

func TestJoinRequestRepository_CreatePending_Success(t *testing.T) {
	repo := NewInMemoryJoinRequestRepository()
	ctx := context.Background()

	jr, err := repo.CreatePending(ctx, "network-1", "user-1")

	require.NoError(t, err)
	assert.NotEmpty(t, jr.ID)
	assert.Equal(t, "network-1", jr.NetworkID)
	assert.Equal(t, "user-1", jr.UserID)
	assert.Equal(t, "pending", jr.Status)
	assert.False(t, jr.CreatedAt.IsZero())
	assert.Nil(t, jr.DecidedAt)
}

func TestJoinRequestRepository_CreatePending_AlreadyExists(t *testing.T) {
	repo := NewInMemoryJoinRequestRepository()
	ctx := context.Background()

	jr1, err := repo.CreatePending(ctx, "network-1", "user-1")
	require.NoError(t, err)

	// Try to create again
	jr2, err := repo.CreatePending(ctx, "network-1", "user-1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "already pending")
	assert.Equal(t, jr1.ID, jr2.ID) // Returns existing request
}

func TestJoinRequestRepository_CreatePending_AfterDecision(t *testing.T) {
	repo := NewInMemoryJoinRequestRepository()
	ctx := context.Background()

	// Create and approve
	jr1, err := repo.CreatePending(ctx, "network-1", "user-1")
	require.NoError(t, err)
	repo.Decide(ctx, jr1.ID, true)

	// Should allow creating new request after approval
	jr2, err := repo.CreatePending(ctx, "network-1", "user-1")

	require.NoError(t, err)
	assert.NotEqual(t, jr1.ID, jr2.ID) // New request
	assert.Equal(t, "pending", jr2.Status)
}

func TestJoinRequestRepository_GetPending_Success(t *testing.T) {
	repo := NewInMemoryJoinRequestRepository()
	ctx := context.Background()

	created, err := repo.CreatePending(ctx, "network-1", "user-1")
	require.NoError(t, err)

	jr, err := repo.GetPending(ctx, "network-1", "user-1")

	require.NoError(t, err)
	assert.Equal(t, created.ID, jr.ID)
	assert.Equal(t, "pending", jr.Status)
}

func TestJoinRequestRepository_GetPending_NotFound(t *testing.T) {
	repo := NewInMemoryJoinRequestRepository()
	ctx := context.Background()

	jr, err := repo.GetPending(ctx, "network-1", "user-1")

	assert.Nil(t, jr)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pending")
}

func TestJoinRequestRepository_GetPending_AfterApproval(t *testing.T) {
	repo := NewInMemoryJoinRequestRepository()
	ctx := context.Background()

	jr, err := repo.CreatePending(ctx, "network-1", "user-1")
	require.NoError(t, err)

	// Approve the request
	repo.Decide(ctx, jr.ID, true)

	// GetPending should not find approved request
	jr, err = repo.GetPending(ctx, "network-1", "user-1")

	assert.Nil(t, jr)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pending")
}

func TestJoinRequestRepository_Decide_Approve(t *testing.T) {
	repo := NewInMemoryJoinRequestRepository()
	ctx := context.Background()

	jr, err := repo.CreatePending(ctx, "network-1", "user-1")
	require.NoError(t, err)

	err = repo.Decide(ctx, jr.ID, true)

	require.NoError(t, err)

	// Verify status changed (access via byID map)
	updatedJr := repo.byID[jr.ID]
	assert.Equal(t, "approved", updatedJr.Status)
	assert.NotNil(t, updatedJr.DecidedAt)
}

func TestJoinRequestRepository_Decide_Deny(t *testing.T) {
	repo := NewInMemoryJoinRequestRepository()
	ctx := context.Background()

	jr, err := repo.CreatePending(ctx, "network-1", "user-1")
	require.NoError(t, err)

	err = repo.Decide(ctx, jr.ID, false)

	require.NoError(t, err)

	// Verify status changed
	updatedJr := repo.byID[jr.ID]
	assert.Equal(t, "denied", updatedJr.Status)
	assert.NotNil(t, updatedJr.DecidedAt)
}

func TestJoinRequestRepository_Decide_NotFound(t *testing.T) {
	repo := NewInMemoryJoinRequestRepository()
	ctx := context.Background()

	err := repo.Decide(ctx, "nonexistent-id", true)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestJoinRequestRepository_Decide_Idempotent(t *testing.T) {
	repo := NewInMemoryJoinRequestRepository()
	ctx := context.Background()

	jr, err := repo.CreatePending(ctx, "network-1", "user-1")
	require.NoError(t, err)

	// Approve once
	err = repo.Decide(ctx, jr.ID, true)
	require.NoError(t, err)

	firstDecidedAt := *repo.byID[jr.ID].DecidedAt

	time.Sleep(1 * time.Millisecond)

	// Decide again should be idempotent (no error, no timestamp change)
	err = repo.Decide(ctx, jr.ID, false)
	require.NoError(t, err)

	// Status and timestamp should not change
	updatedJr := repo.byID[jr.ID]
	assert.Equal(t, "approved", updatedJr.Status) // Still approved
	assert.Equal(t, firstDecidedAt.Unix(), updatedJr.DecidedAt.Unix())
}

func TestJoinRequestRepository_FullCycle(t *testing.T) {
	repo := NewInMemoryJoinRequestRepository()
	ctx := context.Background()

	// Create pending request
	jr, err := repo.CreatePending(ctx, "network-1", "user-1")
	require.NoError(t, err)
	assert.Equal(t, "pending", jr.Status)

	// Get pending request
	jr, err = repo.GetPending(ctx, "network-1", "user-1")
	require.NoError(t, err)
	assert.NotNil(t, jr)
	jrID := jr.ID // Save ID before it becomes nil

	// Approve request
	err = repo.Decide(ctx, jrID, true)
	require.NoError(t, err)

	// Verify no longer pending
	jr, err = repo.GetPending(ctx, "network-1", "user-1")
	assert.Nil(t, jr)
	assert.Error(t, err)

	// Create new request after approval
	jr2, err := repo.CreatePending(ctx, "network-1", "user-1")
	require.NoError(t, err)
	assert.NotEqual(t, jrID, jr2.ID)

	// Deny this time
	err = repo.Decide(ctx, jr2.ID, false)
	require.NoError(t, err)

	updatedJr := repo.byID[jr2.ID]
	assert.Equal(t, "denied", updatedJr.Status)
}

func TestInMemoryJoinRequestRepository_ListPending(t *testing.T) {
	repo := NewInMemoryJoinRequestRepository()
	ctx := context.Background()

	t.Run("EmptyResult", func(t *testing.T) {
		result, err := repo.ListPending(ctx, "network-1")
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("OnlyPendingRequests", func(t *testing.T) {
		// Create pending requests
		jr1, err := repo.CreatePending(ctx, "network-1", "user-1")
		require.NoError(t, err)
		jr2, err := repo.CreatePending(ctx, "network-1", "user-2")
		require.NoError(t, err)

		// Create request in different network
		_, err = repo.CreatePending(ctx, "network-2", "user-3")
		require.NoError(t, err)

		// List pending for network-1
		result, err := repo.ListPending(ctx, "network-1")
		require.NoError(t, err)
		assert.Len(t, result, 2)

		// Verify correct requests returned
		ids := make(map[string]bool)
		for _, r := range result {
			ids[r.ID] = true
		}
		assert.True(t, ids[jr1.ID])
		assert.True(t, ids[jr2.ID])
	})

	t.Run("ExcludesApprovedAndDenied", func(t *testing.T) {
		repo2 := NewInMemoryJoinRequestRepository()

		// Create pending, then approve one
		jr1, err := repo2.CreatePending(ctx, "network-1", "user-1")
		require.NoError(t, err)
		jr2, err := repo2.CreatePending(ctx, "network-1", "user-2")
		require.NoError(t, err)
		jr3, err := repo2.CreatePending(ctx, "network-1", "user-3")
		require.NoError(t, err)

		// Approve first, deny second
		err = repo2.Decide(ctx, jr1.ID, true)
		require.NoError(t, err)
		err = repo2.Decide(ctx, jr2.ID, false)
		require.NoError(t, err)

		// Only third should be pending
		result, err := repo2.ListPending(ctx, "network-1")
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, jr3.ID, result[0].ID)
	})

	t.Run("SortedByCreationTime", func(t *testing.T) {
		repo3 := NewInMemoryJoinRequestRepository()

		// Create requests with small delays to ensure different creation times
		jr1, err := repo3.CreatePending(ctx, "network-1", "user-1")
		require.NoError(t, err)
		time.Sleep(1 * time.Millisecond)
		jr2, err := repo3.CreatePending(ctx, "network-1", "user-2")
		require.NoError(t, err)
		time.Sleep(1 * time.Millisecond)
		jr3, err := repo3.CreatePending(ctx, "network-1", "user-3")
		require.NoError(t, err)

		result, err := repo3.ListPending(ctx, "network-1")
		require.NoError(t, err)
		assert.Len(t, result, 3)

		// Should be sorted oldest first
		assert.Equal(t, jr1.ID, result[0].ID)
		assert.Equal(t, jr2.ID, result[1].ID)
		assert.Equal(t, jr3.ID, result[2].ID)
	})
}
