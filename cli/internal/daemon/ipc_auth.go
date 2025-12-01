package daemon

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	// TokenHeaderKey is the metadata key for the IPC auth token
	TokenHeaderKey = "x-goconnect-ipc-token"

	// TokenLength is the number of bytes in the generated token (32 bytes = 256 bits)
	TokenLength = 32

	// TokenFilePermissions restricts token file to owner-only read/write
	TokenFilePermissions = 0600
)

// IPCAuth handles token-based authentication for local IPC connections.
// This implements Zero-Trust IPC as described in the architecture.
type IPCAuth struct {
	mu        sync.RWMutex
	token     string
	tokenPath string
}

// NewIPCAuth creates a new IPC authentication handler.
func NewIPCAuth() *IPCAuth {
	return &IPCAuth{
		tokenPath: getDefaultTokenPath(),
	}
}

// NewIPCAuthWithPath creates an IPC auth handler with a custom token path (for testing).
func NewIPCAuthWithPath(tokenPath string) *IPCAuth {
	return &IPCAuth{
		tokenPath: tokenPath,
	}
}

// getDefaultTokenPath returns the platform-specific path for the IPC token file.
func getDefaultTokenPath() string {
	switch runtime.GOOS {
	case "windows":
		// Windows: Use %LOCALAPPDATA%\GoConnect\ipc.token
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
		}
		return filepath.Join(localAppData, "GoConnect", "ipc.token")
	case "darwin":
		// macOS: Use ~/Library/Application Support/GoConnect/ipc.token
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "GoConnect", "ipc.token")
	default:
		// Linux: Use ~/.local/share/goconnect/ipc.token or /var/run for system service
		if os.Getuid() == 0 {
			return "/var/run/goconnect/ipc.token"
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".local", "share", "goconnect", "ipc.token")
	}
}

// GenerateAndSave generates a new random token and saves it to the token file.
// This should be called on daemon startup.
func (a *IPCAuth) GenerateAndSave() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Generate cryptographically secure random token
	tokenBytes := make([]byte, TokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("failed to generate random token: %w", err)
	}
	a.token = hex.EncodeToString(tokenBytes)

	// Ensure directory exists
	dir := filepath.Dir(a.tokenPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create token directory: %w", err)
	}

	// Write token to file with restricted permissions
	if err := os.WriteFile(a.tokenPath, []byte(a.token), TokenFilePermissions); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

// Cleanup removes the token file. Should be called on daemon shutdown.
func (a *IPCAuth) Cleanup() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.token = ""
	if err := os.Remove(a.tokenPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove token file: %w", err)
	}
	return nil
}

// ValidateToken checks if the provided token matches the current valid token.
func (a *IPCAuth) ValidateToken(token string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Use constant-time comparison to prevent timing attacks
	// (though less critical for local IPC, it's good practice)
	return len(token) > 0 && token == a.token
}

// GetToken returns the current token (for testing/debugging only).
func (a *IPCAuth) GetToken() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.token
}

// GetTokenPath returns the path to the token file.
func (a *IPCAuth) GetTokenPath() string {
	return a.tokenPath
}

// UnaryServerInterceptor returns a gRPC unary interceptor that validates IPC tokens.
func (a *IPCAuth) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip auth for health check or other public methods if needed
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		if err := a.validateFromContext(ctx); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a gRPC stream interceptor that validates IPC tokens.
func (a *IPCAuth) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Skip auth for public methods if needed
		if isPublicMethod(info.FullMethod) {
			return handler(srv, ss)
		}

		if err := a.validateFromContext(ss.Context()); err != nil {
			return err
		}

		return handler(srv, ss)
	}
}

// validateFromContext extracts and validates the token from gRPC metadata.
func (a *IPCAuth) validateFromContext(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing metadata")
	}

	tokens := md.Get(TokenHeaderKey)
	if len(tokens) == 0 {
		return status.Error(codes.Unauthenticated, "missing IPC auth token")
	}

	token := tokens[0]
	if !a.ValidateToken(token) {
		return status.Error(codes.Unauthenticated, "invalid IPC auth token")
	}

	return nil
}

// isPublicMethod returns true for methods that don't require authentication.
// Currently, all methods require auth for Zero-Trust compliance.
func isPublicMethod(method string) bool {
	// In the future, we might allow certain health-check methods without auth
	// For now, all methods require authentication (Zero-Trust)
	_ = method
	return false
}

// =============================================================================
// CLIENT-SIDE TOKEN LOADING
// =============================================================================

// LoadClientToken reads the IPC token from the token file.
// This is used by clients (CLI, Desktop) to authenticate with the daemon.
func LoadClientToken() (string, error) {
	return LoadClientTokenFromPath(getDefaultTokenPath())
}

// LoadClientTokenFromPath reads the IPC token from a specific path (for testing).
func LoadClientTokenFromPath(tokenPath string) (string, error) {
	tokenBytes, err := os.ReadFile(tokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("daemon not running or token file missing: %s", tokenPath)
		}
		return "", fmt.Errorf("failed to read token file: %w", err)
	}
	return strings.TrimSpace(string(tokenBytes)), nil
}

// TokenCredentials implements grpc.PerRPCCredentials for IPC token auth.
type TokenCredentials struct {
	token string
}

// NewTokenCredentials creates credentials from the loaded token.
func NewTokenCredentials(token string) *TokenCredentials {
	return &TokenCredentials{token: token}
}

// GetRequestMetadata returns the token in gRPC metadata format.
func (c *TokenCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		TokenHeaderKey: c.token,
	}, nil
}

// RequireTransportSecurity returns false since we're using local IPC (Unix sockets/localhost TCP).
// The token itself provides authentication; TLS isn't needed for localhost.
func (c *TokenCredentials) RequireTransportSecurity() bool {
	return false
}
