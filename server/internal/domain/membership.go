package domain

import "time"

// MembershipRole defines the role of a member in a network
type MembershipRole string

const (
	RoleMember MembershipRole = "member"
	RoleAdmin  MembershipRole = "admin"
	RoleOwner  MembershipRole = "owner"
)

// MembershipStatus defines membership status
type MembershipStatus string

const (
	StatusPending  MembershipStatus = "pending"
	StatusApproved MembershipStatus = "approved"
	StatusBanned   MembershipStatus = "banned"
)

// Membership represents memberships table
type Membership struct {
	ID        string           `json:"id" db:"id"`
	NetworkID string           `json:"network_id" db:"network_id"`
	UserID    string           `json:"user_id" db:"user_id"`
	Role      MembershipRole   `json:"role" db:"role"`
	Status    MembershipStatus `json:"status" db:"status"`
	JoinedAt  *time.Time       `json:"joined_at,omitempty" db:"joined_at"`
	CreatedAt time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt time.Time        `json:"updated_at" db:"updated_at"`
}

// JoinRequest represents join_requests table
type JoinRequest struct {
	ID        string     `json:"id" db:"id"`
	NetworkID string     `json:"network_id" db:"network_id"`
	UserID    string     `json:"user_id" db:"user_id"`
	Status    string     `json:"status" db:"status"` // pending|approved|denied
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	DecidedAt *time.Time `json:"decided_at,omitempty" db:"decided_at"`
}

// ListMembersRequest filters
type ListMembersRequest struct {
	Status string `form:"status"`
	Limit  int    `form:"limit,default=20"`
	Cursor string `form:"cursor"`
}
