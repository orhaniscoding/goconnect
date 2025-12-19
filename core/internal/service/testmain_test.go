package service

import (
	"os"
	"testing"
)

// Ensure core/service tests are hermetic: NewAuthService panics if JWT_SECRET is not set.
// Tests should not require external env setup.
func TestMain(m *testing.M) {
	if os.Getenv("JWT_SECRET") == "" {
		_ = os.Setenv("JWT_SECRET", "12345678901234567890123456789012")
	}
	os.Exit(m.Run())
}
