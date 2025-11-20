package service

import (
	"context"
	"testing"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

func TestAuthService_Register(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo)

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
				if derr, ok := err.(*domain.Error); ok {
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
	authService := NewAuthService(userRepo, tenantRepo)

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
				if derr, ok := err.(*domain.Error); ok {
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

func TestAuthService_PasswordHashing(t *testing.T) {
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := NewAuthService(userRepo, tenantRepo)

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
	authService := NewAuthService(userRepo, tenantRepo)

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

	if newAuthResp.RefreshToken == authResp.RefreshToken {
		t.Error("new refresh token should be different from old one")
	}

	// Old refresh token should no longer work (rotation)
	_, err = authService.Refresh(context.Background(), refreshReq)
	if err == nil {
		t.Error("expected error when using old refresh token")
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
