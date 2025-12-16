package domain

import (
	"fmt"
	"net"
)

// IPAllocation represents an allocated IP inside a network for a device
// Note: UserID is deprecated, use DeviceID for new allocations
type IPAllocation struct {
	NetworkID string `json:"network_id"`
	UserID    string `json:"user_id,omitempty"`    // Deprecated: use DeviceID
	DeviceID  string `json:"device_id,omitempty"`  // Preferred: device-based allocation
	IP        string `json:"ip"`
	// Offset represents the host offset (starting at 1) inside the CIDR. Not exposed in JSON, used for reuse on relea
	Offset uint32 `json:"-"`
}

// IPAMError codes
const (
	ErrIPExhausted = "ERR_IP_EXHAUSTED"
)

// NextIP calculates the next IP given a base network and an offset (starting from 1 to avoid network address)
// It returns empty string if offset exceeds range or broadcast address.
func NextIP(cidr string, offset uint32) (string, error) {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", fmt.Errorf("invalid cidr: %w", err)
	}
	// Convert IP to uint32
	base := network.IP.To4()
	if base == nil {
		return "", fmt.Errorf("only IPv4 supported for now")
	}
	mask, bits := network.Mask.Size()
	if bits != 32 || mask < 0 || mask > 32 {
		return "", fmt.Errorf("invalid IPv4 mask size: mask=%d bits=%d", mask, bits)
	}
	hostBits := 32 - mask
	if hostBits <= 1 { // no usable host when /31 or /32
		return "", nil
	}
	// Compute total usable hosts using uint64 to avoid overflow concerns
	totalHosts := (uint64(1) << uint(hostBits)) - 2
	if offset == 0 || uint64(offset) > totalHosts { // offset 1 -> first usable
		return "", nil
	}
	ipInt := (uint32(base[0]) << 24) | (uint32(base[1]) << 16) | (uint32(base[2]) << 8) | uint32(base[3])
	ipInt += offset // skip network address by starting offset at 1
	// Build IP
	ip := []byte{byte(ipInt >> 24), byte(ipInt >> 16), byte(ipInt >> 8), byte(ipInt)}
	candidate := net.IP(ip)
	if !network.Contains(candidate) { // safety
		return "", nil
	}
	// Exclude broadcast implicitly by range check above
	return candidate.String(), nil
}
