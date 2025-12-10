package domain

import (
	"encoding/json"
	"testing"
)

func TestError_ToHTTPStatus_CoreCodes(t *testing.T) {
	cases := []struct {
		code string
		want int
	}{
		// 400 Bad Request
		{ErrInvalidRequest, 400},
		{ErrCIDRInvalid, 400},
		{ErrValidation, 400},
		{ErrWeakPassword, 400},
		// 401 Unauthorized
		{ErrUnauthorized, 401},
		{ErrInvalidToken, 401},
		{ErrTokenExpired, 401},
		{ErrInvalidCredentials, 401},
		// 403 Forbidden
		{ErrNotAuthorized, 403},
		{ErrForbidden, 403},
		{ErrUserBanned, 403},
		{ErrUserKicked, 403},
		// 404 Not Found
		{ErrNotFound, 404},
		{ErrUserNotFound, 404},
		{ErrTenantNotFound, 404},
		{ErrNetworkPrivate, 404},
		// 409 Conflict
		{ErrCIDROverlap, 409},
		{ErrIdempotencyConflict, 409},
		{ErrIPExhausted, 409},
		{ErrUserAlreadyExists, 409},
		{ErrEmailAlreadyExists, 409},
		// 429 Too Many Requests
		{ErrRateLimited, 429},
		// 200 OK (special case)
		{ErrAlreadyMember, 200},
		// 202 Accepted (special case)
		{ErrAlreadyRequested, 202},
		// 501 Not Implemented
		{ErrNotImplemented, 501},
		// 500 Internal Server Error (default)
		{"unknown_error_code", 500},
	}
	for _, tc := range cases {
		if got := NewError(tc.code, "", nil).ToHTTPStatus(); got != tc.want {
			t.Fatalf("code %s => status %d, want %d", tc.code, got, tc.want)
		}
	}
}

func TestError_ToJSON(t *testing.T) {
	err := NewError(ErrNotFound, "Resource not found", nil)
	jsonData := err.ToJSON()

	var decoded Error
	if unmarshalErr := json.Unmarshal(jsonData, &decoded); unmarshalErr != nil {
		t.Fatalf("failed to unmarshal JSON: %v", unmarshalErr)
	}

	if decoded.Code != ErrNotFound {
		t.Errorf("expected code %s, got %s", ErrNotFound, decoded.Code)
	}

	if decoded.Message != "Resource not found" {
		t.Errorf("expected message 'Resource not found', got %s", decoded.Message)
	}
}

func TestError_Error(t *testing.T) {
	err := NewError(ErrUnauthorized, "Invalid token", nil)
	result := err.Error()

	if result != "Invalid token" {
		t.Errorf("expected 'Invalid token', got %s", result)
	}
}
