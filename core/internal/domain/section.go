package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ═══════════════════════════════════════════════════════════════════════════
// SECTION (Organizes channels within a tenant)
// ═══════════════════════════════════════════════════════════════════════════

// SectionVisibility defines section visibility options
type SectionVisibility string

const (
	SectionVisibilityVisible  SectionVisibility = "visible"
	SectionVisibilityHidden   SectionVisibility = "hidden"
	SectionVisibilityArchived SectionVisibility = "archived"
)

// Section represents a category/folder for organizing channels
type Section struct {
	ID          string            `json:"id" db:"id"`
	TenantID    string            `json:"tenant_id" db:"tenant_id"`
	Name        string            `json:"name" db:"name"`
	Description string            `json:"description,omitempty" db:"description"`
	Icon        string            `json:"icon,omitempty" db:"icon"`
	Position    int               `json:"position" db:"position"`
	Visibility  SectionVisibility `json:"visibility" db:"visibility"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time        `json:"deleted_at,omitempty" db:"deleted_at"`

	// Enriched fields
	Channels []Channel `json:"channels,omitempty" db:"-"`
}

// IsDeleted checks if the section is soft-deleted
func (s *Section) IsDeleted() bool {
	return s.DeletedAt != nil
}

// CreateSectionRequest for creating a new section
type CreateSectionRequest struct {
	Name        string            `json:"name" binding:"required,min=1,max=100"`
	Description string            `json:"description,omitempty" binding:"max=500"`
	Icon        string            `json:"icon,omitempty" binding:"max=255"`
	Visibility  SectionVisibility `json:"visibility,omitempty"`
}

// ApplyDefaults sets default values
func (r *CreateSectionRequest) ApplyDefaults() {
	if r.Visibility == "" {
		r.Visibility = SectionVisibilityVisible
	}
}

// UpdateSectionRequest for updating a section
type UpdateSectionRequest struct {
	Name        *string            `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Description *string            `json:"description,omitempty" binding:"omitempty,max=500"`
	Icon        *string            `json:"icon,omitempty" binding:"omitempty,max=255"`
	Position    *int               `json:"position,omitempty" binding:"omitempty,min=0"`
	Visibility  *SectionVisibility `json:"visibility,omitempty"`
}

// GenerateSectionID generates a new section ID
func GenerateSectionID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	return fmt.Sprintf("sec_%s", hex.EncodeToString(bytes))
}

// ValidateSectionName validates and sanitizes a section name
func ValidateSectionName(name string) (string, error) {
	name = strings.TrimSpace(name)

	// Collapse consecutive spaces
	spaceRegex := regexp.MustCompile(`\s+`)
	name = spaceRegex.ReplaceAllString(name, " ")

	if len(name) < 1 {
		return "", NewError(ErrValidation, "section name is required", map[string]any{
			"field": "name",
			"min":   1,
		})
	}

	if len(name) > 100 {
		return "", NewError(ErrValidation, "section name must be at most 100 characters", map[string]any{
			"field": "name",
			"max":   100,
		})
	}

	return name, nil
}
