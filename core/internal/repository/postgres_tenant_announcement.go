package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// PostgresTenantAnnouncementRepository implements TenantAnnouncementRepository using PostgreSQL
type PostgresTenantAnnouncementRepository struct {
	db *sql.DB
}

// NewPostgresTenantAnnouncementRepository creates a new PostgreSQL-backed repository
func NewPostgresTenantAnnouncementRepository(db *sql.DB) *PostgresTenantAnnouncementRepository {
	return &PostgresTenantAnnouncementRepository{db: db}
}

func (r *PostgresTenantAnnouncementRepository) Create(ctx context.Context, announcement *domain.TenantAnnouncement) error {
	if announcement.ID == "" {
		announcement.ID = domain.GenerateAnnouncementID()
	}

	query := `
		INSERT INTO tenant_announcements (id, tenant_id, title, content, author_id, is_pinned, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	now := time.Now()
	announcement.CreatedAt = now
	announcement.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		announcement.ID,
		announcement.TenantID,
		announcement.Title,
		announcement.Content,
		announcement.AuthorID,
		announcement.IsPinned,
		announcement.CreatedAt,
		announcement.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create announcement: %w", err)
	}
	return nil
}

func (r *PostgresTenantAnnouncementRepository) GetByID(ctx context.Context, id string) (*domain.TenantAnnouncement, error) {
	query := `
		SELECT a.id, a.tenant_id, a.title, a.content, a.author_id, a.is_pinned, a.created_at, a.updated_at,
		       u.email, u.locale
		FROM tenant_announcements a
		JOIN users u ON u.id = a.author_id
		WHERE a.id = $1
	`
	ann := &domain.TenantAnnouncement{
		Author: &domain.User{},
	}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ann.ID,
		&ann.TenantID,
		&ann.Title,
		&ann.Content,
		&ann.AuthorID,
		&ann.IsPinned,
		&ann.CreatedAt,
		&ann.UpdatedAt,
		&ann.Author.Email,
		&ann.Author.Locale,
	)
	if err == sql.ErrNoRows {
		return nil, domain.NewError(domain.ErrNotFound, "Announcement not found", nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get announcement: %w", err)
	}
	ann.Author.ID = ann.AuthorID
	return ann, nil
}

func (r *PostgresTenantAnnouncementRepository) Update(ctx context.Context, announcement *domain.TenantAnnouncement) error {
	query := `
		UPDATE tenant_announcements
		SET title = $2, content = $3, is_pinned = $4, updated_at = $5
		WHERE id = $1
	`
	announcement.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		announcement.ID,
		announcement.Title,
		announcement.Content,
		announcement.IsPinned,
		announcement.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update announcement: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Announcement not found", nil)
	}
	return nil
}

func (r *PostgresTenantAnnouncementRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tenant_announcements WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete announcement: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Announcement not found", nil)
	}
	return nil
}

func (r *PostgresTenantAnnouncementRepository) ListByTenant(ctx context.Context, tenantID string, pinnedOnly bool, limit int, cursor string) ([]*domain.TenantAnnouncement, string, error) {
	query := `
		SELECT a.id, a.tenant_id, a.title, a.content, a.author_id, a.is_pinned, a.created_at, a.updated_at,
		       u.email, u.locale
		FROM tenant_announcements a
		JOIN users u ON u.id = a.author_id
		WHERE a.tenant_id = $1
	`
	args := []interface{}{tenantID}
	argIdx := 2

	if pinnedOnly {
		query += " AND a.is_pinned = true"
	}

	if cursor != "" {
		query += fmt.Sprintf(" AND a.id < $%d", argIdx)
		args = append(args, cursor)
		// argIdx++ not needed as no more parameters after this
	}

	// Order by: pinned first, then by created_at desc
	query += " ORDER BY a.is_pinned DESC, a.created_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit+1)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list announcements: %w", err)
	}
	defer rows.Close()

	var announcements []*domain.TenantAnnouncement
	for rows.Next() {
		ann := &domain.TenantAnnouncement{
			Author: &domain.User{},
		}
		err := rows.Scan(
			&ann.ID,
			&ann.TenantID,
			&ann.Title,
			&ann.Content,
			&ann.AuthorID,
			&ann.IsPinned,
			&ann.CreatedAt,
			&ann.UpdatedAt,
			&ann.Author.Email,
			&ann.Author.Locale,
		)
		if err != nil {
			return nil, "", fmt.Errorf("failed to scan announcement: %w", err)
		}
		ann.Author.ID = ann.AuthorID
		announcements = append(announcements, ann)
	}

	nextCursor := ""
	if limit > 0 && len(announcements) > limit {
		nextCursor = announcements[limit-1].ID
		announcements = announcements[:limit]
	}

	return announcements, nextCursor, nil
}

func (r *PostgresTenantAnnouncementRepository) SetPinned(ctx context.Context, id string, pinned bool) error {
	query := `UPDATE tenant_announcements SET is_pinned = $2, updated_at = $3 WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id, pinned, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update pinned status: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Announcement not found", nil)
	}
	return nil
}

func (r *PostgresTenantAnnouncementRepository) DeleteAllByTenant(ctx context.Context, tenantID string) error {
	query := `DELETE FROM tenant_announcements WHERE tenant_id = $1`
	_, err := r.db.ExecContext(ctx, query, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete tenant announcements: %w", err)
	}
	return nil
}
