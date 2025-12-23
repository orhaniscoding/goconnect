package daemon_test

import (
	"testing"

	"github.com/orhaniscoding/goconnect/server/internal/auth"
	"github.com/orhaniscoding/goconnect/server/internal/daemon"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tmpDir := t.TempDir()
	keyMgr := auth.NewKeyManager(tmpDir)
	tokenMgr, err := auth.NewTokenManager(tmpDir)
	assert.NoError(t, err)

	d := daemon.New(keyMgr, tokenMgr, "test-version", "")
	assert.NotNil(t, d, "Daemon instance should not be nil")
	assert.Equal(t, "test-version", d.Version)
}
