package wireguard

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os/exec"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// WGClient abstracts WireGuard device control for testability.
type WGClient interface {
	ConfigureDevice(name string, cfg wgtypes.Config) error
	Device(name string) (*wgtypes.Device, error)
	Close() error
}

// CommandRunner abstracts shell command execution for testability.
type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) error
}

// execRunner is the production CommandRunner using os/exec.
type execRunner struct{}

func (e *execRunner) Run(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Run()
}

// Peer represents a WireGuard peer's current state.
type Peer struct {
	PublicKey     string
	AllowedIPs    []string
	LastHandshake time.Time
	TransferRx    int64
	TransferTx    int64
}

// Manager manages WireGuard peers and corresponding iptables port-mapping rules.
// It uses wgctrl-go for programmatic peer management (no shell-out to wg command)
// and iptables for DNAT port mapping.
type Manager struct {
	client        WGClient
	runner        CommandRunner
	interfaceName string // e.g., "wg0"
	logger        *slog.Logger
}

// NewManager creates a new WireGuard peer manager.
// The client controls the WireGuard device. The runner executes iptables commands.
func NewManager(client WGClient, runner CommandRunner, interfaceName string, logger *slog.Logger) *Manager {
	if runner == nil {
		runner = &execRunner{}
	}
	return &Manager{
		client:        client,
		runner:        runner,
		interfaceName: interfaceName,
		logger:        logger,
	}
}

// AddPeer adds a WireGuard peer and the corresponding iptables port-mapping rules.
// If WireGuard peer addition succeeds but iptables fails, it attempts to roll back
// by removing the peer.
func (m *Manager) AddPeer(ctx context.Context, publicKeyBase64 string, tunnelIP net.IP, externalPort int) error {
	// Parse the public key.
	pubKey, err := wgtypes.ParseKey(publicKeyBase64)
	if err != nil {
		return fmt.Errorf("wireguard: manager: parse public key: %w", err)
	}

	// Configure WireGuard peer via wgctrl-go.
	keepalive := 25 * time.Second
	peerCfg := wgtypes.PeerConfig{
		PublicKey:                   pubKey,
		ReplaceAllowedIPs:          true,
		AllowedIPs:                 []net.IPNet{{IP: tunnelIP.To4(), Mask: net.CIDRMask(32, 32)}},
		PersistentKeepaliveInterval: &keepalive,
	}

	err = m.client.ConfigureDevice(m.interfaceName, wgtypes.Config{
		Peers: []wgtypes.PeerConfig{peerCfg},
	})
	if err != nil {
		return fmt.Errorf("wireguard: manager: configure device (add peer): %w", err)
	}

	m.logger.Info("added WireGuard peer",
		slog.String("public_key", publicKeyBase64[:8]+"..."),
		slog.String("tunnel_ip", tunnelIP.String()),
		slog.Int("external_port", externalPort),
	)

	// Add iptables DNAT rule: external port -> tunnel IP:22.
	tunnelIPStr := tunnelIP.String()
	err = m.runner.Run(ctx, "iptables",
		"-t", "nat", "-A", "PREROUTING",
		"-p", "tcp", "--dport", fmt.Sprintf("%d", externalPort),
		"-j", "DNAT", "--to-destination", fmt.Sprintf("%s:22", tunnelIPStr),
	)
	if err != nil {
		m.logger.Error("iptables DNAT rule failed, rolling back WireGuard peer",
			slog.String("error", err.Error()),
			slog.String("tunnel_ip", tunnelIPStr),
		)
		// Rollback: remove the WireGuard peer.
		m.rollbackPeer(pubKey)
		return fmt.Errorf("wireguard: manager: add iptables DNAT rule: %w", err)
	}

	// Add iptables FORWARD rule: allow TCP to tunnel IP:22.
	err = m.runner.Run(ctx, "iptables",
		"-A", "FORWARD",
		"-p", "tcp", "-d", tunnelIPStr, "--dport", "22",
		"-j", "ACCEPT",
	)
	if err != nil {
		m.logger.Error("iptables FORWARD rule failed, rolling back",
			slog.String("error", err.Error()),
			slog.String("tunnel_ip", tunnelIPStr),
		)
		// Rollback: remove DNAT rule and WireGuard peer.
		_ = m.runner.Run(ctx, "iptables",
			"-t", "nat", "-D", "PREROUTING",
			"-p", "tcp", "--dport", fmt.Sprintf("%d", externalPort),
			"-j", "DNAT", "--to-destination", fmt.Sprintf("%s:22", tunnelIPStr),
		)
		m.rollbackPeer(pubKey)
		return fmt.Errorf("wireguard: manager: add iptables FORWARD rule: %w", err)
	}

	m.logger.Info("added iptables port mapping",
		slog.Int("external_port", externalPort),
		slog.String("destination", fmt.Sprintf("%s:22", tunnelIPStr)),
	)

	return nil
}

