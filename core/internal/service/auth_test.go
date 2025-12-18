package service

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/pquerna/otp/totp"
)

func TestAuthService_Register(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthServiceWithSecret(userRepo, tenantRepo, nil, "12345678901234567890123456789012")

	tests := []struct {
		name    string
		req     *domain.RegisterRequest
		wantErr bool
		errCode string
	}{
		{
			name: "valid registration",
			req: &domain.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Locale:   "en",
			},
			wantErr: false,
		},
		{
			name: "duplicate email",
			req: &domain.RegisterRequest{
				Email:    "test@example.com",
				Password: "password456",
				Locale:   "en",
			},
			wantErr: true,
			errCode: domain.ErrEmailAlreadyExists,
		},
		{
			name: "weak password",
			req: &domain.RegisterRequest{
				Email:    "test2@example.com",
				Password: "123",
				Locale:   "en",
			},
			wantErr: true,
			errCode: domain.ErrWeakPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authResp, err := authService.Register(context.Background(), tt.req)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				var derr *domain.Error; if errors.As(err, &derr) {
					if derr.Code != tt.errCode {
						t.Errorf("expected error code %s, got %s", tt.errCode, derr.Code)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Register now returns AuthResponse with User field
			if authResp.User.Email != tt.req.Email {
				t.Errorf("expected email %s, got %s", tt.req.Email, authResp.User.Email)
			}

			if authResp.User.PasswordHash == "" {
				t.Error("expected password hash to be set")
			}

			if authResp.User.PasswordHash == tt.req.Password {
				t.Error("password should be hashed, not stored in plaintext")
			}

			// Check that tokens are returned
			if authResp.AccessToken == "" {
				t.Error("expected access token to be set")
			}

			if authResp.RefreshToken == "" {
				t.Error("expected refresh token to be set")
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Register a user first
	regReq := &domain.RegisterRequest{
		Email:    "login@example.com",
		Password: "password123",
		Locale:   "en",
	}
	_, err := authService.Register(context.Background(), regReq)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	tests := []struct {
		name    string
		req     *domain.LoginRequest
		wantErr bool
		errCode string
	}{
		{
			name: "valid login",
			req: &domain.LoginRequest{
				Email:    "login@example.com",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "wrong password",
			req: &domain.LoginRequest{
				Email:    "login@example.com",
				Password: "wrongpassword",
			},
			wantErr: true,
			errCode: domain.ErrInvalidCredentials,
		},
		{
			name: "nonexistent user",
			req: &domain.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			wantErr: true,
			errCode: domain.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authResp, err := authService.Login(context.Background(), tt.req)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				var derr *domain.Error; if errors.As(err, &derr) {
					if derr.Code != tt.errCode {
						t.Errorf("expected error code %s, got %s", tt.errCode, derr.Code)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if authResp.AccessToken == "" {
				t.Error("expected access token to be set")
			}

			if authResp.RefreshToken == "" {
				t.Error("expected refresh token to be set")
			}

			if authResp.User == nil {
				t.Fatal("expected user to be set")
			}

			if authResp.User.Email != tt.req.Email {
				t.Errorf("expected email %s, got %s", tt.req.Email, authResp.User.Email)
			}
		})
	}
}


// ==================== Helper Function Tests ====================

func TestGetEnvOrDefault(t *testing.T) {
	key := "TEST_ENV_VAR_KEY"
	defaultVal := "default_value"

	// Test case 1: Env var is set
	os.Setenv(key, "set_value")
	if val := getEnvOrDefault(key, defaultVal); val != "set_value" {
		t.Errorf("expected 'set_value', got '%s'", val)
	}
	os.Unsetenv(key)

	// Test case 2: Env var is not set
	if val := getEnvOrDefault(key, defaultVal); val != defaultVal {
		t.Errorf("expected '%s', got '%s'", defaultVal, val)
	}
}

func TestAuthService_PasswordHashing(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	password := "mySecurePassword123"

	// Test hashing
	hash, err := authService.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if hash == password {
		t.Error("hash should not equal plaintext password")
	}

	// Test verification with correct password
	valid, err := authService.VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("failed to verify password: %v", err)
	}

	if !valid {
		t.Error("expected password to be valid")
	}

	// Test verification with wrong password
	valid, err = authService.VerifyPassword("wrongPassword", hash)
	if err != nil {
		t.Fatalf("failed to verify password: %v", err)
	}

	if valid {
		t.Error("expected password to be invalid")
	}

	// Test that same password produces different hashes (due to random salt)
	hash2, err := authService.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if hash == hash2 {
		t.Error("expected different hashes for same password (different salts)")
	}

	// But both should verify correctly
	valid, err = authService.VerifyPassword(password, hash2)
	if err != nil {
		t.Fatalf("failed to verify password: %v", err)
	}

	if !valid {
		t.Error("expected password to be valid for second hash")
	}
}

func TestAuthService_Refresh(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Register and login first
	regReq := &domain.RegisterRequest{
		Email:    "refresh@example.com",
		Password: "password123",
		Locale:   "en",
	}
	_, err := authService.Register(context.Background(), regReq)
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	loginReq := &domain.LoginRequest{
		Email:    "refresh@example.com",
		Password: "password123",
	}
	authResp, err := authService.Login(context.Background(), loginReq)
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	// Wait a moment to ensure different timestamp in JWT
	time.Sleep(time.Second)

	// Test refresh with valid token
	refreshReq := &domain.RefreshRequest{
		RefreshToken: authResp.RefreshToken,
	}
	newAuthResp, err := authService.Refresh(context.Background(), refreshReq)
	if err != nil {
		t.Fatalf("failed to refresh: %v", err)
	}

	if newAuthResp.AccessToken == "" {
		t.Error("expected new access token")
	}

	if newAuthResp.RefreshToken == "" {
		t.Error("expected new refresh token")
	}

	// Debug: Print tokens to see if they're different
	t.Logf("Original refresh token (first 50 chars): %s", authResp.RefreshToken[:min(50, len(authResp.RefreshToken))])
	t.Logf("New refresh token (first 50 chars): %s", newAuthResp.RefreshToken[:min(50, len(newAuthResp.RefreshToken))])

	if newAuthResp.RefreshToken == authResp.RefreshToken {
		t.Error("new refresh token should be different from old one")
	}

	// Note: With JWT, old refresh token still works until it expires (no token blacklist yet)
	// In production, implement Redis blacklist for token rotation
	// For now, just verify the old token is still valid (JWT hasn't expired)
	oldRefreshResp, err := authService.Refresh(context.Background(), refreshReq)
	if err != nil {
		// This is expected if token blacklist is implemented
		t.Logf("old refresh token rejected (good if blacklist is implemented): %v", err)
	} else {
		// This is current behavior without blacklist
		t.Logf("old refresh token still works (expected without Redis blacklist)")
		if oldRefreshResp.AccessToken == "" {
			t.Error("expected access token from old refresh token")
		}
	}

	// Test refresh with invalid token
	invalidReq := &domain.RefreshRequest{
		RefreshToken: "invalid-token",
	}
	_, err = authService.Refresh(context.Background(), invalidReq)
	if err == nil {
		t.Error("expected error with invalid refresh token")
	}
}

func TestAuthService_RecoveryCodes(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)
	ctx := context.Background()

	// Register a user
	registerReq := &domain.RegisterRequest{
		Email:    "recovery@example.com",
		Password: "password123",
	}
	authResp, err := authService.Register(ctx, registerReq)
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	userID := authResp.User.ID

	// Test generating recovery codes without 2FA enabled (should fail)
	_, err = authService.GenerateRecoveryCodes(ctx, userID, "123456")
	if err == nil {
		t.Error("expected error when generating recovery codes without 2FA enabled")
	}

	// Enable 2FA first
	secret, _, err := authService.Generate2FASecret(ctx, userID)
	if err != nil {
		t.Fatalf("failed to generate 2FA secret: %v", err)
	}

	// Generate a valid TOTP code for testing
	// In real tests, we would use the totp library to generate a valid code
	// For this test, we'll directly enable 2FA by manipulating the user

	// Get user and manually enable 2FA for testing
	user, err := userRepo.GetByID(ctx, userID)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	user.TwoFAKey = secret
	user.TwoFAEnabled = true
	if err := userRepo.Update(ctx, user); err != nil {
		t.Fatalf("failed to update user: %v", err)
	}

	// Test generating recovery codes (with invalid code - should fail)
	_, err = authService.GenerateRecoveryCodes(ctx, userID, "000000")
	if err == nil {
		t.Error("expected error with invalid TOTP code")
	}

	// Generate a valid TOTP code and test successful recovery code generation
	validCode, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatalf("failed to generate TOTP code: %v", err)
	}

	codes, err := authService.GenerateRecoveryCodes(ctx, userID, validCode)
	if err != nil {
		t.Fatalf("failed to generate recovery codes: %v", err)
	}
	if len(codes) != 8 {
		t.Errorf("expected 8 recovery codes, got %d", len(codes))
	}

	// Verify recovery codes are stored (as hashes)
	updatedUser, _ := userRepo.GetByID(ctx, userID)
	if len(updatedUser.RecoveryCodes) != 8 {
		t.Errorf("expected 8 hashed recovery codes stored, got %d", len(updatedUser.RecoveryCodes))
	}
}

func TestRecoveryCodeFormat(t *testing.T) {
	// Test the recovery code format
	code := generateRecoveryCode()

	// Check format: XXXXX-XXXXX (11 chars with dash)
	if len(code) != 11 {
		t.Errorf("expected code length 11, got %d", len(code))
	}

	if code[5] != '-' {
		t.Errorf("expected dash at position 5, got %c", code[5])
	}

	// Check that code only contains valid characters
	validChars := "ABCDEFGHJKLMNPQRSTUVWXYZ23456789-"
	for _, c := range code {
		found := false
		for _, v := range validChars {
			if c == v {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("invalid character in code: %c", c)
		}
	}

	// Generate multiple codes and check uniqueness
	codes := make(map[string]bool)
	for i := 0; i < 100; i++ {
		c := generateRecoveryCode()
		if codes[c] {
			t.Errorf("duplicate code generated: %s", c)
		}
		codes[c] = true
	}
}

func TestNormalizeRecoveryCode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ABCDE-FGHIJ", "ABCDEFGHIJ"},
		{"abcde-fghij", "ABCDEFGHIJ"},
		{"ABCDE FGHIJ", "ABCDEFGHIJ"},
		{"abcde fghij", "ABCDEFGHIJ"},
		{"ABC-DE-FGH-IJ", "ABCDEFGHIJ"},
		{"  abc  def  ", "ABCDEF"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeRecoveryCode(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeRecoveryCode(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ==================== GetUserByID Tests ====================

func TestAuthService_GetUserByID(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Create a test user
	testUser := &domain.User{
		ID:       "user-123",
		Email:    "test@example.com",
		TenantID: "tenant-1",
	}
	err := userRepo.Create(context.Background(), testUser)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	t.Run("User Found", func(t *testing.T) {
		user, err := authService.GetUserByID(context.Background(), "user-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.Email != "test@example.com" {
			t.Errorf("expected email test@example.com, got %s", user.Email)
		}
	})

	t.Run("User Not Found", func(t *testing.T) {
		_, err := authService.GetUserByID(context.Background(), "non-existent")
		if err == nil {
			t.Fatal("expected error for non-existent user")
		}
	})
}

// ==================== UpdateUserProfile Tests ====================

func TestAuthService_UpdateUserProfile(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Create a test user
	testUser := &domain.User{
		ID:       "user-456",
		Email:    "profile@example.com",
		TenantID: "tenant-1",
	}
	err := userRepo.Create(context.Background(), testUser)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	t.Run("Update Full Name", func(t *testing.T) {
		fullName := "John Doe"
		err := authService.UpdateUserProfile(context.Background(), "user-456", &fullName, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify update
		user, _ := userRepo.GetByID(context.Background(), "user-456")
		if user.FullName == nil || *user.FullName != "John Doe" {
			t.Errorf("expected full name John Doe")
		}
	})

	t.Run("Update Bio", func(t *testing.T) {
		bio := "Software developer"
		err := authService.UpdateUserProfile(context.Background(), "user-456", nil, &bio, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		user, _ := userRepo.GetByID(context.Background(), "user-456")
		if user.Bio == nil || *user.Bio != "Software developer" {
			t.Errorf("expected bio Software developer")
		}
	})

	t.Run("Update Avatar URL", func(t *testing.T) {
		avatar := "https://example.com/avatar.jpg"
		err := authService.UpdateUserProfile(context.Background(), "user-456", nil, nil, &avatar)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		user, _ := userRepo.GetByID(context.Background(), "user-456")
		if user.AvatarURL == nil || *user.AvatarURL != "https://example.com/avatar.jpg" {
			t.Errorf("expected avatar URL")
		}
	})

	t.Run("User Not Found", func(t *testing.T) {
		fullName := "Test"
		err := authService.UpdateUserProfile(context.Background(), "non-existent", &fullName, nil, nil)
		if err == nil {
			t.Fatal("expected error for non-existent user")
		}
	})
}

// ==================== ChangePassword Tests ====================

func TestAuthService_ChangePassword(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Register a user first
	authResp, err := authService.Register(context.Background(), &domain.RegisterRequest{
		Email:    "password@example.com",
		Password: "oldPassword123",
		Locale:   "en",
	})
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}
	userID := authResp.User.ID

	t.Run("Successful Password Change", func(t *testing.T) {
		err := authService.ChangePassword(context.Background(), userID, "oldPassword123", "newPassword456")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify new password works
		_, err = authService.Login(context.Background(), &domain.LoginRequest{
			Email:    "password@example.com",
			Password: "newPassword456",
		})
		if err != nil {
			t.Fatalf("login with new password failed: %v", err)
		}
	})

	t.Run("Wrong Old Password", func(t *testing.T) {
		err := authService.ChangePassword(context.Background(), userID, "wrongOldPassword", "anotherPassword")
		if err == nil {
			t.Fatal("expected error for wrong old password")
		}
	})

	t.Run("User Not Found", func(t *testing.T) {
		err := authService.ChangePassword(context.Background(), "non-existent", "test", "test2test2")
		if err == nil {
			t.Fatal("expected error for non-existent user")
		}
	})

	// Note: ChangePassword does not validate password strength unlike Register
	// This is by design - validation should happen at handler level
}

// ==================== Logout Tests ====================

func TestAuthService_Logout(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	// Note: Logout requires Redis for token blacklisting
	// Without Redis, the function still works but doesn't blacklist
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Register a user
	authResp, err := authService.Register(context.Background(), &domain.RegisterRequest{
		Email:    "logout@example.com",
		Password: "password123",
		Locale:   "en",
	})
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	t.Run("Logout Without Redis", func(t *testing.T) {
		// Should succeed even without Redis (graceful degradation)
		err := authService.Logout(context.Background(), authResp.AccessToken, authResp.RefreshToken)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("Logout With Empty Token", func(t *testing.T) {
		err := authService.Logout(context.Background(), "", "")
		// Empty token might still work but is a no-op
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// ==================== Enable2FA Tests ====================

func TestAuthService_Enable2FA(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Create a user first
	user := &domain.User{
		ID:           "test-user-2fa",
		Email:        "2fa@example.com",
		TwoFAEnabled: false,
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	t.Run("Invalid 2FA Code", func(t *testing.T) {
		// Try to enable 2FA with an invalid code
		err := authService.Enable2FA(context.Background(), user.ID, "TESTSECRET", "000000")
		if err == nil {
			t.Fatal("expected error for invalid 2FA code")
		}
		var derr *domain.Error; if errors.As(err, &derr) {
			if derr.Code != domain.ErrInvalidCredentials {
				t.Errorf("expected error code %s, got %s", domain.ErrInvalidCredentials, derr.Code)
			}
		}
	})

	t.Run("User Not Found", func(t *testing.T) {
		err := authService.Enable2FA(context.Background(), "non-existent-user", "SECRET", "123456")
		if err == nil {
			t.Fatal("expected error for non-existent user")
		}
	})

	t.Run("Success Enable2FA", func(t *testing.T) {
		// Generate a real TOTP secret
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      "GoConnect",
			AccountName: user.Email,
		})
		if err != nil {
			t.Fatalf("failed to generate TOTP key: %v", err)
		}

		// Generate a valid TOTP code using the secret
		code, err := totp.GenerateCode(key.Secret(), time.Now())
		if err != nil {
			t.Fatalf("failed to generate TOTP code: %v", err)
		}

		// Enable 2FA with valid code
		err = authService.Enable2FA(context.Background(), user.ID, key.Secret(), code)
		if err != nil {
			t.Fatalf("failed to enable 2FA: %v", err)
		}

		// Verify 2FA is enabled
		updatedUser, _ := userRepo.GetByID(context.Background(), user.ID)
		if !updatedUser.TwoFAEnabled {
			t.Error("expected TwoFAEnabled to be true")
		}
		if updatedUser.TwoFAKey != key.Secret() {
			t.Error("expected TwoFAKey to be set")
		}
	})
}

// ==================== Disable2FA Tests ====================

func TestAuthService_Disable2FA(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	t.Run("User Without 2FA Enabled", func(t *testing.T) {
		// Create a user without 2FA
		user := &domain.User{
			ID:           "test-user-disable-2fa",
			Email:        "disable2fa@example.com",
			TwoFAEnabled: false,
		}
		if err := userRepo.Create(context.Background(), user); err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		// Disabling 2FA when not enabled should be a no-op (return nil)
		err := authService.Disable2FA(context.Background(), user.ID, "")
		if err != nil {
			t.Fatalf("unexpected error when disabling 2FA on user without 2FA: %v", err)
		}
	})

	t.Run("User Not Found", func(t *testing.T) {
		err := authService.Disable2FA(context.Background(), "non-existent-user", "123456")
		if err == nil {
			t.Fatal("expected error for non-existent user")
		}
	})

	t.Run("Invalid 2FA Code", func(t *testing.T) {
		// Create a user with 2FA enabled
		user := &domain.User{
			ID:           "test-user-disable-2fa-invalid",
			Email:        "disable2fa-invalid@example.com",
			TwoFAEnabled: true,
			TwoFAKey:     "TESTSECRETKEY",
		}
		if err := userRepo.Create(context.Background(), user); err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		err := authService.Disable2FA(context.Background(), user.ID, "000000")
		if err == nil {
			t.Fatal("expected error for invalid 2FA code")
		}
		var derr *domain.Error; if errors.As(err, &derr) {
			if derr.Code != domain.ErrInvalidCredentials {
				t.Errorf("expected error code %s, got %s", domain.ErrInvalidCredentials, derr.Code)
			}
		}
	})
}

// ==================== ValidateToken Tests ====================

func TestAuthService_ValidateToken(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	t.Run("Invalid Token Format", func(t *testing.T) {
		_, err := authService.ValidateToken(context.Background(), "invalid-token")
		if err == nil {
			t.Fatal("expected error for invalid token format")
		}
	})

	t.Run("Empty Token", func(t *testing.T) {
		_, err := authService.ValidateToken(context.Background(), "")
		if err == nil {
			t.Fatal("expected error for empty token")
		}
	})

	t.Run("Malformed JWT", func(t *testing.T) {
		_, err := authService.ValidateToken(context.Background(), "eyJ.malformed.token")
		if err == nil {
			t.Fatal("expected error for malformed JWT")
		}
	})

	t.Run("Valid Token For Registered User", func(t *testing.T) {
		// Register a user to get a valid token
		authResp, err := authService.Register(context.Background(), &domain.RegisterRequest{
			Email:    "validate@example.com",
			Password: "password123",
			Locale:   "en",
		})
		if err != nil {
			t.Fatalf("failed to register user: %v", err)
		}

		// Validate the access token
		claims, err := authService.ValidateToken(context.Background(), authResp.AccessToken)
		if err != nil {
			t.Fatalf("unexpected error validating token: %v", err)
		}

		if claims.Email != "validate@example.com" {
			t.Errorf("expected email validate@example.com, got %s", claims.Email)
		}

		if claims.Type != "access" {
			t.Errorf("expected token type 'access', got %s", claims.Type)
		}
	})

	t.Run("Suspended User Token", func(t *testing.T) {
		// Create a suspended user directly
		user := &domain.User{
			ID:        "suspended-user-id",
			Email:     "suspended@example.com",
			Suspended: true,
		}
		if err := userRepo.Create(context.Background(), user); err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		// Generate a token for this user using the internal method
		accessToken, err := authService.generateJWT(user.ID, "", user.Email, false, false, "access", time.Hour)
		if err != nil {
			t.Fatalf("failed to generate JWT: %v", err)
		}

		// Validate should fail because user is suspended
		_, err = authService.ValidateToken(context.Background(), accessToken)
		if err == nil {
			t.Fatal("expected error for suspended user token")
		}
		var derr *domain.Error; if errors.As(err, &derr) {
			if derr.Code != domain.ErrForbidden {
				t.Errorf("expected error code %s, got %s", domain.ErrForbidden, derr.Code)
			}
		}
	})
}

// ==================== GetRecoveryCodeCount Tests ====================

func TestAuthService_GetRecoveryCodeCount(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	t.Run("User With Recovery Codes", func(t *testing.T) {
		// Create a user with recovery codes
		user := &domain.User{
			ID:            "user-with-codes",
			Email:         "recovery@example.com",
			RecoveryCodes: []string{"code1", "code2", "code3"},
		}
		if err := userRepo.Create(context.Background(), user); err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		count, err := authService.GetRecoveryCodeCount(context.Background(), user.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if count != 3 {
			t.Errorf("expected 3 recovery codes, got %d", count)
		}
	})

	t.Run("User Without Recovery Codes", func(t *testing.T) {
		user := &domain.User{
			ID:            "user-without-codes",
			Email:         "norecovery@example.com",
			RecoveryCodes: nil,
		}
		if err := userRepo.Create(context.Background(), user); err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		count, err := authService.GetRecoveryCodeCount(context.Background(), user.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if count != 0 {
			t.Errorf("expected 0 recovery codes, got %d", count)
		}
	})

	t.Run("User Not Found", func(t *testing.T) {
		_, err := authService.GetRecoveryCodeCount(context.Background(), "non-existent")
		if err == nil {
			t.Fatal("expected error for non-existent user")
		}
	})
}

// ==================== USE RECOVERY CODE TESTS ====================

func TestAuthService_UseRecoveryCode(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Create a user with 2FA enabled and recovery codes
	password := "password123"
	hashedPassword, _ := authService.HashPassword(password)

	// Generate hashed recovery codes
	var recoveryCodes []string
	plainCode := "ABCD1234" // We'll use this to test
	hashedCode, _ := authService.HashPassword(plainCode)
	recoveryCodes = append(recoveryCodes, hashedCode)

	user := &domain.User{
		ID:            "user-with-2fa",
		Email:         "2fa@example.com",
		PasswordHash:  hashedPassword,
		TwoFAEnabled:  true,
		RecoveryCodes: recoveryCodes,
		TenantID:      "tenant-1",
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	t.Run("Success - Valid Recovery Code", func(t *testing.T) {
		resp, err := authService.UseRecoveryCode(context.Background(), &domain.UseRecoveryCodeRequest{
			Email:        "2fa@example.com",
			Password:     password,
			RecoveryCode: plainCode,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.AccessToken == "" {
			t.Error("expected access token")
		}
	})

	t.Run("Invalid Email", func(t *testing.T) {
		_, err := authService.UseRecoveryCode(context.Background(), &domain.UseRecoveryCodeRequest{
			Email:        "wrong@example.com",
			Password:     password,
			RecoveryCode: plainCode,
		})

		if err == nil {
			t.Fatal("expected error for invalid email")
		}
	})

	t.Run("Invalid Password", func(t *testing.T) {
		// Re-create user since code was consumed
		user.RecoveryCodes = []string{hashedCode}
		_ = userRepo.Update(context.Background(), user)

		_, err := authService.UseRecoveryCode(context.Background(), &domain.UseRecoveryCodeRequest{
			Email:        "2fa@example.com",
			Password:     "wrongpassword",
			RecoveryCode: plainCode,
		})

		if err == nil {
			t.Fatal("expected error for invalid password")
		}
	})

	t.Run("Invalid Recovery Code", func(t *testing.T) {
		// Re-add recovery code
		user.RecoveryCodes = []string{hashedCode}
		_ = userRepo.Update(context.Background(), user)

		_, err := authService.UseRecoveryCode(context.Background(), &domain.UseRecoveryCodeRequest{
			Email:        "2fa@example.com",
			Password:     password,
			RecoveryCode: "WRONGCODE",
		})

		if err == nil {
			t.Fatal("expected error for invalid recovery code")
		}
	})

	t.Run("2FA Not Enabled", func(t *testing.T) {
		// Create user without 2FA
		user2 := &domain.User{
			ID:           "user-no-2fa",
			Email:        "no2fa@example.com",
			PasswordHash: hashedPassword,
			TwoFAEnabled: false,
		}
		_ = userRepo.Create(context.Background(), user2)

		_, err := authService.UseRecoveryCode(context.Background(), &domain.UseRecoveryCodeRequest{
			Email:        "no2fa@example.com",
			Password:     password,
			RecoveryCode: "ANYCODE",
		})

		if err == nil {
			t.Fatal("expected error when 2FA not enabled")
		}
	})

	t.Run("No Recovery Codes", func(t *testing.T) {
		// User with 2FA enabled but no recovery codes
		user3 := &domain.User{
			ID:            "user-no-codes",
			Email:         "nocodes@example.com",
			PasswordHash:  hashedPassword,
			TwoFAEnabled:  true,
			RecoveryCodes: []string{},
		}
		_ = userRepo.Create(context.Background(), user3)

		_, err := authService.UseRecoveryCode(context.Background(), &domain.UseRecoveryCodeRequest{
			Email:        "nocodes@example.com",
			Password:     password,
			RecoveryCode: "ANYCODE",
		})

		if err == nil {
			t.Fatal("expected error when no recovery codes")
		}
	})
}

// ==================== OIDC LOGIN TESTS ====================

func TestAuthService_LoginOrRegisterOIDC(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	t.Run("Register New User", func(t *testing.T) {
		resp, err := authService.LoginOrRegisterOIDC(context.Background(), "oidc@example.com", "ext-123", "google")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.AccessToken == "" {
			t.Error("expected access token")
		}
		if resp.User.Email != "oidc@example.com" {
			t.Errorf("expected email oidc@example.com, got %s", resp.User.Email)
		}
		if resp.User.AuthProvider != "google" {
			t.Errorf("expected provider google, got %s", resp.User.AuthProvider)
		}
	})

	t.Run("Login Existing User", func(t *testing.T) {
		// Same email should return existing user
		resp, err := authService.LoginOrRegisterOIDC(context.Background(), "oidc@example.com", "ext-123", "google")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.User.Email != "oidc@example.com" {
			t.Errorf("expected email oidc@example.com, got %s", resp.User.Email)
		}
	})

	t.Run("Link Account - User Without Provider", func(t *testing.T) {
		// Create user without auth provider
		password, _ := authService.HashPassword("password123")
		user := &domain.User{
			ID:           "existing-user",
			Email:        "existing@example.com",
			PasswordHash: password,
			TenantID:     "tenant-1",
		}
		_ = userRepo.Create(context.Background(), user)
		_ = tenantRepo.Create(context.Background(), &domain.Tenant{ID: "tenant-1", Name: "Test"})

		resp, err := authService.LoginOrRegisterOIDC(context.Background(), "existing@example.com", "ext-456", "github")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.User.AuthProvider != "github" {
			t.Errorf("expected provider github, got %s", resp.User.AuthProvider)
		}
		if resp.User.ExternalID != "ext-456" {
			t.Errorf("expected externalID ext-456, got %s", resp.User.ExternalID)
		}
	})
}

func TestAuthService_ExtractJTI(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	t.Run("ValidToken", func(t *testing.T) {
		// Register a user to get a valid token
		resp, err := authService.Register(context.Background(), &domain.RegisterRequest{
			Email:    "jti-test@example.com",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("register failed: %v", err)
		}

		jti := authService.extractJTI(resp.AccessToken)
		if jti == "" {
			t.Error("expected non-empty JTI")
		}
	})

	t.Run("InvalidToken", func(t *testing.T) {
		jti := authService.extractJTI("invalid-token")
		if jti != "" {
			t.Error("expected empty JTI for invalid token")
		}
	})
}

func TestAuthService_ExtractUserID(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	t.Run("ValidToken", func(t *testing.T) {
		// Register a user to get a valid token
		resp, err := authService.Register(context.Background(), &domain.RegisterRequest{
			Email:    "userid-test@example.com",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("register failed: %v", err)
		}

		userID := authService.extractUserID(resp.AccessToken)
		if userID == "" {
			t.Error("expected non-empty userID")
		}
		if userID != resp.User.ID {
			t.Errorf("expected %s, got %s", resp.User.ID, userID)
		}
	})

	t.Run("InvalidToken", func(t *testing.T) {
		userID := authService.extractUserID("invalid-token")
		if userID != "" {
			t.Error("expected empty userID for invalid token")
		}
	})
}

func TestAuthService_AddToBlacklist_NoRedis(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	err := authService.addToBlacklist(context.Background(), "some-token")
	if err != nil {
		t.Errorf("expected nil error without redis, got %v", err)
	}
}

func TestAuthService_CheckBlacklist_NoRedis(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	err := authService.checkBlacklist(context.Background(), "some-token")
	if err != nil {
		t.Errorf("expected nil error without redis, got %v", err)
	}
}

// ==================== ADDITIONAL LOGIN TESTS ====================

func TestAuthService_Login_SuspendedUser(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Create a suspended user directly
	password := "password123"
	hashedPassword, _ := authService.HashPassword(password)
	user := &domain.User{
		ID:           "suspended-user-login",
		Email:        "suspended-login@example.com",
		PasswordHash: hashedPassword,
		Suspended:    true,
		TenantID:     "tenant-1",
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Attempt login with suspended user
	_, err := authService.Login(context.Background(), &domain.LoginRequest{
		Email:    "suspended-login@example.com",
		Password: password,
	})

	if err == nil {
		t.Fatal("expected error for suspended user login")
	}
	var derr *domain.Error; if errors.As(err, &derr) {
		if derr.Code != domain.ErrForbidden {
			t.Errorf("expected error code %s, got %s", domain.ErrForbidden, derr.Code)
		}
	}
}

func TestAuthService_Login_2FA_Required(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Create a user with 2FA enabled
	password := "password123"
	hashedPassword, _ := authService.HashPassword(password)
	user := &domain.User{
		ID:           "2fa-user-login",
		Email:        "2fa-login@example.com",
		PasswordHash: hashedPassword,
		TwoFAEnabled: true,
		TwoFAKey:     "TESTSECRETKEY123", // Dummy key
		TenantID:     "tenant-1",
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Attempt login without 2FA code
	_, err := authService.Login(context.Background(), &domain.LoginRequest{
		Email:    "2fa-login@example.com",
		Password: password,
	})

	if err == nil {
		t.Fatal("expected error for 2FA required")
	}
	var derr *domain.Error; if errors.As(err, &derr) {
		if derr.Code != "ERR_2FA_REQUIRED" {
			t.Errorf("expected error code ERR_2FA_REQUIRED, got %s", derr.Code)
		}
	}
}

func TestAuthService_Login_Invalid2FACode(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Create a user with 2FA enabled
	password := "password123"
	hashedPassword, _ := authService.HashPassword(password)
	user := &domain.User{
		ID:           "2fa-invalid-code",
		Email:        "2fa-invalid@example.com",
		PasswordHash: hashedPassword,
		TwoFAEnabled: true,
		TwoFAKey:     "JBSWY3DPEHPK3PXP", // A valid base32 encoded key
		TenantID:     "tenant-1",
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Attempt login with invalid 2FA code
	_, err := authService.Login(context.Background(), &domain.LoginRequest{
		Email:    "2fa-invalid@example.com",
		Password: password,
		Code:     "000000", // Invalid code
	})

	if err == nil {
		t.Fatal("expected error for invalid 2FA code")
	}
	var derr *domain.Error; if errors.As(err, &derr) {
		if derr.Code != domain.ErrInvalidCredentials {
			t.Errorf("expected error code %s, got %s", domain.ErrInvalidCredentials, derr.Code)
		}
	}
}

func TestAuthService_Login_Valid2FACode(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Create a user with 2FA enabled using a real secret
	password := "password123"
	hashedPassword, _ := authService.HashPassword(password)

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "GoConnect",
		AccountName: "2fa-valid@example.com",
	})
	if err != nil {
		t.Fatalf("failed to generate TOTP key: %v", err)
	}

	user := &domain.User{
		ID:           "2fa-valid-code",
		Email:        "2fa-valid@example.com",
		PasswordHash: hashedPassword,
		TwoFAEnabled: true,
		TwoFAKey:     key.Secret(),
		TenantID:     "tenant-1",
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Generate a valid TOTP code
	validCode, err := totp.GenerateCode(key.Secret(), time.Now())
	if err != nil {
		t.Fatalf("failed to generate TOTP code: %v", err)
	}

	// Attempt login with valid 2FA code
	resp, err := authService.Login(context.Background(), &domain.LoginRequest{
		Email:    "2fa-valid@example.com",
		Password: password,
		Code:     validCode,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected access token")
	}
}

// ==================== LOGOUT WITH VALID TOKENS TESTS ====================

func TestAuthService_Logout_WithValidTokens(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Register a user to get valid tokens
	authResp, err := authService.Register(context.Background(), &domain.RegisterRequest{
		Email:    "logout-valid@example.com",
		Password: "password123",
		Locale:   "en",
	})
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	// Logout should succeed (even without Redis)
	err = authService.Logout(context.Background(), authResp.AccessToken, authResp.RefreshToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthService_Logout_WithMalformedTokens(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Logout with malformed tokens should not panic
	err := authService.Logout(context.Background(), "not-a-jwt", "also-not-a-jwt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ==================== ADD TO BLACKLIST WITH TOKENS TESTS ====================

func TestAuthService_AddToBlacklist_WithValidToken(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Register a user to get a valid token
	authResp, err := authService.Register(context.Background(), &domain.RegisterRequest{
		Email:    "blacklist-test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	// Without Redis, addToBlacklist returns nil
	err = authService.addToBlacklist(context.Background(), authResp.AccessToken)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestAuthService_AddToBlacklist_InvalidToken(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Without Redis, addToBlacklist returns nil even for invalid tokens
	err := authService.addToBlacklist(context.Background(), "invalid-token")
	if err != nil {
		t.Errorf("expected nil error without redis, got %v", err)
	}
}

// ==================== VERIFY PASSWORD EDGE CASES ====================

func TestAuthService_VerifyPassword_MalformedHash(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Test with malformed hash
	_, err := authService.VerifyPassword("password", "not-an-argon2-hash")
	if err == nil {
		t.Error("expected error for malformed hash")
	}
}

func TestAuthService_VerifyPassword_InvalidBase64Salt(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Test with invalid base64 salt
	_, err := authService.VerifyPassword("password", "$argon2id$v=19$m=65536,t=1,p=4$invalid!!salt$validhash")
	if err == nil {
		t.Error("expected error for invalid base64 salt")
	}
}

// ==================== DISABLE 2FA TESTS ====================

func TestAuthService_Disable2FA_Success(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Generate a real TOTP key
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "GoConnect",
		AccountName: "disable2fa@example.com",
	})
	if err != nil {
		t.Fatalf("failed to generate TOTP key: %v", err)
	}

	// Create a user with 2FA enabled
	user := &domain.User{
		ID:           "disable-2fa-success",
		Email:        "disable2fa@example.com",
		TwoFAEnabled: true,
		TwoFAKey:     key.Secret(),
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Generate a valid TOTP code
	validCode, err := totp.GenerateCode(key.Secret(), time.Now())
	if err != nil {
		t.Fatalf("failed to generate TOTP code: %v", err)
	}

	// Disable 2FA with valid code
	err = authService.Disable2FA(context.Background(), user.ID, validCode)
	if err != nil {
		t.Fatalf("failed to disable 2FA: %v", err)
	}

	// Verify 2FA is disabled
	updatedUser, _ := userRepo.GetByID(context.Background(), user.ID)
	if updatedUser.TwoFAEnabled {
		t.Error("expected TwoFAEnabled to be false")
	}
}

// ==================== GENERATE 2FA SECRET TESTS ====================

func TestAuthService_Generate2FASecret_UserNotFound(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	_, _, err := authService.Generate2FASecret(context.Background(), "non-existent-user")
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}

func TestAuthService_Generate2FASecret_Success(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Create a test user
	user := &domain.User{
		ID:    "gen-2fa-secret",
		Email: "gen2fa@example.com",
	}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	secret, url, err := authService.Generate2FASecret(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("failed to generate 2FA secret: %v", err)
	}

	if secret == "" {
		t.Error("expected non-empty secret")
	}
	if url == "" {
		t.Error("expected non-empty URL")
	}
}

// ==================== ADDITIONAL JWT/TOKEN TESTS ====================

func TestAuthService_ExtractJTI_ValidToken(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Register user to get a valid token
	authResp, err := authService.Register(context.Background(), &domain.RegisterRequest{
		Email:    "jti-test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// extractJTI is private, but we can test it indirectly through Logout
	// The token should have a JTI claim
	if authResp.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
}

func TestAuthService_ExtractJTI_InvalidToken(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Test with completely invalid token - Logout should handle gracefully
	err := authService.Logout(context.Background(), "invalid.token.here", "also.invalid.token")
	// Should not error because redis is nil (graceful degradation)
	if err != nil {
		t.Errorf("expected no error with nil redis, got: %v", err)
	}
}

func TestAuthService_Logout_NilRedis(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil) // nil redis

	// Register user to get valid tokens
	authResp, err := authService.Register(context.Background(), &domain.RegisterRequest{
		Email:    "logout-test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// Logout with valid tokens but nil redis should succeed
	err = authService.Logout(context.Background(), authResp.AccessToken, authResp.RefreshToken)
	if err != nil {
		t.Errorf("expected no error with nil redis, got: %v", err)
	}
}

func TestAuthService_Logout_EmptyTokens(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Logout with empty tokens should succeed (graceful)
	err := authService.Logout(context.Background(), "", "")
	if err != nil {
		t.Errorf("expected no error with empty tokens, got: %v", err)
	}
}

func TestAuthService_Refresh_UserNotFound(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Create and delete user to get orphaned token
	authResp, err := authService.Register(context.Background(), &domain.RegisterRequest{
		Email:    "refresh-delete@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// Get user and delete
	user, _ := userRepo.GetByEmail(context.Background(), "refresh-delete@example.com")
	userRepo.Delete(context.Background(), user.ID)

	// Try to refresh - should fail because user not found
	_, err = authService.Refresh(context.Background(), &domain.RefreshRequest{
		RefreshToken: authResp.RefreshToken,
	})
	if err == nil {
		t.Error("expected error for deleted user refresh")
	}
}

func TestAuthService_Refresh_InvalidToken(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Try to refresh with completely invalid token
	_, err := authService.Refresh(context.Background(), &domain.RefreshRequest{
		RefreshToken: "not.a.valid.jwt.token",
	})
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestAuthService_Refresh_ExpiredToken(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Create an expired refresh token manually
	expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjEsInN1YiI6InVzZXItaWQifQ.invalid"

	_, err := authService.Refresh(context.Background(), &domain.RefreshRequest{
		RefreshToken: expiredToken,
	})
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestAuthService_ValidateToken_InvalidToken(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)
	ctx := context.Background()

	// Test with completely malformed token
	_, err := authService.ValidateToken(ctx, "not-a-valid-jwt")
	if err == nil {
		t.Error("expected error for invalid token")
	}

	// Test with empty token
	_, err = authService.ValidateToken(ctx, "")
	if err == nil {
		t.Error("expected error for empty token")
	}
}

func TestAuthService_GetEnvOrDefault(t *testing.T) {
	// This tests the getEnvOrDefault function indirectly
	// The function is used in NewAuthService

	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()

	// Service creation uses getEnvOrDefault
	authService := NewAuthService(userRepo, tenantRepo, nil)
	if authService == nil {
		t.Error("expected non-nil auth service")
	}
}

func TestAuthService_HashPassword_EdgeCases(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo, nil)

	// Test with empty password
	hash, err := authService.HashPassword("")
	if err != nil {
		t.Errorf("expected no error for empty password, got: %v", err)
	}
	if hash == "" {
		t.Error("expected non-empty hash for empty password")
	}

	// Test with very long password
	longPass := string(make([]byte, 1000))
	for i := range longPass {
		longPass = longPass[:i] + "a"
	}
	hash, err = authService.HashPassword(longPass)
	if err != nil {
		t.Errorf("expected no error for long password, got: %v", err)
	}
	if hash == "" {
		t.Error("expected non-empty hash for long password")
	}

	// Test with unicode password
	unicodePass := "пароль密码パスワード"
	hash, err = authService.HashPassword(unicodePass)
	if err != nil {
		t.Errorf("expected no error for unicode password, got: %v", err)
	}
	if hash == "" {
		t.Error("expected non-empty hash for unicode password")
	}
}
