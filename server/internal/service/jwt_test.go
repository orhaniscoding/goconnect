package service

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetJWTSecret_FromEnv(t *testing.T) {
	// Save original env
	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	// Set custom secret
	testSecret := "test-secret-key-12345"
	os.Setenv("JWT_SECRET", testSecret)

	secret := getJWTSecret()
	assert.Equal(t, []byte(testSecret), secret)
}

func TestGetJWTSecret_DefaultFallback(t *testing.T) {
	// Save original env
	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	// Clear JWT_SECRET
	os.Unsetenv("JWT_SECRET")

	secret := getJWTSecret()
	assert.Equal(t, []byte("dev-secret-DO-NOT-USE-IN-PRODUCTION"), secret)
}

func TestGenerateTokenPair_Success(t *testing.T) {
	accessToken, refreshToken, expiresIn, err := GenerateTokenPair(
		"user-123",
		"tenant-456",
		"user@example.com",
		false,
		false,
	)

	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, int(defaultAccessTTL.Seconds()), expiresIn)
	assert.Equal(t, 900, expiresIn) // 15 minutes = 900 seconds

	// Tokens should be different
	assert.NotEqual(t, accessToken, refreshToken)

	// Tokens should be valid JWT format (3 parts separated by dots)
	assert.Equal(t, 3, len(strings.Split(accessToken, ".")))
	assert.Equal(t, 3, len(strings.Split(refreshToken, ".")))
}

func TestGenerateTokenPair_AdminUser(t *testing.T) {
	accessToken, refreshToken, expiresIn, err := GenerateTokenPair(
		"admin-123",
		"tenant-456",
		"admin@example.com",
		true, // isAdmin
		false,
	)

	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, 900, expiresIn)

	// Validate access token contains admin flag
	claims, err := ValidateToken(accessToken)
	require.NoError(t, err)
	assert.True(t, claims.IsAdmin)
	assert.False(t, claims.IsModerator)
}

func TestGenerateTokenPair_ModeratorUser(t *testing.T) {
	accessToken, refreshToken, _, err := GenerateTokenPair(
		"mod-123",
		"tenant-456",
		"mod@example.com",
		false,
		true, // isModerator
	)

	require.NoError(t, err)

	// Validate access token contains moderator flag
	claims, err := ValidateToken(accessToken)
	require.NoError(t, err)
	assert.False(t, claims.IsAdmin)
	assert.True(t, claims.IsModerator)

	// Validate refresh token also contains flags
	refreshClaims, err := ValidateToken(refreshToken)
	require.NoError(t, err)
	assert.False(t, refreshClaims.IsAdmin)
	assert.True(t, refreshClaims.IsModerator)
}

func TestGenerateTokenPair_TokenTypes(t *testing.T) {
	accessToken, refreshToken, _, err := GenerateTokenPair(
		"user-123",
		"tenant-456",
		"user@example.com",
		false,
		false,
	)

	require.NoError(t, err)

	// Validate token types
	accessClaims, err := ValidateToken(accessToken)
	require.NoError(t, err)
	assert.Equal(t, "access", accessClaims.Type)

	refreshClaims, err := ValidateToken(refreshToken)
	require.NoError(t, err)
	assert.Equal(t, "refresh", refreshClaims.Type)
}

func TestGenerateTokenPair_UniqueTokenIDs(t *testing.T) {
	// Generate multiple token pairs
	tokens := make(map[string]bool)

	for i := 0; i < 10; i++ {
		accessToken, refreshToken, _, err := GenerateTokenPair(
			"user-123",
			"tenant-456",
			"user@example.com",
			false,
			false,
		)
		require.NoError(t, err)

		// Each token should be unique
		assert.False(t, tokens[accessToken], "Access token should be unique")
		assert.False(t, tokens[refreshToken], "Refresh token should be unique")

		tokens[accessToken] = true
		tokens[refreshToken] = true
	}

	// Should have 20 unique tokens (10 pairs)
	assert.Equal(t, 20, len(tokens))
}

func TestJWT_ValidateToken_Success(t *testing.T) {
	accessToken, _, _, err := GenerateTokenPair(
		"user-123",
		"tenant-456",
		"user@example.com",
		false,
		false,
	)
	require.NoError(t, err)

	claims, err := ValidateToken(accessToken)
	require.NoError(t, err)

	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "tenant-456", claims.TenantID)
	assert.Equal(t, "user@example.com", claims.Email)
	assert.False(t, claims.IsAdmin)
	assert.False(t, claims.IsModerator)
	assert.Equal(t, "access", claims.Type)
}

