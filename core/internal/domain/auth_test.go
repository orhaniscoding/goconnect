package domain

import (
	"testing"
	"time"
)

// Test User.Sanitize removes sensitive fields
func TestUser_Sanitize(t *testing.T) {
	now := time.Now()
	user := &User{
		ID:           "user-123",
		TenantID:     "tenant-456",
		Email:        "user@example.com",
		PasswordHash: "sensitive-password-hash-should-be-removed",
		Locale:       "en",
		IsAdmin:      true,
		IsModerator:  false,
		TwoFAKey:     "sensitive-2fa-key-should-be-removed",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	sanitized := user.Sanitize()

	// Verify non-sensitive fields are preserved
	if sanitized.ID != user.ID {
		t.Errorf("expected ID %s, got %s", user.ID, sanitized.ID)
	}
	if sanitized.TenantID != user.TenantID {
		t.Errorf("expected TenantID %s, got %s", user.TenantID, sanitized.TenantID)
	}
	if sanitized.Email != user.Email {
		t.Errorf("expected Email %s, got %s", user.Email, sanitized.Email)
	}
	if sanitized.Locale != user.Locale {
		t.Errorf("expected Locale %s, got %s", user.Locale, sanitized.Locale)
	}
	if sanitized.IsAdmin != user.IsAdmin {
		t.Errorf("expected IsAdmin %v, got %v", user.IsAdmin, sanitized.IsAdmin)
	}
	if sanitized.IsModerator != user.IsModerator {
		t.Errorf("expected IsModerator %v, got %v", user.IsModerator, sanitized.IsModerator)
	}
	if !sanitized.CreatedAt.Equal(user.CreatedAt) {
		t.Errorf("expected CreatedAt %v, got %v", user.CreatedAt, sanitized.CreatedAt)
	}
	if !sanitized.UpdatedAt.Equal(user.UpdatedAt) {
		t.Errorf("expected UpdatedAt %v, got %v", user.UpdatedAt, sanitized.UpdatedAt)
	}

	// Verify sensitive fields are removed
	if sanitized.PasswordHash != "" {
		t.Errorf("expected PasswordHash to be empty, got: %s", sanitized.PasswordHash)
	}
	if sanitized.TwoFAKey != "" {
		t.Errorf("expected TwoFAKey to be empty, got: %s", sanitized.TwoFAKey)
	}
}

// Test User.Sanitize returns a new instance
func TestUser_Sanitize_ReturnsNewInstance(t *testing.T) {
	original := &User{
		ID:           "user-123",
		Email:        "user@example.com",
		PasswordHash: "secret-hash",
	}

	sanitized := original.Sanitize()

	// Verify it's a different instance
	if sanitized == original {
		t.Error("Sanitize should return a new User instance, not modify original")
	}

	// Verify original is unchanged
	if original.PasswordHash != "secret-hash" {
		t.Error("Sanitize should not modify the original User")
	}

	// Verify sanitized has no password
	if sanitized.PasswordHash != "" {
		t.Error("Sanitized user should have empty PasswordHash")
	}
}

// Test User.Sanitize with various user types
func TestUser_Sanitize_VariousUserTypes(t *testing.T) {
	tests := []struct {
		name        string
		user        *User
		checkFields func(*testing.T, *User)
	}{
		{
			name: "Admin user",
			user: &User{
				ID:           "admin-1",
				Email:        "admin@example.com",
				PasswordHash: "admin-hash",
				IsAdmin:      true,
				IsModerator:  false,
			},
			checkFields: func(t *testing.T, sanitized *User) {
				if !sanitized.IsAdmin {
					t.Error("expected IsAdmin to be true")
				}
				if sanitized.IsModerator {
					t.Error("expected IsModerator to be false")
				}
				if sanitized.PasswordHash != "" {
					t.Error("expected PasswordHash to be empty")
				}
			},
		},
		{
			name: "Moderator user",
			user: &User{
				ID:           "mod-1",
				Email:        "mod@example.com",
				PasswordHash: "mod-hash",
				TwoFAKey:     "mod-2fa",
				IsAdmin:      false,
				IsModerator:  true,
			},
			checkFields: func(t *testing.T, sanitized *User) {
				if sanitized.IsAdmin {
					t.Error("expected IsAdmin to be false")
				}
				if !sanitized.IsModerator {
					t.Error("expected IsModerator to be true")
				}
				if sanitized.PasswordHash != "" {
					t.Error("expected PasswordHash to be empty")
				}
				if sanitized.TwoFAKey != "" {
					t.Error("expected TwoFAKey to be empty")
				}
			},
		},
		{
			name: "Regular user",
			user: &User{
				ID:           "user-1",
				Email:        "user@example.com",
				PasswordHash: "user-hash",
				IsAdmin:      false,
				IsModerator:  false,
			},
			checkFields: func(t *testing.T, sanitized *User) {
				if sanitized.IsAdmin {
					t.Error("expected IsAdmin to be false")
				}
				if sanitized.IsModerator {
					t.Error("expected IsModerator to be false")
				}
				if sanitized.PasswordHash != "" {
					t.Error("expected PasswordHash to be empty")
				}
			},
		},
		{
			name: "User with Turkish locale",
			user: &User{
				ID:           "user-tr",
				Email:        "user@example.com",
				PasswordHash: "hash",
				Locale:       "tr",
			},
			checkFields: func(t *testing.T, sanitized *User) {
				if sanitized.Locale != "tr" {
					t.Errorf("expected Locale 'tr', got %s", sanitized.Locale)
				}
				if sanitized.PasswordHash != "" {
					t.Error("expected PasswordHash to be empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized := tt.user.Sanitize()
			tt.checkFields(t, sanitized)

			// Common checks for all tests
			if sanitized.ID != tt.user.ID {
				t.Errorf("expected ID %s, got %s", tt.user.ID, sanitized.ID)
			}
			if sanitized.Email != tt.user.Email {
				t.Errorf("expected Email %s, got %s", tt.user.Email, sanitized.Email)
			}
		})
	}
}

// Test that sanitized user is safe for JSON marshaling
func TestUser_Sanitize_SafeForJSON(t *testing.T) {
	user := &User{
		ID:           "user-123",
		Email:        "user@example.com",
		PasswordHash: "this-should-never-appear-in-json",
		TwoFAKey:     "this-should-never-appear-either",
	}

	sanitized := user.Sanitize()

	// These fields should be safe to expose
	if sanitized.ID == "" {
		t.Error("ID should be present for JSON")
	}
	if sanitized.Email == "" {
		t.Error("Email should be present for JSON")
	}

	// These should NOT be present
	if sanitized.PasswordHash != "" {
		t.Error("PasswordHash should be empty for safe JSON marshaling")
	}
	if sanitized.TwoFAKey != "" {
		t.Error("TwoFAKey should be empty for safe JSON marshaling")
	}
}
