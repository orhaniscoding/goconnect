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

func TestNewPostgresNetworkRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresNetworkRepository(db)
	require.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestPostgresNetworkRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresNetworkRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		network := &domain.Network{
			ID:         "net-123",
			TenantID:   "tenant-456",
			Name:       "Test Network",
			CIDR:       "10.0.0.0/24",
			Visibility: domain.NetworkVisibilityPublic,
			CreatedBy:  "user-001",
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO networks (id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`)).
			WithArgs(network.ID, network.TenantID, network.Name, network.CIDR, network.Visibility, network.CreatedBy, network.CreatedAt, network.UpdatedAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, network)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate key error", func(t *testing.T) {
		network := &domain.Network{
			ID:         "net-dup",
			TenantID:   "tenant-456",
			Name:       "Duplicate Network",
			CIDR:       "10.0.1.0/24",
			Visibility: domain.NetworkVisibilityPrivate,
			CreatedBy:  "user-001",
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO networks (id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`)).
			WithArgs(network.ID, network.TenantID, network.Name, network.CIDR, network.Visibility, network.CreatedBy, network.CreatedAt, network.UpdatedAt).
			WillReturnError(errors.New("duplicate key value"))

		err := repo.Create(ctx, network)
		require.Error(t, err)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrInvalidRequest, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("unique constraint error", func(t *testing.T) {
		network := &domain.Network{
			ID:         "net-uniq",
			TenantID:   "tenant-456",
			Name:       "Unique Network",
			CIDR:       "10.0.2.0/24",
			Visibility: domain.NetworkVisibilityPrivate,
			CreatedBy:  "user-001",
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO networks (id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`)).
			WithArgs(network.ID, network.TenantID, network.Name, network.CIDR, network.Visibility, network.CreatedBy, network.CreatedAt, network.UpdatedAt).
			WillReturnError(errors.New("UNIQUE constraint failed"))

		err := repo.Create(ctx, network)
		require.Error(t, err)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrInvalidRequest, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		network := &domain.Network{
			ID:         "net-error",
			TenantID:   "tenant-456",
			Name:       "Error Network",
			CIDR:       "10.0.3.0/24",
			Visibility: domain.NetworkVisibilityPublic,
			CreatedBy:  "user-001",
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO networks (id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`)).
			WithArgs(network.ID, network.TenantID, network.Name, network.CIDR, network.Visibility, network.CreatedBy, network.CreatedAt, network.UpdatedAt).
			WillReturnError(errors.New("connection refused"))

		err := repo.Create(ctx, network)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create network")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresNetworkRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresNetworkRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success without deleted_at", func(t *testing.T) {
		id := "net-123"

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "cidr", "visibility", "created_by", "created_at", "updated_at", "deleted_at"}).
			AddRow(id, "tenant-456", "Test Network", "10.0.0.0/24", domain.NetworkVisibilityPublic, "user-001", now, now, nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at, deleted_at FROM networks WHERE id = $1 AND deleted_at IS NULL`)).
			WithArgs(id).
			WillReturnRows(rows)

		network, err := repo.GetByID(ctx, id)
		require.NoError(t, err)
		assert.NotNil(t, network)
		assert.Equal(t, id, network.ID)
		assert.Equal(t, "tenant-456", network.TenantID)
		assert.Equal(t, "Test Network", network.Name)
		assert.Equal(t, "10.0.0.0/24", network.CIDR)
		assert.Equal(t, domain.NetworkVisibilityPublic, network.Visibility)
		assert.Nil(t, network.SoftDeletedAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with deleted_at", func(t *testing.T) {
		id := "net-deleted"
		deletedAt := now.Add(-1 * time.Hour)

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "cidr", "visibility", "created_by", "created_at", "updated_at", "deleted_at"}).
			AddRow(id, "tenant-456", "Deleted Network", "10.0.1.0/24", domain.NetworkVisibilityPrivate, "user-001", now, now, deletedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at, deleted_at FROM networks WHERE id = $1 AND deleted_at IS NULL`)).
			WithArgs(id).
			WillReturnRows(rows)

		network, err := repo.GetByID(ctx, id)
		require.NoError(t, err)
		assert.NotNil(t, network)
		assert.NotNil(t, network.SoftDeletedAt)
		assert.Equal(t, deletedAt.Unix(), network.SoftDeletedAt.Unix())
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		id := "nonexistent"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at, deleted_at FROM networks WHERE id = $1 AND deleted_at IS NULL`)).
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		network, err := repo.GetByID(ctx, id)
		require.Error(t, err)
		assert.Nil(t, network)
		assert.Contains(t, err.Error(), domain.ErrNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		id := "error-id"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at, deleted_at FROM networks WHERE id = $1 AND deleted_at IS NULL`)).
			WithArgs(id).
			WillReturnError(errors.New("connection error"))

		network, err := repo.GetByID(ctx, id)
		require.Error(t, err)
		assert.Nil(t, network)
		assert.Contains(t, err.Error(), "failed to get network by ID")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresNetworkRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresNetworkRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success with no filters", func(t *testing.T) {
		filter := NetworkFilter{
			Limit: 10,
		}

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "cidr", "visibility", "created_by", "created_at", "updated_at", "deleted_at"}).
			AddRow("net-1", "tenant-456", "Network 1", "10.0.0.0/24", domain.NetworkVisibilityPublic, "user-001", now, now, nil).
			AddRow("net-2", "tenant-456", "Network 2", "10.0.1.0/24", domain.NetworkVisibilityPrivate, "user-002", now, now, nil)

		mock.ExpectQuery(`SELECT id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at, deleted_at FROM networks WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT`).
			WillReturnRows(rows)

		networks, cursor, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, networks, 2)
		assert.Empty(t, cursor)
		assert.Equal(t, "net-1", networks[0].ID)
		assert.Equal(t, "net-2", networks[1].ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with tenant filter", func(t *testing.T) {
		filter := NetworkFilter{
			TenantID: "tenant-123",
			Limit:    10,
		}

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "cidr", "visibility", "created_by", "created_at", "updated_at", "deleted_at"}).
			AddRow("net-1", "tenant-123", "Network 1", "10.0.0.0/24", domain.NetworkVisibilityPublic, "user-001", now, now, nil)

		mock.ExpectQuery(`SELECT id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at, deleted_at FROM networks WHERE deleted_at IS NULL AND tenant_id = .* ORDER BY created_at DESC LIMIT`).
			WillReturnRows(rows)

		networks, cursor, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, networks, 1)
		assert.Empty(t, cursor)
		assert.Equal(t, "tenant-123", networks[0].TenantID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with public visibility filter", func(t *testing.T) {
		filter := NetworkFilter{
			Visibility: "public",
			Limit:      10,
		}

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "cidr", "visibility", "created_by", "created_at", "updated_at", "deleted_at"}).
			AddRow("net-1", "tenant-456", "Public Network", "10.0.0.0/24", domain.NetworkVisibilityPublic, "user-001", now, now, nil)

		mock.ExpectQuery(`SELECT id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at, deleted_at FROM networks WHERE deleted_at IS NULL AND visibility = .* ORDER BY created_at DESC LIMIT`).
			WillReturnRows(rows)

		networks, cursor, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, networks, 1)
		assert.Empty(t, cursor)
		assert.Equal(t, domain.NetworkVisibilityPublic, networks[0].Visibility)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with mine visibility filter", func(t *testing.T) {
		filter := NetworkFilter{
			Visibility: "mine",
			UserID:     "user-001",
			Limit:      10,
		}

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "cidr", "visibility", "created_by", "created_at", "updated_at", "deleted_at"}).
			AddRow("net-1", "tenant-456", "My Network", "10.0.0.0/24", domain.NetworkVisibilityPrivate, "user-001", now, now, nil)

		mock.ExpectQuery(`SELECT id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at, deleted_at FROM networks WHERE deleted_at IS NULL AND created_by = .* ORDER BY created_at DESC LIMIT`).
			WillReturnRows(rows)

		networks, cursor, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, networks, 1)
		assert.Empty(t, cursor)
		assert.Equal(t, "user-001", networks[0].CreatedBy)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with search filter", func(t *testing.T) {
		filter := NetworkFilter{
			Search: "test",
			Limit:  10,
		}

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "cidr", "visibility", "created_by", "created_at", "updated_at", "deleted_at"}).
			AddRow("net-1", "tenant-456", "Test Network", "10.0.0.0/24", domain.NetworkVisibilityPublic, "user-001", now, now, nil)

		mock.ExpectQuery(`SELECT id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at, deleted_at FROM networks WHERE deleted_at IS NULL AND \(name ILIKE .* OR cidr ILIKE .*\) ORDER BY created_at DESC LIMIT`).
			WillReturnRows(rows)

		networks, cursor, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, networks, 1)
		assert.Empty(t, cursor)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with cursor pagination", func(t *testing.T) {
		filter := NetworkFilter{
			Cursor: "net-cursor",
			Limit:  10,
		}

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "cidr", "visibility", "created_by", "created_at", "updated_at", "deleted_at"}).
			AddRow("net-1", "tenant-456", "Network 1", "10.0.0.0/24", domain.NetworkVisibilityPublic, "user-001", now, now, nil)

		mock.ExpectQuery(`SELECT id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at, deleted_at FROM networks WHERE deleted_at IS NULL AND created_at < \(SELECT created_at FROM networks WHERE id = .*\) ORDER BY created_at DESC LIMIT`).
			WillReturnRows(rows)

		networks, cursor, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, networks, 1)
		assert.Empty(t, cursor)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with next cursor", func(t *testing.T) {
		filter := NetworkFilter{
			Limit: 2,
		}

		// Return 3 rows (limit+1) to indicate more results
		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "cidr", "visibility", "created_by", "created_at", "updated_at", "deleted_at"}).
			AddRow("net-1", "tenant-456", "Network 1", "10.0.0.0/24", domain.NetworkVisibilityPublic, "user-001", now, now, nil).
			AddRow("net-2", "tenant-456", "Network 2", "10.0.1.0/24", domain.NetworkVisibilityPublic, "user-002", now, now, nil).
			AddRow("net-3", "tenant-456", "Network 3", "10.0.2.0/24", domain.NetworkVisibilityPublic, "user-003", now, now, nil)

		mock.ExpectQuery(`SELECT id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at, deleted_at FROM networks WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT`).
			WillReturnRows(rows)

		networks, cursor, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, networks, 2)
		assert.Equal(t, "net-2", cursor)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with deleted_at populated", func(t *testing.T) {
		filter := NetworkFilter{
			Limit: 10,
		}
		deletedAt := now.Add(-1 * time.Hour)

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "cidr", "visibility", "created_by", "created_at", "updated_at", "deleted_at"}).
			AddRow("net-1", "tenant-456", "Network 1", "10.0.0.0/24", domain.NetworkVisibilityPublic, "user-001", now, now, deletedAt)

		mock.ExpectQuery(`SELECT id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at, deleted_at FROM networks WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT`).
			WillReturnRows(rows)

		networks, cursor, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, networks, 1)
		assert.Empty(t, cursor)
		assert.NotNil(t, networks[0].SoftDeletedAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error", func(t *testing.T) {
		filter := NetworkFilter{
			Limit: 10,
		}

		mock.ExpectQuery(`SELECT id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at, deleted_at FROM networks WHERE deleted_at IS NULL`).
			WillReturnError(errors.New("database error"))

		networks, cursor, err := repo.List(ctx, filter)
		require.Error(t, err)
		assert.Nil(t, networks)
		assert.Empty(t, cursor)
		assert.Contains(t, err.Error(), "failed to list networks")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		filter := NetworkFilter{
			Limit: 10,
		}

		// Return wrong number of columns
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow("net-1", "Network 1")

		mock.ExpectQuery(`SELECT id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at, deleted_at FROM networks WHERE deleted_at IS NULL`).
			WillReturnRows(rows)

		networks, cursor, err := repo.List(ctx, filter)
		require.Error(t, err)
		assert.Nil(t, networks)
		assert.Empty(t, cursor)
		assert.Contains(t, err.Error(), "failed to scan network")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows error", func(t *testing.T) {
		filter := NetworkFilter{
			Limit: 10,
		}

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "cidr", "visibility", "created_by", "created_at", "updated_at", "deleted_at"}).
			AddRow("net-1", "tenant-456", "Network 1", "10.0.0.0/24", domain.NetworkVisibilityPublic, "user-001", now, now, nil).
			RowError(0, errors.New("row iteration error"))

		mock.ExpectQuery(`SELECT id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at, deleted_at FROM networks WHERE deleted_at IS NULL`).
			WillReturnRows(rows)

		networks, cursor, err := repo.List(ctx, filter)
		require.Error(t, err)
		assert.Nil(t, networks)
		assert.Empty(t, cursor)
		assert.Contains(t, err.Error(), "failed to iterate networks")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresNetworkRepository_CheckCIDROverlap(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresNetworkRepository(db)
	ctx := context.Background()

	t.Run("no overlap exists", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS( SELECT 1 FROM networks WHERE cidr = $1 AND id != $2 AND tenant_id = $3 AND deleted_at IS NULL )`)).
			WithArgs("10.0.0.0/24", "net-123", "tenant-456").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		overlap, err := repo.CheckCIDROverlap(ctx, "10.0.0.0/24", "net-123", "tenant-456")
		require.NoError(t, err)
		assert.False(t, overlap)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("overlap exists", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS( SELECT 1 FROM networks WHERE cidr = $1 AND id != $2 AND tenant_id = $3 AND deleted_at IS NULL )`)).
			WithArgs("10.0.0.0/24", "net-new", "tenant-456").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		overlap, err := repo.CheckCIDROverlap(ctx, "10.0.0.0/24", "net-new", "tenant-456")
		require.NoError(t, err)
		assert.True(t, overlap)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS( SELECT 1 FROM networks WHERE cidr = $1 AND id != $2 AND tenant_id = $3 AND deleted_at IS NULL )`)).
			WithArgs("10.0.0.0/24", "net-123", "tenant-456").
			WillReturnError(errors.New("connection error"))

		overlap, err := repo.CheckCIDROverlap(ctx, "10.0.0.0/24", "net-123", "tenant-456")
		require.Error(t, err)
		assert.False(t, overlap)
		assert.Contains(t, err.Error(), "failed to check CIDR overlap")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresNetworkRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresNetworkRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		id := "net-123"

		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "cidr", "visibility", "created_by", "created_at", "updated_at", "deleted_at"}).
			AddRow(id, "tenant-456", "Original Name", "10.0.0.0/24", domain.NetworkVisibilityPublic, "user-001", now, now, nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at, deleted_at FROM networks WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`)).
			WithArgs(id).
			WillReturnRows(rows)

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE networks SET name = $1, cidr = $2, visibility = $3, updated_at = $4 WHERE id = $5`)).
			WithArgs("Updated Name", "10.0.0.0/24", domain.NetworkVisibilityPublic, sqlmock.AnyArg(), id).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		network, err := repo.Update(ctx, id, func(n *domain.Network) error {
			n.Name = "Updated Name"
			return nil
		})

		require.NoError(t, err)
		assert.NotNil(t, network)
		assert.Equal(t, "Updated Name", network.Name)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with deleted_at populated", func(t *testing.T) {
		id := "net-with-deleted"
		deletedAt := now.Add(-1 * time.Hour)

		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "cidr", "visibility", "created_by", "created_at", "updated_at", "deleted_at"}).
			AddRow(id, "tenant-456", "Network", "10.0.0.0/24", domain.NetworkVisibilityPublic, "user-001", now, now, deletedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at, deleted_at FROM networks WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`)).
			WithArgs(id).
			WillReturnRows(rows)

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE networks SET name = $1, cidr = $2, visibility = $3, updated_at = $4 WHERE id = $5`)).
			WithArgs("Network", "10.0.0.0/24", domain.NetworkVisibilityPublic, sqlmock.AnyArg(), id).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		network, err := repo.Update(ctx, id, func(n *domain.Network) error {
			return nil
		})

		require.NoError(t, err)
		assert.NotNil(t, network)
		assert.NotNil(t, network.SoftDeletedAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("begin transaction error", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(errors.New("begin failed"))

		network, err := repo.Update(ctx, "net-123", func(n *domain.Network) error {
			return nil
		})

		require.Error(t, err)
		assert.Nil(t, network)
		assert.Contains(t, err.Error(), "failed to begin transaction")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("network not found", func(t *testing.T) {
		id := "nonexistent"

		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at, deleted_at FROM networks WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`)).
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		mock.ExpectRollback()

		network, err := repo.Update(ctx, id, func(n *domain.Network) error {
			return nil
		})

		require.Error(t, err)
		assert.Nil(t, network)
		assert.Contains(t, err.Error(), domain.ErrNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("select error", func(t *testing.T) {
		id := "net-error"

		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at, deleted_at FROM networks WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`)).
			WithArgs(id).
			WillReturnError(errors.New("database error"))

		mock.ExpectRollback()

		network, err := repo.Update(ctx, id, func(n *domain.Network) error {
			return nil
		})

		require.Error(t, err)
		assert.Nil(t, network)
		assert.Contains(t, err.Error(), "failed to get network for update")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("mutate function error", func(t *testing.T) {
		id := "net-123"

		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "cidr", "visibility", "created_by", "created_at", "updated_at", "deleted_at"}).
			AddRow(id, "tenant-456", "Original Name", "10.0.0.0/24", domain.NetworkVisibilityPublic, "user-001", now, now, nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at, deleted_at FROM networks WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`)).
			WithArgs(id).
			WillReturnRows(rows)

		mock.ExpectRollback()

		network, err := repo.Update(ctx, id, func(n *domain.Network) error {
			return errors.New("mutation failed")
		})

		require.Error(t, err)
		assert.Nil(t, network)
		assert.Contains(t, err.Error(), "mutation failed")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update exec error", func(t *testing.T) {
		id := "net-123"

		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "cidr", "visibility", "created_by", "created_at", "updated_at", "deleted_at"}).
			AddRow(id, "tenant-456", "Original Name", "10.0.0.0/24", domain.NetworkVisibilityPublic, "user-001", now, now, nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at, deleted_at FROM networks WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`)).
			WithArgs(id).
			WillReturnRows(rows)

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE networks SET name = $1, cidr = $2, visibility = $3, updated_at = $4 WHERE id = $5`)).
			WithArgs("Original Name", "10.0.0.0/24", domain.NetworkVisibilityPublic, sqlmock.AnyArg(), id).
			WillReturnError(errors.New("update failed"))

		mock.ExpectRollback()

		network, err := repo.Update(ctx, id, func(n *domain.Network) error {
			return nil
		})

		require.Error(t, err)
		assert.Nil(t, network)
		assert.Contains(t, err.Error(), "failed to update network")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("commit error", func(t *testing.T) {
		id := "net-123"

		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "cidr", "visibility", "created_by", "created_at", "updated_at", "deleted_at"}).
			AddRow(id, "tenant-456", "Original Name", "10.0.0.0/24", domain.NetworkVisibilityPublic, "user-001", now, now, nil)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, tenant_id, name, cidr, visibility, created_by, created_at, updated_at, deleted_at FROM networks WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`)).
			WithArgs(id).
			WillReturnRows(rows)

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE networks SET name = $1, cidr = $2, visibility = $3, updated_at = $4 WHERE id = $5`)).
			WithArgs("Original Name", "10.0.0.0/24", domain.NetworkVisibilityPublic, sqlmock.AnyArg(), id).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

		network, err := repo.Update(ctx, id, func(n *domain.Network) error {
			return nil
		})

		require.Error(t, err)
		assert.Nil(t, network)
		assert.Contains(t, err.Error(), "failed to commit transaction")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresNetworkRepository_SoftDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresNetworkRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		id := "net-123"

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE networks SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`)).
			WithArgs(now, id).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.SoftDelete(ctx, id, now)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found or already deleted", func(t *testing.T) {
		id := "nonexistent"

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE networks SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`)).
			WithArgs(now, id).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.SoftDelete(ctx, id, now)
		require.Error(t, err)
		assert.Contains(t, err.Error(), domain.ErrNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		id := "net-error"

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE networks SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`)).
			WithArgs(now, id).
			WillReturnError(errors.New("connection error"))

		err := repo.SoftDelete(ctx, id, now)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to soft delete network")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		id := "net-rows-err"

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE networks SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`)).
			WithArgs(now, id).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err := repo.SoftDelete(ctx, id, now)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get rows affected")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresNetworkRepository_Count(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresNetworkRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM networks WHERE deleted_at IS NULL`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(42))

		count, err := repo.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, 42, count)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("zero count", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM networks WHERE deleted_at IS NULL`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		count, err := repo.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM networks WHERE deleted_at IS NULL`)).
			WillReturnError(errors.New("connection error"))

		count, err := repo.Count(ctx)
		require.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "failed to count networks")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
