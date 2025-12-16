package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPostgresIPAMRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresIPAMRepository(db)
	require.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestPostgresIPAMRepository_GetOrAllocate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresIPAMRepository(db)
	ctx := context.Background()

	// Updated SQL queries to match the new device_id based schema
	existingQuery := `SELECT network_id, COALESCE(device_id, user_id) as subject_id, ip_address
		FROM ip_allocations
		WHERE network_id = $1 AND (device_id = $2 OR user_id = $2)`

	allocatedQuery := `SELECT ip_address FROM ip_allocations
		WHERE network_id = $1
		FOR UPDATE`

	insertQuery := `INSERT INTO ip_allocations (id, network_id, device_id, user_id, ip_address, allocated_at)
		VALUES ($1, $2, $3, $3, $4, NOW())`

	t.Run("returns existing allocation", func(t *testing.T) {
		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"network_id", "subject_id", "ip_address"}).
			AddRow("network-1", "user-1", "10.0.0.2")

		mock.ExpectQuery(regexp.QuoteMeta(existingQuery)).
			WithArgs("network-1", "user-1").
			WillReturnRows(rows)

		mock.ExpectCommit()

		allocation, err := repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/24")
		require.NoError(t, err)
		assert.Equal(t, "network-1", allocation.NetworkID)
		assert.Equal(t, "user-1", allocation.DeviceID) // subjectID is now stored in DeviceID
		assert.Equal(t, "10.0.0.2", allocation.IP)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("allocates new IP when none exists", func(t *testing.T) {
		mock.ExpectBegin()

		// No existing allocation
		mock.ExpectQuery(regexp.QuoteMeta(existingQuery)).
			WithArgs("network-1", "user-2").
			WillReturnError(sql.ErrNoRows)

		// Query allocated IPs (empty for this network)
		allocatedRows := sqlmock.NewRows([]string{"ip_address"})
		mock.ExpectQuery(regexp.QuoteMeta(allocatedQuery)).
			WithArgs("network-1").
			WillReturnRows(allocatedRows)

		// Insert new allocation - first available IP should be 10.0.0.2 (after network + gateway)
		mock.ExpectExec(regexp.QuoteMeta(insertQuery)).
			WithArgs(sqlmock.AnyArg(), "network-1", "user-2", "10.0.0.2").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		allocation, err := repo.GetOrAllocate(ctx, "network-1", "user-2", "10.0.0.0/24")
		require.NoError(t, err)
		assert.Equal(t, "network-1", allocation.NetworkID)
		assert.Equal(t, "user-2", allocation.DeviceID) // subjectID is now stored in DeviceID
		assert.Equal(t, "10.0.0.2", allocation.IP)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("allocates next available IP when some are used", func(t *testing.T) {
		mock.ExpectBegin()

		// No existing allocation
		mock.ExpectQuery(regexp.QuoteMeta(existingQuery)).
			WithArgs("network-1", "user-3").
			WillReturnError(sql.ErrNoRows)

		// Query allocated IPs - 10.0.0.2 and 10.0.0.3 are already used
		allocatedRows := sqlmock.NewRows([]string{"ip_address"}).
			AddRow("10.0.0.2").
			AddRow("10.0.0.3")
		mock.ExpectQuery(regexp.QuoteMeta(allocatedQuery)).
			WithArgs("network-1").
			WillReturnRows(allocatedRows)

		// Insert new allocation - next available should be 10.0.0.4
		mock.ExpectExec(regexp.QuoteMeta(insertQuery)).
			WithArgs(sqlmock.AnyArg(), "network-1", "user-3", "10.0.0.4").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		allocation, err := repo.GetOrAllocate(ctx, "network-1", "user-3", "10.0.0.0/24")
		require.NoError(t, err)
		assert.Equal(t, "10.0.0.4", allocation.IP)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("begin transaction error", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(errors.New("begin error"))

		allocation, err := repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/24")
		require.Error(t, err)
		assert.Nil(t, allocation)
		assert.Contains(t, err.Error(), "failed to begin transaction")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("check existing allocation error", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(existingQuery)).
			WithArgs("network-1", "user-1").
			WillReturnError(errors.New("database error"))

		mock.ExpectRollback()

		allocation, err := repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/24")
		require.Error(t, err)
		assert.Nil(t, allocation)
		assert.Contains(t, err.Error(), "failed to check existing allocation")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("invalid CIDR", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(existingQuery)).
			WithArgs("network-1", "user-1").
			WillReturnError(sql.ErrNoRows)

		mock.ExpectRollback()

		allocation, err := repo.GetOrAllocate(ctx, "network-1", "user-1", "invalid-cidr")
		require.Error(t, err)
		assert.Nil(t, allocation)
		assert.Contains(t, err.Error(), "invalid CIDR")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("CIDR with no usable IPs", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(existingQuery)).
			WithArgs("network-1", "user-1").
			WillReturnError(sql.ErrNoRows)

		mock.ExpectRollback()

		// /31 has only 2 IPs total, after reserving network and broadcast, no usable IPs
		allocation, err := repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/31")
		require.Error(t, err)
		assert.Nil(t, allocation)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrIPExhausted, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query allocated IPs error", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(existingQuery)).
			WithArgs("network-1", "user-1").
			WillReturnError(sql.ErrNoRows)

		mock.ExpectQuery(regexp.QuoteMeta(allocatedQuery)).
			WithArgs("network-1").
			WillReturnError(errors.New("query error"))

		mock.ExpectRollback()

		allocation, err := repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/24")
		require.Error(t, err)
		assert.Nil(t, allocation)
		assert.Contains(t, err.Error(), "failed to query allocations")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan allocated IP error", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(existingQuery)).
			WithArgs("network-1", "user-1").
			WillReturnError(sql.ErrNoRows)

		// Return invalid row data to trigger scan error
		allocatedRows := sqlmock.NewRows([]string{"ip_address", "extra_column"}).
			AddRow("10.0.0.2", "unexpected")
		mock.ExpectQuery(regexp.QuoteMeta(allocatedQuery)).
			WithArgs("network-1").
			WillReturnRows(allocatedRows)

		mock.ExpectRollback()

		allocation, err := repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/24")
		require.Error(t, err)
		assert.Nil(t, allocation)
		assert.Contains(t, err.Error(), "failed to scan allocated IP")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("insert allocation error", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(existingQuery)).
			WithArgs("network-1", "user-1").
			WillReturnError(sql.ErrNoRows)

		allocatedRows := sqlmock.NewRows([]string{"ip_address"})
		mock.ExpectQuery(regexp.QuoteMeta(allocatedQuery)).
			WithArgs("network-1").
			WillReturnRows(allocatedRows)

		mock.ExpectExec(regexp.QuoteMeta(insertQuery)).
			WithArgs(sqlmock.AnyArg(), "network-1", "user-1", "10.0.0.2").
			WillReturnError(errors.New("insert error"))

		mock.ExpectRollback()

		allocation, err := repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/24")
		require.Error(t, err)
		assert.Nil(t, allocation)
		assert.Contains(t, err.Error(), "failed to create allocation")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("commit error", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(existingQuery)).
			WithArgs("network-1", "user-1").
			WillReturnError(sql.ErrNoRows)

		allocatedRows := sqlmock.NewRows([]string{"ip_address"})
		mock.ExpectQuery(regexp.QuoteMeta(allocatedQuery)).
			WithArgs("network-1").
			WillReturnRows(allocatedRows)

		mock.ExpectExec(regexp.QuoteMeta(insertQuery)).
			WithArgs(sqlmock.AnyArg(), "network-1", "user-1", "10.0.0.2").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit().WillReturnError(errors.New("commit error"))

		allocation, err := repo.GetOrAllocate(ctx, "network-1", "user-1", "10.0.0.0/24")
		require.Error(t, err)
		assert.Nil(t, allocation)
		assert.Contains(t, err.Error(), "failed to commit transaction")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("all IPs exhausted", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(existingQuery)).
			WithArgs("network-1", "user-new").
			WillReturnError(sql.ErrNoRows)

		// For /30 network: 4 total IPs, 2 reserved (network + gateway), 1 broadcast = only 1 usable IP (offset 2)
		// usableStart = 2, usableEnd = 3 (totalHosts - 1 = 4 - 1 = 3)
		// So we only have offset 2 as usable
		allocatedRows := sqlmock.NewRows([]string{"ip_address"}).
			AddRow("10.0.0.2") // The only usable IP is already taken
		mock.ExpectQuery(regexp.QuoteMeta(allocatedQuery)).
			WithArgs("network-1").
			WillReturnRows(allocatedRows)

		mock.ExpectRollback()

		allocation, err := repo.GetOrAllocate(ctx, "network-1", "user-new", "10.0.0.0/30")
		require.Error(t, err)
		assert.Nil(t, allocation)
		var domainErr *domain.Error
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, domain.ErrIPExhausted, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresIPAMRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresIPAMRepository(db)
	ctx := context.Background()

	// Query now includes device_id column
	query := `SELECT network_id, device_id, user_id, ip_address
		FROM ip_allocations
		WHERE network_id = $1
		ORDER BY allocated_at ASC`

	t.Run("success with multiple allocations", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"network_id", "device_id", "user_id", "ip_address"}).
			AddRow("network-1", "user-1", "user-1", "10.0.0.2").
			AddRow("network-1", "user-2", "user-2", "10.0.0.3").
			AddRow("network-1", "user-3", "user-3", "10.0.0.4")

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("network-1").
			WillReturnRows(rows)

		allocations, err := repo.List(ctx, "network-1")
		require.NoError(t, err)
		require.Len(t, allocations, 3)
		assert.Equal(t, "user-1", allocations[0].DeviceID)
		assert.Equal(t, "10.0.0.2", allocations[0].IP)
		assert.Equal(t, "user-2", allocations[1].DeviceID)
		assert.Equal(t, "10.0.0.3", allocations[1].IP)
		assert.Equal(t, "user-3", allocations[2].DeviceID)
		assert.Equal(t, "10.0.0.4", allocations[2].IP)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with empty result", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"network_id", "device_id", "user_id", "ip_address"})

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("network-2").
			WillReturnRows(rows)

		allocations, err := repo.List(ctx, "network-2")
		require.NoError(t, err)
		assert.Empty(t, allocations)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("network-1").
			WillReturnError(errors.New("database error"))

		allocations, err := repo.List(ctx, "network-1")
		require.Error(t, err)
		assert.Nil(t, allocations)
		assert.Contains(t, err.Error(), "failed to list allocations")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		// Return wrong number of columns to trigger scan error
		rows := sqlmock.NewRows([]string{"network_id"}).
			AddRow("network-1")

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs("network-1").
			WillReturnRows(rows)

		allocations, err := repo.List(ctx, "network-1")
		require.Error(t, err)
		assert.Nil(t, allocations)
		assert.Contains(t, err.Error(), "failed to scan allocation")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresIPAMRepository_Release(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresIPAMRepository(db)
	ctx := context.Background()

	// Release query now checks both device_id and user_id
	query := `DELETE FROM ip_allocations WHERE network_id = $1 AND (device_id = $2 OR user_id = $2)`

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("network-1", "user-1").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Release(ctx, "network-1", "user-1")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success when no allocation exists (idempotent)", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("network-1", "user-non-existent").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Release(ctx, "network-1", "user-non-existent")
		require.NoError(t, err) // Should not return error due to idempotency
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("exec error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(query)).
			WithArgs("network-1", "user-1").
			WillReturnError(errors.New("database error"))

		err := repo.Release(ctx, "network-1", "user-1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to release allocation")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
