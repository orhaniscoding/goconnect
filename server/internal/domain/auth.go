package domain

import "time"

// User represents a user in the system
type User struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`            // Never expose in JSON
	Locale       string    `json:"locale"`       // "tr" or "en"
	IsAdmin      bool      `json:"is_admin"`     // Global admin flag
	IsModerator  bool      `json:"is_moderator"` // Can moderate chat/content
	TwoFAKey     string    `json:"-"`            // TOTP secret (future)
	TwoFAEnabled bool      `json:"two_fa_enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Tenant represents a tenant (organization) in the system
type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	OwnerID   string    `json:"owner_id"` // User ID of the owner
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RegisterRequest is the request body for user registration
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"` // Argon2id max 72 bytes
	Locale   string `json:"locale" binding:"omitempty,oneof=tr en"`
}

// LoginRequest is the request body for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest is the request body for token refresh
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AuthResponse is the response for successful authentication
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds
	TokenType    string `json:"token_type"` // "Bearer"
	User         *User  `json:"user"`
}

// TokenClaims represents validated JWT token claims
type TokenClaims struct {
	UserID      string `json:"user_id"`
	TenantID    string `json:"tenant_id"`
	Email       string `json:"email"`
	IsAdmin     bool   `json:"is_admin"`
	IsModerator bool   `json:"is_moderator"`
	Type        string `json:"type"` // "access" or "refresh"
	Exp         int64  `json:"exp"`  // Unix timestamp - expiration
	Iat         int64  `json:"iat"`  // Unix timestamp - issued at
}

// Sanitize removes sensitive fields for logging
func (u *User) Sanitize() *User {
	return &User{
		ID:          u.ID,
		TenantID:    u.TenantID,
		Email:       u.Email,
		Locale:      u.Locale,
		IsAdmin:     u.IsAdmin,
		IsModerator: u.IsModerator,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}
