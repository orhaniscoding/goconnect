package service

import (
	"context"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

// ══════════════════════════════════════════════════════════════════════════════
// SECTION SERVICE
// ══════════════════════════════════════════════════════════════════════════════
// Business logic layer for section management with RBAC integration

type SectionService struct {
	sectionRepo        repository.SectionRepository
	permissionResolver *PermissionResolver
}

// NewSectionService creates a new section service
func NewSectionService(
	sectionRepo repository.SectionRepository,
	permissionResolver *PermissionResolver,
) *SectionService {
	return &SectionService{
		sectionRepo:        sectionRepo,
		permissionResolver: permissionResolver,
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// CREATE
// ══════════════════════════════════════════════════════════════════════════════

// CreateSectionInput represents input for creating a section
type CreateSectionInput struct {
	TenantID    string
	UserID      string
	Name        string
	Description string
	Icon        string
	Visibility  domain.SectionVisibility
}

// Create creates a new section
// Permissions: Requires MANAGE_SECTIONS permission
func (s *SectionService) Create(ctx context.Context, input CreateSectionInput) (*domain.Section, error) {
	// Step 1: Check permissions
	permResult, err := s.permissionResolver.CheckPermission(
		ctx,
		input.UserID,
		input.TenantID,
		"", // No channel context
		domain.PermissionManageSections,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}

	if !permResult.Allowed {
		return nil, domain.NewError(domain.ErrForbidden, "insufficient permissions to create section", map[string]any{
			"user_id":    input.UserID,
			"tenant_id":  input.TenantID,
			"permission": domain.PermissionManageSections,
			"reason":     permResult.Reason,
		})
	}

	// Step 2: Validate input
	validatedName, err := domain.ValidateSectionName(input.Name)
	if err != nil {
		return nil, err
	}

	if len(input.Description) > 500 {
		return nil, domain.NewError(domain.ErrValidation, "description too long", map[string]any{
			"field": "description",
			"max":   500,
		})
	}

	// Step 3: Apply defaults
	if input.Visibility == "" {
		input.Visibility = domain.SectionVisibilityVisible
	}

	// Step 4: Get current sections to determine position
	existingSections, err := s.sectionRepo.GetByTenantID(ctx, input.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing sections: %w", err)
	}

	// New section goes to the end
	position := len(existingSections)

	// Step 5: Create section entity
	section := &domain.Section{
		ID:          domain.GenerateSectionID(),
		TenantID:    input.TenantID,
		Name:        validatedName,
		Description: input.Description,
		Icon:        input.Icon,
		Position:    position,
		Visibility:  input.Visibility,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Step 6: Persist to database
	if err := s.sectionRepo.Create(ctx, section); err != nil {
		return nil, fmt.Errorf("failed to create section: %w", err)
	}

	return section, nil
}

// ══════════════════════════════════════════════════════════════════════════════
// READ
// ══════════════════════════════════════════════════════════════════════════════

// GetByID retrieves a section by ID
// Permissions: User must be member of tenant (basic check)
func (s *SectionService) GetByID(ctx context.Context, userID, sectionID string) (*domain.Section, error) {
	// Get section
	section, err := s.sectionRepo.GetByID(ctx, sectionID)
	if err != nil {
		return nil, err
	}

	// Basic permission check: User must have access to tenant
	// (Could be enhanced with more granular checks)
	permResult, err := s.permissionResolver.CheckPermission(
		ctx,
		userID,
		section.TenantID,
		"",
		domain.PermissionViewChannels, // Basic view permission
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}

	if !permResult.Allowed {
		return nil, domain.NewError(domain.ErrForbidden, "insufficient permissions to view section", map[string]any{
			"user_id":    userID,
			"section_id": sectionID,
		})
	}

	return section, nil
}

// ListByTenant retrieves all sections for a tenant
// Permissions: User must be member of tenant
func (s *SectionService) ListByTenant(ctx context.Context, userID, tenantID string) ([]domain.Section, error) {
	// Check permissions
	permResult, err := s.permissionResolver.CheckPermission(
		ctx,
		userID,
		tenantID,
		"",
		domain.PermissionViewChannels,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}

	if !permResult.Allowed {
		return nil, domain.NewError(domain.ErrForbidden, "insufficient permissions to view sections", map[string]any{
			"user_id":   userID,
			"tenant_id": tenantID,
		})
	}

	// Get sections
	sections, err := s.sectionRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sections: %w", err)
	}

	return sections, nil
}

// ══════════════════════════════════════════════════════════════════════════════
// UPDATE
// ══════════════════════════════════════════════════════════════════════════════

// UpdateSectionInput represents input for updating a section
type UpdateSectionInput struct {
	UserID      string
	SectionID   string
	Name        *string
	Description *string
	Icon        *string
	Visibility  *domain.SectionVisibility
}

// Update updates a section
// Permissions: Requires MANAGE_SECTIONS permission
func (s *SectionService) Update(ctx context.Context, input UpdateSectionInput) (*domain.Section, error) {
	// Step 1: Get existing section
	section, err := s.sectionRepo.GetByID(ctx, input.SectionID)
	if err != nil {
		return nil, err
	}

	// Step 2: Check permissions
	permResult, err := s.permissionResolver.CheckPermission(
		ctx,
		input.UserID,
		section.TenantID,
		"",
		domain.PermissionManageSections,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}

	if !permResult.Allowed {
		return nil, domain.NewError(domain.ErrForbidden, "insufficient permissions to update section", map[string]any{
			"user_id":    input.UserID,
			"section_id": input.SectionID,
			"reason":     permResult.Reason,
		})
	}

	// Step 3: Apply updates
	if input.Name != nil {
		validatedName, err := domain.ValidateSectionName(*input.Name)
		if err != nil {
			return nil, err
		}
		section.Name = validatedName
	}

	if input.Description != nil {
		if len(*input.Description) > 500 {
			return nil, domain.NewError(domain.ErrValidation, "description too long", map[string]any{
				"field": "description",
				"max":   500,
			})
		}
		section.Description = *input.Description
	}

	if input.Icon != nil {
		section.Icon = *input.Icon
	}

	if input.Visibility != nil {
		section.Visibility = *input.Visibility
	}

	section.UpdatedAt = time.Now()

	// Step 4: Persist changes
	if err := s.sectionRepo.Update(ctx, section); err != nil {
		return nil, fmt.Errorf("failed to update section: %w", err)
	}

	return section, nil
}

// ══════════════════════════════════════════════════════════════════════════════
// DELETE
// ══════════════════════════════════════════════════════════════════════════════

// Delete deletes a section
// Permissions: Requires MANAGE_SECTIONS permission
func (s *SectionService) Delete(ctx context.Context, userID, sectionID string) error {
	// Step 1: Get existing section
	section, err := s.sectionRepo.GetByID(ctx, sectionID)
	if err != nil {
		return err
	}

	// Step 2: Check permissions
	permResult, err := s.permissionResolver.CheckPermission(
		ctx,
		userID,
		section.TenantID,
		"",
		domain.PermissionManageSections,
	)
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}

	if !permResult.Allowed {
		return domain.NewError(domain.ErrForbidden, "insufficient permissions to delete section", map[string]any{
			"user_id":    userID,
			"section_id": sectionID,
			"reason":     permResult.Reason,
		})
	}

	// Step 3: Delete section (soft delete)
	if err := s.sectionRepo.Delete(ctx, sectionID); err != nil {
		return fmt.Errorf("failed to delete section: %w", err)
	}

	return nil
}

// ══════════════════════════════════════════════════════════════════════════════
// POSITION MANAGEMENT
// ══════════════════════════════════════════════════════════════════════════════

// UpdatePositions updates positions for multiple sections
// Permissions: Requires MANAGE_SECTIONS permission
func (s *SectionService) UpdatePositions(ctx context.Context, userID, tenantID string, positions map[string]int) error {
	// Step 1: Check permissions
	permResult, err := s.permissionResolver.CheckPermission(
		ctx,
		userID,
		tenantID,
		"",
		domain.PermissionManageSections,
	)
	if err != nil {
		return fmt.Errorf("failed to check permissions: %w", err)
	}

	if !permResult.Allowed {
		return domain.NewError(domain.ErrForbidden, "insufficient permissions to reorder sections", map[string]any{
			"user_id":   userID,
			"tenant_id": tenantID,
			"reason":    permResult.Reason,
		})
	}

	// Step 2: Validate positions (must be non-negative)
	for sectionID, position := range positions {
		if position < 0 {
			return domain.NewError(domain.ErrValidation, "invalid position", map[string]any{
				"section_id": sectionID,
				"position":   position,
			})
		}
	}

	// Step 3: Update positions atomically
	if err := s.sectionRepo.UpdatePositions(ctx, tenantID, positions); err != nil {
		return fmt.Errorf("failed to update positions: %w", err)
	}

	return nil
}
