package wireguard

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// mockWGClient implements WGClient for testing.
type mockWGClient struct {
	configureCalls []wgtypes.Config
	configureErr   error
	device         *wgtypes.Device
	deviceErr      error
}

func (m *mockWGClient) ConfigureDevice(name string, cfg wgtypes.Config) error {
	m.configureCalls = append(m.configureCalls, cfg)
	return m.configureErr
}

func (m *mockWGClient) Device(name string) (*wgtypes.Device, error) {
	if m.deviceErr != nil {
		return nil, m.deviceErr
	}
	return m.device, nil
}

func (m *mockWGClient) Close() error { return nil }

// mockCommandRunner implements CommandRunner for testing.
type mockCommandRunner struct {
	calls    []string // Concatenated command strings for inspection.
	callArgs [][]string
	err      error    // Error to return (nil for success).
	errOn    int      // Which call index to fail on (-1 = never).
	callNum  int
}

func newMockRunner() *mockCommandRunner {
	return &mockCommandRunner{errOn: -1}
}

func (m *mockCommandRunner) Run(ctx context.Context, name string, args ...string) error {
	fullCmd := name + " " + strings.Join(args, " ")
	m.calls = append(m.calls, fullCmd)
	allArgs := append([]string{name}, args...)
	m.callArgs = append(m.callArgs, allArgs)

	callIdx := m.callNum
	m.callNum++

	if m.errOn >= 0 && callIdx == m.errOn {
		return m.err
	}
	if m.errOn < 0 && m.err != nil {
		return m.err
	}
	return nil
}

// testKey generates a valid WireGuard key for testing.
func testWGKey(t *testing.T) wgtypes.Key {
	t.Helper()
	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("generate test key: %v", err)
	}
	return key.PublicKey()
}

func managerTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestAddPeer(t *testing.T) {
	client := &mockWGClient{}
	runner := newMockRunner()
	mgr := NewManager(client, runner, "wg0", managerTestLogger())

	pubKey := testWGKey(t)
	tunnelIP := net.ParseIP("10.0.0.5").To4()
	externalPort := 10005

	err := mgr.AddPeer(context.Background(), pubKey.String(), tunnelIP, externalPort)
	if err != nil {
		t.Fatalf("AddPeer() error: %v", err)
	}

	// Verify WireGuard ConfigureDevice was called once (add peer).
	if len(client.configureCalls) != 1 {
		t.Fatalf("ConfigureDevice calls = %d, want 1", len(client.configureCalls))
	}

	cfg := client.configureCalls[0]
	if len(cfg.Peers) != 1 {
		t.Fatalf("peers in config = %d, want 1", len(cfg.Peers))
	}

	peer := cfg.Peers[0]
	if peer.PublicKey != pubKey {
		t.Errorf("peer public key mismatch")
	}
	if !peer.ReplaceAllowedIPs {
		t.Error("ReplaceAllowedIPs should be true")
	}
	if len(peer.AllowedIPs) != 1 {
		t.Fatalf("AllowedIPs count = %d, want 1", len(peer.AllowedIPs))
	}
	if !peer.AllowedIPs[0].IP.Equal(tunnelIP) {
		t.Errorf("AllowedIPs[0].IP = %v, want %v", peer.AllowedIPs[0].IP, tunnelIP)
	}
	ones, bits := peer.AllowedIPs[0].Mask.Size()
	if ones != 32 || bits != 32 {
		t.Errorf("AllowedIPs mask = /%d (of %d), want /32", ones, bits)
	}
	if peer.PersistentKeepaliveInterval == nil || *peer.PersistentKeepaliveInterval != 25*time.Second {
		t.Error("PersistentKeepaliveInterval should be 25s")
	}

	// Verify iptables commands (DNAT + FORWARD).
	if len(runner.calls) != 2 {
		t.Fatalf("iptables calls = %d, want 2", len(runner.calls))
	}

	// DNAT rule.
	if !strings.Contains(runner.calls[0], "-t nat -A PREROUTING") {
		t.Errorf("DNAT rule missing PREROUTING: %s", runner.calls[0])
	}
	if !strings.Contains(runner.calls[0], "--dport 10005") {
		t.Errorf("DNAT rule missing correct port: %s", runner.calls[0])
	}
	if !strings.Contains(runner.calls[0], "--to-destination 10.0.0.5:22") {
		t.Errorf("DNAT rule missing destination: %s", runner.calls[0])
	}

	// FORWARD rule.
	if !strings.Contains(runner.calls[1], "-A FORWARD") {
		t.Errorf("FORWARD rule missing: %s", runner.calls[1])
	}
	if !strings.Contains(runner.calls[1], "-d 10.0.0.5") {
		t.Errorf("FORWARD rule missing destination IP: %s", runner.calls[1])
	}
	if !strings.Contains(runner.calls[1], "--dport 22") {
		t.Errorf("FORWARD rule missing SSH port: %s", runner.calls[1])
	}
}

