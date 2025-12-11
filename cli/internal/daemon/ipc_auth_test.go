package daemon_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/orhaniscoding/goconnect/client-daemon/internal/daemon"
)

func TestIPCAuth_GenerateAndSave(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "test.token")

	// Create auth with custom path
	auth := daemon.NewIPCAuthWithPath(tokenPath)

	// Generate and save token
	err := auth.GenerateAndSave()
	if err != nil {
		t.Fatalf("GenerateAndSave failed: %v", err)
	}

	// Verify token file exists
	if _, err := os.Stat(tokenPath); os.IsNotExist(err) {
		t.Fatal("Token file was not created")
	}

	// Verify token is non-empty
	token := auth.GetToken()
	if token == "" {
		t.Fatal("Token is empty")
	}

	// Token should be 64 hex chars (32 bytes * 2)
	if len(token) != 64 {
		t.Errorf("Token length = %d, want 64", len(token))
	}

	// Verify file contents match
	fileContent, err := os.ReadFile(tokenPath)
	if err != nil {
		t.Fatalf("Failed to read token file: %v", err)
	}
	if string(fileContent) != token {
		t.Error("Token file content doesn't match internal token")
	}

	// Verify file permissions (should be 0600)
	// On Windows, file permissions work differently, so skip this check
	if runtime.GOOS != "windows" {
		info, _ := os.Stat(tokenPath)
		perm := info.Mode().Perm()
		if perm != 0600 {
			t.Errorf("Token file permissions = %o, want 0600", perm)
		}
	}
}

func TestIPCAuth_ValidateToken(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "test.token")
	auth := daemon.NewIPCAuthWithPath(tokenPath)

	// Generate token
	if err := auth.GenerateAndSave(); err != nil {
		t.Fatalf("GenerateAndSave failed: %v", err)
	}

	validToken := auth.GetToken()

	tests := []struct {
		name  string
		token string
		want  bool
	}{
		{"valid token", validToken, true},
		{"empty token", "", false},
		{"wrong token", "wrongtoken123", false},
		{"partial token", validToken[:10], false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := auth.ValidateToken(tt.token); got != tt.want {
				t.Errorf("ValidateToken(%q) = %v, want %v", tt.token, got, tt.want)
			}
		})
	}
}

func TestIPCAuth_Cleanup(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "test.token")
	auth := daemon.NewIPCAuthWithPath(tokenPath)

	// Generate token
	if err := auth.GenerateAndSave(); err != nil {
		t.Fatalf("GenerateAndSave failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tokenPath); os.IsNotExist(err) {
		t.Fatal("Token file should exist before cleanup")
	}

	// Cleanup
	if err := auth.Cleanup(); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify file is deleted
	if _, err := os.Stat(tokenPath); !os.IsNotExist(err) {
		t.Error("Token file should be deleted after cleanup")
	}

	// Token should be cleared
	if auth.GetToken() != "" {
		t.Error("Token should be empty after cleanup")
	}

	// Cleanup on non-existent file should not error
	if err := auth.Cleanup(); err != nil {
		t.Errorf("Cleanup on non-existent file should not error: %v", err)
	}
}

func TestIPCAuth_TokenUniqueness(t *testing.T) {
	tmpDir := t.TempDir()

	// Generate multiple tokens and ensure they're unique
	tokens := make(map[string]bool)

	for i := 0; i < 10; i++ {
		tokenPath := filepath.Join(tmpDir, "test.token")
		auth := daemon.NewIPCAuthWithPath(tokenPath)

		if err := auth.GenerateAndSave(); err != nil {
			t.Fatalf("GenerateAndSave failed: %v", err)
		}

		token := auth.GetToken()
		if tokens[token] {
			t.Error("Duplicate token generated")
		}
		tokens[token] = true

		auth.Cleanup()
	}
}

func TestLoadClientToken(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "test.token")

	// Create auth and generate token
	auth := daemon.NewIPCAuthWithPath(tokenPath)
	if err := auth.GenerateAndSave(); err != nil {
		t.Fatalf("GenerateAndSave failed: %v", err)
	}

	expectedToken := auth.GetToken()

	// Load token using client function
	loadedToken, err := daemon.LoadClientTokenFromPath(tokenPath)
	if err != nil {
		t.Fatalf("LoadClientTokenFromPath failed: %v", err)
	}

	if loadedToken != expectedToken {
		t.Errorf("Loaded token = %q, want %q", loadedToken, expectedToken)
	}
}

