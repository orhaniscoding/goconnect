package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// SQLitePeerRepository implements PeerRepository for SQLite.
type SQLitePeerRepository struct {
	db *sql.DB
}

func NewSQLitePeerRepository(db *sql.DB) *SQLitePeerRepository {
	return &SQLitePeerRepository{db: db}
}

func (r *SQLitePeerRepository) Create(ctx context.Context, peer *domain.Peer) error {
	now := time.Now()
	if peer.ID == "" {
		peer.ID = domain.GenerateNetworkID()
	}
	peer.CreatedAt = now
	peer.UpdatedAt = now
	allowedIPs := strings.Join(peer.AllowedIPs, ",")
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO peers (
			id, network_id, device_id, tenant_id, public_key, preshared_key,
			endpoint, allowed_ips, persistent_keepalive, last_handshake,
			rx_bytes, tx_bytes, active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, peer.ID, peer.NetworkID, peer.DeviceID, peer.TenantID, peer.PublicKey, peer.PresharedKey, peer.Endpoint,
		allowedIPs, peer.PersistentKeepalive, peer.LastHandshake, peer.RxBytes, peer.TxBytes, peer.Active, peer.CreatedAt, peer.UpdatedAt)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return domain.NewError(domain.ErrConflict, "peer already exists (public key or device-network combo)", nil)
		}
		return fmt.Errorf("failed to create peer: %w", err)
	}
	return nil
}

func (r *SQLitePeerRepository) GetByID(ctx context.Context, peerID string) (*domain.Peer, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, network_id, device_id, tenant_id, public_key, preshared_key,
		       endpoint, allowed_ips, persistent_keepalive, last_handshake,
		       rx_bytes, tx_bytes, active, created_at, updated_at, disabled_at
		FROM peers
		WHERE id = ? AND disabled_at IS NULL
	`, peerID)
	return scanPeer(row)
}

func (r *SQLitePeerRepository) GetByNetworkID(ctx context.Context, networkID string) ([]*domain.Peer, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, network_id, device_id, tenant_id, public_key, preshared_key,
		       endpoint, allowed_ips, persistent_keepalive, last_handshake,
		       rx_bytes, tx_bytes, active, created_at, updated_at, disabled_at
		FROM peers
		WHERE network_id = ? AND disabled_at IS NULL
		ORDER BY created_at DESC
	`, networkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPeers(rows)
}

func (r *SQLitePeerRepository) GetByDeviceID(ctx context.Context, deviceID string) ([]*domain.Peer, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, network_id, device_id, tenant_id, public_key, preshared_key,
		       endpoint, allowed_ips, persistent_keepalive, last_handshake,
		       rx_bytes, tx_bytes, active, created_at, updated_at, disabled_at
		FROM peers
		WHERE device_id = ? AND disabled_at IS NULL
		ORDER BY created_at DESC
	`, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPeers(rows)
}

func (r *SQLitePeerRepository) GetByNetworkAndDevice(ctx context.Context, networkID, deviceID string) (*domain.Peer, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, network_id, device_id, tenant_id, public_key, preshared_key,
		       endpoint, allowed_ips, persistent_keepalive, last_handshake,
		       rx_bytes, tx_bytes, active, created_at, updated_at, disabled_at
		FROM peers
		WHERE network_id = ? AND device_id = ? AND disabled_at IS NULL
	`, networkID, deviceID)
	return scanPeer(row)
}

func (r *SQLitePeerRepository) GetByPublicKey(ctx context.Context, publicKey string) (*domain.Peer, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, network_id, device_id, tenant_id, public_key, preshared_key,
		       endpoint, allowed_ips, persistent_keepalive, last_handshake,
		       rx_bytes, tx_bytes, active, created_at, updated_at, disabled_at
		FROM peers
		WHERE public_key = ? AND disabled_at IS NULL
	`, publicKey)
	return scanPeer(row)
}

func (r *SQLitePeerRepository) GetActivePeers(ctx context.Context, networkID string) ([]*domain.Peer, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, network_id, device_id, tenant_id, public_key, preshared_key,
		       endpoint, allowed_ips, persistent_keepalive, last_handshake,
		       rx_bytes, tx_bytes, active, created_at, updated_at, disabled_at
		FROM peers
		WHERE network_id = ? AND disabled_at IS NULL AND active = 1
	`, networkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPeers(rows)
}

func (r *SQLitePeerRepository) GetAllActive(ctx context.Context) ([]*domain.Peer, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, network_id, device_id, tenant_id, public_key, preshared_key,
		       endpoint, allowed_ips, persistent_keepalive, last_handshake,
		       rx_bytes, tx_bytes, active, created_at, updated_at, disabled_at
		FROM peers
		WHERE disabled_at IS NULL
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPeers(rows)
}

