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
	ErrUnauthorized        = "ERR_UNAUTHORIZED"
	ErrNotAuthorized       = "ERR_NOT_AUTHORIZED" // outward unified code for authz failures
	ErrForbidden           = "ERR_FORBIDDEN"
	ErrRateLimited         = "ERR_RATE_LIMITED"
	ErrNotFound            = "ERR_NOT_FOUND"
	ErrCIDROverlap         = "ERR_CIDR_OVERLAP"
	ErrCIDRInvalid         = "ERR_CIDR_INVALID"
	ErrIdempotencyConflict = "ERR_IDEMPOTENCY_CONFLICT"
	ErrInternalServer      = "ERR_INTERNAL_SERVER"
	ErrNotImplemented      = "ERR_NOT_IMPLEMENTED"
	// Membership/Join flow specific
	ErrNetworkPrivate   = "ERR_NETWORK_PRIVATE"
	ErrAlreadyMember    = "ERR_ALREADY_MEMBER"
	ErrAlreadyRequested = "ERR_ALREADY_REQUESTED"
	ErrUserBanned       = "ERR_USER_BANNED"
	ErrUserKicked       = "ERR_USER_KICKED"
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
	case ErrInvalidRequest, ErrCIDRInvalid:
		return http.StatusBadRequest
	case ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrNotAuthorized:
		return http.StatusForbidden
	case ErrForbidden:
		return http.StatusForbidden
	case ErrRateLimited:
		return http.StatusTooManyRequests
	case ErrNotFound:
		return http.StatusNotFound
	case ErrCIDROverlap, ErrIdempotencyConflict, ErrIPExhausted:
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
