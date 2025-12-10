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

func TestNewPostgresIPRuleRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresIPRuleRepository(db)
	require.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestPostgresIPRuleRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresIPRuleRepository(db)
	ctx := context.Background()
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	t.Run("success", func(t *testing.T) {
		rule := &domain.IPRule{
			ID:          "rule-1",
			TenantID:    "tenant-1",
			Type:        domain.IPRuleTypeAllow,
			CIDR:        "192.168.1.0/24",
			Description: "Test rule",
			CreatedBy:   "user-1",
			CreatedAt:   now,
			UpdatedAt:   now,
			ExpiresAt:   &expiresAt,
		}

		query := `INSERT INTO ip_rules (id, tenant_id, type, cidr, description, created_by, created_at, updated_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(
				rule.ID,
				rule.TenantID,
				rule.Type,
				rule.CIDR,
				rule.Description,
				rule.CreatedBy,
				rule.CreatedAt,
				rule.UpdatedAt,
				rule.ExpiresAt,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, rule)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success without expires_at", func(t *testing.T) {
		rule := &domain.IPRule{
			ID:          "rule-2",
			TenantID:    "tenant-1",
			Type:        domain.IPRuleTypeDeny,
			CIDR:        "10.0.0.0/8",
			Description: "Block internal",
			CreatedBy:   "user-1",
			CreatedAt:   now,
			UpdatedAt:   now,
			ExpiresAt:   nil,
		}

		query := `INSERT INTO ip_rules (id, tenant_id, type, cidr, description, created_by, created_at, updated_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(
				rule.ID,
				rule.TenantID,
				rule.Type,
				rule.CIDR,
				rule.Description,
				rule.CreatedBy,
				rule.CreatedAt,
				rule.UpdatedAt,
				nil,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, rule)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		rule := &domain.IPRule{
			ID:          "rule-3",
			TenantID:    "tenant-1",
			Type:        domain.IPRuleTypeAllow,
			CIDR:        "192.168.1.0/24",
			Description: "Test rule",
			CreatedBy:   "user-1",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		query := `INSERT INTO ip_rules (id, tenant_id, type, cidr, description, created_by, created_at, updated_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(
				rule.ID,
				rule.TenantID,
				rule.Type,
				rule.CIDR,
				rule.Description,
				rule.CreatedBy,
				rule.CreatedAt,
				rule.UpdatedAt,
				nil,
			).
			WillReturnError(errors.New("database error"))

		err := repo.Create(ctx, rule)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create IP rule")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresIPRuleRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresIPRuleRepository(db)
	ctx := context.Background()
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	query := `SELECT id, tenant_id, type, cidr, description, created_by, created_at, updated_at, expires_at
		FROM ip_rules
		WHERE id = $1`

	t.Run("success with expires_at", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "type", "cidr", "description", "created_by", "created_at", "updated_at", "expires_at",
		}).AddRow(
			"rule-1", "tenant-1", "allow", "192.168.1.0/24", "Test rule", "user-1", now, now, expiresAt,
		)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("rule-1").
			WillReturnRows(rows)

		rule, err := repo.GetByID(ctx, "rule-1")
		require.NoError(t, err)
		assert.Equal(t, "rule-1", rule.ID)
		assert.Equal(t, "tenant-1", rule.TenantID)
		assert.Equal(t, domain.IPRuleType("allow"), rule.Type)
		assert.Equal(t, "192.168.1.0/24", rule.CIDR)
		assert.Equal(t, "Test rule", rule.Description)
		assert.Equal(t, "user-1", rule.CreatedBy)
		require.NotNil(t, rule.ExpiresAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success without expires_at", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "type", "cidr", "description", "created_by", "created_at", "updated_at", "expires_at",
		}).AddRow(
			"rule-2", "tenant-1", "deny", "10.0.0.0/8", nil, "user-1", now, now, nil,
		)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("rule-2").
			WillReturnRows(rows)

		rule, err := repo.GetByID(ctx, "rule-2")
		require.NoError(t, err)
		assert.Equal(t, "rule-2", rule.ID)
		assert.Equal(t, domain.IPRuleType("deny"), rule.Type)
		assert.Equal(t, "", rule.Description)
		assert.Nil(t, rule.ExpiresAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("non-existent").
			WillReturnError(sql.ErrNoRows)

		rule, err := repo.GetByID(ctx, "non-existent")
		require.Error(t, err)
		assert.Nil(t, rule)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("rule-1").
			WillReturnError(errors.New("database error"))

		rule, err := repo.GetByID(ctx, "rule-1")
		require.Error(t, err)
		assert.Nil(t, rule)
		assert.Contains(t, err.Error(), "failed to get IP rule")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresIPRuleRepository_ListByTenant(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresIPRuleRepository(db)
	ctx := context.Background()
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	query := `SELECT id, tenant_id, type, cidr, description, created_by, created_at, updated_at, expires_at
		FROM ip_rules
		WHERE tenant_id = $1 AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY created_at DESC`

	t.Run("success with multiple rules", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "type", "cidr", "description", "created_by", "created_at", "updated_at", "expires_at",
		}).
			AddRow("rule-1", "tenant-1", "allow", "192.168.1.0/24", "Allow internal", "user-1", now, now, expiresAt).
			AddRow("rule-2", "tenant-1", "deny", "10.0.0.0/8", nil, "user-1", now, now, nil)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("tenant-1").
			WillReturnRows(rows)

		rules, err := repo.ListByTenant(ctx, "tenant-1")
		require.NoError(t, err)
		require.Len(t, rules, 2)
		assert.Equal(t, "rule-1", rules[0].ID)
		assert.Equal(t, "rule-2", rules[1].ID)
		require.NotNil(t, rules[0].ExpiresAt)
		assert.Nil(t, rules[1].ExpiresAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with empty result", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "type", "cidr", "description", "created_by", "created_at", "updated_at", "expires_at",
		})

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("tenant-2").
			WillReturnRows(rows)

		rules, err := repo.ListByTenant(ctx, "tenant-2")
		require.NoError(t, err)
		assert.Empty(t, rules)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("tenant-1").
			WillReturnError(errors.New("database error"))

		rules, err := repo.ListByTenant(ctx, "tenant-1")
		require.Error(t, err)
		assert.Nil(t, rules)
		assert.Contains(t, err.Error(), "failed to list IP rules")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		// Return wrong number of columns to trigger scan error
		rows := sqlmock.NewRows([]string{"id", "tenant_id"}).
			AddRow("rule-1", "tenant-1")

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("tenant-1").
			WillReturnRows(rows)

		rules, err := repo.ListByTenant(ctx, "tenant-1")
		require.Error(t, err)
		assert.Nil(t, rules)
		assert.Contains(t, err.Error(), "failed to scan IP rule")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresIPRuleRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresIPRuleRepository(db)
	ctx := context.Background()

	query := `DELETE FROM ip_rules WHERE id = $1`

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("rule-1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(ctx, "rule-1")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("non-existent").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(ctx, "non-existent")
		require.Error(t, err)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("exec error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("rule-1").
			WillReturnError(errors.New("database error"))

		err := repo.Delete(ctx, "rule-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete IP rule")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("rule-1").
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err := repo.Delete(ctx, "rule-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get rows affected")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresIPRuleRepository_DeleteExpired(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresIPRuleRepository(db)
	ctx := context.Background()

	query := `DELETE FROM ip_rules WHERE expires_at IS NOT NULL AND expires_at < NOW()`

	t.Run("success with deleted rows", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WillReturnResult(sqlmock.NewResult(0, 5))

		count, err := repo.DeleteExpired(ctx)
		require.NoError(t, err)
		assert.Equal(t, 5, count)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with no expired rules", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WillReturnResult(sqlmock.NewResult(0, 0))

		count, err := repo.DeleteExpired(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("exec error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WillReturnError(errors.New("database error"))

		count, err := repo.DeleteExpired(ctx)
		require.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "failed to delete expired IP rules")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		count, err := repo.DeleteExpired(ctx)
		require.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Contains(t, err.Error(), "failed to get rows affected")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresIPRuleRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresIPRuleRepository(db)
	ctx := context.Background()
	expiresAt := time.Now().Add(48 * time.Hour)

	query := `UPDATE ip_rules 
		SET type = $1, cidr = $2, description = $3, updated_at = $4, expires_at = $5
		WHERE id = $6`

	t.Run("success", func(t *testing.T) {
		rule := &domain.IPRule{
			ID:          "rule-1",
			TenantID:    "tenant-1",
			Type:        domain.IPRuleTypeDeny,
			CIDR:        "172.16.0.0/12",
			Description: "Updated description",
			ExpiresAt:   &expiresAt,
		}

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(
				rule.Type,
				rule.CIDR,
				rule.Description,
				sqlmock.AnyArg(), // updated_at is set dynamically
				rule.ExpiresAt,
				rule.ID,
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(ctx, rule)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success without expires_at", func(t *testing.T) {
		rule := &domain.IPRule{
			ID:          "rule-2",
			TenantID:    "tenant-1",
			Type:        domain.IPRuleTypeAllow,
			CIDR:        "192.168.0.0/16",
			Description: "No expiration",
			ExpiresAt:   nil,
		}

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(
				rule.Type,
				rule.CIDR,
				rule.Description,
				sqlmock.AnyArg(),
				nil,
				rule.ID,
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(ctx, rule)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		rule := &domain.IPRule{
			ID:          "non-existent",
			Type:        domain.IPRuleTypeAllow,
			CIDR:        "192.168.0.0/16",
			Description: "Test",
		}

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(
				rule.Type,
				rule.CIDR,
				rule.Description,
				sqlmock.AnyArg(),
				nil,
				rule.ID,
			).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Update(ctx, rule)
		require.Error(t, err)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("exec error", func(t *testing.T) {
		rule := &domain.IPRule{
			ID:          "rule-1",
			Type:        domain.IPRuleTypeAllow,
			CIDR:        "192.168.0.0/16",
			Description: "Test",
		}

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(
				rule.Type,
				rule.CIDR,
				rule.Description,
				sqlmock.AnyArg(),
				nil,
				rule.ID,
			).
			WillReturnError(errors.New("database error"))

		err := repo.Update(ctx, rule)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update IP rule")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		rule := &domain.IPRule{
			ID:          "rule-1",
			Type:        domain.IPRuleTypeAllow,
			CIDR:        "192.168.0.0/16",
			Description: "Test",
		}

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs(
				rule.Type,
				rule.CIDR,
				rule.Description,
				sqlmock.AnyArg(),
				nil,
				rule.ID,
			).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err := repo.Update(ctx, rule)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get rows affected")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
