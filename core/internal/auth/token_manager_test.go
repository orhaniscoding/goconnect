package auth_test

import (
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenManager_SaveLoadSession(t *testing.T) {
	tmpDir := t.TempDir()
	tm, err := auth.NewTokenManager(tmpDir)
	require.NoError(t, err)

	// JSON marshaling of time uses RFC3339 which has second precision usually, 
	// or nano if present but comparison might be tricky across marshal/unmarshal.
	expiry := time.Now().Add(1 * time.Hour).Round(time.Millisecond)

	session := &auth.TokenSession{
		AccessToken:  "acc-token-123",
		RefreshToken: "ref-token-456",
		Expiry:       expiry,
	}

	err = tm.SaveSession(session)
	require.NoError(t, err)

	loaded, err := tm.LoadSession()
	require.NoError(t, err)
	assert.Equal(t, session.AccessToken, loaded.AccessToken)
	assert.Equal(t, session.RefreshToken, loaded.RefreshToken)
	// Check time is close enough (handling timezone/parsing diffs)
	assert.WithinDuration(t, session.Expiry, loaded.Expiry, time.Second)

	// Test persistence across instances (simulating restart)
	tm2, err := auth.NewTokenManager(tmpDir)
	require.NoError(t, err)
	loaded2, err := tm2.LoadSession()
	require.NoError(t, err)
	assert.Equal(t, session.AccessToken, loaded2.AccessToken)
}

func TestTokenManager_ClearSession(t *testing.T) {
	tmpDir := t.TempDir()
	tm, err := auth.NewTokenManager(tmpDir)
	require.NoError(t, err)

	err = tm.SaveSession(&auth.TokenSession{AccessToken: "a"})
	require.NoError(t, err)

	err = tm.ClearSession()
	require.NoError(t, err)
	
	_, err = tm.LoadSession()
	assert.Error(t, err)
}
