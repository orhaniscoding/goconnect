package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

type stubVerifier struct {
	token *oidc.IDToken
	err   error
	calls int
}

func (s *stubVerifier) Verify(_ context.Context, rawIDToken string) (*oidc.IDToken, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return s.token, nil
}

type stubProvider struct {
	userInfo *oidc.UserInfo
	err      error
	calls    int
}

func (s *stubProvider) UserInfo(ctx context.Context, tokenSource oauth2.TokenSource) (*oidc.UserInfo, error) {
	_ = tokenSource // unused in stub
	s.calls++
	return s.userInfo, s.err
}

func (s *stubProvider) Verifier(_ *oidc.Config) *oidc.IDTokenVerifier {
	return nil
}

// newIDTokenWithClaims constructs an oidc.IDToken with provided claims using reflection to set raw claims.
func newIDTokenWithClaims(t *testing.T, claims map[string]any) *oidc.IDToken {
	t.Helper()
	raw, err := json.Marshal(claims)
	require.NoError(t, err)
	token := &oidc.IDToken{}
	val := reflect.ValueOf(token).Elem().FieldByName("claims")
	require.True(t, val.IsValid(), "claims field not found on oidc.IDToken")
	// Allow setting unexported field
	reflect.NewAt(val.Type(), unsafe.Pointer(val.UnsafeAddr())).Elem().SetBytes(raw)
	return token
}

func TestOIDCService_ExchangeCode_Success(t *testing.T) {
	claims := map[string]any{"email": "user@example.com", "email_verified": true, "name": "Test User", "sub": "sub123"}
	idToken := newIDTokenWithClaims(t, claims)

	verifier := &stubVerifier{token: idToken}
	svc := &OIDCService{
		verifier: verifier,
		config: oauth2.Config{
			ClientID:    "cid",
			RedirectURL: "http://localhost/callback",
		},
	}

	// Token endpoint stub
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.Form.Get("code") == "bad" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "access",
			"expires_in":   3600,
			"id_token":     "dummy-id-token",
			"token_type":   "Bearer",
		})
	}))
	defer ts.Close()
	svc.config.Endpoint = oauth2.Endpoint{TokenURL: ts.URL}

	token, info, err := svc.ExchangeCode(context.Background(), "good")

	require.NoError(t, err)
	require.NotNil(t, token)
	require.NotNil(t, info)
	assert.Equal(t, "user@example.com", info.Email)
	assert.Equal(t, "Test User", info.Name)
	assert.Equal(t, "sub123", info.Sub)
	assert.Equal(t, 1, verifier.calls)
}

