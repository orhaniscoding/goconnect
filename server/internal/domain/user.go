package domain

import "time"

// User represents a user in the system
type User struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never expose in JSON
	Locale       string    `json:"locale"`
	TwoFAEnabled bool      `json:"two_fa_enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Tenant represents a tenant/organization
type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Locale   string `json:"locale" binding:"omitempty,oneof=tr en"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds
	TokenType    string `json:"token_type"`
	User         *User  `json:"user"`
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"is_admin"`
	Exp      int64  `json:"exp"`
	Iat      int64  `json:"iat"`
}
