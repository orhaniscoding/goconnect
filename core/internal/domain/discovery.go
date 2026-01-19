package domain

import (
	"regexp"
	"strings"
	"time"
)

// ═══════════════════════════════════════════════════════════════════════════
// SERVER DISCOVERY (Public server listing)
// ═══════════════════════════════════════════════════════════════════════════

// DiscoveryCategory represents server categories
type DiscoveryCategory string

const (
	DiscoveryCategoryGaming        DiscoveryCategory = "gaming"
	DiscoveryCategoryEducation     DiscoveryCategory = "education"
	DiscoveryCategoryMusic         DiscoveryCategory = "music"
	DiscoveryCategoryTech          DiscoveryCategory = "tech"
	DiscoveryCategoryArt           DiscoveryCategory = "art"
	DiscoveryCategoryScience       DiscoveryCategory = "science"
	DiscoveryCategoryEntertainment DiscoveryCategory = "entertainment"
	DiscoveryCategorySports        DiscoveryCategory = "sports"
	DiscoveryCategoryFinance       DiscoveryCategory = "finance"
	DiscoveryCategoryCrypto        DiscoveryCategory = "crypto"
	DiscoveryCategorySocial        DiscoveryCategory = "social"
	DiscoveryCategoryCommunity     DiscoveryCategory = "community"
	DiscoveryCategoryOther         DiscoveryCategory = "other"
)

// AllDiscoveryCategories returns all valid categories
func AllDiscoveryCategories() []DiscoveryCategory {
	return []DiscoveryCategory{
		DiscoveryCategoryGaming,
		DiscoveryCategoryEducation,
		DiscoveryCategoryMusic,
		DiscoveryCategoryTech,
		DiscoveryCategoryArt,
		DiscoveryCategoryScience,
		DiscoveryCategoryEntertainment,
		DiscoveryCategorySports,
		DiscoveryCategoryFinance,
		DiscoveryCategoryCrypto,
		DiscoveryCategorySocial,
		DiscoveryCategoryCommunity,
		DiscoveryCategoryOther,
	}
}

// ServerDiscovery represents discovery settings for a server
type ServerDiscovery struct {
	TenantID         string             `json:"tenant_id" db:"tenant_id"`
	Enabled          bool               `json:"enabled" db:"enabled"`
	Category         *DiscoveryCategory `json:"category,omitempty" db:"category"`
	Tags             []string           `json:"tags" db:"tags"`
	ShortDescription *string            `json:"short_description,omitempty" db:"short_description"`
	MemberCount      int                `json:"member_count" db:"member_count"`
	OnlineCount      int                `json:"online_count" db:"online_count"`
	Featured         bool               `json:"featured" db:"featured"`
	Verified         bool               `json:"verified" db:"verified"`
	UpdatedAt        time.Time          `json:"updated_at" db:"updated_at"`

	// Enriched fields
	Tenant *Tenant `json:"tenant,omitempty" db:"-"`
}

// IsDiscoverable checks if the server is publicly discoverable
func (d *ServerDiscovery) IsDiscoverable() bool {
	return d.Enabled
}

// UpdateDiscoveryRequest for updating discovery settings
type UpdateDiscoveryRequest struct {
	Enabled          *bool              `json:"enabled,omitempty"`
	Category         *DiscoveryCategory `json:"category,omitempty"`
	Tags             []string           `json:"tags,omitempty" binding:"max=5,dive,max=50"`
	ShortDescription *string            `json:"short_description,omitempty" binding:"max=300"`
}

// SearchDiscoveryRequest for searching discoverable servers
type SearchDiscoveryRequest struct {
	Query    string             `form:"q"`
	Category *DiscoveryCategory `form:"category"`
	Tags     []string           `form:"tags"`
	Featured *bool              `form:"featured"`
	Verified *bool              `form:"verified"`
	Sort     string             `form:"sort,default=member_count"` // member_count, online_count, created_at
	Limit    int                `form:"limit,default=20"`
	Cursor   string             `form:"cursor"`
}

// ═══════════════════════════════════════════════════════════════════════════
// SERVER VANITY URL (Custom short URLs)
// ═══════════════════════════════════════════════════════════════════════════

// ServerVanityURL represents a custom URL for a server
type ServerVanityURL struct {
	TenantID   string    `json:"tenant_id" db:"tenant_id"`
	VanityCode string    `json:"vanity_code" db:"vanity_code"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// SetVanityURLRequest for setting a vanity URL
type SetVanityURLRequest struct {
	VanityCode string `json:"vanity_code" binding:"required,min=4,max=32"`
}

// ValidateVanityCode validates a vanity code
func ValidateVanityCode(code string) (string, error) {
	code = strings.ToLower(strings.TrimSpace(code))

	if len(code) < 4 {
		return "", NewError(ErrValidation, "vanity code must be at least 4 characters", map[string]any{
			"field": "vanity_code",
			"min":   4,
		})
	}

	if len(code) > 32 {
		return "", NewError(ErrValidation, "vanity code must be at most 32 characters", map[string]any{
			"field": "vanity_code",
			"max":   32,
		})
	}

	// Must start and end with alphanumeric, can contain hyphens
	validPattern := regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)
	if !validPattern.MatchString(code) {
		return "", NewError(ErrValidation, "vanity code must start and end with a letter or number, and can only contain letters, numbers, and hyphens", map[string]any{
			"field": "vanity_code",
		})
	}

	// No consecutive hyphens
	if strings.Contains(code, "--") {
		return "", NewError(ErrValidation, "vanity code cannot contain consecutive hyphens", map[string]any{
			"field": "vanity_code",
		})
	}

	return code, nil
}

// ═══════════════════════════════════════════════════════════════════════════
// DISCOVERY RESPONSE TYPES
// ═══════════════════════════════════════════════════════════════════════════

// DiscoveryServerResponse is the public view of a discoverable server
type DiscoveryServerResponse struct {
	ID               string             `json:"id"`
	Name             string             `json:"name"`
	Description      string             `json:"description,omitempty"`
	Icon             *string            `json:"icon,omitempty"`
	Banner           *string            `json:"banner,omitempty"`
	Category         *DiscoveryCategory `json:"category,omitempty"`
	Tags             []string           `json:"tags"`
	ShortDescription *string            `json:"short_description,omitempty"`
	MemberCount      int                `json:"member_count"`
	OnlineCount      int                `json:"online_count"`
	Featured         bool               `json:"featured"`
	Verified         bool               `json:"verified"`
	VanityURL        *string            `json:"vanity_url,omitempty"`
}
