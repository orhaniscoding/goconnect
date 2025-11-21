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
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
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

// RegisterAuthRoutes registers all auth-related routes
func RegisterAuthRoutes(r *gin.Engine, handler *AuthHandler, authMiddleware gin.HandlerFunc) {
	v1 := r.Group("/v1")
	v1.Use(RequestIDMiddleware())
	v1.Use(CORSMiddleware())

	auth := v1.Group("/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
		auth.POST("/refresh", handler.Refresh)
		auth.POST("/logout", handler.Logout)
	}

	authProtected := v1.Group("/auth")
	authProtected.Use(authMiddleware)
	{
		authProtected.POST("/2fa/generate", handler.Generate2FA)
		authProtected.POST("/2fa/enable", handler.Enable2FA)
		authProtected.POST("/2fa/disable", handler.Disable2FA)
	}
}
