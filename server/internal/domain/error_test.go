package domain

import "testing"

func TestError_ToHTTPStatus_CoreCodes(t *testing.T) {
	cases := []struct {
		code string
		want int
	}{
		{ErrCIDRInvalid, 400},
		{ErrWeakPassword, 400},
		{ErrCIDROverlap, 409},
		{ErrUnauthorized, 401},
		{ErrInvalidCredentials, 401},
		{ErrInvalidToken, 401},
		{ErrTokenExpired, 401},
		{ErrSessionExpired, 401},
		{ErrNotAuthorized, 403},
		{ErrNotFound, 404},
		{ErrTenantNotFound, 404},
		{ErrUserNotFound, 404},
		{ErrRateLimited, 429},
		{ErrEmailAlreadyExists, 409},
		{ErrNotImplemented, 501},
	}
	for _, tc := range cases {
		if got := NewError(tc.code, "", nil).ToHTTPStatus(); got != tc.want {
			t.Fatalf("code %s => status %d, want %d", tc.code, got, tc.want)
		}
	}
}
