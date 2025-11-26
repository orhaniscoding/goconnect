package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
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
