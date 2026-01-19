package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// SQLiteDeviceRepository implements DeviceRepository for SQLite.
type SQLiteDeviceRepository struct {
	db *sql.DB
}

func NewSQLiteDeviceRepository(db *sql.DB) *SQLiteDeviceRepository {
	return &SQLiteDeviceRepository{db: db}
}

func (r *SQLiteDeviceRepository) Create(ctx context.Context, device *domain.Device) error {
	now := time.Now()
	if device.ID == "" {
		device.ID = domain.GenerateNetworkID()
	}
	if device.CreatedAt.IsZero() {
		device.CreatedAt = now
	}
	if device.UpdatedAt.IsZero() {
		device.UpdatedAt = device.CreatedAt
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO devices (
			id, user_id, tenant_id, name, platform, pubkey, last_seen, active,
			ip_address, daemon_ver, os_version, hostname, created_at, updated_at, disabled_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, device.ID, device.UserID, device.TenantID, device.Name, device.Platform, device.PubKey,
		device.LastSeen, device.Active, nullIfBlank(device.IPAddress), nullIfBlank(device.DaemonVer),
		nullIfBlank(device.OSVersion), nullIfBlank(device.HostName), device.CreatedAt, device.UpdatedAt, nullableTime(device.DisabledAt))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return domain.NewError(domain.ErrConflict, "Device with this public key already exists", map[string]string{
				"pubkey": device.PubKey,
			})
		}
		return fmt.Errorf("failed to create device: %w", err)
	}
	return nil
}

func (r *SQLiteDeviceRepository) GetByID(ctx context.Context, id string) (*domain.Device, error) {
	query := `
		SELECT id, user_id, tenant_id, name, platform, pubkey, last_seen, active,
		       ip_address, daemon_ver, os_version, hostname, created_at, updated_at, disabled_at
		FROM devices
		WHERE id = ?
	`
	return r.scanDevice(ctx, query, id)
}

func (r *SQLiteDeviceRepository) GetByPubKey(ctx context.Context, pubkey string) (*domain.Device, error) {
	query := `
		SELECT id, user_id, tenant_id, name, platform, pubkey, last_seen, active,
		       ip_address, daemon_ver, os_version, hostname, created_at, updated_at, disabled_at
		FROM devices
		WHERE pubkey = ?
		LIMIT 1
	`
	return r.scanDevice(ctx, query, pubkey)
}

func (r *SQLiteDeviceRepository) List(ctx context.Context, filter domain.DeviceFilter) ([]*domain.Device, string, error) {
	args := []interface{}{}
	query := `
		SELECT id, user_id, tenant_id, name, platform, pubkey, last_seen, active,
		       ip_address, daemon_ver, os_version, hostname, created_at, updated_at, disabled_at
		FROM devices
		WHERE 1=1
	`
	if filter.UserID != "" {
		query += " AND user_id = ?"
		args = append(args, filter.UserID)
	}
	if filter.TenantID != "" {
		query += " AND tenant_id = ?"
		args = append(args, filter.TenantID)
	}
	if filter.Platform != "" {
		query += " AND platform = ?"
		args = append(args, filter.Platform)
	}
	if filter.Active != nil {
		query += " AND active = ?"
		args = append(args, *filter.Active)
	}
	if filter.Search != "" {
		query += " AND (LOWER(name) LIKE ? OR LOWER(hostname) LIKE ?)"
		term := "%" + strings.ToLower(filter.Search) + "%"
		args = append(args, term, term)
	}
	if filter.Cursor != "" {
		query += " AND id < ?"
		args = append(args, filter.Cursor)
	}
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	query += " ORDER BY id DESC LIMIT ?"
	args = append(args, filter.Limit+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list devices: %w", err)
	}
	defer rows.Close()

	var result []*domain.Device
	for rows.Next() {
		dev, err := scanDeviceRow(rows)
		if err != nil {
			return nil, "", err
		}
		result = append(result, dev)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("failed to iterate devices: %w", err)
	}
	next := ""
	if len(result) > filter.Limit {
		// Return the last item's ID from the page as cursor for next page
		next = result[filter.Limit-1].ID
		result = result[:filter.Limit]
	}
	return result, next, nil
}

