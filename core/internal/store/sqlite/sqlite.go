package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/orhaniscoding/goconnect/server/internal/store"
	"github.com/orhaniscoding/goconnect/server/internal/store/migrations"
	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func New(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) Migrate() error {
	driver, err := sqlite.WithInstance(s.db, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	sourceDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("failed to create iofs source driver: %w", err)
	}

	m, err := migrate.NewWithInstance(
		"iofs",
		sourceDriver,
		"sqlite",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// Node operations
func (s *SQLiteStore) GetNode(ctx context.Context) (*store.Node, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, public_key, private_key, name FROM nodes LIMIT 1")
	var node store.Node
	if err := row.Scan(&node.ID, &node.PublicKey, &node.PrivateKey, &node.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &node, nil
}

func (s *SQLiteStore) SaveNode(ctx context.Context, node *store.Node) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO nodes (id, public_key, private_key, name)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			public_key = excluded.public_key,
			private_key = excluded.private_key,
			name = excluded.name
	`, node.ID, node.PublicKey, node.PrivateKey, node.Name)
	return err
}

// Peer operations
func (s *SQLiteStore) ListPeers(ctx context.Context) ([]*store.Peer, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, public_key, endpoint, allowed_ips FROM peers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var peers []*store.Peer
	for rows.Next() {
		var peer store.Peer
		if err := rows.Scan(&peer.ID, &peer.PublicKey, &peer.Endpoint, &peer.AllowedIPs); err != nil {
			return nil, err
		}
		peers = append(peers, &peer)
	}
	return peers, nil
}

func (s *SQLiteStore) SavePeer(ctx context.Context, peer *store.Peer) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO peers (id, public_key, endpoint, allowed_ips)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			public_key = excluded.public_key,
			endpoint = excluded.endpoint,
			allowed_ips = excluded.allowed_ips
	`, peer.ID, peer.PublicKey, peer.Endpoint, peer.AllowedIPs)
	return err
}

func (s *SQLiteStore) DeletePeer(ctx context.Context, publicKey string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM peers WHERE public_key = ?", publicKey)
	return err
}
