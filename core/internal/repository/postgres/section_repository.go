package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/database"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/repository"
)

// ══════════════════════════════════════════════════════════════════════════════
// SECTION REPOSITORY POSTGRES IMPLEMENTATION
// ══════════════════════════════════════════════════════════════════════════════

type SectionRepositoryPostgres struct {
	db *sql.DB
}

// NewSectionRepository creates a new PostgreSQL section repository
func NewSectionRepository(db *sql.DB) repository.SectionRepository {
	return &SectionRepositoryPostgres{db: db}
}

// ══════════════════════════════════════════════════════════════════════════════
// CREATE
// ══════════════════════════════════════════════════════════════════════════════

func (r *SectionRepositoryPostgres) Create(ctx context.Context, section *domain.Section) error {
	query := `
		INSERT INTO sections (
			id, tenant_id, name, description, icon, position, visibility,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`

	now := time.Now()
	section.CreatedAt = now
	section.UpdatedAt = now

	_, err := r.db.ExecContext(
		ctx,
		query,
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

// ══════════════════════════════════════════════════════════════════════════════
// READ
// ══════════════════════════════════════════════════════════════════════════════

func (r *SectionRepositoryPostgres) GetByID(ctx context.Context, id string) (*domain.Section, error) {
	query := `
		SELECT
			id, tenant_id, name, description, icon, position, visibility,
			created_at, updated_at, deleted_at
		FROM sections
		WHERE id = $1 AND deleted_at IS NULL
	`

	var section domain.Section
	row := r.db.QueryRowContext(ctx, query, id)
	err := database.ScanRow(row, &section)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewError(domain.ErrNotFound, "section not found", map[string]any{
				"section_id": id,
			})
		}
		return nil, fmt.Errorf("failed to get section: %w", err)
	}

	return &section, nil
}

func (r *SectionRepositoryPostgres) GetByTenantID(ctx context.Context, tenantID string) ([]domain.Section, error) {
	query := `
		SELECT
			id, tenant_id, name, description, icon, position, visibility,
			created_at, updated_at, deleted_at
		FROM sections
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY position ASC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sections: %w", err)
	}
	defer rows.Close()

	var sections []domain.Section
	if err := database.ScanRows(rows, &sections); err != nil {
		return nil, fmt.Errorf("failed to scan sections: %w", err)
	}

	// Return empty slice instead of nil
	if sections == nil {
		sections = []domain.Section{}
	}

	return sections, nil
}

// ══════════════════════════════════════════════════════════════════════════════
// UPDATE
// ══════════════════════════════════════════════════════════════════════════════

func (r *SectionRepositoryPostgres) Update(ctx context.Context, section *domain.Section) error {
	query := `
		UPDATE sections
		SET
			name = $2,
			description = $3,
			icon = $4,
			position = $5,
			visibility = $6,
			updated_at = $7
		WHERE id = $1 AND deleted_at IS NULL
	`

	section.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
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
		return domain.NewError(domain.ErrNotFound, "section not found or already deleted", map[string]any{
			"section_id": section.ID,
		})
	}

	return nil
}

func (r *SectionRepositoryPostgres) UpdatePositions(ctx context.Context, tenantID string, positions map[string]int) error {
	// Start transaction for atomic batch update
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
		result, err := tx.ExecContext(ctx, query, sectionID, position, now, tenantID)
		if err != nil {
			return fmt.Errorf("failed to update section position: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return domain.NewError(domain.ErrNotFound, "section not found in tenant", map[string]any{
				"section_id": sectionID,
				"tenant_id":  tenantID,
			})
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ══════════════════════════════════════════════════════════════════════════════
// DELETE (Soft Delete)
// ══════════════════════════════════════════════════════════════════════════════

func (r *SectionRepositoryPostgres) Delete(ctx context.Context, id string) error {
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
		return domain.NewError(domain.ErrNotFound, "section not found or already deleted", map[string]any{
			"section_id": id,
		})
	}

	return nil
}
