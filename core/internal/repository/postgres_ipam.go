package repository

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// PostgresIPAMRepository implements IPAMRepository using PostgreSQL
type PostgresIPAMRepository struct {
	db *sql.DB
}

// NewPostgresIPAMRepository creates a new PostgreSQL-backed IPAM repository
func NewPostgresIPAMRepository(db *sql.DB) *PostgresIPAMRepository {
	return &PostgresIPAMRepository{db: db}
}

// GetOrAllocate returns existing allocation for user or allocates next available IP
func (r *PostgresIPAMRepository) GetOrAllocate(ctx context.Context, networkID, userID, cidr string) (*domain.IPAllocation, error) {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if user already has allocation
	existingQuery := `
		SELECT network_id, user_id, ip_address
		FROM ip_allocations
		WHERE network_id = $1 AND user_id = $2
	`
	allocation := &domain.IPAllocation{}
	err = tx.QueryRowContext(ctx, existingQuery, networkID, userID).Scan(
		&allocation.NetworkID,
		&allocation.UserID,
		&allocation.IP,
	)
	if err == nil {
		// Existing allocation found
		return allocation, tx.Commit()
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing allocation: %w", err)
	}

	// Parse CIDR
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR: %w", err)
	}

	// Calculate total hosts
	ones, bits := ipNet.Mask.Size()
	totalHosts := uint32(1 << uint(bits-ones))

	// Reserve first 2 IPs (network + gateway) and last IP (broadcast)
	usableStart := uint32(2)
	usableEnd := totalHosts - 1

	if usableEnd <= usableStart {
		return nil, domain.NewError(domain.ErrIPExhausted,
			"No usable IP addresses in this CIDR range",
			map[string]string{"cidr": cidr})
	}

	// Get all allocated IPs in this network
	allocatedQuery := `
		SELECT ip_address FROM ip_allocations
		WHERE network_id = $1
		FOR UPDATE
	`
	rows, err := tx.QueryContext(ctx, allocatedQuery, networkID)
	if err != nil {
		return nil, fmt.Errorf("failed to query allocations: %w", err)
	}
	defer rows.Close()

	usedIPs := make(map[string]bool)
	for rows.Next() {
		var ip string
		if err := rows.Scan(&ip); err != nil {
			return nil, fmt.Errorf("failed to scan allocated IP: %w", err)
		}
		usedIPs[ip] = true
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate allocated IPs: %w", err)
	}

	// Find next available IP
	baseIP := ipNet.IP.To4()
	if baseIP == nil {
		return nil, fmt.Errorf("only IPv4 is supported")
	}

	var allocatedIP string
	for offset := usableStart; offset < usableEnd; offset++ {
		candidateIP := make(net.IP, 4)
		copy(candidateIP, baseIP)

		// Add offset to IP
		val := uint32(candidateIP[0])<<24 | uint32(candidateIP[1])<<16 |
			uint32(candidateIP[2])<<8 | uint32(candidateIP[3])
		val += offset

		candidateIP[0] = byte(val >> 24)
		candidateIP[1] = byte(val >> 16)
		candidateIP[2] = byte(val >> 8)
		candidateIP[3] = byte(val)

		ipStr := candidateIP.String()
		if !usedIPs[ipStr] {
			allocatedIP = ipStr
			break
		}
	}

	if allocatedIP == "" {
		return nil, domain.NewError(domain.ErrIPExhausted,
			"All IP addresses in network are allocated",
			map[string]string{"network_id": networkID})
	}

	// Create allocation
	allocation = &domain.IPAllocation{
		NetworkID: networkID,
		UserID:    userID,
		IP:        allocatedIP,
	}

	insertQuery := `
		INSERT INTO ip_allocations (id, network_id, user_id, ip_address, allocated_at)
		VALUES ($1, $2, $3, $4, NOW())
	`
	_, err = tx.ExecContext(ctx, insertQuery,
		domain.GenerateNetworkID(), // Generate ID for database
		allocation.NetworkID,
		allocation.UserID,
		allocation.IP,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create allocation: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return allocation, nil
}

// List returns all allocations for a network
func (r *PostgresIPAMRepository) List(ctx context.Context, networkID string) ([]*domain.IPAllocation, error) {
	query := `
		SELECT network_id, user_id, ip_address
		FROM ip_allocations
		WHERE network_id = $1
		ORDER BY allocated_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, networkID)
	if err != nil {
		return nil, fmt.Errorf("failed to list allocations: %w", err)
	}
	defer rows.Close()

	var allocations []*domain.IPAllocation
	for rows.Next() {
		allocation := &domain.IPAllocation{}
		err := rows.Scan(
			&allocation.NetworkID,
			&allocation.UserID,
			&allocation.IP,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan allocation: %w", err)
		}
		allocations = append(allocations, allocation)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate allocations: %w", err)
	}

	return allocations, nil
}

// Release removes a user's allocation (idempotent)
func (r *PostgresIPAMRepository) Release(ctx context.Context, networkID, userID string) error {
	query := `DELETE FROM ip_allocations WHERE network_id = $1 AND user_id = $2`
	result, err := r.db.ExecContext(ctx, query, networkID, userID)
	if err != nil {
		return fmt.Errorf("failed to release allocation: %w", err)
	}

	// Idempotent: no error if already released
	rows, _ := result.RowsAffected()
	if rows == 0 {
		// Already released or never existed - this is OK for idempotency
		return nil
	}

	return nil
}
