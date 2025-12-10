package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPostgresTenantRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantRepository(db)
	require.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestPostgresTenantRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		tenant := &domain.Tenant{
			ID:        "tenant-123",
			Name:      "Test Tenant",
			CreatedAt: now,
			UpdatedAt: now,
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenants (id, name, created_at, updated_at)`)).
			WithArgs(tenant.ID, tenant.Name, tenant.CreatedAt, tenant.UpdatedAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, tenant)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		tenant := &domain.Tenant{
			ID:        "tenant-456",
			Name:      "Test Tenant 2",
			CreatedAt: now,
			UpdatedAt: now,
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenants (id, name, created_at, updated_at)`)).
			WithArgs(tenant.ID, tenant.Name, tenant.CreatedAt, tenant.UpdatedAt).
			WillReturnError(errors.New("database error"))

		err := repo.Create(ctx, tenant)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create tenant")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		tenantID := "tenant-123"
		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "icon_url", "visibility", "access_type",
			"password_hash", "max_members", "owner_id", "created_at", "updated_at",
		}).AddRow(
			tenantID, "Test Tenant", "A test tenant", "https://example.com/icon.png",
			domain.TenantVisibilityPublic, domain.TenantAccessOpen,
			"", 100, "owner-123", now, now,
		)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT 
			id,
			name,
			description,
			icon_url,
			visibility,
			access_type,
			password_hash,
			max_members,
			owner_id,
			created_at,
			updated_at
		FROM tenants
		WHERE id = $1`)).
			WithArgs(tenantID).
			WillReturnRows(rows)

		tenant, err := repo.GetByID(ctx, tenantID)
		require.NoError(t, err)
		assert.Equal(t, tenantID, tenant.ID)
		assert.Equal(t, "Test Tenant", tenant.Name)
		assert.Equal(t, "A test tenant", tenant.Description)
		assert.Equal(t, "https://example.com/icon.png", tenant.IconURL)
		assert.Equal(t, domain.TenantVisibilityPublic, tenant.Visibility)
		assert.Equal(t, domain.TenantAccessOpen, tenant.AccessType)
		assert.Equal(t, 100, tenant.MaxMembers)
		assert.Equal(t, "owner-123", tenant.OwnerID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with null fields", func(t *testing.T) {
		tenantID := "tenant-456"
		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "icon_url", "visibility", "access_type",
			"password_hash", "max_members", "owner_id", "created_at", "updated_at",
		}).AddRow(
			tenantID, "Minimal Tenant", nil, nil,
			domain.TenantVisibilityPrivate, domain.TenantAccessInviteOnly,
			nil, 50, nil, now, now,
		)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT 
			id,
			name,
			description,
			icon_url,
			visibility,
			access_type,
			password_hash,
			max_members,
			owner_id,
			created_at,
			updated_at
		FROM tenants
		WHERE id = $1`)).
			WithArgs(tenantID).
			WillReturnRows(rows)

		tenant, err := repo.GetByID(ctx, tenantID)
		require.NoError(t, err)
		assert.Equal(t, tenantID, tenant.ID)
		assert.Equal(t, "Minimal Tenant", tenant.Name)
		assert.Empty(t, tenant.Description)
		assert.Empty(t, tenant.IconURL)
		assert.Empty(t, tenant.PasswordHash)
		assert.Empty(t, tenant.OwnerID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		tenantID := "non-existent"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT 
			id,
			name,
			description,
			icon_url,
			visibility,
			access_type,
			password_hash,
			max_members,
			owner_id,
			created_at,
			updated_at
		FROM tenants
		WHERE id = $1`)).
			WithArgs(tenantID).
			WillReturnError(sql.ErrNoRows)

		tenant, err := repo.GetByID(ctx, tenantID)
		require.Error(t, err)
		assert.Nil(t, tenant)
		assert.Contains(t, err.Error(), domain.ErrNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		tenantID := "tenant-error"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT 
			id,
			name,
			description,
			icon_url,
			visibility,
			access_type,
			password_hash,
			max_members,
			owner_id,
			created_at,
			updated_at
		FROM tenants
		WHERE id = $1`)).
			WithArgs(tenantID).
			WillReturnError(errors.New("database error"))

		tenant, err := repo.GetByID(ctx, tenantID)
		require.Error(t, err)
		assert.Nil(t, tenant)
		assert.Contains(t, err.Error(), "failed to get tenant by ID")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		tenant := &domain.Tenant{
			ID:   "tenant-123",
			Name: "Updated Tenant",
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE tenants
		SET name = $1, updated_at = $2
		WHERE id = $3`)).
			WithArgs(tenant.Name, sqlmock.AnyArg(), tenant.ID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(ctx, tenant)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		tenant := &domain.Tenant{
			ID:   "non-existent",
			Name: "Updated Tenant",
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE tenants
		SET name = $1, updated_at = $2
		WHERE id = $3`)).
			WithArgs(tenant.Name, sqlmock.AnyArg(), tenant.ID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Update(ctx, tenant)
		require.Error(t, err)
		assert.Contains(t, err.Error(), domain.ErrNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		tenant := &domain.Tenant{
			ID:   "tenant-error",
			Name: "Updated Tenant",
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE tenants
		SET name = $1, updated_at = $2
		WHERE id = $3`)).
			WithArgs(tenant.Name, sqlmock.AnyArg(), tenant.ID).
			WillReturnError(errors.New("database error"))

		err := repo.Update(ctx, tenant)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update tenant")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		tenant := &domain.Tenant{
			ID:   "tenant-rows-error",
			Name: "Updated Tenant",
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE tenants
		SET name = $1, updated_at = $2
		WHERE id = $3`)).
			WithArgs(tenant.Name, sqlmock.AnyArg(), tenant.ID).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err := repo.Update(ctx, tenant)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get rows affected")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		tenantID := "tenant-123"

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenants WHERE id = $1`)).
			WithArgs(tenantID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(ctx, tenantID)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		tenantID := "non-existent"

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenants WHERE id = $1`)).
			WithArgs(tenantID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(ctx, tenantID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), domain.ErrNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		tenantID := "tenant-error"

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenants WHERE id = $1`)).
			WithArgs(tenantID).
			WillReturnError(errors.New("database error"))

		err := repo.Delete(ctx, tenantID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete tenant")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		tenantID := "tenant-rows-error"

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenants WHERE id = $1`)).
			WithArgs(tenantID).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err := repo.Delete(ctx, tenantID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get rows affected")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "icon_url", "visibility", "access_type",
			"password_hash", "max_members", "owner_id", "created_at", "updated_at",
		}).
			AddRow("tenant-1", "Tenant 1", "Description 1", "https://example.com/1.png",
				domain.TenantVisibilityPublic, domain.TenantAccessOpen, "", 100, "owner-1", now, now).
			AddRow("tenant-2", "Tenant 2", nil, nil,
				domain.TenantVisibilityPrivate, domain.TenantAccessInviteOnly, nil, 50, nil, now, now)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, description, icon_url, visibility, access_type, password_hash, max_members, owner_id, created_at, updated_at
		FROM tenants
		ORDER BY created_at DESC`)).
			WillReturnRows(rows)

		tenants, err := repo.List(ctx)
		require.NoError(t, err)
		assert.Len(t, tenants, 2)
		assert.Equal(t, "tenant-1", tenants[0].ID)
		assert.Equal(t, "Tenant 1", tenants[0].Name)
		assert.Equal(t, "Description 1", tenants[0].Description)
		assert.Equal(t, "tenant-2", tenants[1].ID)
		assert.Empty(t, tenants[1].Description)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "icon_url", "visibility", "access_type",
			"password_hash", "max_members", "owner_id", "created_at", "updated_at",
		})

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, description, icon_url, visibility, access_type, password_hash, max_members, owner_id, created_at, updated_at
		FROM tenants
		ORDER BY created_at DESC`)).
			WillReturnRows(rows)

		tenants, err := repo.List(ctx)
		require.NoError(t, err)
		assert.Len(t, tenants, 0)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, description, icon_url, visibility, access_type, password_hash, max_members, owner_id, created_at, updated_at
		FROM tenants
		ORDER BY created_at DESC`)).
			WillReturnError(errors.New("database error"))

		tenants, err := repo.List(ctx)
		require.Error(t, err)
		assert.Nil(t, tenants)
		assert.Contains(t, err.Error(), "failed to list tenants")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "icon_url", "visibility", "access_type",
			"password_hash", "max_members", "owner_id", "created_at", "updated_at",
		}).AddRow("tenant-1", "Tenant 1", "Description", "icon", "public", "open", "", "invalid-int", "owner", now, now)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, description, icon_url, visibility, access_type, password_hash, max_members, owner_id, created_at, updated_at
		FROM tenants
		ORDER BY created_at DESC`)).
			WillReturnRows(rows)

		tenants, err := repo.List(ctx)
		require.Error(t, err)
		assert.Nil(t, tenants)
		assert.Contains(t, err.Error(), "failed to scan tenant")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "icon_url", "visibility", "access_type",
			"password_hash", "max_members", "owner_id", "created_at", "updated_at",
		}).
			AddRow("tenant-1", "Tenant 1", "Desc", "icon", "public", "open", "", 100, "owner", now, now).
			RowError(0, errors.New("row iteration error"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, description, icon_url, visibility, access_type, password_hash, max_members, owner_id, created_at, updated_at
		FROM tenants
		ORDER BY created_at DESC`)).
			WillReturnRows(rows)

		tenants, err := repo.List(ctx)
		require.Error(t, err)
		assert.Nil(t, tenants)
		assert.Contains(t, err.Error(), "failed to iterate tenants")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantRepository_ListAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success without query filter", func(t *testing.T) {
		limit := 10
		offset := 0

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM tenants`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "icon_url", "visibility", "access_type",
			"password_hash", "max_members", "owner_id", "created_at", "updated_at",
		}).
			AddRow("tenant-1", "Tenant 1", "Description 1", "https://example.com/1.png",
				domain.TenantVisibilityPublic, domain.TenantAccessOpen, "", 100, "owner-1", now, now).
			AddRow("tenant-2", "Tenant 2", nil, nil,
				domain.TenantVisibilityPrivate, domain.TenantAccessInviteOnly, nil, 50, nil, now, now)

		mock.ExpectQuery(`SELECT id, name, description, icon_url, visibility, access_type, password_hash, max_members, owner_id, created_at, updated_at`).
			WithArgs(limit, offset).
			WillReturnRows(rows)

		tenants, total, err := repo.ListAll(ctx, limit, offset, "")
		require.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, tenants, 2)
		assert.Equal(t, "tenant-1", tenants[0].ID)
		assert.Equal(t, "tenant-2", tenants[1].ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with query filter", func(t *testing.T) {
		limit := 10
		offset := 0
		query := "test"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM tenants WHERE name ILIKE $1`)).
			WithArgs("%test%").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "icon_url", "visibility", "access_type",
			"password_hash", "max_members", "owner_id", "created_at", "updated_at",
		}).AddRow("tenant-1", "Test Tenant", "Description", nil,
			domain.TenantVisibilityPublic, domain.TenantAccessOpen, "", 100, "owner-1", now, now)

		mock.ExpectQuery(`SELECT id, name, description, icon_url, visibility, access_type, password_hash, max_members, owner_id, created_at, updated_at`).
			WithArgs("%test%", limit, offset).
			WillReturnRows(rows)

		tenants, total, err := repo.ListAll(ctx, limit, offset, query)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, tenants, 1)
		assert.Equal(t, "Test Tenant", tenants[0].Name)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with pagination", func(t *testing.T) {
		limit := 5
		offset := 5

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM tenants`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "icon_url", "visibility", "access_type",
			"password_hash", "max_members", "owner_id", "created_at", "updated_at",
		}).AddRow("tenant-6", "Tenant 6", nil, nil,
			domain.TenantVisibilityPublic, domain.TenantAccessOpen, "", 100, nil, now, now)

		mock.ExpectQuery(`SELECT id, name, description, icon_url, visibility, access_type, password_hash, max_members, owner_id, created_at, updated_at`).
			WithArgs(limit, offset).
			WillReturnRows(rows)

		tenants, total, err := repo.ListAll(ctx, limit, offset, "")
		require.NoError(t, err)
		assert.Equal(t, 10, total)
		assert.Len(t, tenants, 1)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("count query error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM tenants`)).
			WillReturnError(errors.New("count error"))

		tenants, total, err := repo.ListAll(ctx, 10, 0, "")
		require.Error(t, err)
		assert.Nil(t, tenants)
		assert.Equal(t, 0, total)
		assert.Contains(t, err.Error(), "failed to count tenants")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list query error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM tenants`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

		mock.ExpectQuery(`SELECT id, name, description, icon_url, visibility, access_type, password_hash, max_members, owner_id, created_at, updated_at`).
			WillReturnError(errors.New("list error"))

		tenants, total, err := repo.ListAll(ctx, 10, 0, "")
		require.Error(t, err)
		assert.Nil(t, tenants)
		assert.Equal(t, 0, total)
		assert.Contains(t, err.Error(), "failed to list tenants")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM tenants`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "icon_url", "visibility", "access_type",
			"password_hash", "max_members", "owner_id", "created_at", "updated_at",
		}).AddRow("tenant-1", "Tenant", "Desc", "icon", "public", "open", "", "invalid", "owner", now, now)

		mock.ExpectQuery(`SELECT id, name, description, icon_url, visibility, access_type, password_hash, max_members, owner_id, created_at, updated_at`).
			WillReturnRows(rows)

		tenants, total, err := repo.ListAll(ctx, 10, 0, "")
		require.Error(t, err)
		assert.Nil(t, tenants)
		assert.Equal(t, 0, total)
		assert.Contains(t, err.Error(), "failed to scan tenant")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM tenants`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{
			"id", "name", "description", "icon_url", "visibility", "access_type",
			"password_hash", "max_members", "owner_id", "created_at", "updated_at",
		}).
			AddRow("tenant-1", "Tenant", "Desc", "icon", "public", "open", "", 100, "owner", now, now).
			RowError(0, errors.New("iteration error"))

		mock.ExpectQuery(`SELECT id, name, description, icon_url, visibility, access_type, password_hash, max_members, owner_id, created_at, updated_at`).
			WillReturnRows(rows)

		tenants, total, err := repo.ListAll(ctx, 10, 0, "")
		require.Error(t, err)
		assert.Nil(t, tenants)
		assert.Equal(t, 0, total)
		assert.Contains(t, err.Error(), "failed to iterate tenants")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresTenantRepository_Count(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresTenantRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM tenants`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(42))

		count, err := repo.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, 42, count)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("zero count", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM tenants`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		count, err := repo.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM tenants`)).
			WillReturnError(errors.New("database error"))

		count, err := repo.Count(ctx)
		require.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "failed to count tenants")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
