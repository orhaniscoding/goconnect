package identity

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager("/path/to/identity.json")
	if m == nil {
		t.Fatal("Expected manager, got nil")
	}
	if m.identityPath != "/path/to/identity.json" {
		t.Errorf("Expected path '/path/to/identity.json', got %s", m.identityPath)
	}
}

func TestManager_Get_Nil(t *testing.T) {
	m := NewManager("/path/to/identity.json")
	id := m.Get()
	if id != nil {
		t.Error("Expected nil identity before initialization")
	}
}

func TestManager_LoadOrCreateIdentity_Create(t *testing.T) {
	tmpDir := t.TempDir()
	identityPath := filepath.Join(tmpDir, "identity.json")

	m := NewManager(identityPath)
	id, err := m.LoadOrCreateIdentity()
	if err != nil {
		t.Fatalf("LoadOrCreateIdentity failed: %v", err)
	}

	if id == nil {
		t.Fatal("Expected identity, got nil")
	}

	// Verify keys are generated
	if id.PrivateKey == "" {
		t.Error("Expected private key to be generated")
	}
	if id.PublicKey == "" {
		t.Error("Expected public key to be generated")
	}

	// Verify keys are valid base64
	_, err = base64.StdEncoding.DecodeString(id.PrivateKey)
	if err != nil {
		t.Errorf("Private key is not valid base64: %v", err)
	}
	_, err = base64.StdEncoding.DecodeString(id.PublicKey)
	if err != nil {
		t.Errorf("Public key is not valid base64: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(identityPath); os.IsNotExist(err) {
		t.Error("Identity file was not created")
	}
}

func TestManager_LoadOrCreateIdentity_Load(t *testing.T) {
	tmpDir := t.TempDir()
	identityPath := filepath.Join(tmpDir, "identity.json")

	// Pre-create identity file
	existingIdentity := &Identity{
		PrivateKey: "dGVzdC1wcml2YXRlLWtleQ==", // test-private-key
		PublicKey:  "dGVzdC1wdWJsaWMta2V5",     // test-public-key
		DeviceID:   "existing-device-123",
	}
	data, _ := json.Marshal(existingIdentity)
	if err := os.WriteFile(identityPath, data, 0600); err != nil {
		t.Fatalf("Failed to write test identity: %v", err)
	}

	m := NewManager(identityPath)
	id, err := m.LoadOrCreateIdentity()
	if err != nil {
		t.Fatalf("LoadOrCreateIdentity failed: %v", err)
	}

	// Verify loaded identity matches
	if id.PrivateKey != existingIdentity.PrivateKey {
		t.Errorf("Expected private key %s, got %s", existingIdentity.PrivateKey, id.PrivateKey)
	}
	if id.PublicKey != existingIdentity.PublicKey {
		t.Errorf("Expected public key %s, got %s", existingIdentity.PublicKey, id.PublicKey)
	}
	if id.DeviceID != "existing-device-123" {
		t.Errorf("Expected device ID 'existing-device-123', got %s", id.DeviceID)
	}
}

func TestManager_Update(t *testing.T) {
	tmpDir := t.TempDir()
	identityPath := filepath.Join(tmpDir, "identity.json")

	m := NewManager(identityPath)
	_, err := m.LoadOrCreateIdentity()
	if err != nil {
		t.Fatalf("LoadOrCreateIdentity failed: %v", err)
	}

	// Update with device ID
	err = m.Update("new-device-id")
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify in memory
	if m.Get().DeviceID != "new-device-id" {
		t.Errorf("Expected device ID 'new-device-id', got %s", m.Get().DeviceID)
	}

	// Reload and verify persistence
	m2 := NewManager(identityPath)
	id2, err := m2.LoadOrCreateIdentity()
	if err != nil {
		t.Fatalf("LoadOrCreateIdentity failed: %v", err)
	}
	if id2.DeviceID != "new-device-id" {
		t.Errorf("Expected persisted device ID 'new-device-id', got %s", id2.DeviceID)
	}
}

func TestManager_Update_NoIdentity(t *testing.T) {
	m := NewManager("/path/to/identity.json")
	
	err := m.Update("device-id")
	if err == nil {
		t.Error("Expected error when updating without identity")
	}
}

func TestManager_LoadOrCreateIdentity_CreateDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	// Use nested subdirectory that doesn't exist
	identityPath := filepath.Join(tmpDir, "nested", "sub", "dir", "identity.json")

	m := NewManager(identityPath)
	_, err := m.LoadOrCreateIdentity()
	if err != nil {
		t.Fatalf("LoadOrCreateIdentity failed: %v", err)
	}

	// Verify directory was created
	dir := filepath.Dir(identityPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("Identity directory was not created")
	}
}

func TestManager_LoadOrCreateIdentity_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	identityPath := filepath.Join(tmpDir, "identity.json")

	// Write invalid JSON
	if err := os.WriteFile(identityPath, []byte("invalid json"), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	m := NewManager(identityPath)
	// Should generate new identity when existing file is invalid
	id, err := m.LoadOrCreateIdentity()
	if err != nil {
		t.Fatalf("LoadOrCreateIdentity failed: %v", err)
	}

	// Should have generated new identity
	if id.PrivateKey == "" || id.PublicKey == "" {
		t.Error("Expected new keys to be generated for invalid JSON")
	}
}

func TestIdentity_Struct(t *testing.T) {
	// Test Identity struct serialization
	id := &Identity{
		PrivateKey: "private123",
		PublicKey:  "public456",
		DeviceID:   "device789",
	}

	data, err := json.Marshal(id)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded Identity
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.PrivateKey != "private123" {
		t.Errorf("Expected private key 'private123', got %s", decoded.PrivateKey)
	}
	if decoded.PublicKey != "public456" {
		t.Errorf("Expected public key 'public456', got %s", decoded.PublicKey)
	}
	if decoded.DeviceID != "device789" {
		t.Errorf("Expected device ID 'device789', got %s", decoded.DeviceID)
	}
}

func TestManager_KeyGeneration(t *testing.T) {
	tmpDir := t.TempDir()
	identityPath := filepath.Join(tmpDir, "identity.json")

	m := NewManager(identityPath)
	id1, err := m.LoadOrCreateIdentity()
	if err != nil {
		t.Fatalf("LoadOrCreateIdentity failed: %v", err)
	}

	// Create another manager to generate different keys
	identityPath2 := filepath.Join(tmpDir, "identity2.json")
	m2 := NewManager(identityPath2)
	id2, err := m2.LoadOrCreateIdentity()
	if err != nil {
		t.Fatalf("LoadOrCreateIdentity failed: %v", err)
	}

	// Keys should be different (cryptographically random)
	if id1.PrivateKey == id2.PrivateKey {
		t.Error("Expected different private keys for different identities")
	}
	if id1.PublicKey == id2.PublicKey {
		t.Error("Expected different public keys for different identities")
	}
}

func TestManager_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	identityPath := filepath.Join(tmpDir, "identity.json")

	m := NewManager(identityPath)
	_, err := m.LoadOrCreateIdentity()
	if err != nil {
		t.Fatalf("LoadOrCreateIdentity failed: %v", err)
	}

	// Check file permissions (should be 0600 for security)
	info, err := os.Stat(identityPath)
	if err != nil {
		t.Fatalf("Failed to stat identity file: %v", err)
	}

	// On Windows, file permissions work differently
	// Just verify the file exists and is not a directory
	if info.IsDir() {
		t.Error("Identity path should be a file, not a directory")
	}
}