func TestLoadClientToken_NotFound(t *testing.T) {
	_, err := daemon.LoadClientTokenFromPath("/nonexistent/path/token")
	if err == nil {
		t.Error("LoadClientTokenFromPath should fail for non-existent file")
	}
}

func TestIPCAuth_GetTokenPath(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "test.token")
	auth := daemon.NewIPCAuthWithPath(tokenPath)

	if got := auth.GetTokenPath(); got != tokenPath {
		t.Errorf("GetTokenPath() = %v, want %v", got, tokenPath)
	}
}

func TestNewIPCAuth(t *testing.T) {
	// Test that NewIPCAuth creates an auth with the default path
	auth := daemon.NewIPCAuth()
	if auth == nil {
		t.Fatal("NewIPCAuth returned nil")
	}

	// Token path should not be empty
	if auth.GetTokenPath() == "" {
		t.Error("Token path should not be empty")
	}
}

func TestTokenCredentials(t *testing.T) {
	token := "test-token-12345"
	creds := daemon.NewTokenCredentials(token)

	t.Run("GetRequestMetadata", func(t *testing.T) {
		md, err := creds.GetRequestMetadata(nil)
		if err != nil {
			t.Fatalf("GetRequestMetadata failed: %v", err)
		}
		if md["x-goconnect-ipc-token"] != token {
			t.Errorf("Token mismatch: got %v, want %v", md["x-goconnect-ipc-token"], token)
		}
	})

	t.Run("RequireTransportSecurity", func(t *testing.T) {
		if creds.RequireTransportSecurity() {
			t.Error("RequireTransportSecurity should return false for IPC")
		}
	})
}

func TestIPCAuth_StreamServerInterceptor(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "test.token")
	auth := daemon.NewIPCAuthWithPath(tokenPath)

	if err := auth.GenerateAndSave(); err != nil {
		t.Fatalf("GenerateAndSave failed: %v", err)
	}

	// Get the interceptor - just ensure it's not nil
	interceptor := auth.StreamServerInterceptor()
	if interceptor == nil {
		t.Fatal("StreamServerInterceptor returned nil")
	}
}

func TestIPCAuth_UnaryServerInterceptor(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "test.token")
	auth := daemon.NewIPCAuthWithPath(tokenPath)

	if err := auth.GenerateAndSave(); err != nil {
		t.Fatalf("GenerateAndSave failed: %v", err)
	}

	// Get the interceptor - just ensure it's not nil
	interceptor := auth.UnaryServerInterceptor()
	if interceptor == nil {
		t.Fatal("UnaryServerInterceptor returned nil")
	}
}

func TestIPCAuth_GenerateAndSave_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	// Nested path that doesn't exist
	tokenPath := filepath.Join(tmpDir, "nested", "dir", "test.token")
	auth := daemon.NewIPCAuthWithPath(tokenPath)

	err := auth.GenerateAndSave()
	if err != nil {
		t.Fatalf("GenerateAndSave should create directory: %v", err)
	}

	// Verify directory was created
	dir := filepath.Dir(tokenPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("Directory should have been created")
	}

	// Verify token file exists
	if _, err := os.Stat(tokenPath); os.IsNotExist(err) {
		t.Error("Token file should exist")
	}
}

func TestLoadClientTokenFromPath_WithWhitespace(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "test.token")

	// Write token with trailing whitespace
	tokenWithWhitespace := "abc123def456\n  \t"
	if err := os.WriteFile(tokenPath, []byte(tokenWithWhitespace), 0600); err != nil {
		t.Fatalf("Failed to write token file: %v", err)
	}

	// LoadClientTokenFromPath should trim whitespace
	token, err := daemon.LoadClientTokenFromPath(tokenPath)
	if err != nil {
		t.Fatalf("LoadClientTokenFromPath failed: %v", err)
	}

	if token != "abc123def456" {
		t.Errorf("Token should be trimmed, got %q", token)
	}
}
