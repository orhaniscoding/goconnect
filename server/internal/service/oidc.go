package service

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OIDCService handles OpenID Connect authentication
type OIDCService struct {
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
	config   oauth2.Config
}

// NewOIDCService creates a new OIDC service
func NewOIDCService(ctx context.Context) (*OIDCService, error) {
	issuer := os.Getenv("OIDC_ISSUER")
	clientID := os.Getenv("OIDC_CLIENT_ID")
	clientSecret := os.Getenv("OIDC_CLIENT_SECRET")
	redirectURL := os.Getenv("OIDC_REDIRECT_URL")

	if issuer == "" || clientID == "" || clientSecret == "" || redirectURL == "" {
		// Return nil error to allow server to start without OIDC configured
		// The handlers should check if service is nil or ready
		return nil, nil
	}

	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to query OIDC provider: %w", err)
	}

	oidcConfig := &oidc.Config{
		ClientID: clientID,
	}
	verifier := provider.Verifier(oidcConfig)

	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &OIDCService{
		provider: provider,
		verifier: verifier,
		config:   config,
	}, nil
}

// GetLoginURL generates the login URL for the OIDC provider
func (s *OIDCService) GetLoginURL(state string) string {
	return s.config.AuthCodeURL(state)
}

// ExchangeCode exchanges the authorization code for a token and extracts user info
func (s *OIDCService) ExchangeCode(ctx context.Context, code string) (*oidc.IDToken, *UserInfo, error) {
	oauth2Token, err := s.config.Exchange(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, nil, errors.New("no id_token field in oauth2 token")
	}

	idToken, err := s.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to verify ID Token: %w", err)
	}

	var claims struct {
		Email    string `json:"email"`
		Verified bool   `json:"email_verified"`
		Name     string `json:"name"`
		Sub      string `json:"sub"`
	}

	if err := idToken.Claims(&claims); err != nil {
		return nil, nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	userInfo := &UserInfo{
		Email: claims.Email,
		Name:  claims.Name,
		Sub:   claims.Sub,
	}

	return idToken, userInfo, nil
}

// UserInfo holds extracted user information
type UserInfo struct {
	Email string
	Name  string
	Sub   string
}
