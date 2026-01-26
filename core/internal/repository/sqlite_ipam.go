package repository

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// SQLiteIPAMRepository implements IPAMRepository using SQLite.
type SQLiteIPAMRepository struct {
	db *sql.DB
}

func NewSQLiteIPAM() *SQLiteIPAMRepository {
	return &SQLiteIPAMRepository{}
}

func NewSQLiteIPAMRepository(db *sql.DB) *SQLiteIPAMRepository {
	return &SQLiteIPAMRepository{db: db}
}

// GetOrAllocate returns existing allocation for subject or allocates next available.
// subjectID can be either a user ID (legacy) or device ID (preferred for device-based allocation).
func (r *SQLiteIPAMRepository) GetOrAllocate(ctx context.Context, networkID, subjectID, cidr string) (*domain.IPAllocation, error) {
	allocs, err := r.ListAllocations(ctx, networkID)
	if err != nil {
		return nil, err
	}
	// Check both DeviceID and UserID for backward compatibility
	for _, a := range allocs {
		if a.DeviceID == subjectID || a.UserID == subjectID {
			return a, nil
		}
	}
	ip, err := r.AllocateIP(ctx, networkID, subjectID, cidr)
	if err != nil {
		return nil, err
	}
	return &domain.IPAllocation{NetworkID: networkID, DeviceID: subjectID, IP: ip}, nil
}

func (r *SQLiteIPAMRepository) AllocateIP(ctx context.Context, networkID, userID string, cidr string) (string, error) {
	if r.db == nil {
		return "", fmt.Errorf("ipam repository not initialized")
	}
	// naive allocator: pick next available host in CIDR not already allocated
	allocs, err := r.ListAllocations(ctx, networkID)
	if err != nil {
		return "", err
	}
	used := map[string]struct{}{}
	for _, a := range allocs {
		used[a.IP] = struct{}{}
	}

	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", domain.NewError(domain.ErrCIDRInvalid, "invalid CIDR", nil)
	}

	ip := ipnet.IP.To4()
	if ip == nil {
		return "", domain.NewError(domain.ErrCIDRInvalid, "only IPv4 supported for IPAM allocation", nil)
	}

	// start from first host (skip network address)
	ip = ip.Mask(ipnet.Mask)
	for i := 1; i < 1<<16; i++ { // cap search for safety
		next := incIP(ip, i)
		if !ipnet.Contains(next) {
			break
		}
		ipStr := next.String()
		if _, ok := used[ipStr]; ok {
			continue
		}
		if err := r.saveAllocation(ctx, networkID, userID, ipStr); err != nil {
			return "", err
		}
		return ipStr, nil
	}
	return "", domain.NewError(domain.ErrIPExhausted, "no available IP addresses", nil)
}

func (r *SQLiteIPAMRepository) saveAllocation(ctx context.Context, networkID, subjectID, ip string) error {
	// Insert with both device_id and user_id for compatibility
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO ip_allocations (id, network_id, device_id, user_id, ip_address, allocated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, domain.GenerateNetworkID(), networkID, subjectID, subjectID, ip, time.Now())
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return domain.NewError(domain.ErrIPExhausted, "IP already allocated", nil)
		}
		return fmt.Errorf("failed to save allocation: %w", err)
	}
	return nil
}

func (r *SQLiteIPAMRepository) ReleaseIP(ctx context.Context, networkID, userID string) error {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM ip_allocations WHERE network_id = ? AND user_id = ?
	`, networkID, userID)
	if err != nil {
		return fmt.Errorf("failed to release ip: %w", err)
	}
	if rows, err := res.RowsAffected(); err != nil || rows == 0 {
		return domain.NewError(domain.ErrNotFound, "IP allocation not found", nil)
	}
	return nil
}

func (r *SQLiteIPAMRepository) ListAllocations(ctx context.Context, networkID string) ([]*domain.IPAllocation, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT network_id, COALESCE(device_id, user_id), ip_address
		FROM ip_allocations
		WHERE network_id = ?
	`, networkID)
	if err != nil {
		return nil, fmt.Errorf("failed to list allocations: %w", err)
	}
	defer rows.Close()

	var result []*domain.IPAllocation
	for rows.Next() {
		var alloc domain.IPAllocation
		var subjectID, ip string
		if err := rows.Scan(&alloc.NetworkID, &subjectID, &ip); err != nil {
			return nil, fmt.Errorf("failed to scan allocation: %w", err)
		}
		alloc.DeviceID = subjectID // Use DeviceID for new allocations
		alloc.IP = ip
		result = append(result, &alloc)
	}
	return result, rows.Err()
}

func incIP(ip net.IP, n int) net.IP {
	res := make(net.IP, len(ip))
	copy(res, ip)
	for i := len(res) - 1; i >= 0 && n > 0; i-- {
		sum := int(res[i]) + n
		res[i] = byte(sum % 256)
		n = sum / 256
	}
	return res
}

func (r *SQLiteIPAMRepository) List(ctx context.Context, networkID string) ([]*domain.IPAllocation, error) {
	return r.ListAllocations(ctx, networkID)
}

// Release removes a subject's allocation (idempotent, supports both device_id and user_id)
func (r *SQLiteIPAMRepository) Release(ctx context.Context, networkID, subjectID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM ip_allocations WHERE network_id = ? AND (device_id = ? OR user_id = ?)`, networkID, subjectID, subjectID)
	if err != nil {
		return fmt.Errorf("failed to release ip: %w", err)
	}
	return nil
}
