package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== SuspendUserRequest Tests ====================

func TestSuspendUserRequest_Validate(t *testing.T) {
	t.Run("Valid Request", func(t *testing.T) {
		req := &SuspendUserRequest{
			UserID: "user123",
			Reason: "Violated terms of service repeatedly",
		}
		err := req.Validate()
		assert.NoError(t, err)
	})

	t.Run("Reason Too Short", func(t *testing.T) {
		req := &SuspendUserRequest{
			UserID: "user123",
			Reason: "short",
		}
		err := req.Validate()
		require.Error(t, err)
		domErr, ok := err.(*Error)
		require.True(t, ok)
		assert.Equal(t, ErrInvalidRequest, domErr.Code)
		assert.Contains(t, domErr.Message, "at least 10 characters")
	})

	t.Run("Reason Too Long", func(t *testing.T) {
		req := &SuspendUserRequest{
			UserID: "user123",
			Reason: string(make([]byte, 501)), // 501 characters
		}
		err := req.Validate()
		require.Error(t, err)
		domErr, ok := err.(*Error)
		require.True(t, ok)
		assert.Contains(t, domErr.Message, "cannot exceed 500")
	})

	t.Run("Reason Exactly 10 Characters", func(t *testing.T) {
		req := &SuspendUserRequest{
			UserID: "user123",
			Reason: "1234567890", // exactly 10
		}
		err := req.Validate()
		assert.NoError(t, err)
	})

	t.Run("Reason Exactly 500 Characters", func(t *testing.T) {
		req := &SuspendUserRequest{
			UserID: "user123",
			Reason: string(make([]byte, 500)), // exactly 500
		}
		err := req.Validate()
		assert.NoError(t, err)
	})
}

// ==================== Struct Tests ====================

func TestUserListItem(t *testing.T) {
	t.Run("Has All Fields", func(t *testing.T) {
		now := time.Now()
		username := "testuser"
		item := UserListItem{
			ID:          "user123",
			Email:       "test@example.com",
			Username:    &username,
			TenantID:    "tenant123",
			IsAdmin:     true,
			IsModerator: false,
			Suspended:   false,
			CreatedAt:   now,
			LastSeen:    &now,
		}

		assert.Equal(t, "user123", item.ID)
		assert.Equal(t, "test@example.com", item.Email)
		assert.Equal(t, "testuser", *item.Username)
		assert.True(t, item.IsAdmin)
	})
}

func TestUpdateUserRoleRequest(t *testing.T) {
	t.Run("Has All Fields", func(t *testing.T) {
		isAdmin := true
		isModerator := false
		req := UpdateUserRoleRequest{
			UserID:      "user123",
			IsAdmin:     &isAdmin,
			IsModerator: &isModerator,
		}

		assert.Equal(t, "user123", req.UserID)
		assert.True(t, *req.IsAdmin)
		assert.False(t, *req.IsModerator)
	})
}

func TestSystemStats(t *testing.T) {
	t.Run("Has All Fields", func(t *testing.T) {
		stats := SystemStats{
			TotalUsers:     100,
			TotalTenants:   5,
			TotalNetworks:  20,
			TotalDevices:   150,
			ActivePeers:    75,
			AdminUsers:     3,
			ModeratorUsers: 10,
			SuspendedUsers: 2,
			LastUpdated:    time.Now(),
		}

		assert.Equal(t, 100, stats.TotalUsers)
		assert.Equal(t, 5, stats.TotalTenants)
		assert.Equal(t, 20, stats.TotalNetworks)
	})
}

func TestUserActivity(t *testing.T) {
	t.Run("Has All Fields", func(t *testing.T) {
		activity := UserActivity{
			ID:        1,
			UserID:    "user123",
			Action:    "login",
			Resource:  "session",
			Details:   "Successful login",
			IPAddress: "192.168.1.1",
			CreatedAt: time.Now(),
		}

		assert.Equal(t, int64(1), activity.ID)
		assert.Equal(t, "login", activity.Action)
	})
}

func TestUserFilters(t *testing.T) {
	t.Run("Has All Fields", func(t *testing.T) {
		filters := UserFilters{
			Role:     "admin",
			Status:   "active",
			TenantID: "tenant123",
			Search:   "test@example.com",
		}

		assert.Equal(t, "admin", filters.Role)
		assert.Equal(t, "active", filters.Status)
	})
}

func TestPaginationParams(t *testing.T) {
	t.Run("Has All Fields", func(t *testing.T) {
		params := PaginationParams{
			Page:    1,
			PerPage: 20,
		}

		assert.Equal(t, 1, params.Page)
		assert.Equal(t, 20, params.PerPage)
	})
}
