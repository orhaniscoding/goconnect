package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestCategory_String(t *testing.T) {
	tests := []struct {
		category Category
		expected string
	}{
		{CategoryUnknown, "Unknown"},
		{CategoryNetwork, "Network"},
		{CategoryAuth, "Authentication"},
		{CategoryConfig, "Configuration"},
		{CategoryStorage, "Storage"},
		{CategoryValidation, "Validation"},
		{CategoryNotFound, "Not Found"},
		{CategoryConflict, "Conflict"},
		{CategoryPermission, "Permission"},
		{CategoryInternal, "Internal"},
		{CategoryUnavailable, "Unavailable"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.category.String(); got != tt.expected {
				t.Errorf("Category.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAppError_Error(t *testing.T) {
	// Without cause
	err := New(CodeInvalidInput, CategoryValidation, "Invalid input")
	if got := err.Error(); got != "Invalid input" {
		t.Errorf("Error() = %v, want %v", got, "Invalid input")
	}

	// With cause
	cause := fmt.Errorf("underlying error")
	err = Wrap(cause, CodeInvalidInput, CategoryValidation, "Invalid input")
	expected := "Invalid input: underlying error"
	if got := err.Error(); got != expected {
		t.Errorf("Error() = %v, want %v", got, expected)
	}
}

func TestAppError_Unwrap(t *testing.T) {
	cause := fmt.Errorf("underlying error")
	err := Wrap(cause, CodeInternalError, CategoryInternal, "Wrapped error")

	if unwrapped := err.Unwrap(); unwrapped != cause {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
	}

	// Test with errors.Unwrap
	if unwrapped := errors.Unwrap(err); unwrapped != cause {
		t.Errorf("errors.Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestAppError_Is(t *testing.T) {
	err1 := New(CodeInvalidInput, CategoryValidation, "Error 1")
	err2 := New(CodeInvalidInput, CategoryValidation, "Error 2")
	err3 := New(CodeInternalError, CategoryInternal, "Error 3")

	if !err1.Is(err2) {
		t.Error("Expected errors with same code to match")
	}
	if err1.Is(err3) {
		t.Error("Expected errors with different codes to not match")
	}

	// Test with errors.Is
	if !errors.Is(err1, err2) {
		t.Error("errors.Is should match errors with same code")
	}
}

func TestAppError_UserMessage(t *testing.T) {
	err := New(CodeDaemonUnavailable, CategoryUnavailable, "Daemon not running")
	if got := err.UserMessage(); got != "Daemon not running" {
		t.Errorf("UserMessage() = %v, want %v", got, "Daemon not running")
	}
}

func TestAppError_LogMessage(t *testing.T) {
	// Without details
	err := New(CodeInvalidInput, CategoryValidation, "Invalid input")
	expected := "[INVALID_INPUT] Invalid input"
	if got := err.LogMessage(); got != expected {
		t.Errorf("LogMessage() = %v, want %v", got, expected)
	}

	// With details
	err = WrapWithDetails(nil, CodeInvalidInput, CategoryValidation, "Invalid input", "field: name")
	expected = "[INVALID_INPUT] Invalid input: field: name"
	if got := err.LogMessage(); got != expected {
		t.Errorf("LogMessage() = %v, want %v", got, expected)
	}
}

func TestNew(t *testing.T) {
	err := New(CodeNetworkNotFound, CategoryNotFound, "Network not found")

	if err.Code != CodeNetworkNotFound {
		t.Errorf("Code = %v, want %v", err.Code, CodeNetworkNotFound)
	}
	if err.Category != CategoryNotFound {
		t.Errorf("Category = %v, want %v", err.Category, CategoryNotFound)
	}
	if err.Message != "Network not found" {
		t.Errorf("Message = %v, want %v", err.Message, "Network not found")
	}
}

func TestWrap(t *testing.T) {
	cause := fmt.Errorf("root cause")
	err := Wrap(cause, CodeDatabaseError, CategoryStorage, "Database error")

	if err.Code != CodeDatabaseError {
		t.Errorf("Code = %v, want %v", err.Code, CodeDatabaseError)
	}
	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}
}

func TestWrapWithDetails(t *testing.T) {
	cause := fmt.Errorf("root cause")
	err := WrapWithDetails(cause, CodeConnectionFailed, CategoryNetwork, "Connection failed", "target: localhost:8080")

	if err.Details != "target: localhost:8080" {
		t.Errorf("Details = %v, want %v", err.Details, "target: localhost:8080")
	}
}

// Test convenience constructors

func TestErrConnectionFailed(t *testing.T) {
	cause := fmt.Errorf("connection refused")
	err := ErrConnectionFailed(cause, "localhost:8080")

	if err.Code != CodeConnectionFailed {
		t.Errorf("Code = %v, want %v", err.Code, CodeConnectionFailed)
	}
	if err.Category != CategoryNetwork {
		t.Errorf("Category = %v, want %v", err.Category, CategoryNetwork)
	}
}

func TestErrDaemonUnavailable(t *testing.T) {
	cause := fmt.Errorf("connection refused")
	err := ErrDaemonUnavailable(cause)

	if err.Code != CodeDaemonUnavailable {
		t.Errorf("Code = %v, want %v", err.Code, CodeDaemonUnavailable)
	}
	if !containsString(err.Message, "daemon") {
		t.Error("Message should mention daemon")
	}
}

func TestErrUnauthorized(t *testing.T) {
	err := ErrUnauthorized(nil)

	if err.Code != CodeUnauthorized {
		t.Errorf("Code = %v, want %v", err.Code, CodeUnauthorized)
	}
	if err.Category != CategoryAuth {
		t.Errorf("Category = %v, want %v", err.Category, CategoryAuth)
	}
}

func TestErrInvalidInput(t *testing.T) {
	err := ErrInvalidInput("email", "must be a valid email address")

	if err.Code != CodeInvalidInput {
		t.Errorf("Code = %v, want %v", err.Code, CodeInvalidInput)
	}
	if !containsString(err.Message, "email") {
		t.Error("Message should contain field name")
	}
}

func TestErrMissingRequired(t *testing.T) {
	err := ErrMissingRequired("username")

	if err.Code != CodeMissingRequired {
		t.Errorf("Code = %v, want %v", err.Code, CodeMissingRequired)
	}
	if !containsString(err.Message, "username") {
		t.Error("Message should contain field name")
	}
}

func TestErrNetworkNotFound(t *testing.T) {
	err := ErrNetworkNotFound("net-123")

	if err.Code != CodeNetworkNotFound {
		t.Errorf("Code = %v, want %v", err.Code, CodeNetworkNotFound)
	}
	if err.Category != CategoryNotFound {
		t.Errorf("Category = %v, want %v", err.Category, CategoryNotFound)
	}
}

func TestErrAlreadyJoined(t *testing.T) {
	err := ErrAlreadyJoined("My Network")

	if err.Code != CodeAlreadyJoined {
		t.Errorf("Code = %v, want %v", err.Code, CodeAlreadyJoined)
	}
	if !containsString(err.Message, "My Network") {
		t.Error("Message should contain network name")
	}
}

func TestErrAccessDenied(t *testing.T) {
	err := ErrAccessDenied("kick peers")

	if err.Code != CodeAccessDenied {
		t.Errorf("Code = %v, want %v", err.Code, CodeAccessDenied)
	}
	if !containsString(err.Message, "kick peers") {
		t.Error("Message should contain action")
	}
}

func TestErrBanned(t *testing.T) {
	err := ErrBanned("Test Network")

	if err.Code != CodeBanned {
		t.Errorf("Code = %v, want %v", err.Code, CodeBanned)
	}
	if !containsString(err.Message, "banned") {
		t.Error("Message should mention banned")
	}
}

// Test inspection helpers

func TestIsCategory(t *testing.T) {
	err := New(CodeInvalidInput, CategoryValidation, "Invalid")

	if !IsCategory(err, CategoryValidation) {
		t.Error("IsCategory should return true for matching category")
	}
	if IsCategory(err, CategoryNetwork) {
		t.Error("IsCategory should return false for non-matching category")
	}

	// Test with non-AppError
	stdErr := fmt.Errorf("standard error")
	if IsCategory(stdErr, CategoryValidation) {
		t.Error("IsCategory should return false for non-AppError")
	}
}

func TestIsCode(t *testing.T) {
	err := New(CodeNetworkNotFound, CategoryNotFound, "Not found")

	if !IsCode(err, CodeNetworkNotFound) {
		t.Error("IsCode should return true for matching code")
	}
	if IsCode(err, CodePeerNotFound) {
		t.Error("IsCode should return false for non-matching code")
	}

	// Test with non-AppError
	stdErr := fmt.Errorf("standard error")
	if IsCode(stdErr, CodeNetworkNotFound) {
		t.Error("IsCode should return false for non-AppError")
	}
}

func TestGetCode(t *testing.T) {
	err := New(CodeDatabaseError, CategoryStorage, "DB error")

	if got := GetCode(err); got != CodeDatabaseError {
		t.Errorf("GetCode() = %v, want %v", got, CodeDatabaseError)
	}

	// Test with non-AppError
	stdErr := fmt.Errorf("standard error")
	if got := GetCode(stdErr); got != "" {
		t.Errorf("GetCode() for non-AppError = %v, want empty", got)
	}
}

func TestGetUserMessage(t *testing.T) {
	err := New(CodeInvalidInput, CategoryValidation, "Please enter a valid email")

	if got := GetUserMessage(err); got != "Please enter a valid email" {
		t.Errorf("GetUserMessage() = %v, want %v", got, "Please enter a valid email")
	}

	// Test with non-AppError
	stdErr := fmt.Errorf("standard error message")
	if got := GetUserMessage(stdErr); got != "standard error message" {
		t.Errorf("GetUserMessage() for non-AppError = %v, want %v", got, "standard error message")
	}
}

func TestToAppError(t *testing.T) {
	// Test with nil
	if got := ToAppError(nil); got != nil {
		t.Error("ToAppError(nil) should return nil")
	}

	// Test with AppError
	appErr := New(CodeInvalidInput, CategoryValidation, "Invalid")
	if got := ToAppError(appErr); got != appErr {
		t.Error("ToAppError should return same AppError")
	}

	// Test with standard error
	stdErr := fmt.Errorf("standard error")
	got := ToAppError(stdErr)
	if got == nil {
		t.Fatal("ToAppError should wrap standard error")
	}
	if got.Code != CodeUnexpected {
		t.Errorf("Code = %v, want %v", got.Code, CodeUnexpected)
	}
	if got.Category != CategoryInternal {
		t.Errorf("Category = %v, want %v", got.Category, CategoryInternal)
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Additional tests for 0% coverage functions

func TestErrServerUnreachable(t *testing.T) {
	cause := fmt.Errorf("connection timeout")
	err := ErrServerUnreachable(cause, "https://api.goconnect.io")

	if err.Code != CodeServerUnreachable {
		t.Errorf("Code = %v, want %v", err.Code, CodeServerUnreachable)
	}
	if err.Category != CategoryNetwork {
		t.Errorf("Category = %v, want %v", err.Category, CategoryNetwork)
	}
	if !containsString(err.Details, "api.goconnect.io") {
		t.Error("Details should contain server URL")
	}
	if !containsString(err.Message, "Cannot reach") {
		t.Error("Message should mention unreachable")
	}
}

func TestErrPeerUnreachable(t *testing.T) {
	cause := fmt.Errorf("peer offline")
	err := ErrPeerUnreachable(cause, "peer-12345")

	if err.Code != CodePeerUnreachable {
		t.Errorf("Code = %v, want %v", err.Code, CodePeerUnreachable)
	}
	if err.Category != CategoryNetwork {
		t.Errorf("Category = %v, want %v", err.Category, CategoryNetwork)
	}
	if !containsString(err.Details, "peer-12345") {
		t.Error("Details should contain peer ID")
	}
}

func TestErrInvalidToken(t *testing.T) {
	cause := fmt.Errorf("token expired")
	err := ErrInvalidToken(cause)

	if err.Code != CodeInvalidToken {
		t.Errorf("Code = %v, want %v", err.Code, CodeInvalidToken)
	}
	if err.Category != CategoryAuth {
		t.Errorf("Category = %v, want %v", err.Category, CategoryAuth)
	}
	if !containsString(err.Message, "expired") {
		t.Error("Message should mention expired")
	}
}

func TestErrInvalidCredentials(t *testing.T) {
	cause := fmt.Errorf("wrong password")
	err := ErrInvalidCredentials(cause)

	if err.Code != CodeInvalidCredentials {
		t.Errorf("Code = %v, want %v", err.Code, CodeInvalidCredentials)
	}
	if err.Category != CategoryAuth {
		t.Errorf("Category = %v, want %v", err.Category, CategoryAuth)
	}
	if !containsString(err.Message, "Invalid username or password") {
		t.Error("Message should mention invalid credentials")
	}
}

func TestErrConfigNotFound(t *testing.T) {
	err := ErrConfigNotFound("/home/user/.config/goconnect/config.yaml")

	if err.Code != CodeConfigNotFound {
		t.Errorf("Code = %v, want %v", err.Code, CodeConfigNotFound)
	}
	if err.Category != CategoryConfig {
		t.Errorf("Category = %v, want %v", err.Category, CategoryConfig)
	}
	if !containsString(err.Details, "/home/user/.config/goconnect/config.yaml") {
		t.Error("Details should contain config path")
	}
	if !containsString(err.Message, "Configuration file not found") {
		t.Error("Message should mention not found")
	}
}

func TestErrConfigInvalid(t *testing.T) {
	cause := fmt.Errorf("yaml parse error")
	err := ErrConfigInvalid(cause, "/etc/goconnect/config.yaml")

	if err.Code != CodeConfigInvalid {
		t.Errorf("Code = %v, want %v", err.Code, CodeConfigInvalid)
	}
	if err.Category != CategoryConfig {
		t.Errorf("Category = %v, want %v", err.Category, CategoryConfig)
	}
	if !containsString(err.Details, "/etc/goconnect/config.yaml") {
		t.Error("Details should contain config path")
	}
}

func TestErrDatabaseError(t *testing.T) {
	cause := fmt.Errorf("connection refused")
	err := ErrDatabaseError(cause, "insert_user")

	if err.Code != CodeDatabaseError {
		t.Errorf("Code = %v, want %v", err.Code, CodeDatabaseError)
	}
	if err.Category != CategoryStorage {
		t.Errorf("Category = %v, want %v", err.Category, CategoryStorage)
	}
	if !containsString(err.Details, "insert_user") {
		t.Error("Details should contain operation name")
	}
}

func TestErrPeerNotFound(t *testing.T) {
	err := ErrPeerNotFound("peer-nonexistent")

	if err.Code != CodePeerNotFound {
		t.Errorf("Code = %v, want %v", err.Code, CodePeerNotFound)
	}
	if err.Category != CategoryNotFound {
		t.Errorf("Category = %v, want %v", err.Category, CategoryNotFound)
	}
	if !containsString(err.Details, "peer-nonexistent") {
		t.Error("Details should contain peer ID")
	}
}

func TestErrTransferNotFound(t *testing.T) {
	err := ErrTransferNotFound("transfer-12345")

	if err.Code != CodeTransferNotFound {
		t.Errorf("Code = %v, want %v", err.Code, CodeTransferNotFound)
	}
	if err.Category != CategoryNotFound {
		t.Errorf("Category = %v, want %v", err.Category, CategoryNotFound)
	}
	if !containsString(err.Details, "transfer-12345") {
		t.Error("Details should contain transfer ID")
	}
}

func TestErrTransferInProgress(t *testing.T) {
	err := ErrTransferInProgress("transfer-active")

	if err.Code != CodeTransferInProgress {
		t.Errorf("Code = %v, want %v", err.Code, CodeTransferInProgress)
	}
	if err.Category != CategoryConflict {
		t.Errorf("Category = %v, want %v", err.Category, CategoryConflict)
	}
	if !containsString(err.Details, "transfer-active") {
		t.Error("Details should contain transfer ID")
	}
}

func TestAppError_Is_WithNonAppError(t *testing.T) {
	appErr := New(CodeInvalidInput, CategoryValidation, "Invalid")
	stdErr := fmt.Errorf("standard error")

	// AppError.Is should return false for non-AppError
	if appErr.Is(stdErr) {
		t.Error("AppError.Is should return false for non-AppError")
	}
}

func TestAppError_Is_WithNil(t *testing.T) {
	appErr := New(CodeInvalidInput, CategoryValidation, "Invalid")

	// AppError.Is should return false for nil
	if appErr.Is(nil) {
		t.Error("AppError.Is should return false for nil")
	}
}
