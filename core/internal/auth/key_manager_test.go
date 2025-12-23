package auth

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyManager_LoadOrGenerate(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "goconnect-auth-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	km := NewKeyManager(tempDir)

	// 1. First run: Generate new key
	err = km.LoadOrGenerate()
	require.NoError(t, err)

	pub1 := km.GetPublicKey()
	require.NotEmpty(t, pub1.String())

	// Verify file exists
	keyPath := filepath.Join(tempDir, "device.key")
	_, err = os.Stat(keyPath)
	assert.NoError(t, err)

	// Verify permissions on Unix
	if os.PathSeparator == '/' {
		info, err := os.Stat(keyPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
	}

	// 2. Second run: Load existing key
	km2 := NewKeyManager(tempDir)
	err = km2.LoadOrGenerate()
	require.NoError(t, err)

	pub2 := km2.GetPublicKey()
	assert.Equal(t, pub1, pub2, "Public keys should be identical on reload")
}

func TestKeyManager_InvalidKey(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "goconnect-auth-test-invalid-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	keyPath := filepath.Join(tempDir, "device.key")
	err = os.MkdirAll(tempDir, 0700)
	require.NoError(t, err)

	// Write garbage to key file
	err = os.WriteFile(keyPath, []byte("not-a-key"), 0600)
	require.NoError(t, err)

	km := NewKeyManager(tempDir)
	err = km.LoadOrGenerate()
	assert.Error(t, err, "Should fail with invalid key data")
}