func TestAddPeerInvalidKey(t *testing.T) {
	client := &mockWGClient{}
	runner := newMockRunner()
	mgr := NewManager(client, runner, "wg0", managerTestLogger())

	tunnelIP := net.ParseIP("10.0.0.5").To4()

	err := mgr.AddPeer(context.Background(), "not-a-valid-base64-key!!!", tunnelIP, 10005)
	if err == nil {
		t.Error("AddPeer() with invalid key should return error, got nil")
	}

	// No WireGuard or iptables calls should have been made.
	if len(client.configureCalls) != 0 {
		t.Errorf("ConfigureDevice calls = %d, want 0", len(client.configureCalls))
	}
	if len(runner.calls) != 0 {
		t.Errorf("iptables calls = %d, want 0", len(runner.calls))
	}
}

func TestAddPeerWGFailure(t *testing.T) {
	client := &mockWGClient{
		configureErr: fmt.Errorf("device not found"),
	}
	runner := newMockRunner()
	mgr := NewManager(client, runner, "wg0", managerTestLogger())

	pubKey := testWGKey(t)
	tunnelIP := net.ParseIP("10.0.0.5").To4()

	err := mgr.AddPeer(context.Background(), pubKey.String(), tunnelIP, 10005)
	if err == nil {
		t.Error("AddPeer() with WG failure should return error, got nil")
	}

	// iptables should NOT have been called since WireGuard failed first.
	if len(runner.calls) != 0 {
		t.Errorf("iptables calls = %d, want 0 (WG failed, no iptables needed)", len(runner.calls))
	}
}

func TestAddPeerIptablesFailure(t *testing.T) {
	// WireGuard succeeds on first call, but should also be called for rollback.
	client := &mockWGClient{}
	runner := newMockRunner()
	runner.err = fmt.Errorf("iptables: permission denied")
	runner.errOn = 0 // Fail on first iptables call (DNAT rule).
	mgr := NewManager(client, runner, "wg0", managerTestLogger())

	pubKey := testWGKey(t)
	tunnelIP := net.ParseIP("10.0.0.5").To4()

	err := mgr.AddPeer(context.Background(), pubKey.String(), tunnelIP, 10005)
	if err == nil {
		t.Error("AddPeer() with iptables failure should return error, got nil")
	}

	// WireGuard should have been called twice: once to add, once to remove (rollback).
	if len(client.configureCalls) != 2 {
		t.Fatalf("ConfigureDevice calls = %d, want 2 (add + rollback)", len(client.configureCalls))
	}

	// Second call should be a remove.
	rollbackCfg := client.configureCalls[1]
	if len(rollbackCfg.Peers) != 1 || !rollbackCfg.Peers[0].Remove {
		t.Error("rollback call should have Remove=true")
	}
}

func TestRemovePeer(t *testing.T) {
	client := &mockWGClient{}
	runner := newMockRunner()
	mgr := NewManager(client, runner, "wg0", managerTestLogger())

	pubKey := testWGKey(t)
	tunnelIP := net.ParseIP("10.0.0.5").To4()
	externalPort := 10005

	err := mgr.RemovePeer(context.Background(), pubKey.String(), tunnelIP, externalPort)
	if err != nil {
		t.Fatalf("RemovePeer() error: %v", err)
	}

	// Verify iptables removal called before WireGuard peer removal.
	if len(runner.calls) != 2 {
		t.Fatalf("iptables calls = %d, want 2", len(runner.calls))
	}

	// DNAT removal.
	if !strings.Contains(runner.calls[0], "-t nat -D PREROUTING") {
		t.Errorf("DNAT removal missing -D PREROUTING: %s", runner.calls[0])
	}
	if !strings.Contains(runner.calls[0], "--dport 10005") {
		t.Errorf("DNAT removal missing correct port: %s", runner.calls[0])
	}
	if !strings.Contains(runner.calls[0], "--to-destination 10.0.0.5:22") {
		t.Errorf("DNAT removal missing destination: %s", runner.calls[0])
	}

	// FORWARD removal.
	if !strings.Contains(runner.calls[1], "-D FORWARD") {
		t.Errorf("FORWARD removal missing -D: %s", runner.calls[1])
	}

	// WireGuard peer removal.
	if len(client.configureCalls) != 1 {
		t.Fatalf("ConfigureDevice calls = %d, want 1", len(client.configureCalls))
	}
	cfg := client.configureCalls[0]
	if len(cfg.Peers) != 1 || !cfg.Peers[0].Remove {
		t.Error("peer config should have Remove=true")
	}
}