func (r *SQLitePeerRepository) Update(ctx context.Context, peer *domain.Peer) error {
	peer.UpdatedAt = time.Now()
	allowedIPs := strings.Join(peer.AllowedIPs, ",")
	res, err := r.db.ExecContext(ctx, `
		UPDATE peers
		SET network_id = ?, device_id = ?, tenant_id = ?, public_key = ?, preshared_key = ?,
		    endpoint = ?, allowed_ips = ?, persistent_keepalive = ?, last_handshake = ?,
		    rx_bytes = ?, tx_bytes = ?, active = ?, updated_at = ?
		WHERE id = ? AND disabled_at IS NULL
	`, peer.NetworkID, peer.DeviceID, peer.TenantID, peer.PublicKey, peer.PresharedKey, peer.Endpoint,
		allowedIPs, peer.PersistentKeepalive, peer.LastHandshake, peer.RxBytes, peer.TxBytes, peer.Active, peer.UpdatedAt, peer.ID)
	if err != nil {
		return fmt.Errorf("failed to update peer: %w", err)
	}
	if rows, err := res.RowsAffected(); err != nil || rows == 0 {
		return domain.NewError(domain.ErrNotFound, "peer not found", nil)
	}
	return nil
}

func (r *SQLitePeerRepository) UpdateStats(ctx context.Context, peerID string, stats *domain.UpdatePeerStatsRequest) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE peers
		SET endpoint = COALESCE(NULLIF(?, ''), endpoint),
		    last_handshake = COALESCE(?, last_handshake),
		    rx_bytes = ?,
		    tx_bytes = ?,
		    active = CASE 
				WHEN ? IS NOT NULL AND (strftime('%s','now') - strftime('%s',?)) < 180 THEN 1
				ELSE active
			END,
		    updated_at = ?
		WHERE id = ? AND disabled_at IS NULL
	`, stats.Endpoint, stats.LastHandshake, stats.RxBytes, stats.TxBytes, stats.LastHandshake, stats.LastHandshake, time.Now(), peerID)
	if err != nil {
		return fmt.Errorf("failed to update peer stats: %w", err)
	}
	if rows, err := res.RowsAffected(); err != nil || rows == 0 {
		return domain.NewError(domain.ErrNotFound, "peer not found", nil)
	}
	return nil
}

func (r *SQLitePeerRepository) Delete(ctx context.Context, peerID string) error {
	now := time.Now()
	res, err := r.db.ExecContext(ctx, `
		UPDATE peers
		SET disabled_at = ?, active = 0, updated_at = ?
		WHERE id = ? AND disabled_at IS NULL
	`, now, now, peerID)
	if err != nil {
		return fmt.Errorf("failed to delete peer: %w", err)
	}
	if rows, err := res.RowsAffected(); err != nil || rows == 0 {
		return domain.NewError(domain.ErrNotFound, "peer not found", nil)
	}
	return nil
}

func (r *SQLitePeerRepository) HardDelete(ctx context.Context, peerID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM peers WHERE id = ?`, peerID)
	if err != nil {
		return fmt.Errorf("failed to hard delete peer: %w", err)
	}
	if rows, err := res.RowsAffected(); err != nil || rows == 0 {
		return domain.NewError(domain.ErrNotFound, "peer not found", nil)
	}
	return nil
}

func (r *SQLitePeerRepository) ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Peer, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, network_id, device_id, tenant_id, public_key, preshared_key,
		       endpoint, allowed_ips, persistent_keepalive, last_handshake,
		       rx_bytes, tx_bytes, active, created_at, updated_at, disabled_at
		FROM peers
		WHERE tenant_id = ? AND disabled_at IS NULL
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPeers(rows)
}

// helpers
func scanPeer(row interface{ Scan(dest ...any) error }) (*domain.Peer, error) {
	var p domain.Peer
	var allowedIPsStr sql.NullString
	var lastHandshake sql.NullTime
	var disabledAt sql.NullTime
	if err := row.Scan(
		&p.ID,
		&p.NetworkID,
		&p.DeviceID,
		&p.TenantID,
		&p.PublicKey,
		&p.PresharedKey,
		&p.Endpoint,
		&allowedIPsStr,
		&p.PersistentKeepalive,
		&lastHandshake,
		&p.RxBytes,
		&p.TxBytes,
		&p.Active,
		&p.CreatedAt,
		&p.UpdatedAt,
		&disabledAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewError(domain.ErrNotFound, "peer not found", nil)
		}
		return nil, fmt.Errorf("failed to scan peer: %w", err)
	}
	if allowedIPsStr.Valid {
		if allowedIPsStr.String != "" {
			p.AllowedIPs = strings.Split(allowedIPsStr.String, ",")
		}
	}
	if lastHandshake.Valid {
		p.LastHandshake = &lastHandshake.Time
	}
	if disabledAt.Valid {
		p.DisabledAt = &disabledAt.Time
	}
	return &p, nil
}

func scanPeers(rows *sql.Rows) ([]*domain.Peer, error) {
	var peers []*domain.Peer
	for rows.Next() {
		p, err := scanPeer(rows)
		if err != nil {
			return nil, err
		}
		peers = append(peers, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return peers, nil
}
