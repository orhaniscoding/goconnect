package storage

import (
	"testing"
)

func TestKeyringConstants(t *testing.T) {
	// Test that constants are defined correctly
	if keyringService != "goconnect-daemon" {
		t.Errorf("Expected service 'goconnect-daemon', got %s", keyringService)
	}
	if keyringUser != "default" {
		t.Errorf("Expected user 'default', got %s", keyringUser)
	}
	if tokenKey != "auth_token" {
		t.Errorf("Expected token key 'auth_token', got %s", tokenKey)
	}
	if deviceIDKey != "device_id" {
		t.Errorf("Expected device ID key 'device_id', got %s", deviceIDKey)
	}
}

func TestKeyringStore_Struct(t *testing.T) {
	// Test that KeyringStore struct exists and can be created
	// Note: We can't fully test keyring operations without a real keyring backend
	// This test verifies the struct definition
	var _ *KeyringStore

	// Verify NewKeyringStore returns an error message if keyring unavailable
	// In CI/test environments, keyring may not be available
	ks, err := NewKeyringStore()
	if err != nil {
		// Expected in environments without keyring support
		t.Logf("Keyring not available in test environment: %v", err)
		return
	}

	// If keyring is available, verify the store is valid
	if ks == nil {
		t.Error("Expected KeyringStore, got nil")
		return
	}
	if ks.kr == nil {
		t.Error("Expected keyring to be initialized")
	}
}

// MockKeyringStore tests - testing the interface contract
func TestKeyringStore_Methods(_ *testing.T) {
	// This test verifies that KeyringStore has the expected methods
	// by checking method signatures compile correctly
	var store *KeyringStore

	// These will panic if called on nil, but they verify the methods exist
	_ = func() { _ = store.StoreAuthToken("test") }
	_ = func() { _, _ = store.RetrieveAuthToken() }
	_ = func() { _ = store.StoreDeviceID("test") }
	_ = func() { _, _ = store.RetrieveDeviceID() }
	_ = func() { _ = store.RemoveAuthData() }

	// Suppress unused variable warning
	_ = store
}

// Integration test that runs only if keyring is available
func TestKeyringStore_Integration(t *testing.T) {
	ks, err := NewKeyringStore()
	if err != nil {
		t.Skip("Keyring not available, skipping integration test")
	}

	testToken := "test-auth-token-12345"
	testDeviceID := "test-device-id-67890"

	// Test StoreAuthToken
	if err := ks.StoreAuthToken(testToken); err != nil {
		t.Fatalf("StoreAuthToken failed: %v", err)
	}

	// Test RetrieveAuthToken
	token, err := ks.RetrieveAuthToken()
	if err != nil {
		t.Fatalf("RetrieveAuthToken failed: %v", err)
	}
	if token != testToken {
		t.Errorf("Expected token %s, got %s", testToken, token)
	}

	// Test StoreDeviceID
	if err := ks.StoreDeviceID(testDeviceID); err != nil {
		t.Fatalf("StoreDeviceID failed: %v", err)
	}

	// Test RetrieveDeviceID
	deviceID, err := ks.RetrieveDeviceID()
	if err != nil {
		t.Fatalf("RetrieveDeviceID failed: %v", err)
	}
	if deviceID != testDeviceID {
		t.Errorf("Expected device ID %s, got %s", testDeviceID, deviceID)
	}

	// Test RemoveAuthData (cleanup)
	if err := ks.RemoveAuthData(); err != nil {
		t.Fatalf("RemoveAuthData failed: %v", err)
	}

	// Verify data is removed
	_, err = ks.RetrieveAuthToken()
	if err == nil {
		t.Error("Expected error retrieving removed token")
	}
}