// rollbackPeer removes a WireGuard peer as part of error rollback.
func (m *Manager) rollbackPeer(pubKey wgtypes.Key) {
	err := m.client.ConfigureDevice(m.interfaceName, wgtypes.Config{
		Peers: []wgtypes.PeerConfig{{
			PublicKey: pubKey,
			Remove:   true,
		}},
	})
	if err != nil {
		m.logger.Error("failed to rollback WireGuard peer",
			slog.String("error", err.Error()),
		)
	}
}

// RemovePeer removes the iptables port-mapping rules and the WireGuard peer.
// iptables removal errors are logged but do not prevent peer removal (best-effort cleanup).
func (m *Manager) RemovePeer(ctx context.Context, publicKeyBase64 string, tunnelIP net.IP, externalPort int) error {
	// Parse the public key.
	pubKey, err := wgtypes.ParseKey(publicKeyBase64)
	if err != nil {
		return fmt.Errorf("wireguard: manager: parse public key: %w", err)
	}

	tunnelIPStr := tunnelIP.String()

	// Remove iptables DNAT rule (best-effort).
	err = m.runner.Run(ctx, "iptables",
		"-t", "nat", "-D", "PREROUTING",
		"-p", "tcp", "--dport", fmt.Sprintf("%d", externalPort),
		"-j", "DNAT", "--to-destination", fmt.Sprintf("%s:22", tunnelIPStr),
	)
	if err != nil {
		m.logger.Warn("failed to remove iptables DNAT rule (best-effort)",
			slog.String("error", err.Error()),
			slog.String("tunnel_ip", tunnelIPStr),
			slog.Int("external_port", externalPort),
		)
	}

	// Remove iptables FORWARD rule (best-effort).
	err = m.runner.Run(ctx, "iptables",
		"-D", "FORWARD",
		"-p", "tcp", "-d", tunnelIPStr, "--dport", "22",
		"-j", "ACCEPT",
	)
	if err != nil {
		m.logger.Warn("failed to remove iptables FORWARD rule (best-effort)",
			slog.String("error", err.Error()),
			slog.String("tunnel_ip", tunnelIPStr),
		)
	}

	// Remove WireGuard peer.
	err = m.client.ConfigureDevice(m.interfaceName, wgtypes.Config{
		Peers: []wgtypes.PeerConfig{{
			PublicKey: pubKey,
			Remove:   true,
		}},
	})
	if err != nil {
		return fmt.Errorf("wireguard: manager: configure device (remove peer): %w", err)
	}

	m.logger.Info("removed WireGuard peer and iptables rules",
		slog.String("public_key", publicKeyBase64[:8]+"..."),
		slog.String("tunnel_ip", tunnelIPStr),
		slog.Int("external_port", externalPort),
	)

	return nil
}

// ListPeers returns the current WireGuard peers on the managed interface.
func (m *Manager) ListPeers(ctx context.Context) ([]Peer, error) {
	dev, err := m.client.Device(m.interfaceName)
	if err != nil {
		return nil, fmt.Errorf("wireguard: manager: get device: %w", err)
	}

	peers := make([]Peer, 0, len(dev.Peers))
	for _, p := range dev.Peers {
		allowedIPs := make([]string, 0, len(p.AllowedIPs))
		for _, ipNet := range p.AllowedIPs {
			allowedIPs = append(allowedIPs, ipNet.String())
		}

		peers = append(peers, Peer{
			PublicKey:     p.PublicKey.String(),
			AllowedIPs:    allowedIPs,
			LastHandshake: p.LastHandshakeTime,
			TransferRx:    p.ReceiveBytes,
			TransferTx:    p.TransmitBytes,
		})
	}

	return peers, nil
}

// PortFromTunnelIP derives the external SSH port from a tunnel IP address.
// Formula: 10000 + int(ip[2])*256 + int(ip[3]).
// Maps the full /16 range (10.0.0.2 to 10.0.255.255) to ports 10002-75535.
// Examples: 10.0.0.5 -> 10005, 10.0.1.0 -> 10256.
func PortFromTunnelIP(ip net.IP) int {
	ip4 := ip.To4()
	if ip4 == nil {
		return 0
	}
	return 10000 + int(ip4[2])*256 + int(ip4[3])
}
