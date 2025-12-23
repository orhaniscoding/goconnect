package domain

import "time"

// User represents a user in the system
type User struct {
	ID              string     `json:"id"`
	TenantID        string     `json:"tenant_id"`
	Email           string     `json:"email"`
	PasswordHash    string     `json:"-"`            // Never expose in JSON
	Username        *string    `json:"username"`     // Username for display
	FullName        *string    `json:"full_name"`    // Full name
	Bio             *string    `json:"bio"`          // Bio/description
	AvatarURL       *string    `json:"avatar_url"`   // Avatar image URL
	Locale          string     `json:"locale"`       // "tr" or "en"
	IsAdmin         bool       `json:"is_admin"`     // Global admin flag
	IsModerator     bool       `json:"is_moderator"` // Can moderate chat/content
	TwoFAKey        string     `json:"-"`            // TOTP secret
	TwoFAEnabled    bool       `json:"two_fa_enabled"`
	RecoveryCodes   []string   `json:"-"`                // Hashed recovery codes (8 codes, one-time use)
	AuthProvider    string     `json:"auth_provider"`    // "local", "google", "github", "oidc"
	ExternalID      string     `json:"external_id"`      // ID from the provider
	Suspended       bool       `json:"suspended"`        // Whether the user account is suspended
	SuspendedAt     *time.Time `json:"suspended_at"`     // When the user was suspended
	SuspendedReason *string    `json:"suspended_reason"` // Reason for suspension
	SuspendedBy     *string    `json:"suspended_by"`     // User ID of the admin who suspended this user
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// Tenant represents a tenant (organization) in the system
type Tenant struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	IconURL      string           `json:"icon_url"`
	Visibility   TenantVisibility `json:"visibility"`
	AccessType   TenantAccessType `json:"access_type"`
	PasswordHash string           `json:"-"`
	MaxMembers   int              `json:"max_members"`
	OwnerID      string           `json:"owner_id"` // User ID of the owner
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
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
	Code     string `json:"code"` // TOTP Code (optional, required if 2FA enabled)
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

// Enable2FARequest is the request body for enabling 2FA
type Enable2FARequest struct {
	Secret string `json:"secret" binding:"required"`
	Code   string `json:"code" binding:"required,len=6"`
}

// Disable2FARequest is the request body for disabling 2FA
type Disable2FARequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

// RecoveryCodeResponse is the response when generating recovery codes
type RecoveryCodeResponse struct {
	Codes []string `json:"codes"` // 8 plaintext codes shown once to user
}

// UseRecoveryCodeRequest is the request body for using a recovery code to login
type UseRecoveryCodeRequest struct {
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required"`
	RecoveryCode string `json:"recovery_code" binding:"required,len=10"`
}

// RegenerateRecoveryCodesRequest is the request body for regenerating recovery codes
type RegenerateRecoveryCodesRequest struct {
	Code string `json:"code" binding:"required,len=6"` // Current TOTP code to verify
}

// DeviceCodeRequest is the request to initiate device flow from a client
type DeviceCodeRequest struct {
	ClientID string `json:"client_id"` // Optional: identifying the client type (cli, desktop, etc)
}

// DeviceCodeResponse matches the OAuth2 Device Authorization Response (RFC 8628)
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"` // seconds
	Interval        int    `json:"interval"`   // seconds
}

// DeviceTokenRequest is used by the client to poll for the token using the device code
type DeviceTokenRequest struct {
	DeviceCode string `json:"device_code" binding:"required"`
}

// DeviceVerifyRequest is used by the authenticated user to approve the device flow
type DeviceVerifyRequest struct {
	UserCode string `json:"user_code" binding:"required"`
}