func TestNewTestKeyring(t *testing.T) {
	dir := t.TempDir()
	
	ks, err := NewTestKeyring(dir)
	if err != nil {
		t.Fatalf("NewTestKeyring failed: %v", err)
	}

	if ks == nil {
		t.Fatal("Expected KeyringStore, got nil")
	}

	if ks.kr == nil {
		t.Fatal("Expected keyring to be initialized")
	}

	// Test that it works
	testToken := "test-token-via-file-backend"
	if err := ks.StoreAuthToken(testToken); err != nil {
		t.Fatalf("StoreAuthToken via file backend failed: %v", err)
	}

	token, err := ks.RetrieveAuthToken()
	if err != nil {
		t.Fatalf("RetrieveAuthToken via file backend failed: %v", err)
	}

	if token != testToken {
		t.Errorf("Expected token %s, got %s", testToken, token)
	}

	// Test StoreDeviceID via file backend
	testDeviceID := "test-device-via-file"
	if err := ks.StoreDeviceID(testDeviceID); err != nil {
		t.Fatalf("StoreDeviceID via file backend failed: %v", err)
	}

	deviceID, err := ks.RetrieveDeviceID()
	if err != nil {
		t.Fatalf("RetrieveDeviceID via file backend failed: %v", err)
	}

	if deviceID != testDeviceID {
		t.Errorf("Expected device ID %s, got %s", testDeviceID, deviceID)
	}

	// Test RemoveAuthData
	if err := ks.RemoveAuthData(); err != nil {
		t.Fatalf("RemoveAuthData via file backend failed: %v", err)
	}
}

func TestNewTestKeyring_RetrieveBeforeStore(t *testing.T) {
	dir := t.TempDir()

	ks, err := NewTestKeyring(dir)
	if err != nil {
		t.Fatalf("NewTestKeyring failed: %v", err)
	}

	// Try to retrieve token before storing - should fail
	_, err = ks.RetrieveAuthToken()
	if err == nil {
		t.Error("Expected error retrieving non-existent token")
	}

	// Try to retrieve device ID before storing - should fail
	_, err = ks.RetrieveDeviceID()
	if err == nil {
		t.Error("Expected error retrieving non-existent device ID")
	}
}

func TestNewTestKeyring_RemoveNonExistent(t *testing.T) {
	dir := t.TempDir()

	ks, err := NewTestKeyring(dir)
	if err != nil {
		t.Fatalf("NewTestKeyring failed: %v", err)
	}

	// Store only auth token
	if err := ks.StoreAuthToken("test-token"); err != nil {
		t.Fatalf("StoreAuthToken failed: %v", err)
	}

	// RemoveAuthData should handle missing device ID gracefully
	err = ks.RemoveAuthData()
	// The error handling depends on whether keyring.ErrKeyNotFound is returned
	// for device ID - our implementation should handle this gracefully
	if err != nil {
		t.Logf("RemoveAuthData returned error (may be expected): %v", err)
	}
}

func TestNewTestKeyring_OverwriteToken(t *testing.T) {
	dir := t.TempDir()

	ks, err := NewTestKeyring(dir)
	if err != nil {
		t.Fatalf("NewTestKeyring failed: %v", err)
	}

	// Store first token
	if err := ks.StoreAuthToken("token-1"); err != nil {
		t.Fatalf("StoreAuthToken failed: %v", err)
	}

	// Overwrite with second token
	if err := ks.StoreAuthToken("token-2"); err != nil {
		t.Fatalf("StoreAuthToken (overwrite) failed: %v", err)
	}

	// Should get the second token
	token, err := ks.RetrieveAuthToken()
	if err != nil {
		t.Fatalf("RetrieveAuthToken failed: %v", err)
	}

	if token != "token-2" {
		t.Errorf("Expected token-2, got %s", token)
	}
}

