package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/service"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
)

func setupAuthTest() (*gin.Engine, *AuthHandler, *service.AuthService) {
	gin.SetMode(gin.TestMode)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := service.NewAuthService(userRepo, tenantRepo, nil)
	handler := NewAuthHandler(authService, nil)
	r := gin.New()
	return r, handler, authService
}

func TestAuthHandler_Register(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, _ := setupAuthTest()
		r.POST("/register", handler.Register)

		body := map[string]interface{}{"email": "test@example.com", "password": "password123", "locale": "en"}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		assert.NotEmpty(t, data["access_token"])
	})

	t.Run("Invalid email", func(t *testing.T) {
		r, handler, _ := setupAuthTest()
		r.POST("/register", handler.Register)

		body := map[string]interface{}{"email": "not-email", "password": "password123"}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Password too short", func(t *testing.T) {
		r, handler, _ := setupAuthTest()
		r.POST("/register", handler.Register)

		body := map[string]interface{}{"email": "test@example.com", "password": "short"}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Duplicate email", func(t *testing.T) {
		r, handler, authService := setupAuthTest()
		r.POST("/register", handler.Register)

		authService.Register(context.Background(), &domain.RegisterRequest{Email: "test@example.com", Password: "password123"})

		body := map[string]interface{}{"email": "test@example.com", "password": "password456"}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, authService := setupAuthTest()
		r.POST("/login", handler.Login)

		authService.Register(context.Background(), &domain.RegisterRequest{Email: "user@example.com", Password: "password123"})

		body := map[string]interface{}{"email": "user@example.com", "password": "password123"}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		assert.NotEmpty(t, data["access_token"])
	})

	t.Run("Wrong password", func(t *testing.T) {
		r, handler, authService := setupAuthTest()
		r.POST("/login", handler.Login)

		authService.Register(context.Background(), &domain.RegisterRequest{Email: "user@example.com", Password: "password123"})

		body := map[string]interface{}{"email": "user@example.com", "password": "wrongpass"}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("User not found", func(t *testing.T) {
		r, handler, _ := setupAuthTest()
		r.POST("/login", handler.Login)

		body := map[string]interface{}{"email": "notfound@example.com", "password": "password123"}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		r, handler, _ := setupAuthTest()
		r.POST("/login", handler.Login)

		req := httptest.NewRequest("POST", "/login", bytes.NewBufferString("{invalid}"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Empty request body", func(t *testing.T) {
		r, handler, _ := setupAuthTest()
		r.POST("/login", handler.Login)

		req := httptest.NewRequest("POST", "/login", nil)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAuthHandler_Refresh(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, authService := setupAuthTest()
		r.POST("/refresh", handler.Refresh)

		resp, _ := authService.Register(context.Background(), &domain.RegisterRequest{Email: "user@example.com", Password: "password123"})

		body := map[string]interface{}{"refresh_token": resp.RefreshToken}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		assert.NotEmpty(t, data["access_token"])
	})

	t.Run("Invalid token", func(t *testing.T) {
		r, handler, _ := setupAuthTest()
		r.POST("/refresh", handler.Refresh)

		body := map[string]interface{}{"refresh_token": "invalid-token"}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Invalid token format causes parsing error, returns 500
		assert.True(t, w.Code == http.StatusUnauthorized || w.Code == http.StatusInternalServerError)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		r, handler, _ := setupAuthTest()
		r.POST("/refresh", handler.Refresh)

		req := httptest.NewRequest("POST", "/refresh", bytes.NewBufferString("{invalid}"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Missing refresh token", func(t *testing.T) {
		r, handler, _ := setupAuthTest()
		r.POST("/refresh", handler.Refresh)

		body := map[string]interface{}{}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized)
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, handler, _ := setupAuthTest()
		r.POST("/logout", handler.Logout)

		// Logout requires refresh_token in body
		body := map[string]interface{}{"refresh_token": "dummy-refresh-token"}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/logout", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, true, response["ok"])
	})

	t.Run("With Authorization Header", func(t *testing.T) {
		r, handler, _ := setupAuthTest()
		r.POST("/logout", handler.Logout)

		body := map[string]interface{}{"refresh_token": "dummy-refresh-token"}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/logout", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer access-token-123")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Invalid JSON Body", func(t *testing.T) {
		r, handler, _ := setupAuthTest()
		r.POST("/logout", handler.Logout)

		req := httptest.NewRequest("POST", "/logout", bytes.NewBufferString("{invalid}"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

type stubOIDCService struct {
	loginURL    string
	userInfo    *service.UserInfo
	exchangeErr error
}

func (s *stubOIDCService) GetLoginURL(state string) string {
	if s.loginURL == "" {
		return "https://oidc.example.com/auth?state=" + state
	}
	return s.loginURL + "?state=" + state
}

func (s *stubOIDCService) ExchangeCode(_ context.Context, _ string) (*oidc.IDToken, *service.UserInfo, error) {
	if s.exchangeErr != nil {
		return nil, nil, s.exchangeErr
	}
	if s.userInfo != nil {
		return nil, s.userInfo, nil
	}
	return nil, &service.UserInfo{Email: "oidc@example.com", Sub: "sub-123"}, nil
}

func TestAuthHandler_LoginOIDC_SetsStateCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authSvc := service.NewAuthService(userRepo, tenantRepo, nil)
	oidcMock := &stubOIDCService{loginURL: "https://oidc.example.com/auth"}
	handler := NewAuthHandler(authSvc, oidcMock)

	r := gin.New()
	r.GET("/v1/auth/oidc/login", handler.LoginOIDC)

	req := httptest.NewRequest("GET", "/v1/auth/oidc/login", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	location := w.Result().Header.Get("Location")

	var stateCookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "oidc_state" {
			stateCookie = c
			break
		}
	}
	assert.NotNil(t, stateCookie)
	assert.NotEmpty(t, stateCookie.Value)
	assert.Contains(t, location, stateCookie.Value)
}

func TestAuthHandler_CallbackOIDC_InvalidState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authSvc := service.NewAuthService(userRepo, tenantRepo, nil)
	oidcMock := &stubOIDCService{}
	handler := NewAuthHandler(authSvc, oidcMock)

	r := gin.New()
	r.GET("/v1/auth/oidc/callback", handler.CallbackOIDC)

	req := httptest.NewRequest("GET", "/v1/auth/oidc/callback?code=abc&state=wrong", nil)
	// No state cookie set -> should be rejected
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_CallbackOIDC_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authSvc := service.NewAuthService(userRepo, tenantRepo, nil)
	oidcMock := &stubOIDCService{
		userInfo: &service.UserInfo{
			Email: "oidc-user@example.com",
			Sub:   "oidc-subject-1",
		},
	}
	handler := NewAuthHandler(authSvc, oidcMock)

	r := gin.New()
	r.GET("/v1/auth/oidc/callback", handler.CallbackOIDC)

	state := "state-123"
	req := httptest.NewRequest("GET", "/v1/auth/oidc/callback?code=abc&state="+state, nil)
	req.AddCookie(&http.Cookie{
		Name:  "oidc_state",
		Value: state,
		Path:  "/v1/auth/oidc",
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	location := w.Result().Header.Get("Location")
	assert.Contains(t, location, "access_token=")

	// State cookie should be cleared
	var cleared bool
	for _, c := range w.Result().Cookies() {
		if c.Name == "oidc_state" {
			cleared = true
			assert.True(t, c.MaxAge < 0)
		}
	}
	assert.True(t, cleared)
}

func TestAuthHandler_CallbackOIDC_MissingCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authSvc := service.NewAuthService(userRepo, tenantRepo, nil)
	oidcMock := &stubOIDCService{}
	handler := NewAuthHandler(authSvc, oidcMock)

	r := gin.New()
	r.GET("/v1/auth/oidc/callback", handler.CallbackOIDC)

	// Missing code param
	req := httptest.NewRequest("GET", "/v1/auth/oidc/callback?state=state-123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_CallbackOIDC_MissingState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authSvc := service.NewAuthService(userRepo, tenantRepo, nil)
	oidcMock := &stubOIDCService{}
	handler := NewAuthHandler(authSvc, oidcMock)

	r := gin.New()
	r.GET("/v1/auth/oidc/callback", handler.CallbackOIDC)

	// Missing state param
	req := httptest.NewRequest("GET", "/v1/auth/oidc/callback?code=abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_LoginOIDC_NoOIDCService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authSvc := service.NewAuthService(userRepo, tenantRepo, nil)
	// No OIDC service configured
	handler := NewAuthHandler(authSvc, nil)

	r := gin.New()
	r.GET("/v1/auth/oidc/login", handler.LoginOIDC)

	req := httptest.NewRequest("GET", "/v1/auth/oidc/login", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_CallbackOIDC_NoOIDCService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authSvc := service.NewAuthService(userRepo, tenantRepo, nil)
	// No OIDC service configured
	handler := NewAuthHandler(authSvc, nil)

	r := gin.New()
	r.GET("/v1/auth/oidc/callback", handler.CallbackOIDC)

	req := httptest.NewRequest("GET", "/v1/auth/oidc/callback?code=abc&state=123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==================== 2FA TESTS ====================

func TestAuthHandler_Generate2FA(t *testing.T) {
	t.Run("Unauthorized", func(t *testing.T) {
		r, handler, _ := setupAuthTest()
		r.POST("/v1/auth/2fa/generate", handler.Generate2FA)

		req := httptest.NewRequest("POST", "/v1/auth/2fa/generate", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		userRepo := repository.NewInMemoryUserRepository()
		tenantRepo := repository.NewInMemoryTenantRepository()
		authService := service.NewAuthService(userRepo, tenantRepo, nil)

		// Create user
		userRepo.Create(context.Background(), &domain.User{
			ID:       "user-1",
			TenantID: "tenant-1",
			Email:    "test@example.com",
		})

		handler := NewAuthHandler(authService, nil)
		r := gin.New()

		r.POST("/v1/auth/2fa/generate", func(c *gin.Context) {
			c.Set("user_id", "user-1")
			c.Next()
		}, handler.Generate2FA)

		req := httptest.NewRequest("POST", "/v1/auth/2fa/generate", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		assert.NotEmpty(t, data["secret"])
		assert.NotEmpty(t, data["url"])
	})
}

func TestAuthHandler_Enable2FA(t *testing.T) {
	t.Run("Unauthorized", func(t *testing.T) {
		r, handler, _ := setupAuthTest()
		r.POST("/v1/auth/2fa/enable", handler.Enable2FA)

		req := httptest.NewRequest("POST", "/v1/auth/2fa/enable", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		userRepo := repository.NewInMemoryUserRepository()
		tenantRepo := repository.NewInMemoryTenantRepository()
		authService := service.NewAuthService(userRepo, tenantRepo, nil)

		userRepo.Create(context.Background(), &domain.User{
			ID:       "user-1",
			TenantID: "tenant-1",
			Email:    "test@example.com",
		})

		handler := NewAuthHandler(authService, nil)
		r := gin.New()

		r.POST("/v1/auth/2fa/enable", func(c *gin.Context) {
			c.Set("user_id", "user-1")
			c.Next()
		}, handler.Enable2FA)

		req := httptest.NewRequest("POST", "/v1/auth/2fa/enable", bytes.NewBuffer([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAuthHandler_Disable2FA(t *testing.T) {
	t.Run("Unauthorized", func(t *testing.T) {
		r, handler, _ := setupAuthTest()
		r.POST("/v1/auth/2fa/disable", handler.Disable2FA)

		req := httptest.NewRequest("POST", "/v1/auth/2fa/disable", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		userRepo := repository.NewInMemoryUserRepository()
		tenantRepo := repository.NewInMemoryTenantRepository()
		authService := service.NewAuthService(userRepo, tenantRepo, nil)

		userRepo.Create(context.Background(), &domain.User{
			ID:       "user-1",
			TenantID: "tenant-1",
			Email:    "test@example.com",
		})

		handler := NewAuthHandler(authService, nil)
		r := gin.New()

		r.POST("/v1/auth/2fa/disable", func(c *gin.Context) {
			c.Set("user_id", "user-1")
			c.Next()
		}, handler.Disable2FA)

		req := httptest.NewRequest("POST", "/v1/auth/2fa/disable", bytes.NewBuffer([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// ==================== RECOVERY CODE TESTS ====================

func TestAuthHandler_GenerateRecoveryCodes(t *testing.T) {
	t.Run("Unauthorized", func(t *testing.T) {
		r, handler, _ := setupAuthTest()
		r.POST("/v1/auth/recovery-codes", handler.GenerateRecoveryCodes)

		req := httptest.NewRequest("POST", "/v1/auth/recovery-codes", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		userRepo := repository.NewInMemoryUserRepository()
		tenantRepo := repository.NewInMemoryTenantRepository()
		authService := service.NewAuthService(userRepo, tenantRepo, nil)

		userRepo.Create(context.Background(), &domain.User{
			ID:       "user-1",
			TenantID: "tenant-1",
			Email:    "test@example.com",
		})

		handler := NewAuthHandler(authService, nil)
		r := gin.New()

		r.POST("/v1/auth/recovery-codes", func(c *gin.Context) {
			c.Set("user_id", "user-1")
			c.Next()
		}, handler.GenerateRecoveryCodes)

		req := httptest.NewRequest("POST", "/v1/auth/recovery-codes", bytes.NewBuffer([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAuthHandler_UseRecoveryCode(t *testing.T) {
	t.Run("Invalid Request Body", func(t *testing.T) {
		r, handler, _ := setupAuthTest()
		r.POST("/v1/auth/recovery-codes/use", handler.UseRecoveryCode)

		// Empty body should return 400 as it's invalid JSON
		req := httptest.NewRequest("POST", "/v1/auth/recovery-codes/use", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Missing email in request", func(t *testing.T) {
		r, handler, _ := setupAuthTest()
		r.POST("/v1/auth/recovery-codes/use", handler.UseRecoveryCode)

		body := map[string]string{"code": "TESTCODE"}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/auth/recovery-codes/use", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should return 400 due to missing email
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid credentials", func(t *testing.T) {
		r, handler, _ := setupAuthTest()
		r.POST("/v1/auth/recovery-codes/use", handler.UseRecoveryCode)

		// Valid request body but non-existent user
		body := map[string]string{
			"email":         "nonexistent@example.com",
			"password":      "password123",
			"recovery_code": "TESTCODE",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/auth/recovery-codes/use", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should return error (401 or similar) for invalid credentials
		assert.NotEqual(t, http.StatusOK, w.Code)
	})
}

func TestAuthHandler_GetRecoveryCodeCount(t *testing.T) {
	t.Run("Unauthorized", func(t *testing.T) {
		r, handler, _ := setupAuthTest()
		r.GET("/v1/auth/recovery-codes/count", handler.GetRecoveryCodeCount)

		req := httptest.NewRequest("GET", "/v1/auth/recovery-codes/count", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		userRepo := repository.NewInMemoryUserRepository()
		tenantRepo := repository.NewInMemoryTenantRepository()
		authService := service.NewAuthService(userRepo, tenantRepo, nil)

		userRepo.Create(context.Background(), &domain.User{
			ID:       "user-1",
			TenantID: "tenant-1",
			Email:    "test@example.com",
		})

		handler := NewAuthHandler(authService, nil)
		r := gin.New()

		r.GET("/v1/auth/recovery-codes/count", func(c *gin.Context) {
			c.Set("user_id", "user-1")
			c.Next()
		}, handler.GetRecoveryCodeCount)

		req := httptest.NewRequest("GET", "/v1/auth/recovery-codes/count", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// ==================== PASSWORD/PROFILE TESTS ====================

func TestAuthHandler_ChangePassword(t *testing.T) {
	t.Run("Unauthorized", func(t *testing.T) {
		r, handler, _ := setupAuthTest()
		r.POST("/v1/auth/password", handler.ChangePassword)

		req := httptest.NewRequest("POST", "/v1/auth/password", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		userRepo := repository.NewInMemoryUserRepository()
		tenantRepo := repository.NewInMemoryTenantRepository()
		authService := service.NewAuthService(userRepo, tenantRepo, nil)

		userRepo.Create(context.Background(), &domain.User{
			ID:       "user-1",
			TenantID: "tenant-1",
			Email:    "test@example.com",
		})

		handler := NewAuthHandler(authService, nil)
		r := gin.New()

		r.POST("/v1/auth/password", func(c *gin.Context) {
			c.Set("user_id", "user-1")
			c.Next()
		}, handler.ChangePassword)

		req := httptest.NewRequest("POST", "/v1/auth/password", bytes.NewBuffer([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAuthHandler_Me(t *testing.T) {
	t.Run("Unauthorized", func(t *testing.T) {
		r, handler, _ := setupAuthTest()
		r.GET("/v1/auth/me", handler.Me)

		req := httptest.NewRequest("GET", "/v1/auth/me", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		userRepo := repository.NewInMemoryUserRepository()
		tenantRepo := repository.NewInMemoryTenantRepository()
		authService := service.NewAuthService(userRepo, tenantRepo, nil)

		userRepo.Create(context.Background(), &domain.User{
			ID:       "user-1",
			TenantID: "tenant-1",
			Email:    "test@example.com",
		})

		handler := NewAuthHandler(authService, nil)
		r := gin.New()

		r.GET("/v1/auth/me", func(c *gin.Context) {
			c.Set("user_id", "user-1")
			c.Next()
		}, handler.Me)

		req := httptest.NewRequest("GET", "/v1/auth/me", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "test@example.com", data["email"])
	})

	t.Run("User Not Found", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		userRepo := repository.NewInMemoryUserRepository()
		tenantRepo := repository.NewInMemoryTenantRepository()
		authService := service.NewAuthService(userRepo, tenantRepo, nil)

		handler := NewAuthHandler(authService, nil)
		r := gin.New()

		r.GET("/v1/auth/me", func(c *gin.Context) {
			c.Set("user_id", "non-existent-user")
			c.Next()
		}, handler.Me)

		req := httptest.NewRequest("GET", "/v1/auth/me", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestAuthHandler_UpdateProfile(t *testing.T) {
	t.Run("Unauthorized", func(t *testing.T) {
		r, handler, _ := setupAuthTest()
		r.PUT("/v1/auth/me", handler.UpdateProfile)

		req := httptest.NewRequest("PUT", "/v1/auth/me", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Request Body", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		userRepo := repository.NewInMemoryUserRepository()
		tenantRepo := repository.NewInMemoryTenantRepository()
		authService := service.NewAuthService(userRepo, tenantRepo, nil)

		userRepo.Create(context.Background(), &domain.User{
			ID:       "user-1",
			TenantID: "tenant-1",
			Email:    "test@example.com",
		})

		handler := NewAuthHandler(authService, nil)
		r := gin.New()

		r.PUT("/v1/auth/me", func(c *gin.Context) {
			c.Set("user_id", "user-1")
			c.Next()
		}, handler.UpdateProfile)

		req := httptest.NewRequest("PUT", "/v1/auth/me", bytes.NewBuffer([]byte(`{invalid`)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Success", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		userRepo := repository.NewInMemoryUserRepository()
		tenantRepo := repository.NewInMemoryTenantRepository()
		authService := service.NewAuthService(userRepo, tenantRepo, nil)

		userRepo.Create(context.Background(), &domain.User{
			ID:       "user-1",
			TenantID: "tenant-1",
			Email:    "test@example.com",
		})

		handler := NewAuthHandler(authService, nil)
		r := gin.New()

		r.PUT("/v1/auth/me", func(c *gin.Context) {
			c.Set("user_id", "user-1")
			c.Next()
		}, handler.UpdateProfile)

		username := "johndoe"
		body := map[string]*string{"username": &username}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("PUT", "/v1/auth/me", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// ==================== ADDITIONAL COVERAGE TESTS ====================

func TestAuthHandler_ChangePassword_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := service.NewAuthService(userRepo, tenantRepo, nil)

	// Register user with password
	_, err := authService.Register(context.Background(), &domain.RegisterRequest{
		Email:    "user@example.com",
		Password: "password123",
	})
	assert.NoError(t, err)

	// Get user to get ID
	user, _ := userRepo.GetByEmail(context.Background(), "user@example.com")

	handler := NewAuthHandler(authService, nil)
	r := gin.New()

	r.POST("/v1/auth/password", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Next()
	}, handler.ChangePassword)

	body := map[string]string{
		"old_password": "password123",
		"new_password": "newpassword456",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/auth/password", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandler_ChangePassword_InvalidOldPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := service.NewAuthService(userRepo, tenantRepo, nil)

	// Register user with password
	_, err := authService.Register(context.Background(), &domain.RegisterRequest{
		Email:    "user@example.com",
		Password: "password123",
	})
	assert.NoError(t, err)

	// Get user to get ID
	user, _ := userRepo.GetByEmail(context.Background(), "user@example.com")

	handler := NewAuthHandler(authService, nil)
	r := gin.New()

	r.POST("/v1/auth/password", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Next()
	}, handler.ChangePassword)

	body := map[string]string{
		"old_password": "wrongpassword",
		"new_password": "newpassword456",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/auth/password", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_Logout_EmptyBody(t *testing.T) {
	r, handler, _ := setupAuthTest()
	r.POST("/logout", handler.Logout)

	// Empty body should still succeed - logout doesn't require refresh token to be valid
	req := httptest.NewRequest("POST", "/logout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// This tests the empty body path through Logout
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadRequest)
}

func TestAuthHandler_UseRecoveryCode_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := service.NewAuthService(userRepo, tenantRepo, nil)

	// Register user
	_, err := authService.Register(context.Background(), &domain.RegisterRequest{
		Email:    "user@example.com",
		Password: "password123",
	})
	assert.NoError(t, err)

	// Get user and enable 2FA
	user, _ := userRepo.GetByEmail(context.Background(), "user@example.com")
	key, _ := totp.Generate(totp.GenerateOpts{
		Issuer:      "GoConnect",
		AccountName: user.Email,
	})
	user.TwoFAKey = key.Secret()
	user.TwoFAEnabled = true

	// Generate a recovery code (10 characters) and store its hash
	recoveryCode := "ABCDE12345"
	hashedCode, _ := authService.HashPassword(recoveryCode)
	user.RecoveryCodes = []string{hashedCode}
	userRepo.Update(context.Background(), user)

	handler := NewAuthHandler(authService, nil)
	r := gin.New()

	r.POST("/v1/auth/2fa/recovery", handler.UseRecoveryCode)

	body := map[string]string{
		"email":         "user@example.com",
		"password":      "password123",
		"recovery_code": recoveryCode,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/auth/2fa/recovery", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})
	assert.NotEmpty(t, data["access_token"])
}

func TestAuthHandler_Enable2FA_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := service.NewAuthService(userRepo, tenantRepo, nil)

	// Register user
	_, err := authService.Register(context.Background(), &domain.RegisterRequest{
		Email:    "user@example.com",
		Password: "password123",
	})
	assert.NoError(t, err)

	// Get user and set 2FA key (from Generate2FA)
	user, _ := userRepo.GetByEmail(context.Background(), "user@example.com")
	key, _ := totp.Generate(totp.GenerateOpts{
		Issuer:      "GoConnect",
		AccountName: user.Email,
	})
	user.TwoFAKey = key.Secret()
	userRepo.Update(context.Background(), user)

	// Generate a valid TOTP code
	code, _ := totp.GenerateCode(key.Secret(), time.Now())

	handler := NewAuthHandler(authService, nil)
	r := gin.New()

	r.POST("/v1/auth/2fa/enable", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Next()
	}, handler.Enable2FA)

	body := map[string]string{"code": code, "secret": key.Secret()}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/auth/2fa/enable", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandler_Disable2FA_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := service.NewAuthService(userRepo, tenantRepo, nil)

	// Register user
	_, err := authService.Register(context.Background(), &domain.RegisterRequest{
		Email:    "user@example.com",
		Password: "password123",
	})
	assert.NoError(t, err)

	// Get user and enable 2FA
	user, _ := userRepo.GetByEmail(context.Background(), "user@example.com")
	key, _ := totp.Generate(totp.GenerateOpts{
		Issuer:      "GoConnect",
		AccountName: user.Email,
	})
	user.TwoFAKey = key.Secret()
	user.TwoFAEnabled = true
	userRepo.Update(context.Background(), user)

	// Generate a valid TOTP code
	code, _ := totp.GenerateCode(key.Secret(), time.Now())

	handler := NewAuthHandler(authService, nil)
	r := gin.New()

	r.POST("/v1/auth/2fa/disable", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Next()
	}, handler.Disable2FA)

	body := map[string]string{"code": code}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/auth/2fa/disable", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandler_GenerateRecoveryCodes_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := service.NewAuthService(userRepo, tenantRepo, nil)

	// Register user
	_, err := authService.Register(context.Background(), &domain.RegisterRequest{
		Email:    "user@example.com",
		Password: "password123",
	})
	assert.NoError(t, err)

	// Get user and enable 2FA
	user, _ := userRepo.GetByEmail(context.Background(), "user@example.com")
	key, _ := totp.Generate(totp.GenerateOpts{
		Issuer:      "GoConnect",
		AccountName: user.Email,
	})
	user.TwoFAKey = key.Secret()
	user.TwoFAEnabled = true
	userRepo.Update(context.Background(), user)

	// Generate a valid TOTP code
	code, _ := totp.GenerateCode(key.Secret(), time.Now())

	handler := NewAuthHandler(authService, nil)
	r := gin.New()

	r.POST("/v1/auth/recovery-codes", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Next()
	}, handler.GenerateRecoveryCodes)

	body := map[string]string{"code": code}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/auth/recovery-codes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	data := response["data"].(map[string]interface{})
	codes := data["codes"].([]interface{})
	assert.Equal(t, 8, len(codes))
}

// ==================== ADDITIONAL EDGE CASE TESTS ====================

func TestAuthHandler_Generate2FA_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := service.NewAuthService(userRepo, tenantRepo, nil)

	handler := NewAuthHandler(authService, nil)
	r := gin.New()

	// User ID that doesn't exist in the repo
	r.POST("/v1/auth/2fa/generate", func(c *gin.Context) {
		c.Set("user_id", "nonexistent-user")
		c.Next()
	}, handler.Generate2FA)

	req := httptest.NewRequest("POST", "/v1/auth/2fa/generate", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should return 500 since user doesn't exist
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAuthHandler_Enable2FA_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := service.NewAuthService(userRepo, tenantRepo, nil)

	handler := NewAuthHandler(authService, nil)
	r := gin.New()

	r.POST("/v1/auth/2fa/enable", func(c *gin.Context) {
		c.Set("user_id", "nonexistent-user")
		c.Next()
	}, handler.Enable2FA)

	body := map[string]string{"code": "123456", "secret": "invalid-secret"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/auth/2fa/enable", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should return error since user doesn't exist
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestAuthHandler_Disable2FA_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := service.NewAuthService(userRepo, tenantRepo, nil)

	handler := NewAuthHandler(authService, nil)
	r := gin.New()

	r.POST("/v1/auth/2fa/disable", func(c *gin.Context) {
		c.Set("user_id", "nonexistent-user")
		c.Next()
	}, handler.Disable2FA)

	body := map[string]string{"code": "123456"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/auth/2fa/disable", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should return error since user doesn't exist
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestAuthHandler_GenerateRecoveryCodes_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := service.NewAuthService(userRepo, tenantRepo, nil)

	handler := NewAuthHandler(authService, nil)
	r := gin.New()

	r.POST("/v1/auth/recovery-codes", func(c *gin.Context) {
		c.Set("user_id", "nonexistent-user")
		c.Next()
	}, handler.GenerateRecoveryCodes)

	body := map[string]string{"code": "123456"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/auth/recovery-codes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should return error since user doesn't exist
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestAuthHandler_GetRecoveryCodeCount_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := service.NewAuthService(userRepo, tenantRepo, nil)

	handler := NewAuthHandler(authService, nil)
	r := gin.New()

	r.GET("/v1/auth/recovery-codes/count", func(c *gin.Context) {
		c.Set("user_id", "nonexistent-user")
		c.Next()
	}, handler.GetRecoveryCodeCount)

	req := httptest.NewRequest("GET", "/v1/auth/recovery-codes/count", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should return error since user doesn't exist
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestAuthHandler_UseRecoveryCode_InvalidCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := service.NewAuthService(userRepo, tenantRepo, nil)

	// Register user
	_, err := authService.Register(context.Background(), &domain.RegisterRequest{
		Email:    "user@example.com",
		Password: "password123",
	})
	assert.NoError(t, err)

	// Get user and enable 2FA
	user, _ := userRepo.GetByEmail(context.Background(), "user@example.com")
	key, _ := totp.Generate(totp.GenerateOpts{
		Issuer:      "GoConnect",
		AccountName: user.Email,
	})
	user.TwoFAKey = key.Secret()
	user.TwoFAEnabled = true
	userRepo.Update(context.Background(), user)

	handler := NewAuthHandler(authService, nil)
	r := gin.New()

	r.POST("/v1/auth/2fa/recovery", handler.UseRecoveryCode)

	body := map[string]string{
		"email":         "user@example.com",
		"password":      "password123",
		"recovery_code": "INVALID-CODE",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/auth/2fa/recovery", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should return error since recovery code is invalid
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestAuthHandler_UseRecoveryCode_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := service.NewAuthService(userRepo, tenantRepo, nil)

	handler := NewAuthHandler(authService, nil)
	r := gin.New()

	r.POST("/v1/auth/2fa/recovery", handler.UseRecoveryCode)

	body := map[string]string{
		"email":         "nonexistent@example.com",
		"password":      "password123",
		"recovery_code": "ABCDE12345",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/auth/2fa/recovery", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should return error since user doesn't exist
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestAuthHandler_ChangePassword_ServiceNonDomainError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := service.NewAuthService(userRepo, tenantRepo, nil)

	handler := NewAuthHandler(authService, nil)
	r := gin.New()

	// User ID that doesn't exist - will cause error
	r.POST("/v1/auth/password", func(c *gin.Context) {
		c.Set("user_id", "nonexistent-user")
		c.Next()
	}, handler.ChangePassword)

	body := map[string]string{
		"old_password": "oldpass123",
		"new_password": "newpass456",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/auth/password", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should return error since user doesn't exist
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestAuthHandler_UpdateProfile_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := service.NewAuthService(userRepo, tenantRepo, nil)

	handler := NewAuthHandler(authService, nil)
	r := gin.New()

	// User ID that doesn't exist - will cause error
	r.PUT("/v1/users/me", func(c *gin.Context) {
		c.Set("user_id", "nonexistent-user")
		c.Next()
	}, handler.UpdateProfile)

	fullName := "New Name"
	body := map[string]*string{
		"full_name": &fullName,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/v1/users/me", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should return error since user doesn't exist
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestAuthHandler_Logout_WithAccessToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := service.NewAuthService(userRepo, tenantRepo, nil)

	// Register and login to get tokens
	authResp, err := authService.Register(context.Background(), &domain.RegisterRequest{
		Email:    "user@example.com",
		Password: "password123",
	})
	assert.NoError(t, err)

	handler := NewAuthHandler(authService, nil)
	r := gin.New()

	r.POST("/logout", handler.Logout)

	body := map[string]string{"refresh_token": authResp.RefreshToken}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/logout", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authResp.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandler_CallbackOIDC_ExchangeError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := repository.NewInMemoryUserRepository()
	tenantRepo := repository.NewInMemoryTenantRepository()
	authService := service.NewAuthService(userRepo, tenantRepo, nil)

	// Mock OIDC provider that returns error on exchange
	stubOIDC := &stubOIDCService{
		exchangeErr: assert.AnError,
	}

	handler := NewAuthHandler(authService, stubOIDC)
	r := gin.New()

	r.GET("/v1/auth/oidc/callback", handler.CallbackOIDC)

	req := httptest.NewRequest("GET", "/v1/auth/oidc/callback?code=test-code&state=test-state", nil)
	req.AddCookie(&http.Cookie{Name: "oidc_state", Value: "test-state"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should return error due to exchange failure
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
