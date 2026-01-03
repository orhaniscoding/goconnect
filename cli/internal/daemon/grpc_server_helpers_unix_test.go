//go:build !windows

package daemon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ==================== IsPipeSupported Test (Unix) ====================

func TestIsPipeSupported_Unix(t *testing.T) {
	// On Unix, named pipes are not supported (Windows only)
	supported := IsPipeSupported()
	assert.False(t, supported)
}

func TestCreateWindowsListener_Unix(t *testing.T) {
	// On Unix, this returns nil, nil (no error, no listener)
	listener, err := CreateWindowsListener()
	assert.NoError(t, err)
	assert.Nil(t, listener)
}
