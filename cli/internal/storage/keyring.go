package storage

import (
	"fmt"

	"github.com/99designs/keyring"
)

const (
	keyringService = "goconnect-daemon"
	keyringUser    = "default" // Using a default user for simplicity
	tokenKey       = "auth_token"
	deviceIDKey    = "device_id"
)

// KeyringStore provides an interface to store and retrieve sensitive data using OS keyring.
type KeyringStore struct {
	kr keyring.Keyring
}

// NewKeyringStore creates a new KeyringStore instance.
func NewKeyringStore() (*KeyringStore, error) {
	kr, err := keyring.Open(keyring.Config{
		ServiceName: keyringService,
		AllowedBackends: []keyring.BackendType{
			keyring.SecretServiceBackend,
			keyring.KeychainBackend,
			keyring.WinCredBackend,
			keyring.PassBackend,
		},
		// For 99designs/keyring, there isn't a direct equivalent to KeychainNotTrustUserPrompt.
		// This might need to be handled via specific backend configurations if available.
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open keyring: %w", err)
	}
	return &KeyringStore{kr: kr}, nil
}

// StoreAuthToken stores the authentication token in the keyring.
func (ks *KeyringStore) StoreAuthToken(token string) error {
	return ks.kr.Set(keyring.Item{
		Key:         tokenKey,
		Data:        []byte(token),
		Label:       "GoConnect Daemon Auth Token",
		Description: "Authentication token for GoConnect server API access",
	})
}

// RetrieveAuthToken retrieves the authentication token from the keyring.
func (ks *KeyringStore) RetrieveAuthToken() (string, error) {
	item, err := ks.kr.Get(tokenKey)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve auth token: %w", err)
	}
	return string(item.Data), nil
}

// StoreDeviceID stores the device ID in the keyring (optional, could be in identity.json).
func (ks *KeyringStore) StoreDeviceID(id string) error {
	return ks.kr.Set(keyring.Item{
		Key:         deviceIDKey,
		Data:        []byte(id),
		Label:       "GoConnect Device ID",
		Description: "Unique identifier for this GoConnect client device",
	})
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
