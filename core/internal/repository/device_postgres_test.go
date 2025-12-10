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

func TestNewPostgresDeviceRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresDeviceRepository(db)
	require.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestPostgresDeviceRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresDeviceRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		device := &domain.Device{
			ID:        "device-123",
			UserID:    "user-123",
			TenantID:  "tenant-123",
			Name:      "Test Device",
			Platform:  "linux",
			PubKey:    "test-pubkey-12345678901234567890123456789012",
			Active:    true,
			IPAddress: "192.168.1.1",
			DaemonVer: "1.0.0",
			OSVersion: "Ubuntu 22.04",
			HostName:  "test-host",
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO devices`)).
			WithArgs(
				device.ID, device.UserID, device.TenantID, device.Name, device.Platform,
				device.PubKey, sqlmock.AnyArg(), device.Active, device.IPAddress,
				device.DaemonVer, device.OSVersion, device.HostName,
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, device)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with auto-generated ID", func(t *testing.T) {
		device := &domain.Device{
			UserID:   "user-123",
			TenantID: "tenant-123",
			Name:     "Test Device",
			Platform: "linux",
			PubKey:   "test-pubkey-12345678901234567890123456789012",
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO devices`)).
			WithArgs(
				sqlmock.AnyArg(), device.UserID, device.TenantID, device.Name, device.Platform,
				device.PubKey, sqlmock.AnyArg(), device.Active, device.IPAddress,
				device.DaemonVer, device.OSVersion, device.HostName,
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, device)
		require.NoError(t, err)
		assert.NotEmpty(t, device.ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate pubkey error", func(t *testing.T) {
		device := &domain.Device{
			ID:       "device-123",
			UserID:   "user-123",
			TenantID: "tenant-123",
			Name:     "Test Device",
			Platform: "linux",
			PubKey:   "duplicate-pubkey-1234567890123456789012",
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO devices`)).
			WithArgs(
				device.ID, device.UserID, device.TenantID, device.Name, device.Platform,
				device.PubKey, sqlmock.AnyArg(), device.Active, device.IPAddress,
				device.DaemonVer, device.OSVersion, device.HostName,
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnError(errors.New("pq: duplicate key value violates unique constraint \"devices_pubkey_key\""))

		err := repo.Create(ctx, device)
		require.Error(t, err)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrConflict, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		device := &domain.Device{
			ID:       "device-123",
			UserID:   "user-123",
			TenantID: "tenant-123",
			Name:     "Test Device",
			Platform: "linux",
			PubKey:   "test-pubkey-12345678901234567890123456789012",
		}

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO devices`)).
			WithArgs(
				device.ID, device.UserID, device.TenantID, device.Name, device.Platform,
				device.PubKey, sqlmock.AnyArg(), device.Active, device.IPAddress,
				device.DaemonVer, device.OSVersion, device.HostName,
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnError(errors.New("database connection error"))

		err := repo.Create(ctx, device)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database connection error")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresDeviceRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresDeviceRepository(db)
	ctx := context.Background()
	now := time.Now()

	columns := []string{
		"id", "user_id", "tenant_id", "name", "platform", "pubkey",
		"last_seen", "active", "ip_address", "daemon_ver", "os_version",
		"hostname", "created_at", "updated_at", "disabled_at",
	}

	t.Run("success", func(t *testing.T) {
		deviceID := "device-123"

		rows := sqlmock.NewRows(columns).
			AddRow(
				deviceID, "user-123", "tenant-123", "Test Device", "linux",
				"test-pubkey", now, true, "192.168.1.1", "1.0.0", "Ubuntu 22.04",
				"test-host", now, now, nil,
			)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, tenant_id, name, platform, pubkey`)).
			WithArgs(deviceID).
			WillReturnRows(rows)

		device, err := repo.GetByID(ctx, deviceID)
		require.NoError(t, err)
		require.NotNil(t, device)
		assert.Equal(t, deviceID, device.ID)
		assert.Equal(t, "user-123", device.UserID)
		assert.Equal(t, "tenant-123", device.TenantID)
		assert.Equal(t, "Test Device", device.Name)
		assert.Equal(t, "linux", device.Platform)
		assert.Equal(t, "test-pubkey", device.PubKey)
		assert.True(t, device.Active)
		assert.Equal(t, "192.168.1.1", device.IPAddress)
		assert.Equal(t, "1.0.0", device.DaemonVer)
		assert.Equal(t, "Ubuntu 22.04", device.OSVersion)
		assert.Equal(t, "test-host", device.HostName)
		assert.Nil(t, device.DisabledAt)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with disabled device", func(t *testing.T) {
		deviceID := "device-disabled"
		disabledAt := now.Add(-24 * time.Hour)

		rows := sqlmock.NewRows(columns).
			AddRow(
				deviceID, "user-123", "tenant-123", "Disabled Device", "linux",
				"test-pubkey", now, false, "192.168.1.1", "1.0.0", "Ubuntu 22.04",
				"test-host", now, now, disabledAt,
			)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, tenant_id, name, platform, pubkey`)).
			WithArgs(deviceID).
			WillReturnRows(rows)

		device, err := repo.GetByID(ctx, deviceID)
		require.NoError(t, err)
		require.NotNil(t, device)
		require.NotNil(t, device.DisabledAt)
		assert.Equal(t, disabledAt.Unix(), device.DisabledAt.Unix())
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with null optional fields", func(t *testing.T) {
		deviceID := "device-minimal"

		rows := sqlmock.NewRows(columns).
			AddRow(
				deviceID, "user-123", "tenant-123", "Minimal Device", "linux",
				"test-pubkey", now, true, nil, nil, nil,
				nil, now, now, nil,
			)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, tenant_id, name, platform, pubkey`)).
			WithArgs(deviceID).
			WillReturnRows(rows)

		device, err := repo.GetByID(ctx, deviceID)
		require.NoError(t, err)
		require.NotNil(t, device)
		assert.Empty(t, device.IPAddress)
		assert.Empty(t, device.DaemonVer)
		assert.Empty(t, device.OSVersion)
		assert.Empty(t, device.HostName)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		deviceID := "non-existent"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, tenant_id, name, platform, pubkey`)).
			WithArgs(deviceID).
			WillReturnError(sql.ErrNoRows)

		device, err := repo.GetByID(ctx, deviceID)
		require.Error(t, err)
		assert.Nil(t, device)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		deviceID := "device-123"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, tenant_id, name, platform, pubkey`)).
			WithArgs(deviceID).
			WillReturnError(errors.New("database error"))

		device, err := repo.GetByID(ctx, deviceID)
		require.Error(t, err)
		assert.Nil(t, device)
		assert.Contains(t, err.Error(), "database error")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresDeviceRepository_GetByPubKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresDeviceRepository(db)
	ctx := context.Background()
	now := time.Now()

	columns := []string{
		"id", "user_id", "tenant_id", "name", "platform", "pubkey",
		"last_seen", "active", "ip_address", "daemon_ver", "os_version",
		"hostname", "created_at", "updated_at", "disabled_at",
	}

	t.Run("success", func(t *testing.T) {
		pubkey := "test-pubkey-12345678901234567890123456789012"

		rows := sqlmock.NewRows(columns).
			AddRow(
				"device-123", "user-123", "tenant-123", "Test Device", "linux",
				pubkey, now, true, "192.168.1.1", "1.0.0", "Ubuntu 22.04",
				"test-host", now, now, nil,
			)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, tenant_id, name, platform, pubkey`)).
			WithArgs(pubkey).
			WillReturnRows(rows)

		device, err := repo.GetByPubKey(ctx, pubkey)
		require.NoError(t, err)
		require.NotNil(t, device)
		assert.Equal(t, "device-123", device.ID)
		assert.Equal(t, pubkey, device.PubKey)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		pubkey := "non-existent-pubkey"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, tenant_id, name, platform, pubkey`)).
			WithArgs(pubkey).
			WillReturnError(sql.ErrNoRows)

		device, err := repo.GetByPubKey(ctx, pubkey)
		require.Error(t, err)
		assert.Nil(t, device)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		pubkey := "test-pubkey"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, tenant_id, name, platform, pubkey`)).
			WithArgs(pubkey).
			WillReturnError(errors.New("database error"))

		device, err := repo.GetByPubKey(ctx, pubkey)
		require.Error(t, err)
		assert.Nil(t, device)
		assert.Contains(t, err.Error(), "database error")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresDeviceRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresDeviceRepository(db)
	ctx := context.Background()
	now := time.Now()

	columns := []string{
		"id", "user_id", "tenant_id", "name", "platform", "pubkey",
		"last_seen", "active", "ip_address", "daemon_ver", "os_version",
		"hostname", "created_at", "updated_at", "disabled_at",
	}

	t.Run("success with no filters", func(t *testing.T) {
		filter := domain.DeviceFilter{}

		rows := sqlmock.NewRows(columns).
			AddRow(
				"device-1", "user-123", "tenant-123", "Device 1", "linux",
				"pubkey-1", now, true, "192.168.1.1", "1.0.0", "Ubuntu 22.04",
				"host-1", now, now, nil,
			).
			AddRow(
				"device-2", "user-123", "tenant-123", "Device 2", "windows",
				"pubkey-2", now, false, "192.168.1.2", "1.0.1", "Windows 11",
				"host-2", now, now, nil,
			)

		mock.ExpectQuery(`SELECT id, user_id, tenant_id, name, platform, pubkey`).
			WillReturnRows(rows)

		devices, cursor, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, devices, 2)
		assert.Empty(t, cursor)
		assert.Equal(t, "device-1", devices[0].ID)
		assert.Equal(t, "device-2", devices[1].ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with user filter", func(t *testing.T) {
		filter := domain.DeviceFilter{UserID: "user-123"}

		rows := sqlmock.NewRows(columns).
			AddRow(
				"device-1", "user-123", "tenant-123", "Device 1", "linux",
				"pubkey-1", now, true, "192.168.1.1", "1.0.0", "Ubuntu 22.04",
				"host-1", now, now, nil,
			)

		mock.ExpectQuery(`SELECT id, user_id, tenant_id, name, platform, pubkey`).
			WithArgs("user-123", sqlmock.AnyArg()).
			WillReturnRows(rows)

		devices, cursor, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, devices, 1)
		assert.Empty(t, cursor)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with tenant filter", func(t *testing.T) {
		filter := domain.DeviceFilter{TenantID: "tenant-123"}

		rows := sqlmock.NewRows(columns).
			AddRow(
				"device-1", "user-123", "tenant-123", "Device 1", "linux",
				"pubkey-1", now, true, "192.168.1.1", "1.0.0", "Ubuntu 22.04",
				"host-1", now, now, nil,
			)

		mock.ExpectQuery(`SELECT id, user_id, tenant_id, name, platform, pubkey`).
			WithArgs("tenant-123", sqlmock.AnyArg()).
			WillReturnRows(rows)

		devices, cursor, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, devices, 1)
		assert.Empty(t, cursor)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with platform filter", func(t *testing.T) {
		filter := domain.DeviceFilter{Platform: "linux"}

		rows := sqlmock.NewRows(columns).
			AddRow(
				"device-1", "user-123", "tenant-123", "Device 1", "linux",
				"pubkey-1", now, true, "192.168.1.1", "1.0.0", "Ubuntu 22.04",
				"host-1", now, now, nil,
			)

		mock.ExpectQuery(`SELECT id, user_id, tenant_id, name, platform, pubkey`).
			WithArgs("linux", sqlmock.AnyArg()).
			WillReturnRows(rows)

		devices, cursor, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, devices, 1)
		assert.Equal(t, "linux", devices[0].Platform)
		assert.Empty(t, cursor)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with active filter", func(t *testing.T) {
		active := true
		filter := domain.DeviceFilter{Active: &active}

		rows := sqlmock.NewRows(columns).
			AddRow(
				"device-1", "user-123", "tenant-123", "Device 1", "linux",
				"pubkey-1", now, true, "192.168.1.1", "1.0.0", "Ubuntu 22.04",
				"host-1", now, now, nil,
			)

		mock.ExpectQuery(`SELECT id, user_id, tenant_id, name, platform, pubkey`).
			WithArgs(true, sqlmock.AnyArg()).
			WillReturnRows(rows)

		devices, cursor, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, devices, 1)
		assert.True(t, devices[0].Active)
		assert.Empty(t, cursor)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with search filter", func(t *testing.T) {
		filter := domain.DeviceFilter{Search: "test"}

		rows := sqlmock.NewRows(columns).
			AddRow(
				"device-1", "user-123", "tenant-123", "Test Device", "linux",
				"pubkey-1", now, true, "192.168.1.1", "1.0.0", "Ubuntu 22.04",
				"host-1", now, now, nil,
			)

		mock.ExpectQuery(`SELECT id, user_id, tenant_id, name, platform, pubkey`).
			WithArgs("%test%", sqlmock.AnyArg()).
			WillReturnRows(rows)

		devices, cursor, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, devices, 1)
		assert.Empty(t, cursor)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with pagination cursor", func(t *testing.T) {
		filter := domain.DeviceFilter{
			Cursor: "device-cursor",
			Limit:  10,
		}

		rows := sqlmock.NewRows(columns).
			AddRow(
				"device-1", "user-123", "tenant-123", "Device 1", "linux",
				"pubkey-1", now, true, "192.168.1.1", "1.0.0", "Ubuntu 22.04",
				"host-1", now, now, nil,
			)

		mock.ExpectQuery(`SELECT id, user_id, tenant_id, name, platform, pubkey`).
			WithArgs("device-cursor", sqlmock.AnyArg()).
			WillReturnRows(rows)

		devices, cursor, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, devices, 1)
		assert.Empty(t, cursor)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with next cursor when more results available", func(t *testing.T) {
		filter := domain.DeviceFilter{Limit: 2}

		// Return 3 rows (limit + 1) to indicate more results available
		rows := sqlmock.NewRows(columns).
			AddRow(
				"device-1", "user-123", "tenant-123", "Device 1", "linux",
				"pubkey-1", now, true, nil, nil, nil, nil, now, now, nil,
			).
			AddRow(
				"device-2", "user-123", "tenant-123", "Device 2", "linux",
				"pubkey-2", now, true, nil, nil, nil, nil, now, now, nil,
			).
			AddRow(
				"device-3", "user-123", "tenant-123", "Device 3", "linux",
				"pubkey-3", now, true, nil, nil, nil, nil, now, now, nil,
			)

		mock.ExpectQuery(`SELECT id, user_id, tenant_id, name, platform, pubkey`).
			WillReturnRows(rows)

		devices, cursor, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, devices, 2)
		assert.Equal(t, "device-2", cursor)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("limit defaults to 50", func(t *testing.T) {
		filter := domain.DeviceFilter{Limit: 0}

		rows := sqlmock.NewRows(columns)

		mock.ExpectQuery(`SELECT id, user_id, tenant_id, name, platform, pubkey`).
			WithArgs(51). // Default limit of 50 + 1
			WillReturnRows(rows)

		devices, cursor, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, devices, 0)
		assert.Empty(t, cursor)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("limit capped at 100", func(t *testing.T) {
		filter := domain.DeviceFilter{Limit: 200}

		rows := sqlmock.NewRows(columns)

		mock.ExpectQuery(`SELECT id, user_id, tenant_id, name, platform, pubkey`).
			WithArgs(101). // Capped limit of 100 + 1
			WillReturnRows(rows)

		devices, cursor, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, devices, 0)
		assert.Empty(t, cursor)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error", func(t *testing.T) {
		filter := domain.DeviceFilter{}

		mock.ExpectQuery(`SELECT id, user_id, tenant_id, name, platform, pubkey`).
			WillReturnError(errors.New("database error"))

		devices, cursor, err := repo.List(ctx, filter)
		require.Error(t, err)
		assert.Nil(t, devices)
		assert.Empty(t, cursor)
		assert.Contains(t, err.Error(), "database error")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		filter := domain.DeviceFilter{}

		// Return wrong number of columns to trigger scan error
		rows := sqlmock.NewRows([]string{"id", "user_id"}).
			AddRow("device-1", "user-123")

		mock.ExpectQuery(`SELECT id, user_id, tenant_id, name, platform, pubkey`).
			WillReturnRows(rows)

		devices, cursor, err := repo.List(ctx, filter)
		require.Error(t, err)
		assert.Nil(t, devices)
		assert.Empty(t, cursor)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresDeviceRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresDeviceRepository(db)
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		device := &domain.Device{
			ID:        "device-123",
			Name:      "Updated Device",
			Platform:  "linux",
			PubKey:    "updated-pubkey",
			LastSeen:  now,
			Active:    true,
			IPAddress: "192.168.1.100",
			DaemonVer: "2.0.0",
			OSVersion: "Ubuntu 24.04",
			HostName:  "updated-host",
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE devices`)).
			WithArgs(
				device.Name, device.Platform, device.PubKey,
				device.LastSeen, device.Active, device.IPAddress,
				device.DaemonVer, device.OSVersion, device.HostName,
				sqlmock.AnyArg(), device.DisabledAt, device.ID,
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(ctx, device)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		device := &domain.Device{
			ID:       "non-existent",
			Name:     "Test Device",
			Platform: "linux",
			PubKey:   "test-pubkey",
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE devices`)).
			WithArgs(
				device.Name, device.Platform, device.PubKey,
				device.LastSeen, device.Active, device.IPAddress,
				device.DaemonVer, device.OSVersion, device.HostName,
				sqlmock.AnyArg(), device.DisabledAt, device.ID,
			).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Update(ctx, device)
		require.Error(t, err)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate pubkey error", func(t *testing.T) {
		device := &domain.Device{
			ID:       "device-123",
			Name:     "Test Device",
			Platform: "linux",
			PubKey:   "duplicate-pubkey",
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE devices`)).
			WithArgs(
				device.Name, device.Platform, device.PubKey,
				device.LastSeen, device.Active, device.IPAddress,
				device.DaemonVer, device.OSVersion, device.HostName,
				sqlmock.AnyArg(), device.DisabledAt, device.ID,
			).
			WillReturnError(errors.New("pq: duplicate key value violates unique constraint \"devices_pubkey_key\""))

		err := repo.Update(ctx, device)
		require.Error(t, err)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrConflict, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		device := &domain.Device{
			ID:       "device-123",
			Name:     "Test Device",
			Platform: "linux",
			PubKey:   "test-pubkey",
		}

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE devices`)).
			WithArgs(
				device.Name, device.Platform, device.PubKey,
				device.LastSeen, device.Active, device.IPAddress,
				device.DaemonVer, device.OSVersion, device.HostName,
				sqlmock.AnyArg(), device.DisabledAt, device.ID,
			).
			WillReturnError(errors.New("database error"))

		err := repo.Update(ctx, device)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresDeviceRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresDeviceRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		deviceID := "device-123"

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM devices WHERE id = $1`)).
			WithArgs(deviceID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(ctx, deviceID)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		deviceID := "non-existent"

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM devices WHERE id = $1`)).
			WithArgs(deviceID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(ctx, deviceID)
		require.Error(t, err)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		deviceID := "device-123"

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM devices WHERE id = $1`)).
			WithArgs(deviceID).
			WillReturnError(errors.New("database error"))

		err := repo.Delete(ctx, deviceID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresDeviceRepository_UpdateHeartbeat(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresDeviceRepository(db)
	ctx := context.Background()

	t.Run("success with IP address", func(t *testing.T) {
		deviceID := "device-123"
		ipAddress := "192.168.1.100"

		mock.ExpectExec(`UPDATE devices`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), ipAddress, deviceID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateHeartbeat(ctx, deviceID, ipAddress)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success without IP address", func(t *testing.T) {
		deviceID := "device-123"
		ipAddress := ""

		mock.ExpectExec(`UPDATE devices`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), deviceID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateHeartbeat(ctx, deviceID, ipAddress)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		deviceID := "non-existent"
		ipAddress := "192.168.1.100"

		mock.ExpectExec(`UPDATE devices`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), ipAddress, deviceID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.UpdateHeartbeat(ctx, deviceID, ipAddress)
		require.Error(t, err)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		deviceID := "device-123"
		ipAddress := "192.168.1.100"

		mock.ExpectExec(`UPDATE devices`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), ipAddress, deviceID).
			WillReturnError(errors.New("database error"))

		err := repo.UpdateHeartbeat(ctx, deviceID, ipAddress)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresDeviceRepository_MarkInactive(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresDeviceRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		deviceID := "device-123"

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE devices SET active = FALSE, updated_at = $1 WHERE id = $2`)).
			WithArgs(sqlmock.AnyArg(), deviceID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.MarkInactive(ctx, deviceID)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		deviceID := "non-existent"

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE devices SET active = FALSE, updated_at = $1 WHERE id = $2`)).
			WithArgs(sqlmock.AnyArg(), deviceID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.MarkInactive(ctx, deviceID)
		require.Error(t, err)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrNotFound, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		deviceID := "device-123"

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE devices SET active = FALSE, updated_at = $1 WHERE id = $2`)).
			WithArgs(sqlmock.AnyArg(), deviceID).
			WillReturnError(errors.New("database error"))

		err := repo.MarkInactive(ctx, deviceID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresDeviceRepository_Count(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresDeviceRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM devices")).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(42))

		count, err := repo.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, 42, count)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM devices")).
			WillReturnError(errors.New("database error"))

		count, err := repo.Count(ctx)
		require.Error(t, err)
		assert.Equal(t, 0, count)
		domainErr, ok := err.(*domain.Error)
		require.True(t, ok)
		assert.Equal(t, domain.ErrInternalServer, domainErr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresDeviceRepository_GetStaleDevices(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresDeviceRepository(db)
	ctx := context.Background()
	now := time.Now()
	threshold := 5 * time.Minute

	columns := []string{
		"id", "user_id", "tenant_id", "name", "platform", "pubkey",
		"last_seen", "active", "ip_address", "daemon_ver", "os_version",
		"hostname", "created_at", "updated_at", "disabled_at",
	}

	t.Run("success with stale devices", func(t *testing.T) {
		rows := sqlmock.NewRows(columns).
			AddRow(
				"device-1", "user-123", "tenant-123", "Stale Device 1", "linux",
				"pubkey-1", now.Add(-10*time.Minute), true, "192.168.1.1", "1.0.0", "Ubuntu 22.04",
				"host-1", now, now, nil,
			).
			AddRow(
				"device-2", "user-456", "tenant-123", "Stale Device 2", "windows",
				"pubkey-2", now.Add(-15*time.Minute), true, "192.168.1.2", "1.0.1", "Windows 11",
				"host-2", now, now, nil,
			)

		mock.ExpectQuery(`SELECT.*id, user_id, tenant_id, name, platform, pubkey.*FROM devices.*WHERE active = TRUE AND last_seen`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(rows)

		devices, err := repo.GetStaleDevices(ctx, threshold)
		require.NoError(t, err)
		assert.Len(t, devices, 2)
		assert.Equal(t, "device-1", devices[0].ID)
		assert.Equal(t, "device-2", devices[1].ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with no stale devices", func(t *testing.T) {
		rows := sqlmock.NewRows(columns)

		mock.ExpectQuery(`SELECT.*id, user_id, tenant_id, name, platform, pubkey.*FROM devices.*WHERE active = TRUE AND last_seen`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(rows)

		devices, err := repo.GetStaleDevices(ctx, threshold)
		require.NoError(t, err)
		assert.Len(t, devices, 0)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT.*id, user_id, tenant_id, name, platform, pubkey.*FROM devices.*WHERE active = TRUE AND last_seen`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnError(errors.New("database error"))

		devices, err := repo.GetStaleDevices(ctx, threshold)
		require.Error(t, err)
		assert.Nil(t, devices)
		assert.Contains(t, err.Error(), "database error")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		// Return wrong number of columns to trigger scan error
		rows := sqlmock.NewRows([]string{"id", "user_id"}).
			AddRow("device-1", "user-123")

		mock.ExpectQuery(`SELECT.*id, user_id, tenant_id, name, platform, pubkey.*FROM devices.*WHERE active = TRUE AND last_seen`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(rows)

		devices, err := repo.GetStaleDevices(ctx, threshold)
		require.Error(t, err)
		assert.Nil(t, devices)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows error", func(t *testing.T) {
		rows := sqlmock.NewRows(columns).
			AddRow(
				"device-1", "user-123", "tenant-123", "Stale Device 1", "linux",
				"pubkey-1", now.Add(-10*time.Minute), true, "192.168.1.1", "1.0.0", "Ubuntu 22.04",
				"host-1", now, now, nil,
			).
			RowError(0, errors.New("row error"))

		mock.ExpectQuery(`SELECT.*id, user_id, tenant_id, name, platform, pubkey.*FROM devices.*WHERE active = TRUE AND last_seen`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(rows)

		devices, err := repo.GetStaleDevices(ctx, threshold)
		require.Error(t, err)
		assert.Nil(t, devices)
		assert.Contains(t, err.Error(), "row error")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestIntToString(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{10, "10"},
		{100, "100"},
		{-1, "-1"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := intToString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
