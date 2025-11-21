package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

type PostgresPeerRepository struct {
	db *sql.DB
}

func NewPostgresPeerRepository(db *sql.DB) *PostgresPeerRepository {
	return &PostgresPeerRepository{db: db}
}

func (r *PostgresPeerRepository) Create(ctx context.Context, peer *domain.Peer) error {
	query := `
		INSERT INTO peers (
			id, network_id, device_id, tenant_id, public_key, preshared_key,
			endpoint, allowed_ips, persistent_keepalive, last_handshake,
			rx_bytes, tx_bytes, active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		peer.ID,
		peer.NetworkID,
		peer.DeviceID,
		peer.TenantID,
		peer.PublicKey,
		peer.PresharedKey,
		peer.Endpoint,
		pq.Array(peer.AllowedIPs),
		peer.PersistentKeepalive,
		peer.LastHandshake,
		peer.RxBytes,
		peer.TxBytes,
		peer.Active,
		time.Now(),
		time.Now(),
	).Scan(&peer.ID, &peer.CreatedAt, &peer.UpdatedAt)

	if err != nil {
		// Check for unique constraint violations
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				return domain.NewError(domain.ErrConflict, "peer with this public key or device-network combination already exists", nil)
			}
		}
		return err
	}

	return nil
}

func (r *PostgresPeerRepository) GetByID(ctx context.Context, peerID string) (*domain.Peer, error) {
	query := `
		SELECT id, network_id, device_id, tenant_id, public_key, preshared_key,
			   endpoint, allowed_ips, persistent_keepalive, last_handshake,
			   rx_bytes, tx_bytes, active, created_at, updated_at, disabled_at
		FROM peers
		WHERE id = $1 AND disabled_at IS NULL
	`

	peer := &domain.Peer{}
	var allowedIPs pq.StringArray
	var lastHandshake sql.NullTime

	err := r.db.QueryRowContext(ctx, query, peerID).Scan(
		&peer.ID,
		&peer.NetworkID,
		&peer.DeviceID,
		&peer.TenantID,
		&peer.PublicKey,
		&peer.PresharedKey,
		&peer.Endpoint,
		&allowedIPs,
		&peer.PersistentKeepalive,
		&lastHandshake,
		&peer.RxBytes,
		&peer.TxBytes,
		&peer.Active,
		&peer.CreatedAt,
		&peer.UpdatedAt,
		&peer.DisabledAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.NewError(domain.ErrNotFound, "peer not found", nil)
	}
	if err != nil {
		return nil, err
	}

	peer.AllowedIPs = allowedIPs
	if lastHandshake.Valid {
		peer.LastHandshake = &lastHandshake.Time
	}

	return peer, nil
}

func (r *PostgresPeerRepository) GetByNetworkID(ctx context.Context, networkID string) ([]*domain.Peer, error) {
	query := `
		SELECT id, network_id, device_id, tenant_id, public_key, preshared_key,
			   endpoint, allowed_ips, persistent_keepalive, last_handshake,
			   rx_bytes, tx_bytes, active, created_at, updated_at, disabled_at
		FROM peers
		WHERE network_id = $1 AND disabled_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, networkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPeers(rows)
}

