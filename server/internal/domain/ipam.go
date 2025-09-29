package domain

import (
	"fmt"
	"net"
)

// IPAllocation represents an allocated IP inside a network for a user
type IPAllocation struct {
	NetworkID string `json:"network_id"`
	UserID    string `json:"user_id"`
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
	hostBits := uint32(bits - mask)
	totalHosts := uint32(1<<hostBits) - 2   // exclude network & broadcast
	if offset == 0 || offset > totalHosts { // offset 1 -> first usable
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
	// Exclude broadcast: recompute last usable
	if offset == totalHosts+1 { // would be broadcast
		return "", nil
	}
	return candidate.String(), nil
}
