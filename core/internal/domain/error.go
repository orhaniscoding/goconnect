package domain

import (
	"encoding/json"
	"net/http"
)

// Error represents the standard error response format
type Error struct {
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	RetryAfter int         `json:"retry_after,omitempty"`
}

// Error implements the error interface
func (e *Error) Error() string {
	return e.Message
}

// Error codes following ERR_SNAKE_CASE convention
const (
	ErrInvalidRequest      = "ERR_INVALID_REQUEST"
	ErrValidation          = "ERR_VALIDATION"
	ErrUnauthorized        = "ERR_UNAUTHORIZED"
	ErrNotAuthorized       = "ERR_NOT_AUTHORIZED" // outward unified code for authz failures
	ErrForbidden           = "ERR_FORBIDDEN"
	ErrRateLimited         = "ERR_RATE_LIMITED"
	ErrNotFound            = "ERR_NOT_FOUND"
	ErrCIDROverlap         = "ERR_CIDR_OVERLAP"
	ErrCIDRInvalid         = "ERR_CIDR_INVALID"
	ErrIdempotencyConflict = "ERR_IDEMPOTENCY_CONFLICT"
	ErrConflict            = "ERR_CONFLICT" // Generic conflict (duplicate resource)
	ErrInternalServer      = "ERR_INTERNAL_SERVER"
	ErrNotImplemented      = "ERR_NOT_IMPLEMENTED"
	// Membership/Join flow specific
	ErrNetworkPrivate   = "ERR_NETWORK_PRIVATE"
	ErrAlreadyMember    = "ERR_ALREADY_MEMBER"
	ErrAlreadyRequested = "ERR_ALREADY_REQUESTED"
	ErrUserBanned       = "ERR_USER_BANNED"
	ErrUserKicked       = "ERR_USER_KICKED"
	// Authentication/User specific
	ErrInvalidCredentials = "ERR_INVALID_CREDENTIALS"  // Wrong email/password
	ErrEmailAlreadyExists = "ERR_EMAIL_ALREADY_EXISTS" // Email already registered (alias for ErrUserAlreadyExists)
	ErrUserAlreadyExists  = "ERR_USER_ALREADY_EXISTS"  // Email already registered
	ErrInvalidToken       = "ERR_INVALID_TOKEN"        // Malformed or invalid JWT
	ErrTokenExpired       = "ERR_TOKEN_EXPIRED"        // JWT expired
	ErrUserNotFound       = "ERR_USER_NOT_FOUND"       // User ID not found
	ErrTenantNotFound     = "ERR_TENANT_NOT_FOUND"     // Tenant ID not found
	ErrWeakPassword       = "ERR_WEAK_PASSWORD"        // Password too weak
	ErrSessionExpired     = "ERR_SESSION_EXPIRED"      // Session expired
	ErrRefreshTokenReuse  = "ERR_REFRESH_TOKEN_REUSE"  // Refresh token reuse detected
	// Invite token specific
	ErrInviteTokenExpired  = "ERR_INVITE_TOKEN_EXPIRED"   // Invite token expired or exhausted
	ErrInviteTokenNotFound = "ERR_INVITE_TOKEN_NOT_FOUND" // Invite token not found
	ErrInviteTokenRevoked  = "ERR_INVITE_TOKEN_REVOKED"   // Invite token was revoked
)

// NewError creates a new domain error
func NewError(code, message string, details interface{}) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// ToHTTPStatus maps domain error codes to HTTP status codes
func (e *Error) ToHTTPStatus() int {
	switch e.Code {
	case ErrInvalidRequest, ErrCIDRInvalid, ErrValidation, ErrWeakPassword:
		return http.StatusBadRequest
	case ErrUnauthorized, ErrInvalidToken, ErrTokenExpired, ErrInvalidCredentials:
		return http.StatusUnauthorized
	case ErrNotAuthorized:
		return http.StatusForbidden
	case ErrForbidden:
		return http.StatusForbidden
	case ErrRateLimited:
		return http.StatusTooManyRequests
	case ErrNotFound, ErrUserNotFound, ErrTenantNotFound:
		return http.StatusNotFound
	case ErrCIDROverlap, ErrIdempotencyConflict, ErrIPExhausted, ErrUserAlreadyExists, ErrEmailAlreadyExists:
		return http.StatusConflict
	case ErrNetworkPrivate:
		return http.StatusNotFound // hide private resource existence
	case ErrAlreadyMember:
		return http.StatusOK
	case ErrAlreadyRequested:
		return http.StatusAccepted
	case ErrUserBanned, ErrUserKicked:
		return http.StatusForbidden
	case ErrNotImplemented:
		return http.StatusNotImplemented
	default:
		return http.StatusInternalServerError
	}
}

// ToJSON converts error to JSON response
func (e *Error) ToJSON() []byte {
	data, _ := json.Marshal(e)
	return data
}