func TestJWT_ValidateToken_InvalidFormat(t *testing.T) {
	invalidTokens := []string{
		"",
		"invalid",
		"invalid.token",
		"invalid.token.format.extra",
		"not-a-jwt-token",
	}

	for _, token := range invalidTokens {
		_, err := ValidateToken(token)
		assert.Error(t, err, "Token: %s", token)

		// Should return domain error
		domainErr, ok := err.(*domain.Error)
		assert.True(t, ok, "Should be domain.Error")
		if ok {
			assert.Equal(t, domain.ErrInvalidToken, domainErr.Code)
		}
	}
}

func TestJWT_ValidateToken_InvalidSignature(t *testing.T) {
	// Generate token with one secret
	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	os.Setenv("JWT_SECRET", "secret-1")
	accessToken, _, _, err := GenerateTokenPair(
		"user-123",
		"tenant-456",
		"user@example.com",
		false,
		false,
	)
	require.NoError(t, err)

	// Try to validate with different secret
	os.Setenv("JWT_SECRET", "secret-2")
	_, err = ValidateToken(accessToken)
	assert.Error(t, err)

	domainErr, ok := err.(*domain.Error)
	assert.True(t, ok)
	if ok {
		assert.Equal(t, domain.ErrInvalidToken, domainErr.Code)
	}
}

func TestJWT_ValidateToken_ExpiredToken(t *testing.T) {
	// Create token with custom claims (expired)
	secret := getJWTSecret()
	now := time.Now()

	expiredClaims := JWTClaims{
		UserID:   "user-123",
		TenantID: "tenant-456",
		Email:    "user@example.com",
		Type:     "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			Issuer:    "goconnect",
			Subject:   "user-123",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	tokenString, err := token.SignedString(secret)
	require.NoError(t, err)

	// Validate should fail - JWT library catches expiration during parse
	_, err = ValidateToken(tokenString)
	assert.Error(t, err)

	// Should return a domain error (invalid token since JWT lib catches it first)
	domainErr, ok := err.(*domain.Error)
	assert.True(t, ok)
	if ok {
		// JWT library returns generic error which gets wrapped as ErrInvalidToken
		assert.Equal(t, domain.ErrInvalidToken, domainErr.Code)
		if details, ok := domainErr.Details.(map[string]string); ok {
			assert.Contains(t, details["error"], "expired")
		}
	}
}

func TestJWT_ValidateToken_WrongSigningMethod(t *testing.T) {
	// Create token with RSA instead of HMAC
	claims := JWTClaims{
		UserID:   "user-123",
		TenantID: "tenant-456",
		Type:     "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			Issuer:    "goconnect",
		},
	}

	// This will create a token with "none" signing method which should fail
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	_, err = ValidateToken(tokenString)
	assert.Error(t, err)
}

func TestJWT_ValidateToken_AllFields(t *testing.T) {
	accessToken, _, _, err := GenerateTokenPair(
		"user-123",
		"tenant-456",
		"user@example.com",
		true, // isAdmin
		true, // isModerator
	)
	require.NoError(t, err)

	claims, err := ValidateToken(accessToken)
	require.NoError(t, err)

	// Verify all fields
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "tenant-456", claims.TenantID)
	assert.Equal(t, "user@example.com", claims.Email)
	assert.True(t, claims.IsAdmin)
	assert.True(t, claims.IsModerator)
	assert.Equal(t, "access", claims.Type)
}

func TestGenerateSecureToken_Success(t *testing.T) {
	lengths := []int{8, 16, 32, 64}

	for _, length := range lengths {
		token, err := GenerateSecureToken(length)
		require.NoError(t, err, "Length: %d", length)
		assert.NotEmpty(t, token)

		// Base64 URL encoding of N bytes produces string of different length
		// Just verify it's not empty and is URL-safe base64
		assert.NotContains(t, token, "+")
		assert.NotContains(t, token, "/")
		assert.NotContains(t, token, "=") // RawURLEncoding doesn't use padding
	}
}

func TestGenerateSecureToken_Uniqueness(t *testing.T) {
	tokens := make(map[string]bool)
	length := 16

	// Generate 100 tokens
	for i := 0; i < 100; i++ {
		token, err := GenerateSecureToken(length)
		require.NoError(t, err)

		// Each token should be unique
		assert.False(t, tokens[token], "Token should be unique")
		tokens[token] = true
	}

	assert.Equal(t, 100, len(tokens), "All tokens should be unique")
}

func TestGenerateSecureToken_ZeroLength(t *testing.T) {
	token, err := GenerateSecureToken(0)
	require.NoError(t, err)
	assert.Empty(t, token)
}

