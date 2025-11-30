package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// SQLiteTenantAnnouncementRepository implements TenantAnnouncementRepository using SQLite.
type SQLiteTenantAnnouncementRepository struct {
	db *sql.DB
}

func NewSQLiteTenantAnnouncementRepository(db *sql.DB) *SQLiteTenantAnnouncementRepository {
	return &SQLiteTenantAnnouncementRepository{db: db}
}

func (r *SQLiteTenantAnnouncementRepository) Create(ctx context.Context, ann *domain.TenantAnnouncement) error {
	if ann.ID == "" {
		ann.ID = domain.GenerateAnnouncementID()
	}
	now := time.Now()
	if ann.CreatedAt.IsZero() {
		ann.CreatedAt = now
	}
	ann.UpdatedAt = now
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tenant_announcements (id, tenant_id, title, content, author_id, is_pinned, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, ann.ID, ann.TenantID, ann.Title, ann.Content, ann.AuthorID, ann.IsPinned, ann.CreatedAt, ann.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create tenant announcement: %w", err)
	}
	return nil
}

func (r *SQLiteTenantAnnouncementRepository) GetByID(ctx context.Context, id string) (*domain.TenantAnnouncement, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, title, content, author_id, is_pinned, created_at, updated_at
		FROM tenant_announcements
		WHERE id = ?
	`, id)
	var ann domain.TenantAnnouncement
	if err := row.Scan(
		&ann.ID, &ann.TenantID, &ann.Title, &ann.Content, &ann.AuthorID,
		&ann.IsPinned, &ann.CreatedAt, &ann.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewError(domain.ErrNotFound, "Announcement not found", nil)
		}
		return nil, fmt.Errorf("failed to get announcement: %w", err)
	}
	return &ann, nil
}

func (r *SQLiteTenantAnnouncementRepository) Update(ctx context.Context, ann *domain.TenantAnnouncement) error {
	ann.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx, `
		UPDATE tenant_announcements
		SET title = ?, content = ?, is_pinned = ?, updated_at = ?
		WHERE id = ?
	`, ann.Title, ann.Content, ann.IsPinned, ann.UpdatedAt, ann.ID)
	if err != nil {
		return fmt.Errorf("failed to update announcement: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Announcement not found", nil)
	}
	return nil
}

func (r *SQLiteTenantAnnouncementRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM tenant_announcements WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete announcement: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Announcement not found", nil)
	}
	return nil
}

func (r *SQLiteTenantAnnouncementRepository) ListByTenant(ctx context.Context, tenantID string, pinnedOnly bool, limit int, cursor string) ([]*domain.TenantAnnouncement, string, error) {
	args := []interface{}{tenantID}
	query := `
		SELECT id, tenant_id, title, content, author_id, is_pinned, created_at, updated_at
		FROM tenant_announcements
		WHERE tenant_id = ?
	`
	if pinnedOnly {
		query += " AND is_pinned = 1"
	}
	if cursor != "" {
		query += " AND id > ?"
		args = append(args, cursor)
	}
	if limit <= 0 {
		limit = 20
	}
	query += " ORDER BY created_at DESC, id DESC LIMIT ?"
	args = append(args, limit+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list announcements: %w", err)
	}
	defer rows.Close()

	var out []*domain.TenantAnnouncement
	for rows.Next() {
		var ann domain.TenantAnnouncement
		if err := rows.Scan(
			&ann.ID, &ann.TenantID, &ann.Title, &ann.Content, &ann.AuthorID,
			&ann.IsPinned, &ann.CreatedAt, &ann.UpdatedAt,
		); err != nil {
			return nil, "", fmt.Errorf("failed to scan announcement: %w", err)
		}
		out = append(out, &ann)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("failed to iterate announcements: %w", err)
	}
	next := ""
	if len(out) > limit {
		next = out[limit].ID
		out = out[:limit]
	}
	return out, next, nil
}

func (r *SQLiteTenantAnnouncementRepository) SetPinned(ctx context.Context, id string, pinned bool) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE tenant_announcements
		SET is_pinned = ?, updated_at = ?
		WHERE id = ?
	`, pinned, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to set pinned: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Announcement not found", nil)
	}
	return nil
}

func (r *SQLiteTenantAnnouncementRepository) DeleteAllByTenant(ctx context.Context, tenantID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM tenant_announcements WHERE tenant_id = ?`, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete tenant announcements: %w", err)
	}
	return nil
}