func (r *SQLiteDeviceRepository) Update(ctx context.Context, device *domain.Device) error {
	device.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx, `
		UPDATE devices
		SET name = ?, platform = ?, pubkey = ?, last_seen = ?, active = ?, ip_address = ?, daemon_ver = ?, os_version = ?, hostname = ?, updated_at = ?, disabled_at = ?
		WHERE id = ?
	`, device.Name, device.Platform, device.PubKey, device.LastSeen, device.Active, nullIfBlank(device.IPAddress),
		nullIfBlank(device.DaemonVer), nullIfBlank(device.OSVersion), nullIfBlank(device.HostName), device.UpdatedAt, nullableTime(device.DisabledAt), device.ID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return domain.NewError(domain.ErrConflict, "Device with this public key already exists", map[string]string{
				"pubkey": device.PubKey,
			})
		}
		return fmt.Errorf("failed to update device: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if affected == 0 {
		return domain.NewError(domain.ErrNotFound, "Device not found", map[string]string{"device_id": device.ID})
	}
	return nil
}

func (r *SQLiteDeviceRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM devices WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if affected == 0 {
		return domain.NewError(domain.ErrNotFound, "Device not found", map[string]string{"device_id": id})
	}
	return nil
}

func (r *SQLiteDeviceRepository) UpdateHeartbeat(ctx context.Context, id string, ipAddress string) error {
	now := time.Now()
	res, err := r.db.ExecContext(ctx, `
		UPDATE devices
		SET last_seen = ?, active = 1, ip_address = COALESCE(NULLIF(?, ''), ip_address), updated_at = ?
		WHERE id = ?
	`, now, ipAddress, now, id)
	if err != nil {
		return fmt.Errorf("failed to update heartbeat: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if affected == 0 {
		return domain.NewError(domain.ErrNotFound, "Device not found", map[string]string{"device_id": id})
	}
	return nil
}

func (r *SQLiteDeviceRepository) MarkInactive(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE devices
		SET active = 0, updated_at = ?
		WHERE id = ?
	`, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to mark inactive: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if affected == 0 {
		return domain.NewError(domain.ErrNotFound, "Device not found", map[string]string{"device_id": id})
	}
	return nil
}

func (r *SQLiteDeviceRepository) Count(ctx context.Context) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM devices`).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count devices: %w", err)
	}
	return count, nil
}

func (r *SQLiteDeviceRepository) GetStaleDevices(ctx context.Context, threshold time.Duration) ([]*domain.Device, error) {
	cutoff := time.Now().Add(-threshold)
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, tenant_id, name, platform, pubkey, last_seen, active,
		       ip_address, daemon_ver, os_version, hostname, created_at, updated_at, disabled_at
		FROM devices
		WHERE active = 1 AND disabled_at IS NULL AND last_seen < ?
	`, cutoff)
	if err != nil {
		return nil, fmt.Errorf("failed to get stale devices: %w", err)
	}
	defer rows.Close()

	var result []*domain.Device
	for rows.Next() {
		dev, err := scanDeviceRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, dev)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate stale devices: %w", err)
	}
	return result, nil
}

func (r *SQLiteDeviceRepository) scanDevice(ctx context.Context, query string, arg interface{}) (*domain.Device, error) {
	row := r.db.QueryRowContext(ctx, query, arg)
	dev, err := scanDeviceRow(row)
	if err != nil {
		return nil, err
	}
	return dev, nil
}

type deviceScanner interface {
	Scan(dest ...any) error
}

func scanDeviceRow(row deviceScanner) (*domain.Device, error) {
	var dev domain.Device
	var ip, daemonVer, osVer, host sql.NullString
	var disabledAt sql.NullTime
	if err := row.Scan(
		&dev.ID,
		&dev.UserID,
		&dev.TenantID,
		&dev.Name,
		&dev.Platform,
		&dev.PubKey,
		&dev.LastSeen,
		&dev.Active,
		&ip,
		&daemonVer,
		&osVer,
		&host,
		&dev.CreatedAt,
		&dev.UpdatedAt,
		&disabledAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewError(domain.ErrNotFound, "Device not found", nil)
		}
		return nil, fmt.Errorf("failed to scan device: %w", err)
	}
	if ip.Valid {
		dev.IPAddress = ip.String
	}
	if daemonVer.Valid {
		dev.DaemonVer = daemonVer.String
	}
	if osVer.Valid {
		dev.OSVersion = osVer.String
	}
	if host.Valid {
		dev.HostName = host.String
	}
	if disabledAt.Valid {
		dev.DisabledAt = &disabledAt.Time
	}
	return &dev, nil
}

func nullIfBlank(s string) interface{} {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func nullableTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}
