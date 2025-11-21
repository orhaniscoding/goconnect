package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/argon2"
)

// AuthService handles authentication operations
type AuthService struct {
	userRepo   repository.UserRepository
	tenantRepo repository.TenantRepository
	jwtSecret  []byte // Secret key for JWT signing
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo repository.UserRepository, tenantRepo repository.TenantRepository) *AuthService {
	// Get JWT secret from environment or use default (NOT for production!)
	jwtSecret := []byte(getEnvOrDefault("JWT_SECRET", "dev-secret-change-in-production"))

	return &AuthService{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
		jwtSecret:  jwtSecret,
	}
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
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

// generateJWT generates a JWT token with the given claims
func (s *AuthService) generateJWT(userID, tenantID, email string, isAdmin, isModerator bool, tokenType string, expiryDuration time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id":      userID,
		"tenant_id":    tenantID,
		"email":        email,
		"is_admin":     isAdmin,
		"is_moderator": isModerator,
		"type":         tokenType, // "access" or "refresh"
		"exp":          now.Add(expiryDuration).Unix(),
		"iat":          now.Unix(),
		"nbf":          now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// Register creates a new user account and returns auth tokens (auto-login)
func (s *AuthService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error) {
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

	// Generate JWT tokens
	accessToken, err := s.generateJWT(user.ID, user.TenantID, user.Email, user.IsAdmin, user.IsModerator, "access", 15*time.Minute)
	if err != nil {
		return nil, domain.NewError(domain.ErrInternalServer, "Failed to generate access token", nil)
	}

	refreshToken, err := s.generateJWT(user.ID, user.TenantID, user.Email, user.IsAdmin, user.IsModerator, "refresh", 7*24*time.Hour)
	if err != nil {
		return nil, domain.NewError(domain.ErrInternalServer, "Failed to generate refresh token", nil)
	}

	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900, // 15 minutes
		TokenType:    "Bearer",
		User:         user,
	}, nil
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

	// Check 2FA
	if user.TwoFAEnabled {
		if req.Code == "" {
			return nil, domain.NewError("ERR_2FA_REQUIRED", "Two-factor authentication required", nil)
		}
		if !totp.Validate(req.Code, user.TwoFAKey) {
			return nil, domain.NewError(domain.ErrInvalidCredentials, "Invalid 2FA code", nil)
		}
	}

	// Generate JWT tokens
	accessToken, err := s.generateJWT(user.ID, user.TenantID, user.Email, user.IsAdmin, user.IsModerator, "access", 15*time.Minute)
	if err != nil {
		return nil, domain.NewError(domain.ErrInternalServer, "Failed to generate access token", nil)
	}

	refreshToken, err := s.generateJWT(user.ID, user.TenantID, user.Email, user.IsAdmin, user.IsModerator, "refresh", 7*24*time.Hour)
	if err != nil {
		return nil, domain.NewError(domain.ErrInternalServer, "Failed to generate refresh token", nil)
	}

	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900, // 15 minutes
		TokenType:    "Bearer",
		User:         user,
	}, nil
}

// Generate2FASecret generates a new TOTP secret for the user
func (s *AuthService) Generate2FASecret(ctx context.Context, userID string) (string, string, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", "", err
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "GoConnect",
		AccountName: user.Email,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	return key.Secret(), key.URL(), nil
}

// Enable2FA verifies the code and enables 2FA for the user
func (s *AuthService) Enable2FA(ctx context.Context, userID, secret, code string) error {
	if !totp.Validate(code, secret) {
		return domain.NewError(domain.ErrInvalidCredentials, "Invalid 2FA code", nil)
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.TwoFAKey = secret
	user.TwoFAEnabled = true
	user.UpdatedAt = time.Now().UTC()

	return s.userRepo.Update(ctx, user)
}

// Disable2FA verifies the code and disables 2FA for the user
func (s *AuthService) Disable2FA(ctx context.Context, userID, code string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if !user.TwoFAEnabled {
		return nil
	}

	if !totp.Validate(code, user.TwoFAKey) {
		return domain.NewError(domain.ErrInvalidCredentials, "Invalid 2FA code", nil)
	}

	user.TwoFAEnabled = false
	user.TwoFAKey = ""
	user.UpdatedAt = time.Now().UTC()

	return s.userRepo.Update(ctx, user)
}

// ValidateToken validates an access token and returns claims
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*domain.TokenClaims, error) {
	// Parse and validate JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, domain.NewError(domain.ErrInvalidToken, "Invalid token", nil)
	}

	if !token.Valid {
		return nil, domain.NewError(domain.ErrInvalidToken, "Invalid token", nil)
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, domain.NewError(domain.ErrInvalidToken, "Invalid token claims", nil)
	}

	// Verify token type (must be "access" token for API requests)
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "access" {
		return nil, domain.NewError(domain.ErrInvalidToken, "Invalid token type", nil)
	}

	// Extract required fields
	userID, _ := claims["user_id"].(string)
	tenantID, _ := claims["tenant_id"].(string)
	email, _ := claims["email"].(string)
	isAdmin, _ := claims["is_admin"].(bool)
	isModerator, _ := claims["is_moderator"].(bool)
	exp, _ := claims["exp"].(float64)
	iat, _ := claims["iat"].(float64)

	return &domain.TokenClaims{
		UserID:      userID,
		TenantID:    tenantID,
		Email:       email,
		IsAdmin:     isAdmin,
		IsModerator: isModerator,
		Type:        tokenType,
		Exp:         int64(exp),
		Iat:         int64(iat),
	}, nil
}

// Refresh generates new tokens using a refresh token
func (s *AuthService) Refresh(ctx context.Context, req *domain.RefreshRequest) (*domain.AuthResponse, error) {
	// Parse and validate refresh token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, domain.NewError(domain.ErrInvalidToken, "Invalid refresh token", nil)
	}

	if !token.Valid {
		return nil, domain.NewError(domain.ErrInvalidToken, "Invalid refresh token", nil)
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, domain.NewError(domain.ErrInvalidToken, "Invalid token claims", nil)
	}

	// Verify token type (must be "refresh" token)
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "refresh" {
		return nil, domain.NewError(domain.ErrInvalidToken, "Invalid token type", nil)
	}

	// Extract user info
	userID, _ := claims["user_id"].(string)

	// Get user to verify still exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Generate new tokens
	newAccessToken, err := s.generateJWT(user.ID, user.TenantID, user.Email, user.IsAdmin, user.IsModerator, "access", 15*time.Minute)
	if err != nil {
		return nil, domain.NewError(domain.ErrInternalServer, "Failed to generate access token", nil)
	}

	newRefreshToken, err := s.generateJWT(user.ID, user.TenantID, user.Email, user.IsAdmin, user.IsModerator, "refresh", 7*24*time.Hour)
	if err != nil {
		return nil, domain.NewError(domain.ErrInternalServer, "Failed to generate refresh token", nil)
	}

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
	// With JWT, we can't truly invalidate tokens without a blacklist (Redis)
	// For now, we just return success. In production, add token to Redis blacklist
	// TODO: Implement token blacklist with Redis
	return nil
}

// ChangePassword changes the user's password
func (s *AuthService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify old password
	valid, err := s.VerifyPassword(oldPassword, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("failed to verify password: %w", err)
	}
	if !valid {
		return domain.NewError(domain.ErrInvalidCredentials, "Invalid current password", nil)
	}

	// Hash new password
	newHash, err := s.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = newHash
	user.UpdatedAt = time.Now().UTC()

	return s.userRepo.Update(ctx, user)
}
