//go:build !windows

package auth

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var (
	ErrKeyNotFound = errors.New("device key not found")
	ErrInvalidKey  = errors.New("invalid device key")
)

// KeyManager defines the interface for managing device identity.
type KeyManager interface {
	LoadOrGenerate() error
	GetPublicKey() wgtypes.Key
	GetPrivateKey() wgtypes.Key
}

// fsKeyManager implements KeyManager using the filesystem.
type fsKeyManager struct {
	configDir  string
	privateKey wgtypes.Key
}

// NewKeyManager creates a new filesystem-based KeyManager.
func NewKeyManager(configDir string) KeyManager {
	return &fsKeyManager{
		configDir: configDir,
	}
}

// LoadOrGenerate loads the key from disk or generates a new one if it doesn't exist.
func (m *fsKeyManager) LoadOrGenerate() error {
	keyPath := filepath.Join(m.configDir, "device.key")

	// Ensure config directory exists with 0700 permissions
	if err := os.MkdirAll(m.configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	privKey, err := loadKey(keyPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Generate new key
			privKey, err = wgtypes.GeneratePrivateKey()
			if err != nil {
				return fmt.Errorf("failed to generate private key: %w", err)
			}

			// Persist new key
			if err := persistKey(keyPath, privKey); err != nil {
				return fmt.Errorf("failed to persist private key: %w", err)
			}
		} else {
			return fmt.Errorf("failed to load private key: %w", err)
		}
	}

	m.privateKey = privKey
	return nil
}

// GetPublicKey returns the public key from the identity.
func (m *fsKeyManager) GetPublicKey() wgtypes.Key {
	return m.privateKey.PublicKey()
}

// GetPrivateKey returns the private key.
func (m *fsKeyManager) GetPrivateKey() wgtypes.Key {
	return m.privateKey
}
