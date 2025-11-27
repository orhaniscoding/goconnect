package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"time"

	"context"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/service"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService *service.AuthService
	oidcService OIDCProvider
}

// OIDCProvider defines the minimal surface required from the OIDC service.
type OIDCProvider interface {
	GetLoginURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*oidc.IDToken, *service.UserInfo, error)
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService, oidcService OIDCProvider) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		oidcService: oidcService,
	}
}

// Register handles POST /v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Invalid request body: "+err.Error(),
			map[string]string{"details": err.Error()}))
		return
	}

	user, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			errorResponse(c, domainErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": user,
	})
}

// Login handles POST /v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Invalid request body: "+err.Error(),
			map[string]string{"details": err.Error()}))
		return
	}

	authResp, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			errorResponse(c, domainErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": authResp,
	})
}

// Refresh handles POST /v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req domain.RefreshRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Invalid request body: "+err.Error(),
			map[string]string{"details": err.Error()}))
		return
	}

	authResp, err := h.authService.Refresh(c.Request.Context(), &req)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			errorResponse(c, domainErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": authResp,
	})
}

// Logout handles POST /v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	var req domain.RefreshRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest,
			"Invalid request body: "+err.Error(),
			map[string]string{"details": err.Error()}))
		return
	}

	if err := h.authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			errorResponse(c, domainErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok": true,
	})
}

// Generate2FA handles POST /v1/auth/2fa/generate
func (h *AuthHandler) Generate2FA(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Unauthorized", nil))
		return
	}

	secret, url, err := h.authService.Generate2FASecret(c.Request.Context(), userID)
	if err != nil {
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Failed to generate 2FA secret", nil))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"secret": secret,
			"url":    url,
		},
	})
}

// Enable2FA handles POST /v1/auth/2fa/enable
func (h *AuthHandler) Enable2FA(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Unauthorized", nil))
		return
	}

	var req domain.Enable2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Invalid request", nil))
		return
	}

	if err := h.authService.Enable2FA(c.Request.Context(), userID, req.Secret, req.Code); err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			errorResponse(c, domainErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, "Failed to enable 2FA", nil))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "2FA enabled successfully",
	})
}

// Disable2FA handles POST /v1/auth/2fa/disable
func (h *AuthHandler) Disable2FA(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Unauthorized", nil))
		return
	}

	var req domain.Disable2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Invalid request", nil))
		return
	}

	if err := h.authService.Disable2FA(c.Request.Context(), userID, req.Code); err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			errorResponse(c, domainErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, "Failed to disable 2FA", nil))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "2FA disabled successfully",
	})
}

// GenerateRecoveryCodes handles POST /v1/auth/2fa/recovery-codes
func (h *AuthHandler) GenerateRecoveryCodes(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Unauthorized", nil))
		return
	}

	var req domain.RegenerateRecoveryCodesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Invalid request: TOTP code required", nil))
		return
	}

	codes, err := h.authService.GenerateRecoveryCodes(c.Request.Context(), userID, req.Code)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			errorResponse(c, domainErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, "Failed to generate recovery codes", nil))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": domain.RecoveryCodeResponse{
			Codes: codes,
		},
		"message": "Recovery codes generated. Store these safely - they will only be shown once!",
	})
}

// UseRecoveryCode handles POST /v1/auth/2fa/recovery
func (h *AuthHandler) UseRecoveryCode(c *gin.Context) {
	var req domain.UseRecoveryCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Invalid request body: "+err.Error(), nil))
		return
	}

	authResp, err := h.authService.UseRecoveryCode(c.Request.Context(), &req)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			errorResponse(c, domainErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    authResp,
		"message": "Login successful. Consider generating new recovery codes.",
	})
}

// GetRecoveryCodeCount handles GET /v1/auth/2fa/recovery-codes/count
func (h *AuthHandler) GetRecoveryCodeCount(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Unauthorized", nil))
		return
	}

	count, err := h.authService.GetRecoveryCodeCount(c.Request.Context(), userID)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			errorResponse(c, domainErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, "Failed to get recovery code count", nil))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"remaining_codes": count,
		},
	})
}

