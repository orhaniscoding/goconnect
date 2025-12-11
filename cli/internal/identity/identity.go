package identity

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log" // Using log for now, will replace with service.Logger
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
	// Token will now be stored in the OS Keyring, not in the identity file.
}

// Manager handles identity storage and generation
type Manager struct {
	identityPath string // Path to the identity file
	identity     *Identity
	mu           sync.RWMutex
}

// NewManager creates a new identity manager.
// It takes the identity file path from the main config.
func NewManager(identityPath string) *Manager {
	return &Manager{
		identityPath: identityPath,
	}
}

// LoadOrCreateIdentity loads the identity from disk or generates a new one if not found.
// This function combines Load and Generate logic for simpler startup.
func (m *Manager) LoadOrCreateIdentity() (*Identity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure the directory for the identity file exists
	dir := filepath.Dir(m.identityPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create identity directory %s: %w", dir, err)
	}

	// Try to load
	if err := m.load(); err == nil {
		log.Printf("Loaded existing device identity from %s", m.identityPath)
		return m.identity, nil
	} else if !os.IsNotExist(err) {
		log.Printf("Failed to load identity from %s: %v", m.identityPath, err)
	}

	// Generate new if load failed or file didn't exist
	log.Printf("Generating new device identity.")
	if err := m.generate(); err != nil {
		return nil, err
	}

	// Save the newly generated identity
	if err := m.save(); err != nil {
		return nil, fmt.Errorf("failed to save new identity to %s: %w", m.identityPath, err)
	}

	log.Printf("Generated and saved new device identity with public key: %s", m.identity.PublicKey)
	return m.identity, nil
}

// Get returns the current identity
func (m *Manager) Get() *Identity {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.identity
}

// Update updates the identity with server details and saves
// Token is no longer updated here, it's handled by KeyringStore.
func (m *Manager) Update(deviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.identity == nil {
		return fmt.Errorf("identity not initialized")
	}

	m.identity.DeviceID = deviceID

	return m.save()
}

func (m *Manager) load() error {
	data, err := os.ReadFile(m.identityPath)
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

	return os.WriteFile(m.identityPath, data, 0600)
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
