// Package errors provides categorized error types with user-friendly messages
// for the GoConnect CLI daemon.
package errors

import (
	"errors"
	"fmt"
)

// Category represents the type/category of an error
type Category int

const (
	// CategoryUnknown is for uncategorized errors
	CategoryUnknown Category = iota
	// CategoryNetwork is for network-related errors (connection, timeout)
	CategoryNetwork
	// CategoryAuth is for authentication/authorization errors
	CategoryAuth
	// CategoryConfig is for configuration errors
	CategoryConfig
	// CategoryStorage is for storage/database errors
	CategoryStorage
	// CategoryValidation is for input validation errors
	CategoryValidation
	// CategoryNotFound is for resource not found errors
	CategoryNotFound
	// CategoryConflict is for resource conflict errors (duplicate, etc.)
	CategoryConflict
	// CategoryPermission is for permission denied errors
	CategoryPermission
	// CategoryInternal is for internal system errors
	CategoryInternal
	// CategoryUnavailable is for service unavailable errors
	CategoryUnavailable
)

// String returns the human-readable category name
func (c Category) String() string {
	switch c {
	case CategoryNetwork:
		return "Network"
	case CategoryAuth:
		return "Authentication"
	case CategoryConfig:
		return "Configuration"
	case CategoryStorage:
		return "Storage"
	case CategoryValidation:
		return "Validation"
	case CategoryNotFound:
		return "Not Found"
	case CategoryConflict:
		return "Conflict"
	case CategoryPermission:
		return "Permission"
	case CategoryInternal:
		return "Internal"
	case CategoryUnavailable:
		return "Unavailable"
	default:
		return "Unknown"
	}
}

// Error codes for specific error conditions
type Code string

const (
	// Network errors
	CodeConnectionFailed    Code = "CONNECTION_FAILED"
	CodeConnectionTimeout   Code = "CONNECTION_TIMEOUT"
	CodeDaemonUnavailable   Code = "DAEMON_UNAVAILABLE"
	CodeServerUnreachable   Code = "SERVER_UNREACHABLE"
	CodePeerUnreachable     Code = "PEER_UNREACHABLE"

	// Auth errors
	CodeInvalidToken        Code = "INVALID_TOKEN"
	CodeTokenExpired        Code = "TOKEN_EXPIRED"
	CodeUnauthorized        Code = "UNAUTHORIZED"
	CodeInvalidCredentials  Code = "INVALID_CREDENTIALS"

	// Config errors
	CodeConfigNotFound      Code = "CONFIG_NOT_FOUND"
	CodeConfigInvalid       Code = "CONFIG_INVALID"
	CodeConfigLoadFailed    Code = "CONFIG_LOAD_FAILED"
	CodeConfigSaveFailed    Code = "CONFIG_SAVE_FAILED"

	// Storage errors
	CodeDatabaseError       Code = "DATABASE_ERROR"
	CodeStorageUnavailable  Code = "STORAGE_UNAVAILABLE"
	CodeDataCorrupted       Code = "DATA_CORRUPTED"

	// Validation errors
	CodeInvalidInput        Code = "INVALID_INPUT"
	CodeMissingRequired     Code = "MISSING_REQUIRED"
	CodeInvalidFormat       Code = "INVALID_FORMAT"

	// Not found errors
	CodeNetworkNotFound     Code = "NETWORK_NOT_FOUND"
	CodePeerNotFound        Code = "PEER_NOT_FOUND"
	CodeFileNotFound        Code = "FILE_NOT_FOUND"
	CodeTransferNotFound    Code = "TRANSFER_NOT_FOUND"

	// Conflict errors
	CodeAlreadyExists       Code = "ALREADY_EXISTS"
	CodeAlreadyJoined       Code = "ALREADY_JOINED"
	CodeTransferInProgress  Code = "TRANSFER_IN_PROGRESS"

	// Permission errors
	CodeAccessDenied        Code = "ACCESS_DENIED"
	CodeNotOwner            Code = "NOT_OWNER"
	CodeBanned              Code = "BANNED"

	// Internal errors
	CodeInternalError       Code = "INTERNAL_ERROR"
	CodeUnexpected          Code = "UNEXPECTED"
)

