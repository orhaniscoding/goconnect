package identity

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/curve25519"
)

// Identity represents the device identity
type Identity struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	DeviceID   string `json:"device_id,omitempty"`
	Token      string `json:"token,omitempty"` // Token to authenticate with server
	mu         sync.RWMutex
}

// Manager handles identity storage and generation
type Manager struct {
	configPath string
	identity   *Identity
	mu         sync.RWMutex
}

// NewManager creates a new identity manager
func NewManager() (*Manager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home dir: %w", err)
	}

	configDir := filepath.Join(home, ".goconnect")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config dir: %w", err)
	}

	return &Manager{
		configPath: filepath.Join(configDir, "device.json"),
	}, nil
}

// LoadOrGenerate loads the identity from disk or generates a new one
func (m *Manager) LoadOrGenerate() (*Identity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Try to load
	if err := m.load(); err == nil {
		return m.identity, nil
	}

	// Generate new
	if err := m.generate(); err != nil {
		return nil, err
	}

	// Save
	if err := m.save(); err != nil {
		return nil, err
	}

	return m.identity, nil
}

// Get returns the current identity
func (m *Manager) Get() *Identity {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.identity
}

// Update updates the identity with server details and saves
func (m *Manager) Update(deviceID, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.identity == nil {
		return fmt.Errorf("identity not initialized")
	}

	m.identity.DeviceID = deviceID
	m.identity.Token = token

	return m.save()
}

func (m *Manager) load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return err
	}

	var id Identity
	if err := json.Unmarshal(data, &id); err != nil {
		return err
	}

	m.identity = &id
	return nil
}

func (m *Manager) save() error {
	if m.identity == nil {
		return fmt.Errorf("no identity to save")
	}

	data, err := json.MarshalIndent(m.identity, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0600)
}

func (m *Manager) generate() error {
	// Generate 32 random bytes for private key
	privateKey := make([]byte, 32)
	if _, err := rand.Read(privateKey); err != nil {
		return fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Clamp the private key (WireGuard requirement)
	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64

	// Derive public key using Curve25519
	publicKey, err := curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		return fmt.Errorf("failed to derive public key: %w", err)
	}

	m.identity = &Identity{
		PrivateKey: base64.StdEncoding.EncodeToString(privateKey),
		PublicKey:  base64.StdEncoding.EncodeToString(publicKey),
	}

	return nil
}
