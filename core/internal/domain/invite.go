package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"
)

// InviteToken represents a network invitation token
type InviteToken struct {
	ID        string     `json:"id" db:"id"`
	NetworkID string     `json:"network_id" db:"network_id"`
	TenantID  string     `json:"tenant_id" db:"tenant_id"`
	Token     string     `json:"token" db:"token"`
	CreatedBy string     `json:"created_by" db:"created_by"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	UsesMax   int        `json:"uses_max" db:"uses_max"`   // 0 = unlimited
	UsesLeft  int        `json:"uses_left" db:"uses_left"` // decrements on use
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
}

// CreateInviteRequest represents the request to create an invite token
type CreateInviteRequest struct {
	ExpiresIn int `json:"expires_in" binding:"omitempty,min=60,max=604800"` // seconds, 1 min to 7 days
	UsesMax   int `json:"uses_max" binding:"omitempty,min=0,max=100"`       // 0 = unlimited
}

// InviteTokenResponse represents the response when creating/viewing an invite
type InviteTokenResponse struct {
	ID        string    `json:"id"`
	NetworkID string    `json:"network_id"`
	Token     string    `json:"token"`
	InviteURL string    `json:"invite_url"`
	ExpiresAt time.Time `json:"expires_at"`
	UsesMax   int       `json:"uses_max"`
	UsesLeft  int       `json:"uses_left"`
	CreatedAt time.Time `json:"created_at"`
	IsActive  bool      `json:"is_active"`
}

// IsValid checks if the invite token is still valid
func (t *InviteToken) IsValid() bool {
	if t.RevokedAt != nil {
		return false
	}
	if time.Now().After(t.ExpiresAt) {
		return false
	}
	if t.UsesMax > 0 && t.UsesLeft <= 0 {
		return false
	}
	return true
}

// DecrementUse decrements the uses left counter
func (t *InviteToken) DecrementUse() error {
	if !t.IsValid() {
		return NewError(ErrInviteTokenExpired, "Invite token is no longer valid", nil)
	}
	if t.UsesMax > 0 {
		t.UsesLeft--
	}
	return nil
}

// GenerateInviteToken generates a secure random token
func GenerateInviteToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateInviteID generates a new ID for invite tokens
func GenerateInviteID() string {
	timestamp := time.Now().Unix()
	n, _ := rand.Int(rand.Reader, big.NewInt(999999))
	return fmt.Sprintf("inv_%d_%06d", timestamp, n.Int64())
}
