package domain

import "time"

// UserListItem represents a user in admin user list
type UserListItem struct {
	ID          string     `json:"id" db:"id"`
	Email       string     `json:"email" db:"email"`
	Username    *string    `json:"username" db:"username"`
	TenantID    string     `json:"tenant_id" db:"tenant_id"`
	IsAdmin     bool       `json:"is_admin" db:"is_admin"`
	IsModerator bool       `json:"is_moderator" db:"is_moderator"`
	Suspended   bool       `json:"suspended" db:"suspended"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	LastSeen    *time.Time `json:"last_seen,omitempty" db:"last_seen"`
}

// UpdateUserRoleRequest represents a request to update user roles
type UpdateUserRoleRequest struct {
	UserID      string `json:"user_id" binding:"required"`
	IsAdmin     *bool  `json:"is_admin,omitempty"`
	IsModerator *bool  `json:"is_moderator,omitempty"`
}

// SuspendUserRequest represents a request to suspend a user
type SuspendUserRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Reason string `json:"reason" binding:"required,min=10,max=500"`
}

// Validate validates the suspension request
func (r *SuspendUserRequest) Validate() error {
	if len(r.Reason) < 10 {
		return NewError(ErrInvalidRequest, "Suspension reason must be at least 10 characters", nil)
	}
	if len(r.Reason) > 500 {
		return NewError(ErrInvalidRequest, "Suspension reason cannot exceed 500 characters", nil)
	}
	return nil
}

// UnsuspendUserRequest represents a request to unsuspend a user
type UnsuspendUserRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// SystemStats represents system-wide statistics
type SystemStats struct {
	TotalUsers     int       `json:"total_users"`
	TotalTenants   int       `json:"total_tenants"`
	TotalNetworks  int       `json:"total_networks"`
	TotalDevices   int       `json:"total_devices"`
	ActivePeers    int       `json:"active_peers"`
	AdminUsers     int       `json:"admin_users"`
	ModeratorUsers int       `json:"moderator_users"`
	SuspendedUsers int       `json:"suspended_users"`
	LastUpdated    time.Time `json:"last_updated"`
}

// UserActivity represents user activity log entry
type UserActivity struct {
	ID        int64     `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Action    string    `json:"action" db:"action"`
	Resource  string    `json:"resource" db:"resource"`
	Details   string    `json:"details" db:"details"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// UserFilters represents filters for user list queries
type UserFilters struct {
	Role     string // "admin", "moderator", "user"
	Status   string // "active", "suspended"
	TenantID string
	Search   string // Email or username search
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page    int
	PerPage int
}