// AppError represents a categorized application error with user-friendly message
type AppError struct {
	Code     Code
	Category Category
	Message  string      // User-friendly message
	Details  string      // Technical details for logging
	Cause    error       // Underlying error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error for errors.Is/As support
func (e *AppError) Unwrap() error {
	return e.Cause
}

// Is implements errors.Is for comparing error codes
func (e *AppError) Is(target error) bool {
	if t, ok := target.(*AppError); ok {
		return e.Code == t.Code
	}
	return false
}

// UserMessage returns a user-friendly message suitable for display
func (e *AppError) UserMessage() string {
	return e.Message
}

// LogMessage returns a detailed message suitable for logging
func (e *AppError) LogMessage() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// New creates a new AppError
func New(code Code, category Category, message string) *AppError {
	return &AppError{
		Code:     code,
		Category: category,
		Message:  message,
	}
}

// Wrap wraps an existing error with application context
func Wrap(err error, code Code, category Category, message string) *AppError {
	return &AppError{
		Code:     code,
		Category: category,
		Message:  message,
		Cause:    err,
	}
}

// WrapWithDetails wraps an error with additional details for logging
func WrapWithDetails(err error, code Code, category Category, message, details string) *AppError {
	return &AppError{
		Code:     code,
		Category: category,
		Message:  message,
		Details:  details,
		Cause:    err,
	}
}

// ============================================================================
// Convenience constructors for common error types
// ============================================================================

// Network errors

// ErrConnectionFailed creates a connection failed error
func ErrConnectionFailed(err error, target string) *AppError {
	return WrapWithDetails(err, CodeConnectionFailed, CategoryNetwork,
		"Failed to connect",
		fmt.Sprintf("target: %s", target))
}

// ErrDaemonUnavailable creates a daemon unavailable error
func ErrDaemonUnavailable(err error) *AppError {
	return Wrap(err, CodeDaemonUnavailable, CategoryUnavailable,
		"GoConnect daemon is not running. Start it with 'goconnect daemon start'")
}

// ErrServerUnreachable creates a server unreachable error
func ErrServerUnreachable(err error, serverURL string) *AppError {
	return WrapWithDetails(err, CodeServerUnreachable, CategoryNetwork,
		"Cannot reach the GoConnect server. Check your internet connection",
		fmt.Sprintf("server: %s", serverURL))
}

// ErrPeerUnreachable creates a peer unreachable error
func ErrPeerUnreachable(err error, peerID string) *AppError {
	return WrapWithDetails(err, CodePeerUnreachable, CategoryNetwork,
		"Cannot reach peer. They may be offline",
		fmt.Sprintf("peer: %s", peerID))
}

// Auth errors

// ErrUnauthorized creates an unauthorized error
func ErrUnauthorized(err error) *AppError {
	return Wrap(err, CodeUnauthorized, CategoryAuth,
		"You are not logged in. Use 'goconnect login' to authenticate")
}

// ErrInvalidToken creates an invalid token error
func ErrInvalidToken(err error) *AppError {
	return Wrap(err, CodeInvalidToken, CategoryAuth,
		"Your session has expired. Please log in again")
}

// ErrInvalidCredentials creates an invalid credentials error
func ErrInvalidCredentials(err error) *AppError {
	return Wrap(err, CodeInvalidCredentials, CategoryAuth,
		"Invalid username or password")
}

// Config errors

// ErrConfigNotFound creates a config not found error
func ErrConfigNotFound(path string) *AppError {
	return WrapWithDetails(nil, CodeConfigNotFound, CategoryConfig,
		"Configuration file not found. Run 'goconnect init' to create one",
		fmt.Sprintf("path: %s", path))
}

// ErrConfigInvalid creates a config invalid error
func ErrConfigInvalid(err error, path string) *AppError {
	return WrapWithDetails(err, CodeConfigInvalid, CategoryConfig,
		"Configuration file is invalid. Check the format and try again",
		fmt.Sprintf("path: %s", path))
}

// Storage errors

// ErrDatabaseError creates a database error
func ErrDatabaseError(err error, operation string) *AppError {
	return WrapWithDetails(err, CodeDatabaseError, CategoryStorage,
		"A database error occurred. Please try again",
		fmt.Sprintf("operation: %s", operation))
}

// Validation errors

// ErrInvalidInput creates an invalid input error
func ErrInvalidInput(field, reason string) *AppError {
	return &AppError{
		Code:     CodeInvalidInput,
		Category: CategoryValidation,
		Message:  fmt.Sprintf("Invalid %s: %s", field, reason),
		Details:  fmt.Sprintf("field: %s, reason: %s", field, reason),
	}
}

// ErrMissingRequired creates a missing required field error
func ErrMissingRequired(field string) *AppError {
	return &AppError{
		Code:     CodeMissingRequired,
		Category: CategoryValidation,
		Message:  fmt.Sprintf("%s is required", field),
		Details:  fmt.Sprintf("missing field: %s", field),
	}
}

// Not found errors

// ErrNetworkNotFound creates a network not found error
func ErrNetworkNotFound(networkID string) *AppError {
	return &AppError{
		Code:     CodeNetworkNotFound,
		Category: CategoryNotFound,
		Message:  "Network not found. Check the network ID or invite code",
		Details:  fmt.Sprintf("network: %s", networkID),
	}
}

// ErrPeerNotFound creates a peer not found error
func ErrPeerNotFound(peerID string) *AppError {
	return &AppError{
		Code:     CodePeerNotFound,
		Category: CategoryNotFound,
		Message:  "Peer not found in this network",
		Details:  fmt.Sprintf("peer: %s", peerID),
	}
}

// ErrTransferNotFound creates a transfer not found error
func ErrTransferNotFound(transferID string) *AppError {
	return &AppError{
		Code:     CodeTransferNotFound,
		Category: CategoryNotFound,
		Message:  "Transfer not found or already completed",
		Details:  fmt.Sprintf("transfer: %s", transferID),
	}
}

// Conflict errors

// ErrAlreadyJoined creates an already joined error
func ErrAlreadyJoined(networkName string) *AppError {
	return &AppError{
		Code:     CodeAlreadyJoined,
		Category: CategoryConflict,
		Message:  fmt.Sprintf("You are already a member of '%s'", networkName),
		Details:  fmt.Sprintf("network: %s", networkName),
	}
}

// ErrTransferInProgress creates a transfer in progress error
func ErrTransferInProgress(transferID string) *AppError {
	return &AppError{
		Code:     CodeTransferInProgress,
		Category: CategoryConflict,
		Message:  "A transfer is already in progress for this file",
		Details:  fmt.Sprintf("transfer: %s", transferID),
	}
}

// Permission errors

// ErrAccessDenied creates an access denied error
func ErrAccessDenied(action string) *AppError {
	return &AppError{
		Code:     CodeAccessDenied,
		Category: CategoryPermission,
		Message:  fmt.Sprintf("You don't have permission to %s", action),
		Details:  fmt.Sprintf("action: %s", action),
	}
}

// ErrBanned creates a banned error
func ErrBanned(networkName string) *AppError {
	return &AppError{
		Code:     CodeBanned,
		Category: CategoryPermission,
		Message:  fmt.Sprintf("You have been banned from '%s'", networkName),
		Details:  fmt.Sprintf("network: %s", networkName),
	}
}

// ============================================================================
// Error inspection helpers
// ============================================================================

// IsCategory checks if an error belongs to a specific category
func IsCategory(err error, category Category) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Category == category
	}
	return false
}

// IsCode checks if an error has a specific error code
func IsCode(err error, code Code) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

// GetCode extracts the error code from an error
func GetCode(err error) Code {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return ""
}

// GetUserMessage extracts a user-friendly message from an error
func GetUserMessage(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.UserMessage()
	}
	return err.Error()
}

// ToAppError converts any error to an AppError
func ToAppError(err error) *AppError {
	if err == nil {
		return nil
	}
	
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	
	// Wrap unknown errors as internal errors
	return Wrap(err, CodeUnexpected, CategoryInternal,
		"An unexpected error occurred")
}
