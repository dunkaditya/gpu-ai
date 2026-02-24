package wireguard

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/jackc/pgx/v5"
)

// ipamLockID is a constant used for the PostgreSQL advisory lock.
// Represents "GPUAIPAM" as a numeric constant.
const ipamLockID int64 = 0x475055414950414D

// IPAM manages IP address allocation from a WireGuard tunnel subnet.
// It uses PostgreSQL advisory locks to prevent concurrent allocation races.
type IPAM struct {
	subnet    *net.IPNet // 10.0.0.0/16
	proxyAddr net.IP     // 10.0.0.1 (reserved)
	logger    *slog.Logger
}

// NewIPAM creates a new IPAM instance for the given subnet CIDR.
// The first usable address (network + 1) is reserved for the proxy server.
func NewIPAM(subnetCIDR string, logger *slog.Logger) (*IPAM, error) {
	_, subnet, err := net.ParseCIDR(subnetCIDR)
	if err != nil {
		return nil, fmt.Errorf("wireguard: ipam: parse CIDR %q: %w", subnetCIDR, err)
	}

	// Reserve the first address after the network address for the proxy.
	proxyAddr := incrementIP(subnet.IP)

	return &IPAM{
		subnet:    subnet,
		proxyAddr: proxyAddr,
		logger:    logger,
	}, nil
}

// AllocateAddress allocates the next available tunnel IP address within the subnet.
// It must be called within a transaction. The function acquires a PostgreSQL
// advisory lock (transaction-scoped) to prevent concurrent allocation races,
// queries the current maximum wg_address, increments it, and returns the next IP.
// The caller is responsible for inserting the address into the instance row
// within the same transaction.
func (ipam *IPAM) AllocateAddress(ctx context.Context, tx pgx.Tx) (net.IP, error) {
	// Acquire transaction-scoped advisory lock to serialize allocations.
	_, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", ipamLockID)
	if err != nil {
		return nil, fmt.Errorf("wireguard: ipam: acquire advisory lock: %w", err)
	}

	// Query the current maximum allocated address.
	var maxAddrStr string
	err = tx.QueryRow(ctx,
		"SELECT COALESCE(host(MAX(wg_address))::text, '10.0.0.1') FROM instances WHERE wg_address IS NOT NULL",
	).Scan(&maxAddrStr)
	if err != nil {
		return nil, fmt.Errorf("wireguard: ipam: query max address: %w", err)
	}

	maxAddr := net.ParseIP(maxAddrStr)
	if maxAddr == nil {
		return nil, fmt.Errorf("wireguard: ipam: invalid max address from database: %q", maxAddrStr)
	}

	nextIP := incrementIP(maxAddr)

	// Verify the next IP is still within the subnet.
	if !ipam.subnet.Contains(nextIP) {
		return nil, fmt.Errorf("wireguard: ipam: subnet %s exhausted (next would be %s)", ipam.subnet.String(), nextIP.String())
	}

	ipam.logger.Info("allocated tunnel IP",
		slog.String("address", nextIP.String()),
		slog.String("subnet", ipam.subnet.String()),
	)

	return nextIP, nil
}

// incrementIP returns the next IP address after the given IP.
// It handles byte carry (e.g., 10.0.0.255 -> 10.0.1.0).
func incrementIP(ip net.IP) net.IP {
	// Work with a 4-byte IPv4 representation.
	ip4 := ip.To4()
	if ip4 == nil {
		// Fallback: try as-is for non-IPv4.
		ip4 = ip
	}

	// Copy to avoid mutation of the original.
	next := make(net.IP, len(ip4))
	copy(next, ip4)

	// Increment from the last byte, carrying over.
	for i := len(next) - 1; i >= 0; i-- {
		next[i]++
		if next[i] != 0 {
			break // No carry needed.
		}
	}

	return next
}

// IsProxyAddress returns true if the given IP is the reserved proxy address.
func (ipam *IPAM) IsProxyAddress(ip net.IP) bool {
	return ipam.proxyAddr.Equal(ip)
}

// SubnetCIDR returns the subnet CIDR string (e.g., "10.0.0.0/16").
func (ipam *IPAM) SubnetCIDR() string {
	return ipam.subnet.String()
}
