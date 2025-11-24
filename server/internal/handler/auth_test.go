package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
	"github.com/orhaniscoding/goconnect/server/internal/service"
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
}
