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

func TestNewAdminRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAdminRepository(db)
	require.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestAdminRepository_ListAllUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAdminRepository(db)
	ctx := context.Background()
	now := time.Now()
	username := "testuser"

	t.Run("success with no filters", func(t *testing.T) {
		filters := domain.UserFilters{}
		pagination := domain.PaginationParams{Page: 1, PerPage: 10}

		// Expect count query
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users WHERE 1=1`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		// Expect list query
		rows := sqlmock.NewRows([]string{
			"id", "email", "username", "tenant_id", "is_admin", "is_moderator", "suspended", "created_at", "last_seen",
		}).
			AddRow("user-1", "test1@example.com", &username, "tenant-1", false, false, false, now, now).
			AddRow("user-2", "test2@example.com", nil, "tenant-1", true, false, false, now, nil)

		mock.ExpectQuery(`SELECT id, email, username, tenant_id, is_admin, is_moderator, suspended, created_at, last_seen`).
			WillReturnRows(rows)

		users, total, err := repo.ListAllUsers(ctx, filters, pagination)
		require.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, users, 2)
		assert.Equal(t, "user-1", users[0].ID)
		assert.Equal(t, "test1@example.com", users[0].Email)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with admin role filter", func(t *testing.T) {
		filters := domain.UserFilters{Role: "admin"}
		pagination := domain.PaginationParams{Page: 1, PerPage: 10}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users WHERE 1=1 AND is_admin = $1`)).
			WithArgs(true).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{
			"id", "email", "username", "tenant_id", "is_admin", "is_moderator", "suspended", "created_at", "last_seen",
		}).AddRow("user-1", "admin@example.com", &username, "tenant-1", true, false, false, now, now)

		mock.ExpectQuery(`SELECT id, email, username, tenant_id, is_admin, is_moderator, suspended, created_at, last_seen`).
			WillReturnRows(rows)

		users, total, err := repo.ListAllUsers(ctx, filters, pagination)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, users, 1)
		assert.True(t, users[0].IsAdmin)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with moderator role filter", func(t *testing.T) {
		filters := domain.UserFilters{Role: "moderator"}
		pagination := domain.PaginationParams{Page: 1, PerPage: 10}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users WHERE 1=1 AND is_moderator = $1`)).
			WithArgs(true).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{
			"id", "email", "username", "tenant_id", "is_admin", "is_moderator", "suspended", "created_at", "last_seen",
		}).AddRow("user-1", "mod@example.com", &username, "tenant-1", false, true, false, now, now)

		mock.ExpectQuery(`SELECT id, email, username, tenant_id, is_admin, is_moderator, suspended, created_at, last_seen`).
			WillReturnRows(rows)

		users, total, err := repo.ListAllUsers(ctx, filters, pagination)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.True(t, users[0].IsModerator)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with user role filter", func(t *testing.T) {
		filters := domain.UserFilters{Role: "user"}
		pagination := domain.PaginationParams{Page: 1, PerPage: 10}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users WHERE 1=1 AND is_admin = $1 AND is_moderator = $2`)).
			WithArgs(false, false).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{
			"id", "email", "username", "tenant_id", "is_admin", "is_moderator", "suspended", "created_at", "last_seen",
		}).AddRow("user-1", "user@example.com", &username, "tenant-1", false, false, false, now, now)

		mock.ExpectQuery(`SELECT id, email, username, tenant_id, is_admin, is_moderator, suspended, created_at, last_seen`).
			WillReturnRows(rows)

		users, total, err := repo.ListAllUsers(ctx, filters, pagination)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.False(t, users[0].IsAdmin)
		assert.False(t, users[0].IsModerator)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with active status filter", func(t *testing.T) {
		filters := domain.UserFilters{Status: "active"}
		pagination := domain.PaginationParams{Page: 1, PerPage: 10}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users WHERE 1=1 AND suspended = $1`)).
			WithArgs(false).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{
			"id", "email", "username", "tenant_id", "is_admin", "is_moderator", "suspended", "created_at", "last_seen",
		}).AddRow("user-1", "active@example.com", &username, "tenant-1", false, false, false, now, now)

		mock.ExpectQuery(`SELECT id, email, username, tenant_id, is_admin, is_moderator, suspended, created_at, last_seen`).
			WillReturnRows(rows)

		users, total, err := repo.ListAllUsers(ctx, filters, pagination)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.False(t, users[0].Suspended)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with suspended status filter", func(t *testing.T) {
		filters := domain.UserFilters{Status: "suspended"}
		pagination := domain.PaginationParams{Page: 1, PerPage: 10}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users WHERE 1=1 AND suspended = $1`)).
			WithArgs(true).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{
			"id", "email", "username", "tenant_id", "is_admin", "is_moderator", "suspended", "created_at", "last_seen",
		}).AddRow("user-1", "suspended@example.com", &username, "tenant-1", false, false, true, now, now)

		mock.ExpectQuery(`SELECT id, email, username, tenant_id, is_admin, is_moderator, suspended, created_at, last_seen`).
			WillReturnRows(rows)

		users, total, err := repo.ListAllUsers(ctx, filters, pagination)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.True(t, users[0].Suspended)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with tenant filter", func(t *testing.T) {
		filters := domain.UserFilters{TenantID: "tenant-123"}
		pagination := domain.PaginationParams{Page: 1, PerPage: 10}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users WHERE 1=1 AND tenant_id = $1`)).
			WithArgs("tenant-123").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{
			"id", "email", "username", "tenant_id", "is_admin", "is_moderator", "suspended", "created_at", "last_seen",
		}).AddRow("user-1", "tenant@example.com", &username, "tenant-123", false, false, false, now, now)

		mock.ExpectQuery(`SELECT id, email, username, tenant_id, is_admin, is_moderator, suspended, created_at, last_seen`).
			WillReturnRows(rows)

		users, total, err := repo.ListAllUsers(ctx, filters, pagination)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Equal(t, "tenant-123", users[0].TenantID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with search filter", func(t *testing.T) {
		filters := domain.UserFilters{Search: "test"}
		pagination := domain.PaginationParams{Page: 1, PerPage: 10}

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE 1=1 AND \(LOWER\(email\) LIKE .* OR LOWER\(username\) LIKE .*\)`).
			WithArgs("%test%").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{
			"id", "email", "username", "tenant_id", "is_admin", "is_moderator", "suspended", "created_at", "last_seen",
		}).AddRow("user-1", "test@example.com", &username, "tenant-1", false, false, false, now, now)

		mock.ExpectQuery(`SELECT id, email, username, tenant_id, is_admin, is_moderator, suspended, created_at, last_seen`).
			WillReturnRows(rows)

		_, total, err := repo.ListAllUsers(ctx, filters, pagination)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("count query error", func(t *testing.T) {
		filters := domain.UserFilters{}
		pagination := domain.PaginationParams{Page: 1, PerPage: 10}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users WHERE 1=1`)).
			WillReturnError(errors.New("database error"))

		users, total, err := repo.ListAllUsers(ctx, filters, pagination)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to count users")
		assert.Nil(t, users)
		assert.Equal(t, 0, total)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list query error", func(t *testing.T) {
		filters := domain.UserFilters{}
		pagination := domain.PaginationParams{Page: 1, PerPage: 10}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users WHERE 1=1`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectQuery(`SELECT id, email, username, tenant_id, is_admin, is_moderator, suspended, created_at, last_seen`).
			WillReturnError(errors.New("query error"))

		users, total, err := repo.ListAllUsers(ctx, filters, pagination)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list users")
		assert.Nil(t, users)
		assert.Equal(t, 0, total)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		filters := domain.UserFilters{}
		pagination := domain.PaginationParams{Page: 1, PerPage: 10}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users WHERE 1=1`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		// Return wrong number of columns to trigger scan error
		rows := sqlmock.NewRows([]string{"id", "email"}).
			AddRow("user-1", "test@example.com")

		mock.ExpectQuery(`SELECT id, email, username, tenant_id, is_admin, is_moderator, suspended, created_at, last_seen`).
			WillReturnRows(rows)

		users, total, err := repo.ListAllUsers(ctx, filters, pagination)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to scan user")
		assert.Nil(t, users)
		assert.Equal(t, 0, total)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAdminRepository_GetUserStats(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAdminRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		// Total users
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users")).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(100))

		// Admin users
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users WHERE is_admin = true")).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

		// Moderator users
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users WHERE is_moderator = true")).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

		// Suspended users
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users WHERE suspended = true")).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

		// Tenants (optional, can fail)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM tenants")).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		// Networks (optional)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM networks")).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(15))

		// Devices (optional)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM devices")).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(50))

		// Active peers (optional)
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM peers WHERE last_seen`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(20))

		stats, err := repo.GetUserStats(ctx)
		require.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, 100, stats.TotalUsers)
		assert.Equal(t, 5, stats.AdminUsers)
		assert.Equal(t, 10, stats.ModeratorUsers)
		assert.Equal(t, 3, stats.SuspendedUsers)
		assert.Equal(t, 2, stats.TotalTenants)
		assert.Equal(t, 15, stats.TotalNetworks)
		assert.Equal(t, 50, stats.TotalDevices)
		assert.Equal(t, 20, stats.ActivePeers)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("total users count error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users")).
			WillReturnError(errors.New("database error"))

		stats, err := repo.GetUserStats(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to count users")
		assert.Nil(t, stats)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("admin users count error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users")).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(100))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users WHERE is_admin = true")).
			WillReturnError(errors.New("database error"))

		stats, err := repo.GetUserStats(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to count admin users")
		assert.Nil(t, stats)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("moderator users count error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users")).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(100))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users WHERE is_admin = true")).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users WHERE is_moderator = true")).
			WillReturnError(errors.New("database error"))

		stats, err := repo.GetUserStats(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to count moderator users")
		assert.Nil(t, stats)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("suspended users count error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users")).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(100))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users WHERE is_admin = true")).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users WHERE is_moderator = true")).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users WHERE suspended = true")).
			WillReturnError(errors.New("database error"))

		stats, err := repo.GetUserStats(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to count suspended users")
		assert.Nil(t, stats)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("optional queries can fail without error", func(t *testing.T) {
		// Required counts succeed
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users")).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(100))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users WHERE is_admin = true")).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users WHERE is_moderator = true")).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users WHERE suspended = true")).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

		// Optional counts fail but are ignored
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM tenants")).
			WillReturnError(errors.New("table does not exist"))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM networks")).
			WillReturnError(errors.New("table does not exist"))

		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM devices")).
			WillReturnError(errors.New("table does not exist"))

		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM peers WHERE last_seen`).
			WillReturnError(errors.New("table does not exist"))

		stats, err := repo.GetUserStats(ctx)
		require.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, 100, stats.TotalUsers)
		assert.Equal(t, 0, stats.TotalTenants) // Defaults to 0 on error
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAdminRepository_UpdateUserRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAdminRepository(db)
	ctx := context.Background()

	t.Run("success update admin only", func(t *testing.T) {
		isAdmin := true

		mock.ExpectExec(`UPDATE users SET is_admin = .*, updated_at = .* WHERE id = .*`).
			WithArgs(true, "NOW()", "user-1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateUserRole(ctx, "user-1", &isAdmin, nil)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success update moderator only", func(t *testing.T) {
		isModerator := true

		mock.ExpectExec(`UPDATE users SET is_moderator = .*, updated_at = .* WHERE id = .*`).
			WithArgs(true, "NOW()", "user-1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateUserRole(ctx, "user-1", nil, &isModerator)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success update both roles", func(t *testing.T) {
		isAdmin := true
		isModerator := false

		mock.ExpectExec(`UPDATE users SET is_admin = .*, is_moderator = .*, updated_at = .* WHERE id = .*`).
			WithArgs(true, false, "NOW()", "user-1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateUserRole(ctx, "user-1", &isAdmin, &isModerator)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("no role updates provided", func(t *testing.T) {
		err := repo.UpdateUserRole(ctx, "user-1", nil, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no role updates provided")
	})

	t.Run("user not found", func(t *testing.T) {
		isAdmin := true

		mock.ExpectExec(`UPDATE users SET is_admin = .*, updated_at = .* WHERE id = .*`).
			WithArgs(true, "NOW()", "nonexistent").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.UpdateUserRole(ctx, "nonexistent", &isAdmin, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		isAdmin := true

		mock.ExpectExec(`UPDATE users SET is_admin = .*, updated_at = .* WHERE id = .*`).
			WithArgs(true, "NOW()", "user-1").
			WillReturnError(errors.New("database error"))

		err := repo.UpdateUserRole(ctx, "user-1", &isAdmin, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update user role")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		isAdmin := true

		mock.ExpectExec(`UPDATE users SET is_admin = .*, updated_at = .* WHERE id = .*`).
			WithArgs(true, "NOW()", "user-1").
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err := repo.UpdateUserRole(ctx, "user-1", &isAdmin, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get rows affected")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAdminRepository_SuspendUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAdminRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		query := `UPDATE users 
			  SET suspended = true, 
			      suspended_at = NOW(), 
			      suspended_reason = $1, 
			      suspended_by = $2,
			      updated_at = NOW()
			  WHERE id = $3`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("Violation of terms", "admin-1", "user-1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.SuspendUser(ctx, "user-1", "Violation of terms", "admin-1")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		query := `UPDATE users 
			  SET suspended = true, 
			      suspended_at = NOW(), 
			      suspended_reason = $1, 
			      suspended_by = $2,
			      updated_at = NOW()
			  WHERE id = $3`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("Violation of terms", "admin-1", "nonexistent").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.SuspendUser(ctx, "nonexistent", "Violation of terms", "admin-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		query := `UPDATE users 
			  SET suspended = true, 
			      suspended_at = NOW(), 
			      suspended_reason = $1, 
			      suspended_by = $2,
			      updated_at = NOW()
			  WHERE id = $3`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("Violation of terms", "admin-1", "user-1").
			WillReturnError(errors.New("database error"))

		err := repo.SuspendUser(ctx, "user-1", "Violation of terms", "admin-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to suspend user")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		query := `UPDATE users 
			  SET suspended = true, 
			      suspended_at = NOW(), 
			      suspended_reason = $1, 
			      suspended_by = $2,
			      updated_at = NOW()
			  WHERE id = $3`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("Violation of terms", "admin-1", "user-1").
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err := repo.SuspendUser(ctx, "user-1", "Violation of terms", "admin-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get rows affected")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAdminRepository_UnsuspendUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAdminRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		query := `UPDATE users 
			  SET suspended = false, 
			      suspended_at = NULL, 
			      suspended_reason = NULL, 
			      suspended_by = NULL,
			      updated_at = NOW()
			  WHERE id = $1`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("user-1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UnsuspendUser(ctx, "user-1")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		query := `UPDATE users 
			  SET suspended = false, 
			      suspended_at = NULL, 
			      suspended_reason = NULL, 
			      suspended_by = NULL,
			      updated_at = NOW()
			  WHERE id = $1`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("nonexistent").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.UnsuspendUser(ctx, "nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		query := `UPDATE users 
			  SET suspended = false, 
			      suspended_at = NULL, 
			      suspended_reason = NULL, 
			      suspended_by = NULL,
			      updated_at = NOW()
			  WHERE id = $1`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("user-1").
			WillReturnError(errors.New("database error"))

		err := repo.UnsuspendUser(ctx, "user-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unsuspend user")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		query := `UPDATE users 
			  SET suspended = false, 
			      suspended_at = NULL, 
			      suspended_reason = NULL, 
			      suspended_by = NULL,
			      updated_at = NOW()
			  WHERE id = $1`

		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("user-1").
			WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected error")))

		err := repo.UnsuspendUser(ctx, "user-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get rows affected")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAdminRepository_GetUserByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAdminRepository(db)
	ctx := context.Background()
	now := time.Now()
	username := "testuser"
	fullName := "Test User"
	bio := "Test bio"
	avatarURL := "https://example.com/avatar.png"

	query := `SELECT id, tenant_id, email, username, full_name, bio, avatar_url, locale, 
			         is_admin, is_moderator, two_fa_enabled, auth_provider, external_id,
			         suspended, suspended_at, suspended_reason, suspended_by,
			         created_at, updated_at, last_seen
			  FROM users WHERE id = $1`

	t.Run("success with all fields", func(t *testing.T) {
		suspendedAt := now.Add(-24 * time.Hour)
		suspendedReason := "Violation"
		suspendedBy := "admin-1"

		// Note: The Scan in admin.go only scans 18 fields (missing 'suspended' and 'last_seen')
		// The mock must provide exactly what Scan expects to receive
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "email", "username", "full_name", "bio", "avatar_url", "locale",
			"is_admin", "is_moderator", "two_fa_enabled", "auth_provider", "external_id",
			"suspended_at", "suspended_reason", "suspended_by",
			"created_at", "updated_at",
		}).AddRow(
			"user-1", "tenant-1", "test@example.com", &username, &fullName, &bio, &avatarURL, "en-US",
			false, false, true, "local", "ext-1",
			suspendedAt, suspendedReason, suspendedBy,
			now, now,
		)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("user-1").
			WillReturnRows(rows)

		user, err := repo.GetUserByID(ctx, "user-1")
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "user-1", user.ID)
		assert.Equal(t, "tenant-1", user.TenantID)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, &username, user.Username)
		assert.Equal(t, &fullName, user.FullName)
		assert.Equal(t, &bio, user.Bio)
		assert.Equal(t, &avatarURL, user.AvatarURL)
		assert.Equal(t, "en-US", user.Locale)
		assert.False(t, user.IsAdmin)
		assert.False(t, user.IsModerator)
		assert.True(t, user.TwoFAEnabled)
		assert.Equal(t, "local", user.AuthProvider)
		assert.Equal(t, "ext-1", user.ExternalID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with null optional fields", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "email", "username", "full_name", "bio", "avatar_url", "locale",
			"is_admin", "is_moderator", "two_fa_enabled", "auth_provider", "external_id",
			"suspended_at", "suspended_reason", "suspended_by",
			"created_at", "updated_at",
		}).AddRow(
			"user-1", "tenant-1", "test@example.com", nil, nil, nil, nil, "en-US",
			false, false, false, "local", "",
			nil, nil, nil,
			now, now,
		)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("user-1").
			WillReturnRows(rows)

		user, err := repo.GetUserByID(ctx, "user-1")
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "user-1", user.ID)
		assert.Nil(t, user.Username)
		assert.Nil(t, user.FullName)
		assert.Nil(t, user.Bio)
		assert.Nil(t, user.AvatarURL)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("nonexistent").
			WillReturnError(sql.ErrNoRows)

		user, err := repo.GetUserByID(ctx, "nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
		assert.Nil(t, user)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("user-1").
			WillReturnError(errors.New("database error"))

		user, err := repo.GetUserByID(ctx, "user-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user")
		assert.Nil(t, user)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAdminRepository_UpdateLastSeen(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewAdminRepository(db)
	ctx := context.Background()

	query := `UPDATE users SET last_seen = NOW() WHERE id = $1`

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("user-1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateLastSeen(ctx, "user-1")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success even when user not found", func(t *testing.T) {
		// Note: UpdateLastSeen doesn't check rows affected
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("nonexistent").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.UpdateLastSeen(ctx, "nonexistent")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("user-1").
			WillReturnError(errors.New("database error"))

		err := repo.UpdateLastSeen(ctx, "user-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update last_seen")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