// ChangePassword handles POST /v1/auth/password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Unauthorized", nil))
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Invalid request: "+err.Error(), nil))
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			errorResponse(c, domainErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, "Failed to change password", nil))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

// LoginOIDC initiates the OIDC login flow
func (h *AuthHandler) LoginOIDC(c *gin.Context) {
	if h.oidcService == nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "OIDC not configured", nil))
		return
	}

	state, err := generateSecureState(32)
	if err != nil {
		errorResponse(c, domain.NewError(domain.ErrInternalServer, "Failed to start OIDC login", nil))
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"oidc_state",
		state,
		int(5*time.Minute/time.Second),
		"/v1/auth/oidc",
		"",
		c.Request.TLS != nil, // Secure only when TLS is present
		true,                 // HttpOnly
	)

	url := h.oidcService.GetLoginURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// CallbackOIDC handles the OIDC callback
func (h *AuthHandler) CallbackOIDC(c *gin.Context) {
	if h.oidcService == nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "OIDC not configured", nil))
		return
	}

	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Missing code or state", nil))
		return
	}

	expectedState, err := c.Cookie("oidc_state")
	if err != nil || expectedState == "" || expectedState != state {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Invalid OIDC state", nil))
		return
	}

	// Clear state cookie after use
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("oidc_state", "", -1, "/v1/auth/oidc", "", c.Request.TLS != nil, true)

	_, userInfo, err := h.oidcService.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidCredentials, "Failed to exchange token", map[string]string{"error": err.Error()}))
		return
	}

	// Login or Register user via AuthService
	authResponse, err := h.authService.LoginOrRegisterOIDC(c.Request.Context(), userInfo.Email, userInfo.Sub, "oidc")
	if err != nil {
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:3000"
		}
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=oidc_failed", frontendURL))
		return
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/auth/callback?access_token=%s&refresh_token=%s",
		frontendURL, authResponse.AccessToken, authResponse.RefreshToken))
}

// Me returns the current authenticated user
func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Unauthorized", nil))
		return
	}

	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		errorResponse(c, domain.NewError(domain.ErrUserNotFound, "User not found", nil))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}

// UpdateProfile handles PUT /v1/users/me
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		errorResponse(c, domain.NewError(domain.ErrUnauthorized, "Unauthorized", nil))
		return
	}

	var req struct {
		FullName  *string `json:"full_name,omitempty"`
		Bio       *string `json:"bio,omitempty"`
		AvatarURL *string `json:"avatar_url,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidRequest, "Invalid request: "+err.Error(), nil))
		return
	}

	// Update user profile via auth service
	if err := h.authService.UpdateUserProfile(c.Request.Context(), userID, req.FullName, req.Bio, req.AvatarURL); err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			errorResponse(c, domainErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, "Failed to update profile", nil))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
	})
}

func generateSecureState(length int) (string, error) {
	if length <= 0 {
		length = 32
	}
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// RegisterAuthRoutes registers authentication routes
func RegisterAuthRoutes(r *gin.Engine, handler *AuthHandler, authMiddleware gin.HandlerFunc) {
	v1 := r.Group("/v1")
	{
		auth := v1.Group("/auth")
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
		auth.POST("/refresh", handler.Refresh)
		auth.POST("/logout", handler.Logout)
		auth.GET("/oidc/login", handler.LoginOIDC)
		auth.GET("/oidc/callback", handler.CallbackOIDC)
		// Recovery code login (no auth required)
		auth.POST("/2fa/recovery", handler.UseRecoveryCode)
	}

	authProtected := v1.Group("/auth")
	authProtected.Use(authMiddleware)
	{
		authProtected.POST("/password", handler.ChangePassword)
		authProtected.GET("/me", handler.Me)
		authProtected.POST("/2fa/generate", handler.Generate2FA)
		authProtected.POST("/2fa/enable", handler.Enable2FA)
		authProtected.POST("/2fa/disable", handler.Disable2FA)
		// Recovery codes (auth required)
		authProtected.POST("/2fa/recovery-codes", handler.GenerateRecoveryCodes)
		authProtected.GET("/2fa/recovery-codes/count", handler.GetRecoveryCodeCount)
	}
}
