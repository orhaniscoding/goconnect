package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/service"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService *service.AuthService
	oidcService *service.OIDCService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService, oidcService *service.OIDCService) *AuthHandler {
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

	// TODO: Generate random state and store in cookie/redis
	state := "random-state-to-be-implemented"

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

	// TODO: Validate state

	_, userInfo, err := h.oidcService.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		errorResponse(c, domain.NewError(domain.ErrInvalidCredentials, "Failed to exchange token", map[string]string{"error": err.Error()}))
		return
	}

	// Login or Register user via AuthService
	authResponse, err := h.authService.LoginOrRegisterOIDC(c.Request.Context(), userInfo.Email, userInfo.Sub, "oidc")
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			errorResponse(c, domainErr)
		} else {
			errorResponse(c, domain.NewError(domain.ErrInternalServer, "Internal server error", nil))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":   true,
		"data": authResponse,
	})
}

// RegisterAuthRoutes registers authentication routes
func RegisterAuthRoutes(r *gin.Engine, handler *AuthHandler, authMiddleware gin.HandlerFunc) {
	v1 := r.Group("/v1")
	{
		auth := v1.Group("/auth")
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
		auth.POST("/refresh", handler.Refresh)
		auth.GET("/oidc/login", handler.LoginOIDC)
		auth.GET("/oidc/callback", handler.CallbackOIDC)
	}

	authProtected := v1.Group("/auth")
	authProtected.Use(authMiddleware)
	{
		authProtected.POST("/password", handler.ChangePassword)
		authProtected.POST("/2fa/generate", handler.Generate2FA)
		authProtected.POST("/2fa/enable", handler.Enable2FA)
		authProtected.POST("/2fa/disable", handler.Disable2FA)
	}
}
