package domain

import (
	"fmt"
	"net"
	"time"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"crypto/rand"
	"math/big"
)

// Network represents a virtual network in GoConnect
type Network struct {
	ID                   string            `json:"id" db:"id"`
	TenantID            string            `json:"tenant_id" db:"tenant_id"`
	Name                string            `json:"name" db:"name"`
	Visibility          NetworkVisibility `json:"visibility" db:"visibility"`
	JoinPolicy          JoinPolicy        `json:"join_policy" db:"join_policy"`
	CIDR                string            `json:"cidr" db:"cidr"`
	DNS                 *string           `json:"dns,omitempty" db:"dns"`
	MTU                 *int              `json:"mtu,omitempty" db:"mtu"`
	SplitTunnel         *bool             `json:"split_tunnel,omitempty" db:"split_tunnel"`
	CreatedBy           string            `json:"created_by" db:"created_by"`
	CreatedAt           time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at" db:"updated_at"`
	SoftDeletedAt       *time.Time        `json:"soft_deleted_at,omitempty" db:"soft_deleted_at"`
	ModerationRedacted  bool              `json:"moderation_redacted" db:"moderation_redacted"`
}

// NetworkVisibility defines network visibility options
type NetworkVisibility string

const (
	NetworkVisibilityPublic  NetworkVisibility = "public"
	NetworkVisibilityPrivate NetworkVisibility = "private"
)

// JoinPolicy defines how users can join the network
type JoinPolicy string

const (
	JoinPolicyOpen     JoinPolicy = "open"
	JoinPolicyInvite   JoinPolicy = "invite"
	JoinPolicyApproval JoinPolicy = "approval"
)

// CreateNetworkRequest represents the request to create a network
type CreateNetworkRequest struct {
	Name        string            `json:"name" binding:"required,min=3,max=64"`
	Visibility  NetworkVisibility `json:"visibility" binding:"required,oneof=public private"`
	JoinPolicy  JoinPolicy        `json:"join_policy" binding:"required,oneof=open invite approval"`
	CIDR        string            `json:"cidr" binding:"required"`
	DNS         *string           `json:"dns,omitempty"`
	MTU         *int              `json:"mtu,omitempty"`
	SplitTunnel *bool             `json:"split_tunnel,omitempty"`
}

// ListNetworksRequest represents query parameters for listing networks
type ListNetworksRequest struct {
	Visibility string `form:"visibility,default=public"` // public|mine|all
	Limit      int    `form:"limit,default=20"`
	Cursor     string `form:"cursor"`
	Search     string `form:"search"`
}

// ValidateCIDR validates if the CIDR is a valid network address
func ValidateCIDR(cidr string) error {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR format: %w", err)
	}
	
	// Ensure it's a network address, not a host address
	if network.String() != cidr {
		return fmt.Errorf("CIDR must be a network address, got %s but expected %s", cidr, network.String())
	}
	
	return nil
}

// CheckCIDROverlap checks if two CIDR blocks overlap
func CheckCIDROverlap(cidr1, cidr2 string) (bool, error) {
	_, net1, err := net.ParseCIDR(cidr1)
	if err != nil {
		return false, fmt.Errorf("invalid CIDR1: %w", err)
	}
	
	_, net2, err := net.ParseCIDR(cidr2)
	if err != nil {
		return false, fmt.Errorf("invalid CIDR2: %w", err)
	}
	
	// Check if either network contains the other's network address
	return net1.Contains(net2.IP) || net2.Contains(net1.IP), nil
}

// GenerateNetworkID generates a new ID for networks (simplified ULID-like)
func GenerateNetworkID() string {
	// Simple timestamp + random suffix for development
	timestamp := time.Now().Unix()
	n, _ := rand.Int(rand.Reader, big.NewInt(999999))
	return fmt.Sprintf("net_%d_%06d", timestamp, n.Int64())
}

// HashRequestBody creates MD5 hash of request body for idempotency
func HashRequestBody(body interface{}) (string, error) {
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(jsonBytes)
	return hex.EncodeToString(sum[:]), nil
}