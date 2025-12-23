package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/99designs/keyring"
)

const (
	keyringService = "goconnect-daemon"
	sessionKey     = "auth_session"
)

// TokenSession holds the authentication tokens and metadata.
type TokenSession struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
}

// TokenManager handles secure storage and retrieval of authentication tokens.
type TokenManager interface {
	SaveSession(session *TokenSession) error
	LoadSession() (*TokenSession, error)
	ClearSession() error
}

type keyringTokenManager struct {
	kr keyring.Keyring
	mu sync.RWMutex
}

// NewTokenManager creates a new TokenManager using the OS keyring or encrypted file fallback.
func NewTokenManager(configDir string) (TokenManager, error) {
	keyringDir := filepath.Join(configDir, "keyring")
	if err := os.MkdirAll(keyringDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create keyring directory: %w", err)
	}

	kr, err := keyring.Open(keyring.Config{
		ServiceName: keyringService,
		AllowedBackends: []keyring.BackendType{
			keyring.SecretServiceBackend,
			keyring.KeychainBackend,
			keyring.WinCredBackend,
			keyring.FileBackend,
		},
		FileDir: keyringDir,
		FilePasswordFunc: func(_ string) (string, error) {
			return "goconnect-daemon-secret", nil
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open keyring: %w", err)
	}

	return &keyringTokenManager{kr: kr}, nil
}

func (tm *keyringTokenManager) SaveSession(session *TokenSession) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := tm.kr.Set(keyring.Item{
		Key:   sessionKey,
		Data:  data,
		Label: "GoConnect Session",
	}); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

func (tm *keyringTokenManager) LoadSession() (*TokenSession, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	item, err := tm.kr.Get(sessionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session TokenSession
	if err := json.Unmarshal(item.Data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

func (tm *keyringTokenManager) ClearSession() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	_ = tm.kr.Remove(sessionKey)
	return nil
}
