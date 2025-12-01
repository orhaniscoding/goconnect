package domain

import "time"

// DeletionRequestStatus represents the status of a deletion request
type DeletionRequestStatus string

const (
	DeletionRequestStatusPending    DeletionRequestStatus = "pending"
	DeletionRequestStatusProcessing DeletionRequestStatus = "processing"
	DeletionRequestStatusCompleted  DeletionRequestStatus = "completed"
	DeletionRequestStatusFailed     DeletionRequestStatus = "failed"
)

// DeletionRequest represents a user deletion request
type DeletionRequest struct {
	ID          string                `json:"id" db:"id"`
	UserID      string                `json:"user_id" db:"user_id"`
	Status      DeletionRequestStatus `json:"status" db:"status"`
	RequestedAt time.Time             `json:"requested_at" db:"requested_at"`
	CompletedAt *time.Time            `json:"completed_at,omitempty" db:"completed_at"`
	Error       string                `json:"error,omitempty" db:"error"`
}