func (r *PostgresPeerRepository) GetByDeviceID(ctx context.Context, deviceID string) ([]*domain.Peer, error) {
	query := `
		SELECT id, network_id, device_id, tenant_id, public_key, preshared_key,
			   endpoint, allowed_ips, persistent_keepalive, last_handshake,
			   rx_bytes, tx_bytes, active, created_at, updated_at, disabled_at
		FROM peers
		WHERE device_id = $1 AND disabled_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPeers(rows)
}

func (r *PostgresPeerRepository) GetByNetworkAndDevice(ctx context.Context, networkID, deviceID string) (*domain.Peer, error) {
	query := `
		SELECT id, network_id, device_id, tenant_id, public_key, preshared_key,
			   endpoint, allowed_ips, persistent_keepalive, last_handshake,
			   rx_bytes, tx_bytes, active, created_at, updated_at, disabled_at
		FROM peers
		WHERE network_id = $1 AND device_id = $2 AND disabled_at IS NULL
	`

	peer := &domain.Peer{}
	var allowedIPs pq.StringArray
	var lastHandshake sql.NullTime

	err := r.db.QueryRowContext(ctx, query, networkID, deviceID).Scan(
		&peer.ID,
		&peer.NetworkID,
		&peer.DeviceID,
		&peer.TenantID,
		&peer.PublicKey,
		&peer.PresharedKey,
		&peer.Endpoint,
		&allowedIPs,
		&peer.PersistentKeepalive,
		&lastHandshake,
		&peer.RxBytes,
		&peer.TxBytes,
		&peer.Active,
		&peer.CreatedAt,
		&peer.UpdatedAt,
		&peer.DisabledAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.NewError(domain.ErrNotFound, "peer not found", nil)
	}
	if err != nil {
		return nil, err
	}

	peer.AllowedIPs = allowedIPs
	if lastHandshake.Valid {
		peer.LastHandshake = &lastHandshake.Time
	}

	return peer, nil
}

func (r *PostgresPeerRepository) GetByPublicKey(ctx context.Context, publicKey string) (*domain.Peer, error) {
	query := `
		SELECT id, network_id, device_id, tenant_id, public_key, preshared_key,
			   endpoint, allowed_ips, persistent_keepalive, last_handshake,
			   rx_bytes, tx_bytes, active, created_at, updated_at, disabled_at
		FROM peers
		WHERE public_key = $1 AND disabled_at IS NULL
	`

	peer := &domain.Peer{}
	var allowedIPs pq.StringArray
	var lastHandshake sql.NullTime

	err := r.db.QueryRowContext(ctx, query, publicKey).Scan(
		&peer.ID,
		&peer.NetworkID,
		&peer.DeviceID,
		&peer.TenantID,
		&peer.PublicKey,
		&peer.PresharedKey,
		&peer.Endpoint,
		&allowedIPs,
		&peer.PersistentKeepalive,
		&lastHandshake,
		&peer.RxBytes,
		&peer.TxBytes,
		&peer.Active,
		&peer.CreatedAt,
		&peer.UpdatedAt,
		&peer.DisabledAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.NewError(domain.ErrNotFound, "peer not found", nil)
	}
	if err != nil {
		return nil, err
	}

	peer.AllowedIPs = allowedIPs
	if lastHandshake.Valid {
		peer.LastHandshake = &lastHandshake.Time
	}

	return peer, nil
}

func (r *PostgresPeerRepository) GetActivePeers(ctx context.Context, networkID string) ([]*domain.Peer, error) {
	query := `
		SELECT id, network_id, device_id, tenant_id, public_key, preshared_key,
			   endpoint, allowed_ips, persistent_keepalive, last_handshake,
			   rx_bytes, tx_bytes, active, created_at, updated_at, disabled_at
		FROM peers
		WHERE network_id = $1 AND disabled_at IS NULL AND active = true
	`

	rows, err := r.db.QueryContext(ctx, query, networkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var peers []*domain.Peer
	for rows.Next() {
		peer := &domain.Peer{}
		var allowedIPs pq.StringArray
		var lastHandshake sql.NullTime

		if err := rows.Scan(
			&peer.ID,
			&peer.NetworkID,
			&peer.DeviceID,
			&peer.TenantID,
			&peer.PublicKey,
			&peer.PresharedKey,
			&peer.Endpoint,
			&allowedIPs,
			&peer.PersistentKeepalive,
			&lastHandshake,
			&peer.RxBytes,
			&peer.TxBytes,
			&peer.Active,
			&peer.CreatedAt,
			&peer.UpdatedAt,
			&peer.DisabledAt,
		); err != nil {
			return nil, err
		}

		peer.AllowedIPs = allowedIPs
		if lastHandshake.Valid {
			peer.LastHandshake = &lastHandshake.Time
		}
		peers = append(peers, peer)
	}

	return peers, nil
}

func (r *PostgresPeerRepository) GetAllActive(ctx context.Context) ([]*domain.Peer, error) {
	query := `
		SELECT id, network_id, device_id, tenant_id, public_key, preshared_key,
			   endpoint, allowed_ips, persistent_keepalive, last_handshake,
			   rx_bytes, tx_bytes, active, created_at, updated_at, disabled_at
		FROM peers
		WHERE disabled_at IS NULL
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var peers []*domain.Peer
	for rows.Next() {
		peer := &domain.Peer{}
		var allowedIPs pq.StringArray
		var lastHandshake sql.NullTime

		if err := rows.Scan(
			&peer.ID,
			&peer.NetworkID,
			&peer.DeviceID,
			&peer.TenantID,
			&peer.PublicKey,
			&peer.PresharedKey,
			&peer.Endpoint,
			&allowedIPs,
			&peer.PersistentKeepalive,
			&lastHandshake,
			&peer.RxBytes,
			&peer.TxBytes,
			&peer.Active,
			&peer.CreatedAt,
			&peer.UpdatedAt,
			&peer.DisabledAt,
		); err != nil {
			return nil, err
		}

		peer.AllowedIPs = allowedIPs
		if lastHandshake.Valid {
			peer.LastHandshake = &lastHandshake.Time
		}
		peers = append(peers, peer)
	}

	return peers, nil
}

