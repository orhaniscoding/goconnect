package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"golang.org/x/crypto/argon2"
)

// AuthService handles authentication operations
type AuthService struct {
	userRepo   repository.UserRepository
	tenantRepo repository.TenantRepository
	// In a real implementation, we'd use a proper JWT library and secret management
	// For now, we'll use a simple token map (sessions)
	sessions map[string]*domain.TokenClaims // refreshToken -> claims
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo repository.UserRepository, tenantRepo repository.TenantRepository) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
		sessions:   make(map[string]*domain.TokenClaims),
	}
}

// HashPassword hashes a password using Argon2id
func (s *AuthService) HashPassword(password string) (string, error) {
	// Argon2id parameters (OWASP recommended)
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	// Encode as: $argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=19$m=65536,t=1,p=4$%s$%s", b64Salt, b64Hash), nil
}

// VerifyPassword verifies a password against a hash
func (s *AuthService) VerifyPassword(password, encodedHash string) (bool, error) {
	// Parse the hash
	var salt, hash []byte
	var version int
	var memory, iterations, parallelism uint32

	_, err := fmt.Sscanf(encodedHash, "$argon2id$v=%d$m=%d,t=%d,p=%d$", &version, &memory, &iterations, &parallelism)
	if err != nil {
		return false, err
	}

	// Extract salt and hash
	parts := []byte(encodedHash)
	lastDollar := 0
	dollarCount := 0
	for i, b := range parts {
		if b == '$' {
			dollarCount++
			if dollarCount == 4 {
				lastDollar = i
				break
			}
		}
	}

	saltAndHash := string(parts[lastDollar+1:])
	parts2 := []byte(saltAndHash)
	dollarPos := 0
	for i, b := range parts2 {
		if b == '$' {
			dollarPos = i
			break
		}
	}

	b64Salt := string(parts2[:dollarPos])
	b64Hash := string(parts2[dollarPos+1:])

	salt, err = base64.RawStdEncoding.DecodeString(b64Salt)
	if err != nil {
		return false, err
	}

	hash, err = base64.RawStdEncoding.DecodeString(b64Hash)
	if err != nil {
		return false, err
	}

	// Validate parameters to prevent overflow
	if parallelism > 255 {
		return false, fmt.Errorf("parallelism too large")
	}
	if len(hash) > 0xFFFFFFFF {
		return false, fmt.Errorf("hash length too large")
	}

	// Hash the input password with the same parameters
	// #nosec G115 - parallelism and hash length are validated above
	computedHash := argon2.IDKey([]byte(password), salt, iterations, memory, uint8(parallelism), uint32(len(hash)))

	// Constant-time comparison
	return subtle.ConstantTimeCompare(hash, computedHash) == 1, nil
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.User, error) {
	// Validate password strength (basic check)
	if len(req.Password) < 8 {
		return nil, domain.NewError(domain.ErrWeakPassword, "Password must be at least 8 characters", nil)
	}

	// Hash password
	hashedPassword, err := s.HashPassword(req.Password)
	if err != nil {
		return nil, domain.NewError(domain.ErrInternalServer, "Failed to hash password", nil)
	}

	// Set default locale
	locale := req.Locale
	if locale == "" {
		locale = "en"
	}

	// Create default tenant for the user
	tenantID := uuid.New().String()
	tenant := &domain.Tenant{
		ID:        tenantID,
		Name:      "Personal", // Default tenant name
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, err
	}

	// Create user
	userID := uuid.New().String()
	user := &domain.User{
		ID:           userID,
		TenantID:     tenantID,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Locale:       locale,
		TwoFAEnabled: false,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Return generic error to prevent user enumeration
		return nil, domain.NewError(domain.ErrInvalidCredentials, "Invalid email or password", nil)
	}

	// Verify password
	valid, err := s.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !valid {
		return nil, domain.NewError(domain.ErrInvalidCredentials, "Invalid email or password", nil)
	}

	// Generate tokens
	accessToken := uuid.New().String()  // In production, use proper JWT
	refreshToken := uuid.New().String() // In production, use proper JWT

	// Store session (simplified; in production use Redis with TTL)
	claims := &domain.TokenClaims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Email:    user.Email,
		IsAdmin:  true, // For now, first user is admin (simplified)
		Exp:      time.Now().Add(15 * time.Minute).Unix(),
		Iat:      time.Now().Unix(),
	}
	s.sessions[refreshToken] = claims

	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900, // 15 minutes
		TokenType:    "Bearer",
		User:         user,
	}, nil
}

// ValidateToken validates an access token and returns claims
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*domain.TokenClaims, error) {
	// In production, decode and validate JWT
	// For now, we'll accept any non-empty token and return mock claims
	// This is a placeholder implementation
	if token == "" {
		return nil, domain.NewError(domain.ErrInvalidToken, "Invalid token", nil)
	}

	// For testing purposes, return mock claims
	// In production, this would decode the JWT and verify signature
	return &domain.TokenClaims{
		UserID:   "test-user",
		TenantID: "test-tenant",
		Email:    "test@example.com",
		IsAdmin:  true,
		Exp:      time.Now().Add(15 * time.Minute).Unix(),
		Iat:      time.Now().Unix(),
	}, nil
}

// Refresh generates new tokens using a refresh token
func (s *AuthService) Refresh(ctx context.Context, req *domain.RefreshRequest) (*domain.AuthResponse, error) {
	// Get claims from session
	claims, exists := s.sessions[req.RefreshToken]
	if !exists {
		return nil, domain.NewError(domain.ErrInvalidToken, "Invalid refresh token", nil)
	}

	// Check if refresh token is expired (in production, check JWT exp)
	if time.Now().Unix() > claims.Exp+86400 { // 24 hours after access token exp
		delete(s.sessions, req.RefreshToken)
		return nil, domain.NewError(domain.ErrTokenExpired, "Refresh token expired", nil)
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	// Generate new tokens
	newAccessToken := uuid.New().String()
	newRefreshToken := uuid.New().String()

	// Update session
	delete(s.sessions, req.RefreshToken) // Remove old refresh token
	newClaims := &domain.TokenClaims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Email:    user.Email,
		IsAdmin:  claims.IsAdmin,
		Exp:      time.Now().Add(15 * time.Minute).Unix(),
		Iat:      time.Now().Unix(),
	}
	s.sessions[newRefreshToken] = newClaims

	return &domain.AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    900,
		TokenType:    "Bearer",
		User:         user,
	}, nil
}

// Logout invalidates a refresh token
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	delete(s.sessions, refreshToken)
	return nil
}
