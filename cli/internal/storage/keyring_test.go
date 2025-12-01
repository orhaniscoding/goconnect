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
	}
	if ks.kr == nil {
		t.Error("Expected keyring to be initialized")
	}
}

// MockKeyringStore tests - testing the interface contract
func TestKeyringStore_Methods(t *testing.T) {
	// This test verifies that KeyringStore has the expected methods
	// by checking method signatures compile correctly
	var store *KeyringStore

	// These will panic if called on nil, but they verify the methods exist
	_ = func() { store.StoreAuthToken("test") }
	_ = func() { store.RetrieveAuthToken() }
	_ = func() { store.StoreDeviceID("test") }
	_ = func() { store.RetrieveDeviceID() }
	_ = func() { store.RemoveAuthData() }

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
