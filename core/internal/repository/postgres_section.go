package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// ═══════════════════════════════════════════════════════════════════════════
// POSTGRES SECTION REPOSITORY
// ═══════════════════════════════════════════════════════════════════════════

// PostgresSectionRepository implements SectionRepository for PostgreSQL
type PostgresSectionRepository struct {
	db *sql.DB
}

// NewPostgresSectionRepository creates a new PostgresSectionRepository
func NewPostgresSectionRepository(db *sql.DB) *PostgresSectionRepository {
	return &PostgresSectionRepository{db: db}
}

// Create creates a new section
func (r *PostgresSectionRepository) Create(ctx context.Context, section *domain.Section) error {
	query := `
		INSERT INTO sections (id, tenant_id, name, description, icon, position, visibility, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	now := time.Now()
	if section.ID == "" {
		section.ID = domain.GenerateSectionID()
	}
	section.CreatedAt = now
	section.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		section.ID,
		section.TenantID,
		section.Name,
		section.Description,
		section.Icon,
		section.Position,
		section.Visibility,
		section.CreatedAt,
		section.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create section: %w", err)
	}

	return nil
}

// GetByID retrieves a section by ID
func (r *PostgresSectionRepository) GetByID(ctx context.Context, id string) (*domain.Section, error) {
	query := `
		SELECT id, tenant_id, name, description, icon, position, visibility, created_at, updated_at, deleted_at
		FROM sections
		WHERE id = $1 AND deleted_at IS NULL
	`

	var section domain.Section
	var description, icon sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&section.ID,
		&section.TenantID,
		&section.Name,
		&description,
		&icon,
		&section.Position,
		&section.Visibility,
		&section.CreatedAt,
		&section.UpdatedAt,
		&section.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get section: %w", err)
	}

	section.Description = description.String
	section.Icon = icon.String

	return &section, nil
}

// GetByTenantID retrieves all sections for a tenant
func (r *PostgresSectionRepository) GetByTenantID(ctx context.Context, tenantID string) ([]domain.Section, error) {
	query := `
		SELECT id, tenant_id, name, description, icon, position, visibility, created_at, updated_at, deleted_at
		FROM sections
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY position ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sections: %w", err)
	}
	defer rows.Close()

	var sections []domain.Section
	for rows.Next() {
		var section domain.Section
		var description, icon sql.NullString

		if err := rows.Scan(
			&section.ID,
			&section.TenantID,
			&section.Name,
			&description,
			&icon,
			&section.Position,
			&section.Visibility,
			&section.CreatedAt,
			&section.UpdatedAt,
			&section.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan section: %w", err)
		}

		section.Description = description.String
		section.Icon = icon.String
		sections = append(sections, section)
	}

	return sections, nil
}

// Update updates a section
func (r *PostgresSectionRepository) Update(ctx context.Context, section *domain.Section) error {
	query := `
		UPDATE sections
		SET name = $2, description = $3, icon = $4, position = $5, visibility = $6, updated_at = $7
		WHERE id = $1 AND deleted_at IS NULL
	`

	section.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		section.ID,
		section.Name,
		section.Description,
		section.Icon,
		section.Position,
		section.Visibility,
		section.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update section: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("section not found")
	}

	return nil
}

// Delete soft-deletes a section
func (r *PostgresSectionRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE sections
		SET deleted_at = $2, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, id, now)
	if err != nil {
		return fmt.Errorf("failed to delete section: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("section not found")
	}

	return nil
}

// UpdatePositions updates positions for multiple sections
func (r *PostgresSectionRepository) UpdatePositions(ctx context.Context, tenantID string, positions map[string]int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		UPDATE sections
		SET position = $2, updated_at = $3
		WHERE id = $1 AND tenant_id = $4 AND deleted_at IS NULL
	`

	now := time.Now()
	for sectionID, position := range positions {
		_, err := tx.ExecContext(ctx, query, sectionID, position, now, tenantID)
		if err != nil {
			return fmt.Errorf("failed to update position for section %s: %w", sectionID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Ensure interface compliance
var _ SectionRepository = (*PostgresSectionRepository)(nil)
