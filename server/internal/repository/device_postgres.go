package repository

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// PostgresDeviceRepository implements DeviceRepository with PostgreSQL
type PostgresDeviceRepository struct {
	db *sql.DB
}

// NewPostgresDeviceRepository creates a new PostgreSQL device repository
func NewPostgresDeviceRepository(db *sql.DB) *PostgresDeviceRepository {
	return &PostgresDeviceRepository{
		db: db,
	}
}

// Create creates a new device
func (r *PostgresDeviceRepository) Create(ctx context.Context, device *domain.Device) error {
	if device.ID == "" {
		device.ID = domain.GenerateNetworkID()
	}

	if device.CreatedAt.IsZero() {
		device.CreatedAt = time.Now()
	}
	if device.UpdatedAt.IsZero() {
		device.UpdatedAt = device.CreatedAt
	}

	query := `
		INSERT INTO devices (
			id, user_id, tenant_id, name, platform, pubkey,
			last_seen, active, ip_address, daemon_ver, os_version,
			hostname, created_at, updated_at, disabled_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := r.db.ExecContext(ctx, query,
		device.ID, device.UserID, device.TenantID, device.Name, device.Platform,
		device.PubKey, device.LastSeen, device.Active, device.IPAddress,
		device.DaemonVer, device.OSVersion, device.HostName,
		device.CreatedAt, device.UpdatedAt, device.DisabledAt,
	)

	if err != nil {
		// Check for unique constraint violation on pubkey
		if err.Error() == "pq: duplicate key value violates unique constraint \"devices_pubkey_key\"" {
			return domain.NewError(domain.ErrConflict, "Device with this public key already exists", map[string]string{
				"pubkey": device.PubKey,
			})
		}
		return err
	}

	return nil
}

// GetByID retrieves a device by ID
func (r *PostgresDeviceRepository) GetByID(ctx context.Context, id string) (*domain.Device, error) {
	query := `
		SELECT id, user_id, tenant_id, name, platform, pubkey,
		       last_seen, active, ip_address, daemon_ver, os_version,
		       hostname, created_at, updated_at, disabled_at
		FROM devices
		WHERE id = $1
	`

	device := &domain.Device{}
	var disabledAt sql.NullTime
	var ipAddress, daemonVer, osVersion, hostname sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&device.ID, &device.UserID, &device.TenantID, &device.Name, &device.Platform,
		&device.PubKey, &device.LastSeen, &device.Active, &ipAddress,
		&daemonVer, &osVersion, &hostname,
		&device.CreatedAt, &device.UpdatedAt, &disabledAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.NewError(domain.ErrNotFound, "Device not found", map[string]string{
			"device_id": id,
		})
	}
	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if ipAddress.Valid {
		device.IPAddress = ipAddress.String
	}
	if daemonVer.Valid {
		device.DaemonVer = daemonVer.String
	}
	if osVersion.Valid {
		device.OSVersion = osVersion.String
	}
	if hostname.Valid {
		device.HostName = hostname.String
	}
	if disabledAt.Valid {
		device.DisabledAt = &disabledAt.Time
	}

	return device, nil
}

// GetByPubKey retrieves a device by public key
func (r *PostgresDeviceRepository) GetByPubKey(ctx context.Context, pubkey string) (*domain.Device, error) {
	query := `
		SELECT id, user_id, tenant_id, name, platform, pubkey,
		       last_seen, active, ip_address, daemon_ver, os_version,
		       hostname, created_at, updated_at, disabled_at
		FROM devices
		WHERE pubkey = $1
	`

	device := &domain.Device{}
	var disabledAt sql.NullTime
	var ipAddress, daemonVer, osVersion, hostname sql.NullString

	err := r.db.QueryRowContext(ctx, query, pubkey).Scan(
		&device.ID, &device.UserID, &device.TenantID, &device.Name, &device.Platform,
		&device.PubKey, &device.LastSeen, &device.Active, &ipAddress,
		&daemonVer, &osVersion, &hostname,
		&device.CreatedAt, &device.UpdatedAt, &disabledAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.NewError(domain.ErrNotFound, "Device not found", map[string]string{
			"pubkey": pubkey,
		})
	}
	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if ipAddress.Valid {
		device.IPAddress = ipAddress.String
	}
	if daemonVer.Valid {
		device.DaemonVer = daemonVer.String
	}
	if osVersion.Valid {
		device.OSVersion = osVersion.String
	}
	if hostname.Valid {
		device.HostName = hostname.String
	}
	if disabledAt.Valid {
		device.DisabledAt = &disabledAt.Time
	}

	return device, nil
}

// List retrieves devices matching the filter
func (r *PostgresDeviceRepository) List(ctx context.Context, filter domain.DeviceFilter) ([]*domain.Device, string, error) {
	query := `
		SELECT id, user_id, tenant_id, name, platform, pubkey,
		       last_seen, active, ip_address, daemon_ver, os_version,
		       hostname, created_at, updated_at, disabled_at
		FROM devices
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	// Apply filters
	if filter.UserID != "" {
		query += ` AND user_id = $` + intToString(argPos)
		args = append(args, filter.UserID)
		argPos++
	}

	if filter.TenantID != "" {
		query += ` AND tenant_id = $` + intToString(argPos)
		args = append(args, filter.TenantID)
		argPos++
	}

	if filter.Platform != "" {
		query += ` AND platform = $` + intToString(argPos)
		args = append(args, filter.Platform)
		argPos++
	}

	if filter.Active != nil {
		query += ` AND active = $` + intToString(argPos)
		args = append(args, *filter.Active)
		argPos++
	}

	// Cursor pagination
	if filter.Cursor != "" {
		query += ` AND created_at < (SELECT created_at FROM devices WHERE id = $` + intToString(argPos) + `)`
		args = append(args, filter.Cursor)
		argPos++
	}

	// Order and limit
	query += ` ORDER BY created_at DESC`

	limit := filter.Limit
	if limit == 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	// Fetch one extra to check if there are more results
	query += ` LIMIT $` + intToString(argPos)
	args = append(args, limit+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	devices := make([]*domain.Device, 0, limit)
	for rows.Next() {
		device := &domain.Device{}
		var disabledAt sql.NullTime
		var ipAddress, daemonVer, osVersion, hostname sql.NullString

		err := rows.Scan(
			&device.ID, &device.UserID, &device.TenantID, &device.Name, &device.Platform,
			&device.PubKey, &device.LastSeen, &device.Active, &ipAddress,
			&daemonVer, &osVersion, &hostname,
			&device.CreatedAt, &device.UpdatedAt, &disabledAt,
		)
		if err != nil {
			return nil, "", err
		}

		// Handle nullable fields
		if ipAddress.Valid {
			device.IPAddress = ipAddress.String
		}
		if daemonVer.Valid {
			device.DaemonVer = daemonVer.String
		}
		if osVersion.Valid {
			device.OSVersion = osVersion.String
		}
		if hostname.Valid {
			device.HostName = hostname.String
		}
		if disabledAt.Valid {
			device.DisabledAt = &disabledAt.Time
		}

		devices = append(devices, device)
	}

	// Check for next cursor
	var nextCursor string
	if len(devices) > limit {
		nextCursor = devices[limit-1].ID
		devices = devices[:limit]
	}

	return devices, nextCursor, nil
}

// Update updates an existing device
func (r *PostgresDeviceRepository) Update(ctx context.Context, device *domain.Device) error {
	device.UpdatedAt = time.Now()

	query := `
		UPDATE devices
		SET name = $1, platform = $2, pubkey = $3,
		    last_seen = $4, active = $5, ip_address = $6,
		    daemon_ver = $7, os_version = $8, hostname = $9,
		    updated_at = $10, disabled_at = $11
		WHERE id = $12
	`

	result, err := r.db.ExecContext(ctx, query,
		device.Name, device.Platform, device.PubKey,
		device.LastSeen, device.Active, device.IPAddress,
		device.DaemonVer, device.OSVersion, device.HostName,
		device.UpdatedAt, device.DisabledAt, device.ID,
	)
	if err != nil {
		// Check for unique constraint violation
		if err.Error() == "pq: duplicate key value violates unique constraint \"devices_pubkey_key\"" {
			return domain.NewError(domain.ErrConflict, "Device with this public key already exists", map[string]string{
				"pubkey": device.PubKey,
			})
		}
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Device not found", map[string]string{
			"device_id": device.ID,
		})
	}

	return nil
}

