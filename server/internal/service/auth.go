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
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/argon2"
)

// AuthService handles authentication operations
type AuthService struct {
	userRepo    repository.UserRepository
	tenantRepo  repository.TenantRepository
	redisClient *redis.Client
	jwtSecret   []byte // Secret key for JWT signing
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo repository.UserRepository, tenantRepo repository.TenantRepository, redisClient *redis.Client) *AuthService {
	// Get JWT secret from environment or use default (NOT for production!)
	jwtSecret := []byte(getEnvOrDefault("JWT_SECRET", "dev-secret-change-in-production"))

	return &AuthService{
		userRepo:    userRepo,
		tenantRepo:  tenantRepo,
		redisClient: redisClient,
		jwtSecret:   jwtSecret,
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

// GenerateRecoveryCodes generates 8 one-time recovery codes for the user
// Returns the plaintext codes (to show once to the user) and stores hashed versions
func (s *AuthService) GenerateRecoveryCodes(ctx context.Context, userID, code string) ([]string, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Verify 2FA is enabled
	if !user.TwoFAEnabled {
		return nil, domain.NewError(domain.ErrInvalidRequest, "2FA must be enabled first", nil)
	}

	// Verify current TOTP code
	if !totp.Validate(code, user.TwoFAKey) {
		return nil, domain.NewError(domain.ErrInvalidCredentials, "Invalid 2FA code", nil)
	}

	// Generate 8 recovery codes (format: XXXXX-XXXXX, 10 chars total)
	const codeCount = 8
	plaintextCodes := make([]string, codeCount)
	hashedCodes := make([]string, codeCount)

	for i := 0; i < codeCount; i++ {
		code := generateRecoveryCode()
		plaintextCodes[i] = code
		hashed, err := s.HashPassword(code)
		if err != nil {
			return nil, fmt.Errorf("failed to hash recovery code: %w", err)
		}
		hashedCodes[i] = hashed
	}

	user.RecoveryCodes = hashedCodes
	user.UpdatedAt = time.Now().UTC()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return plaintextCodes, nil
}

// generateRecoveryCode generates a random recovery code in format XXXXX-XXXXX
func generateRecoveryCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // No O, 0, I, 1 for clarity
	code := make([]byte, 10)
	b := make([]byte, 10)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	for i := range code {
		code[i] = charset[b[i]%byte(len(charset))]
	}
	// Insert dash: XXXXX-XXXXX
	return string(code[:5]) + "-" + string(code[5:])
}

// UseRecoveryCode validates and uses a recovery code for login (one-time use)
func (s *AuthService) UseRecoveryCode(ctx context.Context, req *domain.UseRecoveryCodeRequest) (*domain.AuthResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, domain.NewError(domain.ErrInvalidCredentials, "Invalid email or password", nil)
	}

	// Verify password
	valid, err := s.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !valid {
		return nil, domain.NewError(domain.ErrInvalidCredentials, "Invalid email or password", nil)
	}

	// Verify 2FA is enabled and recovery codes exist
	if !user.TwoFAEnabled {
		return nil, domain.NewError(domain.ErrInvalidRequest, "2FA is not enabled", nil)
	}
	if len(user.RecoveryCodes) == 0 {
		return nil, domain.NewError(domain.ErrInvalidRequest, "No recovery codes available", nil)
	}

	// Normalize recovery code (remove dashes, uppercase)
	normalizedCode := normalizeRecoveryCode(req.RecoveryCode)

	// Check against all hashed recovery codes
	matchedIndex := -1
	for i, hashedCode := range user.RecoveryCodes {
		valid, _ := s.VerifyPassword(normalizedCode, hashedCode)
		if valid {
			matchedIndex = i
			break
		}
	}

	if matchedIndex == -1 {
		return nil, domain.NewError(domain.ErrInvalidCredentials, "Invalid recovery code", nil)
	}

	// Remove used code (one-time use)
	user.RecoveryCodes = append(user.RecoveryCodes[:matchedIndex], user.RecoveryCodes[matchedIndex+1:]...)
	user.UpdatedAt = time.Now().UTC()

	if err := s.userRepo.Update(ctx, user); err != nil {
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
		ExpiresIn:    900,
		TokenType:    "Bearer",
		User:         user,
	}, nil
}

// normalizeRecoveryCode normalizes a recovery code for comparison
func normalizeRecoveryCode(code string) string {
	// Remove dashes and spaces, uppercase
	var result []byte
	for _, c := range []byte(code) {
		if c >= 'a' && c <= 'z' {
			result = append(result, c-32) // to uppercase
		} else if (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			result = append(result, c)
		}
		// Skip dashes, spaces, other chars
	}
	return string(result)
}

// GetRecoveryCodeCount returns the number of remaining recovery codes for a user
func (s *AuthService) GetRecoveryCodeCount(ctx context.Context, userID string) (int, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return len(user.RecoveryCodes), nil
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
	// Check blacklist
	if err := s.checkBlacklist(ctx, req.RefreshToken); err != nil {
		return nil, err
	}

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
	return s.addToBlacklist(ctx, refreshToken)
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

// LoginOrRegisterOIDC handles OIDC login/registration
func (s *AuthService) LoginOrRegisterOIDC(ctx context.Context, email, externalID, provider string) (*domain.AuthResponse, error) {
	// 1. Check if user exists
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Check if it's a "not found" error
		if e, ok := err.(*domain.Error); ok && e.Code == domain.ErrUserNotFound {
			// User not found, proceed to register
			user = nil
		} else {
			// Real error
			return nil, err
		}
	}

	if user == nil {
		// 2. Register new user
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
		user = &domain.User{
			ID:           userID,
			TenantID:     tenantID,
			Email:        email,
			AuthProvider: provider,
			ExternalID:   externalID,
			Locale:       "en", // Default
			TwoFAEnabled: false,
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, err
		}
	} else {
		// 3. Update existing user if needed (link account)
		if user.AuthProvider == "" {
			user.AuthProvider = provider
			user.ExternalID = externalID
			user.UpdatedAt = time.Now().UTC()
			if err := s.userRepo.Update(ctx, user); err != nil {
				return nil, err
			}
		}
	}

	// 4. Generate Tokens
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
		ExpiresIn:    900,
		TokenType:    "Bearer",
		User:         user,
	}, nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// addToBlacklist adds a token to the blacklist
func (s *AuthService) addToBlacklist(ctx context.Context, tokenString string) error {
	if s.redisClient == nil {
		return nil
	}

	// Parse token to get expiration
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("invalid claims")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return fmt.Errorf("invalid expiration")
	}

	expiration := time.Until(time.Unix(int64(exp), 0))
	if expiration < 0 {
		return nil // Already expired
	}

	return s.redisClient.Set(ctx, "blacklist:"+tokenString, "revoked", expiration).Err()
}

// checkBlacklist checks if a token is blacklisted
func (s *AuthService) checkBlacklist(ctx context.Context, tokenString string) error {
	if s.redisClient == nil {
		return nil
	}

	exists, err := s.redisClient.Exists(ctx, "blacklist:"+tokenString).Result()
	if err != nil {
		return err
	}

	if exists > 0 {
		return domain.NewError(domain.ErrInvalidToken, "Token is revoked", nil)
	}

	return nil
}
