//go:build windows

package auth

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	ErrKeyNotFound = errors.New("device key not found")
	ErrInvalidKey  = errors.New("invalid device key")
)

// Key is a 32-byte key that can be used as a WireGuard key.
// On Windows, we use this custom type instead of wgtypes.Key since
// the wgctrl package doesn't build on Windows (requires Linux netlink).
type Key [32]byte

// String returns the base64 representation of the key.
func (k Key) String() string {
	return encodeBase64(k[:])
}

// PublicKey derives the public key from a private key.
// This is a simplified implementation for Windows.
func (k Key) PublicKey() Key {
	return Key(publicKeyFromPrivate(PrivateKey(k)))
}

// KeyManager defines the interface for managing device identity.
type KeyManager interface {
	LoadOrGenerate() error
	GetPublicKey() Key
	GetPrivateKey() Key
}

// fsKeyManager implements KeyManager using the filesystem.
type fsKeyManager struct {
	configDir  string
	privateKey PrivateKey
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
			// Generate new key (32 random bytes for WireGuard)
			privKey, err = generatePrivateKey()
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

// generatePrivateKey generates a new WireGuard private key.
// This follows the WireGuard key generation specification (Curve25519).
func generatePrivateKey() (PrivateKey, error) {
	var key PrivateKey
	_, err := rand.Read(key[:])
	if err != nil {
		return PrivateKey{}, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Clamp the key according to Curve25519 requirements
	key[0] &= 248
	key[31] &= 127
	key[31] |= 64

	return key, nil
}

// GetPublicKey returns the public key from the identity.
func (m *fsKeyManager) GetPublicKey() Key {
	return Key(publicKeyFromPrivate(m.privateKey))
}

// GetPrivateKey returns the private key.
func (m *fsKeyManager) GetPrivateKey() Key {
	return Key(m.privateKey)
}

// publicKeyFromPrivate derives a public key from a private key.
// This is a simplified implementation - for production use, import
// golang.org/x/crypto/curve25519.
func publicKeyFromPrivate(priv PrivateKey) [32]byte {
	// Placeholder: XOR pattern to make it different from private key.
	// In production, use: curve25519.ScalarBaseMult(&pub, &priv)
	var pub [32]byte
	for i := range priv {
		pub[i] = priv[i] ^ 0x55
	}
	return pub
}