func TestGenerateSecureToken_URLSafe(t *testing.T) {
	// Generate many tokens to ensure URL safety
	for i := 0; i < 100; i++ {
		token, err := GenerateSecureToken(32)
		require.NoError(t, err)

		// URL-safe base64 should not contain +, /, or =
		assert.NotContains(t, token, "+", "Token should not contain +")
		assert.NotContains(t, token, "/", "Token should not contain /")
		assert.NotContains(t, token, "=", "Token should not contain =")

		// Should only contain URL-safe characters
		for _, ch := range token {
			assert.True(t,
				(ch >= 'A' && ch <= 'Z') ||
					(ch >= 'a' && ch <= 'z') ||
					(ch >= '0' && ch <= '9') ||
					ch == '-' || ch == '_',
				"Character %c should be URL-safe", ch)
		}
	}
}

func TestJWTClaims_Structure(t *testing.T) {
	// Test that JWTClaims properly embeds RegisteredClaims
	claims := JWTClaims{
		UserID:      "user-123",
		TenantID:    "tenant-456",
		Email:       "user@example.com",
		IsAdmin:     true,
		IsModerator: false,
		Type:        "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "goconnect",
			Subject:   "user-123",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Verify fields are accessible
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "tenant-456", claims.TenantID)
	assert.Equal(t, "user@example.com", claims.Email)
	assert.True(t, claims.IsAdmin)
	assert.False(t, claims.IsModerator)
	assert.Equal(t, "access", claims.Type)
	assert.Equal(t, "goconnect", claims.Issuer)
	assert.Equal(t, "user-123", claims.Subject)
}

func TestTokenPair_ExpirationDifference(t *testing.T) {
	accessToken, refreshToken, _, err := GenerateTokenPair(
		"user-123",
		"tenant-456",
		"user@example.com",
		false,
		false,
	)
	require.NoError(t, err)

	// Parse tokens to check expiration
	secret := getJWTSecret()

	accessTokenObj, err := jwt.ParseWithClaims(accessToken, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	require.NoError(t, err)
	accessClaims := accessTokenObj.Claims.(*JWTClaims)

	refreshTokenObj, err := jwt.ParseWithClaims(refreshToken, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	require.NoError(t, err)
	refreshClaims := refreshTokenObj.Claims.(*JWTClaims)

	// Refresh token should expire much later than access token
	accessExp := accessClaims.ExpiresAt.Time
	refreshExp := refreshClaims.ExpiresAt.Time

	assert.True(t, refreshExp.After(accessExp), "Refresh token should expire after access token")

	// Check approximate durations
	issuedAt := accessClaims.IssuedAt.Time
	accessDuration := accessExp.Sub(issuedAt)
	refreshDuration := refreshExp.Sub(issuedAt)

	assert.InDelta(t, defaultAccessTTL.Seconds(), accessDuration.Seconds(), 5)
	assert.InDelta(t, defaultRefreshTTL.Seconds(), refreshDuration.Seconds(), 5)
}

func TestTokenPair_IssuerAndSubject(t *testing.T) {
	userID := "user-123"
	accessToken, refreshToken, _, err := GenerateTokenPair(
		userID,
		"tenant-456",
		"user@example.com",
		false,
		false,
	)
	require.NoError(t, err)

	secret := getJWTSecret()

	// Check access token
	accessTokenObj, _ := jwt.ParseWithClaims(accessToken, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	accessClaims := accessTokenObj.Claims.(*JWTClaims)
	assert.Equal(t, "goconnect", accessClaims.Issuer)
	assert.Equal(t, userID, accessClaims.Subject)

	// Check refresh token
	refreshTokenObj, _ := jwt.ParseWithClaims(refreshToken, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	refreshClaims := refreshTokenObj.Claims.(*JWTClaims)
	assert.Equal(t, "goconnect", refreshClaims.Issuer)
	assert.Equal(t, userID, refreshClaims.Subject)
}

func TestTokenPair_NotBeforeCheck(t *testing.T) {
	accessToken, _, _, err := GenerateTokenPair(
		"user-123",
		"tenant-456",
		"user@example.com",
		false,
		false,
	)
	require.NoError(t, err)

	secret := getJWTSecret()
	tokenObj, _ := jwt.ParseWithClaims(accessToken, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	claims := tokenObj.Claims.(*JWTClaims)

	// NotBefore should be set to current time or before
	assert.True(t, claims.NotBefore.Time.Before(time.Now().Add(time.Second)) ||
		claims.NotBefore.Time.Equal(time.Now()))
}

func TestConstants(t *testing.T) {
	// Verify constant values are sensible
	assert.Equal(t, 15*time.Minute, defaultAccessTTL, "Access token should be 15 minutes")
	assert.Equal(t, 7*24*time.Hour, defaultRefreshTTL, "Refresh token should be 7 days")
}
