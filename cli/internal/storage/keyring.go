package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/99designs/keyring"
)

const (
	keyringService = "goconnect"
	keyringUser    = "default" // Using a default user for simplicity
	tokenKey       = "auth_token"
	refreshTokenKey = "refresh_token"
	deviceIDKey    = "device_id"
)

// KeyringStore provides an interface to store and retrieve sensitive data using OS keyring.
type KeyringStore struct {
	kr keyring.Keyring
}

// NewKeyringStore creates a new KeyringStore instance.
func NewKeyringStore() (*KeyringStore, error) {
	// Ensure keyring directory exists with secure permissions (0700)
	keyringDir := filepath.Join(defaultDataDir(), "keyring")
	if err := os.MkdirAll(keyringDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create keyring directory: %w", err)
	}

	kr, err := keyring.Open(keyring.Config{
		ServiceName: keyringService,
		AllowedBackends: []keyring.BackendType{
			keyring.SecretServiceBackend,
			keyring.KeychainBackend,
			keyring.WinCredBackend,
			keyring.PassBackend,
			keyring.FileBackend, // Add FileBackend as a last-resort fallback
		},
		// FileBackend configuration (used if system backends are unavailable)
		FileDir: keyringDir,
		FilePasswordFunc: func(_ string) (string, error) {
			// In a real app, this might be a master password.
			// For GoConnect daemon, we'll try to use a persistent machine-specific secret
			// if available, or just a fixed string for basic file-level protection.
			return "goconnect-persistent-key", nil
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open system keyring (tried multiple backends): %w", err)
	}
	return &KeyringStore{kr: kr}, nil
}

// StoreAuthToken stores the authentication token in the keyring.
func (ks *KeyringStore) StoreAuthToken(token string) error {
	if err := ks.kr.Set(keyring.Item{
		Key:         tokenKey,
		Data:        []byte(token),
		Label:       "GoConnect Daemon Auth Token",
		Description: "Authentication token for GoConnect server API access",
	}); err != nil {
		return fmt.Errorf("failed to store auth token in keyring: %w", err)
	}
	return nil
}

// RetrieveAuthToken retrieves the authentication token from the keyring.
func (ks *KeyringStore) RetrieveAuthToken() (string, error) {
	item, err := ks.kr.Get(tokenKey)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve auth token: %w", err)
	}
	return string(item.Data), nil
}

// StoreRefreshToken stores the refresh token in the keyring.
func (ks *KeyringStore) StoreRefreshToken(token string) error {
	if err := ks.kr.Set(keyring.Item{
		Key:         refreshTokenKey,
		Data:        []byte(token),
		Label:       "GoConnect Daemon Refresh Token",
		Description: "Refresh token for GoConnect server API access",
	}); err != nil {
		return fmt.Errorf("failed to store refresh token in keyring: %w", err)
	}
	return nil
}

// RetrieveRefreshToken retrieves the refresh token from the keyring.
func (ks *KeyringStore) RetrieveRefreshToken() (string, error) {
	item, err := ks.kr.Get(refreshTokenKey)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve refresh token: %w", err)
	}
	return string(item.Data), nil
}

// StoreDeviceID stores the device ID in the keyring (optional, could be in identity.json).
func (ks *KeyringStore) StoreDeviceID(id string) error {
	if err := ks.kr.Set(keyring.Item{
		Key:         deviceIDKey,
		Data:        []byte(id),
		Label:       "GoConnect Device ID",
		Description: "Unique identifier for this GoConnect client device",
	}); err != nil {
		return fmt.Errorf("failed to store device ID in keyring: %w", err)
	}
	return nil
}

// RetrieveDeviceID retrieves the device ID from the keyring.
func (ks *KeyringStore) RetrieveDeviceID() (string, error) {
	item, err := ks.kr.Get(deviceIDKey)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve device ID: %w", err)
	}
	return string(item.Data), nil
}

// RemoveAuthData removes all stored authentication data from the keyring.
func (ks *KeyringStore) RemoveAuthData() error {
	if err := ks.kr.Remove(tokenKey); err != nil {
		return fmt.Errorf("failed to remove auth token: %w", err)
	}
	// It's possible deviceIDKey is not always stored, so handle its removal gracefully
	if err := ks.kr.Remove(deviceIDKey); err != nil && err != keyring.ErrKeyNotFound {
		return fmt.Errorf("failed to remove device ID: %w", err)
	}
	return nil
}

// NewTestKeyring creates a new KeyringStore backed by file (for testing).
func NewTestKeyring(dir string) (*KeyringStore, error) {
	kr, err := keyring.Open(keyring.Config{
		AllowedBackends:  []keyring.BackendType{keyring.FileBackend},
		FileDir:          dir,
		FilePasswordFunc: func(_ string) (string, error) { return "test", nil },
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open test keyring: %w", err)
	}
	return &KeyringStore{kr: kr}, nil
}

func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".goconnect")
}