func TestRemovePeerIptablesFailureNonFatal(t *testing.T) {
	client := &mockWGClient{}
	runner := newMockRunner()
	runner.err = fmt.Errorf("iptables: rule not found")
	// All iptables calls fail, but RemovePeer should still succeed.
	mgr := NewManager(client, runner, "wg0", managerTestLogger())

	pubKey := testWGKey(t)
	tunnelIP := net.ParseIP("10.0.0.5").To4()

	err := mgr.RemovePeer(context.Background(), pubKey.String(), tunnelIP, 10005)
	if err != nil {
		t.Fatalf("RemovePeer() should succeed even with iptables failure, got: %v", err)
	}

	// WireGuard peer should still have been removed.
	if len(client.configureCalls) != 1 {
		t.Fatalf("ConfigureDevice calls = %d, want 1", len(client.configureCalls))
	}
	if !client.configureCalls[0].Peers[0].Remove {
		t.Error("peer should be removed even when iptables cleanup fails")
	}
}

func TestListPeers(t *testing.T) {
	handshakeTime := time.Date(2026, 2, 24, 12, 0, 0, 0, time.UTC)
	key1 := testWGKey(t)
	key2 := testWGKey(t)

	client := &mockWGClient{
		device: &wgtypes.Device{
			Name: "wg0",
			Peers: []wgtypes.Peer{
				{
					PublicKey:         key1,
					AllowedIPs:        []net.IPNet{{IP: net.ParseIP("10.0.0.2").To4(), Mask: net.CIDRMask(32, 32)}},
					LastHandshakeTime: handshakeTime,
					ReceiveBytes:      1024,
					TransmitBytes:     2048,
				},
				{
					PublicKey:         key2,
					AllowedIPs:        []net.IPNet{{IP: net.ParseIP("10.0.0.3").To4(), Mask: net.CIDRMask(32, 32)}},
					LastHandshakeTime: handshakeTime.Add(5 * time.Minute),
					ReceiveBytes:      4096,
					TransmitBytes:     8192,
				},
			},
		},
	}
	runner := newMockRunner()
	mgr := NewManager(client, runner, "wg0", managerTestLogger())

	peers, err := mgr.ListPeers(context.Background())
	if err != nil {
		t.Fatalf("ListPeers() error: %v", err)
	}

	if len(peers) != 2 {
		t.Fatalf("peers count = %d, want 2", len(peers))
	}

	// Verify first peer.
	if peers[0].PublicKey != key1.String() {
		t.Errorf("peer[0] public key mismatch")
	}
	if len(peers[0].AllowedIPs) != 1 || peers[0].AllowedIPs[0] != "10.0.0.2/32" {
		t.Errorf("peer[0] AllowedIPs = %v, want [10.0.0.2/32]", peers[0].AllowedIPs)
	}
	if !peers[0].LastHandshake.Equal(handshakeTime) {
		t.Errorf("peer[0] LastHandshake = %v, want %v", peers[0].LastHandshake, handshakeTime)
	}
	if peers[0].TransferRx != 1024 {
		t.Errorf("peer[0] TransferRx = %d, want 1024", peers[0].TransferRx)
	}
	if peers[0].TransferTx != 2048 {
		t.Errorf("peer[0] TransferTx = %d, want 2048", peers[0].TransferTx)
	}

	// Verify second peer.
	if peers[1].PublicKey != key2.String() {
		t.Errorf("peer[1] public key mismatch")
	}
	if peers[1].TransferRx != 4096 {
		t.Errorf("peer[1] TransferRx = %d, want 4096", peers[1].TransferRx)
	}
}

func TestPortFromTunnelIP(t *testing.T) {
	tests := []struct {
		ip   string
		want int
	}{
		{"10.0.0.2", 10002},
		{"10.0.0.5", 10005},
		{"10.0.0.255", 10255},
		{"10.0.1.0", 10256},
		{"10.0.10.5", 12565},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s->%d", tt.ip, tt.want), func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			got := PortFromTunnelIP(ip)
			if got != tt.want {
				t.Errorf("PortFromTunnelIP(%s) = %d, want %d", tt.ip, got, tt.want)
			}
		})
	}
}
