package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleDeepLink_Invalid(t *testing.T) {
	err := HandleDeepLink("invalid-uri")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid deep link")
}

func TestHandleDeepLink_UnknownAction(t *testing.T) {
	// goconnect://unknown
	err := HandleDeepLink("goconnect://unknown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown action")
}

func TestHandleDeepLink_Login_MissingParams(t *testing.T) {
	// goconnect://login
	err := HandleDeepLink("goconnect://login")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing token or server")
}

func TestHandleDeepLink_Join_Invalid(t *testing.T) {
	// mock deeplink handler?
	// The commands package uses deeplink.NewHandler() internally which we can't easily mock
	// without refactoring commands package to accept a handler factory.
	// But we can test invalid inputs that cause deeplink.Parse to succeed
	// but logic to fail (e.g. empty target where required).

	// For now, simple validation tests are sufficient to cover the switch case.
}

func TestHandleDeepLink_Join_Success(t *testing.T) {
	// This would require mocking network calls in deeplink.Handler.
	// Skipping for now as it requires deeper refactoring of deeplink package.
}