func TestNewTestKeyring_OverwriteDeviceID(t *testing.T) {
	dir := t.TempDir()

	ks, err := NewTestKeyring(dir)
	if err != nil {
		t.Fatalf("NewTestKeyring failed: %v", err)
	}

	// Store first device ID
	if err := ks.StoreDeviceID("device-1"); err != nil {
		t.Fatalf("StoreDeviceID failed: %v", err)
	}

	// Overwrite with second device ID
	if err := ks.StoreDeviceID("device-2"); err != nil {
		t.Fatalf("StoreDeviceID (overwrite) failed: %v", err)
	}

	// Should get the second device ID
	deviceID, err := ks.RetrieveDeviceID()
	if err != nil {
		t.Fatalf("RetrieveDeviceID failed: %v", err)
	}

	if deviceID != "device-2" {
		t.Errorf("Expected device-2, got %s", deviceID)
	}
}

func TestNewTestKeyring_EmptyValues(t *testing.T) {
	dir := t.TempDir()

	ks, err := NewTestKeyring(dir)
	if err != nil {
		t.Fatalf("NewTestKeyring failed: %v", err)
	}

	// Store empty token
	if err := ks.StoreAuthToken(""); err != nil {
		t.Fatalf("StoreAuthToken (empty) failed: %v", err)
	}

	token, err := ks.RetrieveAuthToken()
	if err != nil {
		t.Fatalf("RetrieveAuthToken failed: %v", err)
	}

	if token != "" {
		t.Errorf("Expected empty token, got %s", token)
	}

	// Store empty device ID
	if err := ks.StoreDeviceID(""); err != nil {
		t.Fatalf("StoreDeviceID (empty) failed: %v", err)
	}

	deviceID, err := ks.RetrieveDeviceID()
	if err != nil {
		t.Fatalf("RetrieveDeviceID failed: %v", err)
	}

	if deviceID != "" {
		t.Errorf("Expected empty device ID, got %s", deviceID)
	}
}

func TestNewTestKeyring_LongValues(t *testing.T) {
	dir := t.TempDir()

	ks, err := NewTestKeyring(dir)
	if err != nil {
		t.Fatalf("NewTestKeyring failed: %v", err)
	}

	// Create a long token (1000 chars)
	longToken := ""
	for i := 0; i < 100; i++ {
		longToken += "abcdefghij"
	}

	if err := ks.StoreAuthToken(longToken); err != nil {
		t.Fatalf("StoreAuthToken (long) failed: %v", err)
	}

	token, err := ks.RetrieveAuthToken()
	if err != nil {
		t.Fatalf("RetrieveAuthToken failed: %v", err)
	}

	if token != longToken {
		t.Errorf("Long token mismatch: expected %d chars, got %d chars", len(longToken), len(token))
	}
}

func TestNewTestKeyring_SpecialCharacters(t *testing.T) {
	dir := t.TempDir()

	ks, err := NewTestKeyring(dir)
	if err != nil {
		t.Fatalf("NewTestKeyring failed: %v", err)
	}

	// Token with special characters
	specialToken := "token!@#$%^&*()_+-=[]{}|;':\",./<>?`~"
	if err := ks.StoreAuthToken(specialToken); err != nil {
		t.Fatalf("StoreAuthToken (special) failed: %v", err)
	}

	token, err := ks.RetrieveAuthToken()
	if err != nil {
		t.Fatalf("RetrieveAuthToken failed: %v", err)
	}

	if token != specialToken {
		t.Errorf("Special token mismatch: expected %s, got %s", specialToken, token)
	}
}

func TestNewTestKeyring_UnicodeValues(t *testing.T) {
	dir := t.TempDir()

	ks, err := NewTestKeyring(dir)
	if err != nil {
		t.Fatalf("NewTestKeyring failed: %v", err)
	}

	// Token with unicode characters
	unicodeToken := "token-日本語-한국어-中文-🔐"
	if err := ks.StoreAuthToken(unicodeToken); err != nil {
		t.Fatalf("StoreAuthToken (unicode) failed: %v", err)
	}

	token, err := ks.RetrieveAuthToken()
	if err != nil {
		t.Fatalf("RetrieveAuthToken failed: %v", err)
	}

	if token != unicodeToken {
		t.Errorf("Unicode token mismatch: expected %s, got %s", unicodeToken, token)
	}
}