func (r *PostgresPeerRepository) Update(ctx context.Context, peer *domain.Peer) error {
	query := `
		UPDATE peers
		SET network_id = $2,
			device_id = $3,
			tenant_id = $4,
			public_key = $5,
			preshared_key = $6,
			endpoint = $7,
			allowed_ips = $8,
			persistent_keepalive = $9,
			last_handshake = $10,
			rx_bytes = $11,
			tx_bytes = $12,
			active = $13,
			updated_at = $14
		WHERE id = $1 AND disabled_at IS NULL
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		peer.ID,
		peer.NetworkID,
		peer.DeviceID,
		peer.TenantID,
		peer.PublicKey,
		peer.PresharedKey,
		peer.Endpoint,
		pq.Array(peer.AllowedIPs),
		peer.PersistentKeepalive,
		peer.LastHandshake,
		peer.RxBytes,
		peer.TxBytes,
		peer.Active,
		time.Now(),
	).Scan(&peer.UpdatedAt)

	if err == sql.ErrNoRows {
		return domain.NewError(domain.ErrNotFound, "peer not found", nil)
	}
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresPeerRepository) UpdateStats(ctx context.Context, peerID string, stats *domain.UpdatePeerStatsRequest) error {
	query := `
		UPDATE peers
		SET endpoint = COALESCE(NULLIF($2, ''), endpoint),
			last_handshake = COALESCE($3, last_handshake),
			rx_bytes = $4,
			tx_bytes = $5,
			active = CASE 
				WHEN $3 IS NOT NULL AND (NOW() - $3) < INTERVAL '3 minutes' THEN true
				ELSE active
			END,
			updated_at = $6
		WHERE id = $1 AND disabled_at IS NULL
		RETURNING updated_at
	`

	var updatedAt time.Time
	err := r.db.QueryRowContext(
		ctx,
		query,
		peerID,
		stats.Endpoint,
		stats.LastHandshake,
		stats.RxBytes,
		stats.TxBytes,
		time.Now(),
	).Scan(&updatedAt)

	if err == sql.ErrNoRows {
		return domain.NewError(domain.ErrNotFound, "peer not found", nil)
	}
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresPeerRepository) Delete(ctx context.Context, peerID string) error {
	query := `
		UPDATE peers
		SET disabled_at = $2,
			active = false,
			updated_at = $2
		WHERE id = $1 AND disabled_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, peerID, time.Now())
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "peer not found", nil)
	}

	return nil
}

func (r *PostgresPeerRepository) HardDelete(ctx context.Context, peerID string) error {
	query := `DELETE FROM peers WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, peerID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.NewError(domain.ErrNotFound, "peer not found", nil)
	}

	return nil
}

func (r *PostgresPeerRepository) ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Peer, error) {
	query := `
		SELECT id, network_id, device_id, tenant_id, public_key, preshared_key,
			   endpoint, allowed_ips, persistent_keepalive, last_handshake,
			   rx_bytes, tx_bytes, active, created_at, updated_at, disabled_at
		FROM peers
		WHERE tenant_id = $1 AND disabled_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPeers(rows)
}

// scanPeers is a helper function to scan multiple peers from rows
func (r *PostgresPeerRepository) scanPeers(rows *sql.Rows) ([]*domain.Peer, error) {
	var peers []*domain.Peer

	for rows.Next() {
		peer := &domain.Peer{}
		var allowedIPs pq.StringArray
		var lastHandshake sql.NullTime

		err := rows.Scan(
			&peer.ID,
			&peer.NetworkID,
			&peer.DeviceID,
			&peer.TenantID,
			&peer.PublicKey,
			&peer.PresharedKey,
			&peer.Endpoint,
			&allowedIPs,
			&peer.PersistentKeepalive,
			&lastHandshake,
			&peer.RxBytes,
			&peer.TxBytes,
			&peer.Active,
			&peer.CreatedAt,
			&peer.UpdatedAt,
			&peer.DisabledAt,
		)
		if err != nil {
			return nil, err
		}

		peer.AllowedIPs = allowedIPs
		if lastHandshake.Valid {
			peer.LastHandshake = &lastHandshake.Time
		}

		peers = append(peers, peer)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return peers, nil
}
