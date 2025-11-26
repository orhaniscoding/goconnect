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
	provider oidcProvider
	verifier idTokenVerifier
	config   oauth2.Config
}

type oidcProvider interface {
	UserInfo(ctx context.Context, tokenSource oauth2.TokenSource) (*oidc.UserInfo, error)
	Verifier(config *oidc.Config) *oidc.IDTokenVerifier
}

type idTokenVerifier interface {
	Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error)
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
		provider: &realOIDCProvider{provider: provider},
		verifier: &realVerifier{verifier: verifier},
		config:   config,
	}, nil
}

// GetLoginURL generates the login URL for the OIDC provider
func (s *OIDCService) GetLoginURL(state string) string {
	return s.config.AuthCodeURL(state)
}

// ExchangeCode exchanges the authorization code for a token and extracts user info
func (s *OIDCService) ExchangeCode(ctx context.Context, code string) (*oidc.IDToken, *UserInfo, error) {
	if s == nil || s.verifier == nil {
		return nil, nil, errors.New("oidc not configured")
	}

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

// ValidateToken verifies a raw ID token using the configured verifier.
func (s *OIDCService) ValidateToken(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	if s == nil || s.verifier == nil {
		return nil, errors.New("oidc not configured")
	}
	return s.verifier.Verify(ctx, rawIDToken)
}

// GetUserInfo retrieves user info using the provider's userinfo endpoint.
func (s *OIDCService) GetUserInfo(ctx context.Context, token *oauth2.Token) (*oidc.UserInfo, error) {
	if s == nil || s.provider == nil {
		return nil, errors.New("oidc not configured")
	}
	return s.provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
}

type realOIDCProvider struct {
	provider *oidc.Provider
}

func (p *realOIDCProvider) UserInfo(ctx context.Context, tokenSource oauth2.TokenSource) (*oidc.UserInfo, error) {
	return p.provider.UserInfo(ctx, tokenSource)
}

func (p *realOIDCProvider) Verifier(config *oidc.Config) *oidc.IDTokenVerifier {
	return p.provider.Verifier(config)
}

type realVerifier struct {
	verifier *oidc.IDTokenVerifier
}

func (v *realVerifier) Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	return v.verifier.Verify(ctx, rawIDToken)
}

// UserInfo holds extracted user information
type UserInfo struct {
	Email string
	Name  string
	Sub   string
}