func TestOIDCService_ExchangeCode_Failure(t *testing.T) {
	verifier := &stubVerifier{err: errors.New("verify failed")}
	svc := &OIDCService{
		verifier: verifier,
		config: oauth2.Config{
			ClientID:    "cid",
			RedirectURL: "http://localhost/callback",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()
	svc.config.Endpoint = oauth2.Endpoint{TokenURL: ts.URL}

	token, info, err := svc.ExchangeCode(context.Background(), "bad")

	require.Error(t, err)
	assert.Nil(t, token)
	assert.Nil(t, info)
}

func TestOIDCService_ValidateToken(t *testing.T) {
	idToken := newIDTokenWithClaims(t, map[string]any{"sub": "abc"})
	verifier := &stubVerifier{token: idToken}
	svc := &OIDCService{verifier: verifier}

	tok, err := svc.ValidateToken(context.Background(), "raw")

	require.NoError(t, err)
	assert.Equal(t, idToken, tok)
	assert.Equal(t, 1, verifier.calls)
}

func TestOIDCService_ValidateToken_Error(t *testing.T) {
	verifier := &stubVerifier{err: errors.New("boom")}
	svc := &OIDCService{verifier: verifier}

	tok, err := svc.ValidateToken(context.Background(), "raw")

	require.Error(t, err)
	assert.Nil(t, tok)
}

func TestOIDCService_GetUserInfo(t *testing.T) {
	provider := &stubProvider{
		userInfo: &oidc.UserInfo{
			Email:         "a@example.com",
			EmailVerified: true,
			Subject:       "subj",
		},
	}
	svc := &OIDCService{provider: provider}

	info, err := svc.GetUserInfo(context.Background(), &oauth2.Token{AccessToken: "token"})

	require.NoError(t, err)
	assert.Equal(t, "a@example.com", info.Email)
	assert.Equal(t, 1, provider.calls)
}

func TestOIDCService_GetUserInfo_Error(t *testing.T) {
	provider := &stubProvider{err: errors.New("userinfo error")}
	svc := &OIDCService{provider: provider}

	info, err := svc.GetUserInfo(context.Background(), &oauth2.Token{AccessToken: "token"})

	require.Error(t, err)
	assert.Nil(t, info)
	assert.Equal(t, 1, provider.calls)
}

// ==================== NewOIDCService Tests ====================

func TestNewOIDCService_MissingEnvVars(t *testing.T) {
	// Clear any existing env vars
	origIssuer := os.Getenv("OIDC_ISSUER")
	origClientID := os.Getenv("OIDC_CLIENT_ID")
	origClientSecret := os.Getenv("OIDC_CLIENT_SECRET")
	origRedirectURL := os.Getenv("OIDC_REDIRECT_URL")
	defer func() {
		os.Setenv("OIDC_ISSUER", origIssuer)
		os.Setenv("OIDC_CLIENT_ID", origClientID)
		os.Setenv("OIDC_CLIENT_SECRET", origClientSecret)
		os.Setenv("OIDC_REDIRECT_URL", origRedirectURL)
	}()

	t.Run("All env vars missing returns nil service and nil error", func(t *testing.T) {
		os.Unsetenv("OIDC_ISSUER")
		os.Unsetenv("OIDC_CLIENT_ID")
		os.Unsetenv("OIDC_CLIENT_SECRET")
		os.Unsetenv("OIDC_REDIRECT_URL")

		svc, err := NewOIDCService(context.Background())

		assert.Nil(t, svc)
		assert.Nil(t, err)
	})

	t.Run("Missing issuer returns nil service", func(t *testing.T) {
		os.Unsetenv("OIDC_ISSUER")
		os.Setenv("OIDC_CLIENT_ID", "test-client-id")
		os.Setenv("OIDC_CLIENT_SECRET", "test-secret")
		os.Setenv("OIDC_REDIRECT_URL", "http://localhost/callback")

		svc, err := NewOIDCService(context.Background())

		assert.Nil(t, svc)
		assert.Nil(t, err)
	})

	t.Run("Missing client ID returns nil service", func(t *testing.T) {
		os.Setenv("OIDC_ISSUER", "https://issuer.example.com")
		os.Unsetenv("OIDC_CLIENT_ID")
		os.Setenv("OIDC_CLIENT_SECRET", "test-secret")
		os.Setenv("OIDC_REDIRECT_URL", "http://localhost/callback")

		svc, err := NewOIDCService(context.Background())

		assert.Nil(t, svc)
		assert.Nil(t, err)
	})

	t.Run("Missing client secret returns nil service", func(t *testing.T) {
		os.Setenv("OIDC_ISSUER", "https://issuer.example.com")
		os.Setenv("OIDC_CLIENT_ID", "test-client-id")
		os.Unsetenv("OIDC_CLIENT_SECRET")
		os.Setenv("OIDC_REDIRECT_URL", "http://localhost/callback")

		svc, err := NewOIDCService(context.Background())

		assert.Nil(t, svc)
		assert.Nil(t, err)
	})

	t.Run("Missing redirect URL returns nil service", func(t *testing.T) {
		os.Setenv("OIDC_ISSUER", "https://issuer.example.com")
		os.Setenv("OIDC_CLIENT_ID", "test-client-id")
		os.Setenv("OIDC_CLIENT_SECRET", "test-secret")
		os.Unsetenv("OIDC_REDIRECT_URL")

		svc, err := NewOIDCService(context.Background())

		assert.Nil(t, svc)
		assert.Nil(t, err)
	})
}

func TestNewOIDCService_InvalidIssuer(t *testing.T) {
	// Clear any existing env vars
	origIssuer := os.Getenv("OIDC_ISSUER")
	origClientID := os.Getenv("OIDC_CLIENT_ID")
	origClientSecret := os.Getenv("OIDC_CLIENT_SECRET")
	origRedirectURL := os.Getenv("OIDC_REDIRECT_URL")
	defer func() {
		os.Setenv("OIDC_ISSUER", origIssuer)
		os.Setenv("OIDC_CLIENT_ID", origClientID)
		os.Setenv("OIDC_CLIENT_SECRET", origClientSecret)
		os.Setenv("OIDC_REDIRECT_URL", origRedirectURL)
	}()

	t.Run("Invalid issuer URL returns error", func(t *testing.T) {
		os.Setenv("OIDC_ISSUER", "http://invalid-issuer-that-does-not-exist.local")
		os.Setenv("OIDC_CLIENT_ID", "test-client-id")
		os.Setenv("OIDC_CLIENT_SECRET", "test-secret")
		os.Setenv("OIDC_REDIRECT_URL", "http://localhost/callback")

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		svc, err := NewOIDCService(ctx)

		assert.Nil(t, svc)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to query OIDC provider")
	})
}

// ==================== GetLoginURL Tests ====================

func TestOIDCService_GetLoginURL(t *testing.T) {
	svc := &OIDCService{
		config: oauth2.Config{
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost/callback",
			Endpoint: oauth2.Endpoint{
				AuthURL: "https://auth.example.com/authorize",
			},
			Scopes: []string{"openid", "profile", "email"},
		},
	}

	t.Run("Generates login URL with state parameter", func(t *testing.T) {
		state := "test-state-123"
		url := svc.GetLoginURL(state)

		assert.Contains(t, url, "https://auth.example.com/authorize")
		assert.Contains(t, url, "client_id=test-client-id")
		assert.Contains(t, url, "redirect_uri=")
		assert.Contains(t, url, "state=test-state-123")
		assert.Contains(t, url, "response_type=code")
	})

	t.Run("Generates login URL with empty state", func(t *testing.T) {
		url := svc.GetLoginURL("")

		assert.Contains(t, url, "https://auth.example.com/authorize")
		assert.Contains(t, url, "client_id=test-client-id")
	})

	t.Run("URL contains scope parameter", func(t *testing.T) {
		url := svc.GetLoginURL("state")

		assert.Contains(t, url, "scope=")
	})
}

// ==================== Nil Service Tests ====================

func TestOIDCService_ExchangeCode_NilService(t *testing.T) {
	var svc *OIDCService = nil

	token, info, err := svc.ExchangeCode(context.Background(), "code")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "oidc not configured")
	assert.Nil(t, token)
	assert.Nil(t, info)
}

func TestOIDCService_ValidateToken_NilService(t *testing.T) {
	var svc *OIDCService = nil

	token, err := svc.ValidateToken(context.Background(), "raw-token")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "oidc not configured")
	assert.Nil(t, token)
}

func TestOIDCService_GetUserInfo_NilService(t *testing.T) {
	var svc *OIDCService = nil

	info, err := svc.GetUserInfo(context.Background(), &oauth2.Token{AccessToken: "token"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "oidc not configured")
	assert.Nil(t, info)
}

func TestOIDCService_ExchangeCode_NilVerifier(t *testing.T) {
	svc := &OIDCService{
		verifier: nil,
		config: oauth2.Config{
			ClientID:    "cid",
			RedirectURL: "http://localhost/callback",
		},
	}

	token, info, err := svc.ExchangeCode(context.Background(), "code")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "oidc not configured")
	assert.Nil(t, token)
	assert.Nil(t, info)
}

func TestOIDCService_ValidateToken_NilVerifier(t *testing.T) {
	svc := &OIDCService{
		verifier: nil,
	}

	token, err := svc.ValidateToken(context.Background(), "raw-token")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "oidc not configured")
	assert.Nil(t, token)
}

func TestOIDCService_GetUserInfo_NilProvider(t *testing.T) {
	svc := &OIDCService{
		provider: nil,
	}

	info, err := svc.GetUserInfo(context.Background(), &oauth2.Token{AccessToken: "token"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "oidc not configured")
	assert.Nil(t, info)
}
