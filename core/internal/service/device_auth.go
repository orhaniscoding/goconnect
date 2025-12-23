package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/redis/go-redis/v9"
)

const (
	// Device flow constants
	DeviceCodeTTL  = 10 * time.Minute
	DeviceFlowPoll = 5 // seconds
)

// DeviceFlowState represents the state stored in Redis
type DeviceFlowState struct {
	DeviceCode string `json:"device_code"`
	UserCode   string `json:"user_code"`
	ClientID   string `json:"client_id"`
	Status     string `json:"status"` // pending, approved
	UserID     string `json:"user_id,omitempty"`
}

// InitiateDeviceFlow starts the OAuth2 device flow
func (s *AuthService) InitiateDeviceFlow(ctx context.Context, clientID string) (*domain.DeviceCodeResponse, error) {
	if s.redisClient == nil {
		return nil, domain.NewError(domain.ErrInternalServer, "Redis required for device flow", nil)
	}

	// Generate codes
	deviceCode, err := generateRandomString(32)
	if err != nil {
		return nil, err
	}
	userCode, err := generateUserCode()
	if err != nil {
		return nil, err
	}

	state := DeviceFlowState{
		DeviceCode: deviceCode,
		UserCode:   userCode,
		ClientID:   clientID,
		Status:     "pending",
	}

	data, err := json.Marshal(state)
	if err != nil {
		return nil, err
	}

	// Store in Redis using pipeline
	pipe := s.redisClient.TxPipeline()
	// Key by DeviceCode for polling
	pipe.Set(ctx, "device_flow:device:"+deviceCode, data, DeviceCodeTTL)
	// Key by UserCode for approval
	pipe.Set(ctx, "device_flow:user:"+userCode, deviceCode, DeviceCodeTTL)
	
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("failed to store device flow state: %w", err)
	}

	// In a real deployment, VerificationURI would come from config
	// For now, assume it's constructed securely in the frontend
	verificationURI := "/auth/device/verify" 

	return &domain.DeviceCodeResponse{
		DeviceCode:      deviceCode,
		UserCode:        userCode,
		VerificationURI: verificationURI,
		ExpiresIn:       int(DeviceCodeTTL.Seconds()),
		Interval:        DeviceFlowPoll,
	}, nil
}

// ApproveDeviceFlow approves a pending device flow request
func (s *AuthService) ApproveDeviceFlow(ctx context.Context, userCode, userID string) error {
	if s.redisClient == nil {
		return domain.NewError(domain.ErrInternalServer, "Redis required", nil)
	}

	// 1. Look up device code by user code
	deviceCodeCmd := s.redisClient.Get(ctx, "device_flow:user:"+userCode)
	if errors.Is(deviceCodeCmd.Err(), redis.Nil) {
		return domain.NewError(domain.ErrInvalidRequest, "Invalid or expired user code", nil)
	} else if deviceCodeCmd.Err() != nil {
		return fmt.Errorf("redis error: %w", deviceCodeCmd.Err())
	}
	deviceCode := deviceCodeCmd.Val()

	// 2. Get state
	stateCmd := s.redisClient.Get(ctx, "device_flow:device:"+deviceCode)
	if errors.Is(stateCmd.Err(), redis.Nil) {
		return domain.NewError(domain.ErrInvalidRequest, "Session expired", nil)
	}
	
	var state DeviceFlowState
	if err := json.Unmarshal([]byte(stateCmd.Val()), &state); err != nil {
		return err
	}

	if state.Status != "pending" {
		return domain.NewError(domain.ErrInvalidRequest, "Request already processed", nil)
	}

	// 3. Update state to approved
	state.Status = "approved"
	state.UserID = userID
	
	updatedData, _ := json.Marshal(state)

	// Keep remaining TTL
	ttl := s.redisClient.TTL(ctx, "device_flow:device:"+deviceCode).Val()
	if ttl <= 0 {
		ttl = 5 * time.Minute // safety fallback
	}

	if err := s.redisClient.Set(ctx, "device_flow:device:"+deviceCode, updatedData, ttl).Err(); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}
	
	// Delete user code mapping (one-time use for lookup)
	// s.redisClient.Del(ctx, "device_flow:user:"+userCode)  -- Keeping it might be safer to prevent reuse attempts, or just let TTL expire.
	// Actually, better to remove it to prevent re-approval.
	s.redisClient.Del(ctx, "device_flow:user:"+userCode)

	return nil
}

// PollDeviceToken checks if the flow is approved and returns tokens
func (s *AuthService) PollDeviceToken(ctx context.Context, deviceCode string) (*domain.AuthResponse, error) {
	if s.redisClient == nil {
		return nil, domain.NewError(domain.ErrInternalServer, "Redis required", nil)
	}

	stateCmd := s.redisClient.Get(ctx, "device_flow:device:"+deviceCode)
	if errors.Is(stateCmd.Err(), redis.Nil) {
		// RFC 8628: "expired_token"
		return nil, domain.NewError(domain.ErrInvalidRequest, "expired_token", nil)
	} else if stateCmd.Err() != nil {
		return nil, fmt.Errorf("redis error: %w", stateCmd.Err())
	}

	var state DeviceFlowState
	if err := json.Unmarshal([]byte(stateCmd.Val()), &state); err != nil {
		return nil, err
	}

	if state.Status == "pending" {
		// RFC 8628: "authorization_pending"
		return nil, domain.NewError(domain.ErrUnauthorized, "authorization_pending", nil)
	}

	if state.Status != "approved" {
		return nil, domain.NewError(domain.ErrUnauthorized, "access_denied", nil)
	}

	// Approved! Log the user in.
	user, err := s.userRepo.GetByID(ctx, state.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user: %w", err)
	}

	// Generate Tokens
	accessToken, err := s.generateJWT(user.ID, user.TenantID, user.Email, user.IsAdmin, user.IsModerator, "access", 15*time.Minute)
	if err != nil {
		return nil, domain.NewError(domain.ErrInternalServer, "Failed to generate access token", nil)
	}

	refreshToken, err := s.generateJWT(user.ID, user.TenantID, user.Email, user.IsAdmin, user.IsModerator, "refresh", 7*24*time.Hour)
	if err != nil {
		return nil, domain.NewError(domain.ErrInternalServer, "Failed to generate refresh token", nil)
	}

	// Cleanup Redis state (one-time use)
	s.redisClient.Del(ctx, "device_flow:device:"+deviceCode)

	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900,
		TokenType:    "Bearer",
		User:         user,
	}, nil
}

// generateRandomString generates a secure random string of given length
func generateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}
	return string(ret), nil
}

// generateUserCode generates a short, readable code (e.g. ABCD-1234)
func generateUserCode() (string, error) {
	const charset = "BCDFGHJKLMNPQRSTUVWXYZ23456789" // Removed vowels/confusing chars
	
	// 8 chars: XXXX-XXXX
	ret := make([]byte, 8)
	for i := 0; i < 8; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		ret[i] = charset[num.Int64()]
	}
	return string(ret[:4]) + "-" + string(ret[4:]), nil
}