// Delete deletes a device
func (r *PostgresDeviceRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM devices WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Device not found", map[string]string{
			"device_id": id,
		})
	}

	return nil
}

// UpdateHeartbeat updates the last seen timestamp and marks device as active
func (r *PostgresDeviceRepository) UpdateHeartbeat(ctx context.Context, id string, ipAddress string) error {
	now := time.Now()

	query := `
		UPDATE devices
		SET last_seen = $1, active = TRUE, updated_at = $2
	`

	args := []interface{}{now, now}
	argPos := 3

	if ipAddress != "" {
		query += `, ip_address = $` + intToString(argPos)
		args = append(args, ipAddress)
		argPos++
	}

	query += ` WHERE id = $` + intToString(argPos)
	args = append(args, id)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Device not found", map[string]string{
			"device_id": id,
		})
	}

	return nil
}

// MarkInactive marks a device as inactive
func (r *PostgresDeviceRepository) MarkInactive(ctx context.Context, id string) error {
	now := time.Now()

	query := `UPDATE devices SET active = FALSE, updated_at = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "Device not found", map[string]string{
			"device_id": id,
		})
	}

	return nil
}

// Count returns the total number of devices
func (r *PostgresDeviceRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM devices").Scan(&count)
	if err != nil {
		return 0, domain.NewError(domain.ErrInternalServer, "Failed to count devices", map[string]string{"error": err.Error()})
	}
	return count, nil
}

// Helper function to convert int to string for query building
func intToString(i int) string {
	return strconv.Itoa(i)
}
