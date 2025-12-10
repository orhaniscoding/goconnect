package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ==================== DeletionRequestStatus Tests ====================

func TestDeletionRequestStatus_Constants(t *testing.T) {
	t.Run("All Statuses Are Unique", func(t *testing.T) {
		statuses := []DeletionRequestStatus{
			DeletionRequestStatusPending,
			DeletionRequestStatusProcessing,
			DeletionRequestStatusCompleted,
			DeletionRequestStatusFailed,
		}

		seen := make(map[DeletionRequestStatus]bool)
		for _, status := range statuses {
			assert.False(t, seen[status], "Duplicate status: %s", status)
			seen[status] = true
		}
	})

	t.Run("Status Values Are Correct", func(t *testing.T) {
		assert.Equal(t, DeletionRequestStatus("pending"), DeletionRequestStatusPending)
		assert.Equal(t, DeletionRequestStatus("processing"), DeletionRequestStatusProcessing)
		assert.Equal(t, DeletionRequestStatus("completed"), DeletionRequestStatusCompleted)
		assert.Equal(t, DeletionRequestStatus("failed"), DeletionRequestStatusFailed)
	})
}

// ==================== DeletionRequest Tests ====================

func TestDeletionRequest(t *testing.T) {
	t.Run("Has All Fields", func(t *testing.T) {
		now := time.Now()
		completed := now.Add(time.Hour)
		req := DeletionRequest{
			ID:          "del123",
			UserID:      "user123",
			Status:      DeletionRequestStatusPending,
			RequestedAt: now,
			CompletedAt: &completed,
			Error:       "",
		}

		assert.Equal(t, "del123", req.ID)
		assert.Equal(t, "user123", req.UserID)
		assert.Equal(t, DeletionRequestStatusPending, req.Status)
		assert.NotNil(t, req.CompletedAt)
	})

	t.Run("Optional Fields Can Be Nil", func(t *testing.T) {
		req := DeletionRequest{
			ID:          "del123",
			UserID:      "user123",
			Status:      DeletionRequestStatusPending,
			RequestedAt: time.Now(),
			CompletedAt: nil,
			Error:       "",
		}

		assert.Nil(t, req.CompletedAt)
	})

	t.Run("Error Can Be Set", func(t *testing.T) {
		req := DeletionRequest{
			ID:          "del123",
			UserID:      "user123",
			Status:      DeletionRequestStatusFailed,
			RequestedAt: time.Now(),
			Error:       "Failed to delete user data",
		}

		assert.Equal(t, "Failed to delete user data", req.Error)
	})
}
