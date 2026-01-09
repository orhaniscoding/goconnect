package service

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/config"
)

// TURNCredentials represents time-limited TURN credentials
type TURNCredentials struct {
	URL        string `json:"url"`
	Username   string `json:"username"`
	Credential string `json:"credential"`
	TTL        int    `json:"ttl"` // seconds until expiry
}

// ICEConfig represents the full ICE configuration for clients
type ICEConfig struct {
	STUNServers []string         `json:"stun_servers"`
	TURNServer  *TURNCredentials `json:"turn_server,omitempty"`
}

// TURNService generates time-limited TURN credentials
type TURNService struct {
	config config.TURNConfig
}

// NewTURNService creates a new TURN credential service
func NewTURNService(cfg config.TURNConfig) *TURNService {
	return &TURNService{config: cfg}
}

// GenerateCredentials creates time-limited TURN credentials
// Uses TURN REST API credential format:
// - username: timestamp:networkID
// - credential: base64(HMAC-SHA1(secret, username))
func (s *TURNService) GenerateCredentials(networkID string) TURNCredentials {
	ttl := s.config.CredentialTTL
	if ttl == 0 {
		ttl = 10 * time.Minute
	}

	// Username format: expiry_timestamp:network_id
	expiry := time.Now().Add(ttl).Unix()
	username := fmt.Sprintf("%d:%s", expiry, networkID)

	// HMAC-SHA1 signature
	credential := s.computeCredential(username)

	return TURNCredentials{
		URL:        s.config.ServerURL,
		Username:   username,
		Credential: credential,
		TTL:        int(ttl.Seconds()),
	}
}

// computeCredential generates the HMAC-SHA1 credential
func (s *TURNService) computeCredential(username string) string {
	mac := hmac.New(sha1.New, []byte(s.config.Secret))
	mac.Write([]byte(username))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// GetICEConfig returns the full ICE configuration for a network
func (s *TURNService) GetICEConfig(networkID string) ICEConfig {
	cfg := ICEConfig{
		STUNServers: s.config.STUNServers,
	}

	// Only include TURN if secret and URL are configured
	if s.config.Secret != "" && s.config.ServerURL != "" {
		creds := s.GenerateCredentials(networkID)
		cfg.TURNServer = &creds
	}

	return cfg
}

// IsConfigured returns true if TURN is properly configured
func (s *TURNService) IsConfigured() bool {
	return s.config.Secret != "" && s.config.ServerURL != ""
}

// ValidateCredentials checks if credentials are valid (for testing)
func (s *TURNService) ValidateCredentials(username, credential string) bool {
	expected := s.computeCredential(username)
	return hmac.Equal([]byte(expected), []byte(credential))
}
