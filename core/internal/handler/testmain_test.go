package handler

import (
	"os"
	"testing"
)

// Ensure handler tests are hermetic: some tests still use service.NewAuthService which panics
// if JWT_SECRET is not set. Tests should not depend on external env setup.
func TestMain(m *testing.M) {
	if os.Getenv("JWT_SECRET") == "" {
		_ = os.Setenv("JWT_SECRET", "12345678901234567890123456789012")
	}
	os.Exit(m.Run())
}



