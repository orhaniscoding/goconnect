package domain

// Auth-related error codes
// These are error code constants, not actual credentials
// #nosec G101 - False positive: these are not hardcoded credentials, just error code strings
const (
	ErrInvalidCredentials = "ERR_INVALID_CREDENTIALS"
	ErrEmailAlreadyExists = "ERR_EMAIL_ALREADY_EXISTS"
	ErrUserNotFound       = "ERR_USER_NOT_FOUND"
	ErrInvalidToken       = "ERR_INVALID_TOKEN"
	ErrTokenExpired       = "ERR_TOKEN_EXPIRED"
	ErrWeakPassword       = "ERR_WEAK_PASSWORD"
	ErrTenantNotFound     = "ERR_TENANT_NOT_FOUND"
	ErrSessionExpired     = "ERR_SESSION_EXPIRED"
	ErrRefreshTokenReuse  = "ERR_REFRESH_TOKEN_REUSE"
)
