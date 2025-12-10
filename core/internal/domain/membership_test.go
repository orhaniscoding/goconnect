package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ==================== MembershipRole Tests ====================

func TestMembershipRole_Constants(t *testing.T) {
	t.Run("Roles Are Correct", func(t *testing.T) {
		assert.Equal(t, MembershipRole("member"), RoleMember)
		assert.Equal(t, MembershipRole("admin"), RoleAdmin)
		assert.Equal(t, MembershipRole("owner"), RoleOwner)
	})

	t.Run("Roles Are Unique", func(t *testing.T) {
		roles := []MembershipRole{RoleMember, RoleAdmin, RoleOwner}
		seen := make(map[MembershipRole]bool)
		for _, role := range roles {
			assert.False(t, seen[role])
			seen[role] = true
		}
	})
}

// ==================== MembershipStatus Tests ====================

func TestMembershipStatus_Constants(t *testing.T) {
	t.Run("Statuses Are Correct", func(t *testing.T) {
		assert.Equal(t, MembershipStatus("pending"), StatusPending)
		assert.Equal(t, MembershipStatus("approved"), StatusApproved)
		assert.Equal(t, MembershipStatus("banned"), StatusBanned)
	})

	t.Run("Statuses Are Unique", func(t *testing.T) {
		statuses := []MembershipStatus{StatusPending, StatusApproved, StatusBanned}
		seen := make(map[MembershipStatus]bool)
		for _, status := range statuses {
			assert.False(t, seen[status])
			seen[status] = true
		}
	})
}

// ==================== Membership Tests ====================

func TestMembership(t *testing.T) {
	t.Run("Has All Fields", func(t *testing.T) {
		now := time.Now()
		joined := now.Add(-time.Hour)
		membership := Membership{
			ID:                "mem123",
			NetworkID:         "net123",
			UserID:            "user123",
			Role:              RoleMember,
			Status:            StatusApproved,
			JoinedAt:          &joined,
			CreatedAt:         now,
			UpdatedAt:         now,
			OnlineDeviceCount: 2,
		}

		assert.Equal(t, "mem123", membership.ID)
		assert.Equal(t, RoleMember, membership.Role)
		assert.Equal(t, StatusApproved, membership.Status)
		assert.Equal(t, 2, membership.OnlineDeviceCount)
	})

	t.Run("JoinedAt Can Be Nil", func(t *testing.T) {
		membership := Membership{
			ID:        "mem123",
			NetworkID: "net123",
			UserID:    "user123",
			Role:      RoleMember,
			Status:    StatusPending,
			JoinedAt:  nil,
		}

		assert.Nil(t, membership.JoinedAt)
	})
}

// ==================== JoinRequest Tests ====================

func TestJoinRequest(t *testing.T) {
	t.Run("Has All Fields", func(t *testing.T) {
		now := time.Now()
		decided := now.Add(time.Hour)
		req := JoinRequest{
			ID:        "jr123",
			NetworkID: "net123",
			UserID:    "user123",
			Status:    "pending",
			CreatedAt: now,
			DecidedAt: &decided,
		}

		assert.Equal(t, "jr123", req.ID)
		assert.Equal(t, "pending", req.Status)
		assert.NotNil(t, req.DecidedAt)
	})

	t.Run("DecidedAt Can Be Nil", func(t *testing.T) {
		req := JoinRequest{
			ID:        "jr123",
			NetworkID: "net123",
			UserID:    "user123",
			Status:    "pending",
			CreatedAt: time.Now(),
			DecidedAt: nil,
		}

		assert.Nil(t, req.DecidedAt)
	})
}

// ==================== ListMembersRequest Tests ====================

func TestListMembersRequest(t *testing.T) {
	t.Run("Has All Fields", func(t *testing.T) {
		req := ListMembersRequest{
			Status: "approved",
			Limit:  50,
			Cursor: "abc123",
		}

		assert.Equal(t, "approved", req.Status)
		assert.Equal(t, 50, req.Limit)
		assert.Equal(t, "abc123", req.Cursor)
	})

	t.Run("Default Values", func(t *testing.T) {
		req := ListMembersRequest{}
		assert.Equal(t, "", req.Status)
		assert.Equal(t, 0, req.Limit)
		assert.Equal(t, "", req.Cursor)
	})
}
