package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

const (
	defaultAccessTTL  = 15 * time.Minute
	defaultRefreshTTL = 7 * 24 * time.Hour
)

// JWTClaims extends jwt.RegisteredClaims with custom fields
type JWTClaims struct {
	UserID      string `json:"user_id"`
	TenantID    string `json:"tenant_id"`
	Email       string `json:"email"`
	IsAdmin     bool   `json:"is_admin"`
	IsModerator bool   `json:"is_moderator"`
	Type        string `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// getJWTSecret returns the JWT signing secret from environment or generates a development key
// Deprecated: Use AuthService.jwtSecret instead
func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Development fallback - in production this MUST be set via env
		secret = "dev-secret-DO-NOT-USE-IN-PRODUCTION"
	}
	return []byte(secret)
}

// GenerateTokenPair generates access and refresh tokens
// Deprecated: Use AuthService.GenerateTokenPair instead
func GenerateTokenPair(userID, tenantID, email string, isAdmin, isModerator bool) (accessToken, refreshToken string, expiresIn int, err error) {
	secret := getJWTSecret()
	now := time.Now()

	// Generate unique token IDs
	accessTokenID, _ := GenerateSecureToken(16)
	refreshTokenID, _ := GenerateSecureToken(16)

	// Access token (short-lived)
	accessClaims := JWTClaims{
		UserID:      userID,
		TenantID:    tenantID,
		Email:       email,
		IsAdmin:     isAdmin,
		IsModerator: isModerator,
		Type:        "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        accessTokenID,
			ExpiresAt: jwt.NewNumericDate(now.Add(defaultAccessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "goconnect",
			Subject:   userID,
		},
	}

	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessTokenObj.SignedString(secret)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh token (long-lived)
	refreshClaims := JWTClaims{
		UserID:      userID,
		TenantID:    tenantID,
		Email:       email,
		IsAdmin:     isAdmin,
		IsModerator: isModerator,
		Type:        "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshTokenID,
			ExpiresAt: jwt.NewNumericDate(now.Add(defaultRefreshTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "goconnect",
			Subject:   userID,
		},
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString(secret)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return accessToken, refreshToken, int(defaultAccessTTL.Seconds()), nil
}

// ValidateToken validates a JWT token and returns claims
func ValidateToken(tokenString string) (*domain.TokenClaims, error) {
	secret := getJWTSecret()

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, domain.NewError(domain.ErrInvalidToken,
			"Invalid token",
			map[string]string{"error": err.Error()})
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, domain.NewError(domain.ErrInvalidToken,
			"Invalid token claims",
			nil)
	}

	// Check expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, domain.NewError(domain.ErrTokenExpired,
			"Token has expired",
			nil)
	}

	return &domain.TokenClaims{
		UserID:      claims.UserID,
		TenantID:    claims.TenantID,
		Email:       claims.Email,
		IsAdmin:     claims.IsAdmin,
		IsModerator: claims.IsModerator,
		Type:        claims.Type,
	}, nil
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
