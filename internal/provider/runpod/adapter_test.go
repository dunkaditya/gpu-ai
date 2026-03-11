package runpod

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gpuai/gpuctl/internal/provider"
)

// setupTestAdapter creates a test adapter pointing at a local httptest server.
func setupTestAdapter(handler http.HandlerFunc) (*Adapter, *httptest.Server) {
	server := httptest.NewServer(handler)
	adapter := NewAdapter("test-api-key", WithBaseURL(server.URL))
	return adapter, server
}

// mustReadBody reads and returns the request body as a string.
func mustReadBody(r *http.Request) string {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	return string(body)
}

func TestListAvailable(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header.
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-api-key" {
			t.Errorf("unexpected auth header: %s", auth)
		}

		resp := `{
			"data": {
				"gpuTypes": [
					{
						"id": "NVIDIA GeForce RTX 4090",
						"displayName": "NVIDIA GeForce RTX 4090",
						"memoryInGb": 24,
						"secureCloud": true,
						"communityCloud": true,
						"securePrice": 0.74,
						"communityPrice": 0.34,
						"secureSpotPrice": 0.0,
						"communitySpotPrice": 0.34,
						"secureLowestPrice": {
							"minimumBidPrice": 0.34,
							"uninterruptablePrice": 0.74,
							"minVcpu": 16,
							"minMemory": 62,
							"stockStatus": "High",
							"maxUnreservedGpuCount": 15
						}
					},
					{
						"id": "NVIDIA RTX 9090 SUPER",
						"displayName": "NVIDIA RTX 9090 SUPER",
						"memoryInGb": 128,
						"secureCloud": true,
						"communityCloud": false,
						"securePrice": 5.00,
						"communityPrice": 0.0,
						"secureSpotPrice": 0.0,
						"communitySpotPrice": 0.0,
						"secureLowestPrice": {
							"minimumBidPrice": 0.0,
							"uninterruptablePrice": 5.00,
							"minVcpu": 4,
							"minMemory": 8,
							"stockStatus": "Low",
							"maxUnreservedGpuCount": 2
						}
					}
				]
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(resp))
	}

	adapter, server := setupTestAdapter(handler)
	defer server.Close()

	offerings, err := adapter.ListAvailable(context.Background())
	if err != nil {
		t.Fatalf("ListAvailable returned error: %v", err)
	}

	// RTX 4090 is mapped, RTX 9090 SUPER is not. On-demand only.
	// Expected: 1 offering (on-demand for RTX 4090), 0 for unmapped.
	if len(offerings) != 1 {
		t.Fatalf("expected 1 offering, got %d", len(offerings))
	}

	onDemand := offerings[0]
	if onDemand.Tier != provider.TierOnDemand {
		t.Errorf("expected tier on_demand, got %s", onDemand.Tier)
	}
	if onDemand.GPUType != provider.GPUTypeRTX4090 {
		t.Errorf("expected GPU type %s, got %s", provider.GPUTypeRTX4090, onDemand.GPUType)
	}
	if onDemand.PricePerHour != 0.74 {
		t.Errorf("expected on-demand price 0.74, got %f", onDemand.PricePerHour)
	}
	if onDemand.VRAMPerGPUGB != 24 {
		t.Errorf("expected VRAM 24 GB, got %d", onDemand.VRAMPerGPUGB)
	}
	if onDemand.CPUCores != 16 {
		t.Errorf("expected 16 CPU cores, got %d", onDemand.CPUCores)
	}
	if onDemand.RAMGB != 62 {
		t.Errorf("expected 62 GB RAM, got %d", onDemand.RAMGB)
	}
	if onDemand.Provider != "runpod" {
		t.Errorf("expected provider runpod, got %s", onDemand.Provider)
	}
	if onDemand.GPUCount != 1 {
		t.Errorf("expected GPU count 1, got %d", onDemand.GPUCount)
	}
	if onDemand.StockStatus != "High" {
		t.Errorf("expected stock status High, got %s", onDemand.StockStatus)
	}
	if onDemand.AvailableCount != 15 {
		t.Errorf("expected available count 15, got %d", onDemand.AvailableCount)
	}
}

func TestListAvailableGraphQLError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		resp := `{
			"data": null,
			"errors": [
				{"message": "Authentication required. Please add your API key."}
			]
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(resp))
	}

	adapter, server := setupTestAdapter(handler)
	defer server.Close()

	_, err := adapter.ListAvailable(context.Background())
	if err == nil {
		t.Fatal("expected error from GraphQL error response, got nil")
	}
	if !strings.Contains(err.Error(), "Authentication required") {
		t.Errorf("expected auth error message, got: %v", err)
	}
}

func TestProvisionOnDemand(t *testing.T) {
	var receivedBody string
	handler := func(w http.ResponseWriter, r *http.Request) {
		receivedBody = mustReadBody(r)

		resp := `{
			"data": {
				"podFindAndDeployOnDemand": {
					"id": "pod-abc123",
					"costPerHr": 0.74,
					"desiredStatus": "CREATED",
					"lastStatusChange": "2026-02-24T18:00:00Z",
					"machine": {
						"gpuDisplayName": "NVIDIA GeForce RTX 4090",
						"location": "US-TX-3"
					}
				}
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(resp))
	}

	adapter, server := setupTestAdapter(handler)
	defer server.Close()

	result, err := adapter.Provision(context.Background(), provider.ProvisionRequest{
		InstanceID:    "gpu-test-001",
		GPUType:       provider.GPUTypeRTX4090,
		GPUCount:      1,
		Tier:          provider.TierOnDemand,
		SSHPublicKeys: []string{"ssh-rsa AAAA... test@example.com"},
		InternalToken: "token-secret-123",
		CallbackURL:   "https://api.gpu.ai/callbacks",
	})
	if err != nil {
		t.Fatalf("Provision returned error: %v", err)
	}

	// Verify result fields.
	if result.UpstreamID != "pod-abc123" {
		t.Errorf("expected upstream ID pod-abc123, got %s", result.UpstreamID)
	}
	if result.CostPerHour != 0.74 {
		t.Errorf("expected cost 0.74, got %f", result.CostPerHour)
	}
	if result.Provider != "runpod" {
		t.Errorf("expected provider runpod, got %s", result.Provider)
	}
	if result.Status != "creating" {
		t.Errorf("expected status creating, got %s", result.Status)
	}
	if result.DatacenterLocation != "US-TX-3" {
		t.Errorf("expected datacenter US-TX-3, got %s", result.DatacenterLocation)
	}

	// Verify the request body contains correct mutation and params.
	if !strings.Contains(receivedBody, "podFindAndDeployOnDemand") {
		t.Error("expected podFindAndDeployOnDemand mutation in request")
	}
	if !strings.Contains(receivedBody, "SECURE") {
		t.Error("expected cloudType SECURE in request body")
	}
	if !strings.Contains(receivedBody, "NVIDIA GeForce RTX 4090") {
		t.Error("expected GPU type in request body")
	}
	if !strings.Contains(receivedBody, "GPUAI_INSTANCE_ID") {
		t.Error("expected GPUAI_INSTANCE_ID env var in request body")
	}
	if !strings.Contains(receivedBody, "GPUAI_CALLBACK_URL") {
		t.Error("expected GPUAI_CALLBACK_URL env var in request body")
	}
	if !strings.Contains(receivedBody, "GPUAI_INTERNAL_TOKEN") {
		t.Error("expected GPUAI_INTERNAL_TOKEN env var in request body")
	}
	if !strings.Contains(receivedBody, "PUBLIC_KEY") {
		t.Error("expected PUBLIC_KEY env var in request body")
	}
}

func TestProvisionSpot(t *testing.T) {
	var receivedBody string
	handler := func(w http.ResponseWriter, r *http.Request) {
		receivedBody = mustReadBody(r)

		resp := `{
			"data": {
				"podRentInterruptable": {
					"id": "pod-spot-456",
					"costPerHr": 0.34,
					"desiredStatus": "CREATED",
					"lastStatusChange": "2026-02-24T18:00:00Z",
					"machine": {
						"gpuDisplayName": "NVIDIA GeForce RTX 4090",
						"location": "US-CA-1"
					}
				}
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(resp))
	}

	adapter, server := setupTestAdapter(handler)
	defer server.Close()

	result, err := adapter.Provision(context.Background(), provider.ProvisionRequest{
		InstanceID:    "gpu-spot-001",
		GPUType:       provider.GPUTypeRTX4090,
		GPUCount:      1,
		Tier:          provider.TierSpot,
		InternalToken: "token-123",
		CallbackURL:   "https://api.gpu.ai/callbacks",
	})
	if err != nil {
		t.Fatalf("Provision(spot) returned error: %v", err)
	}

	if result.UpstreamID != "pod-spot-456" {
		t.Errorf("expected upstream ID pod-spot-456, got %s", result.UpstreamID)
	}
	if result.CostPerHour != 0.34 {
		t.Errorf("expected cost 0.34, got %f", result.CostPerHour)
	}

	// Verify spot-specific mutation and params.
	if !strings.Contains(receivedBody, "podRentInterruptable") {
		t.Error("expected podRentInterruptable mutation in request")
	}
	if !strings.Contains(receivedBody, "COMMUNITY") {
		t.Error("expected cloudType COMMUNITY in request body")
	}
	if !strings.Contains(receivedBody, "bidPerGpu") {
		t.Error("expected bidPerGpu in request body")
	}
}

func TestProvisionNoCapacity(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		resp := `{
			"data": null,
			"errors": [
				{"message": "There is no available capacity for the requested GPU type"}
			]
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(resp))
	}

	adapter, server := setupTestAdapter(handler)
	defer server.Close()

	_, err := adapter.Provision(context.Background(), provider.ProvisionRequest{
		InstanceID:    "gpu-fail-001",
		GPUType:       provider.GPUTypeRTX4090,
		GPUCount:      1,
		Tier:          provider.TierOnDemand,
		InternalToken: "token-123",
		CallbackURL:   "https://api.gpu.ai/callbacks",
	})
	if err == nil {
		t.Fatal("expected error for no capacity, got nil")
	}
	if !errors.Is(err, provider.ErrNoCapacity) {
		t.Errorf("expected ErrNoCapacity, got: %v", err)
	}
}

func TestGetStatus(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		body := mustReadBody(r)
		if !strings.Contains(body, "pod-running-123") {
			t.Error("expected pod ID in request body")
		}

		resp := `{
			"data": {
				"pod": {
					"id": "pod-running-123",
					"desiredStatus": "RUNNING",
					"lastStatusChange": "2026-02-24T18:00:00Z",
					"costPerHr": 0.74,
					"runtime": {
						"uptimeInSeconds": 3600,
						"ports": [
							{
								"ip": "192.0.2.100",
								"isIpPublic": true,
								"privatePort": 22,
								"publicPort": 22100,
								"type": "tcp"
							},
							{
								"ip": "10.0.0.5",
								"isIpPublic": false,
								"privatePort": 8080,
								"publicPort": 8080,
								"type": "http"
							}
						]
					},
					"machine": {
						"gpuDisplayName": "NVIDIA GeForce RTX 4090",
						"location": "US-TX-3"
					}
				}
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(resp))
	}

	adapter, server := setupTestAdapter(handler)
	defer server.Close()

	status, err := adapter.GetStatus(context.Background(), "pod-running-123")
	if err != nil {
		t.Fatalf("GetStatus returned error: %v", err)
	}

	if status.Status != "running" {
		t.Errorf("expected status running, got %s", status.Status)
	}
	if status.UpstreamID != "pod-running-123" {
		t.Errorf("expected upstream ID pod-running-123, got %s", status.UpstreamID)
	}
	if status.IP != "192.0.2.100" {
		t.Errorf("expected IP 192.0.2.100, got %s", status.IP)
	}
	if status.CostPerHour != 0.74 {
		t.Errorf("expected cost 0.74, got %f", status.CostPerHour)
	}
	if status.UptimeSeconds != 3600 {
		t.Errorf("expected uptime 3600, got %d", status.UptimeSeconds)
	}
	if len(status.Ports) != 2 {
		t.Fatalf("expected 2 ports, got %d", len(status.Ports))
	}
	if status.Ports[0].PublicPort != 22100 {
		t.Errorf("expected public port 22100, got %d", status.Ports[0].PublicPort)
	}
}

func TestTerminate(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		body := mustReadBody(r)
		if !strings.Contains(body, "pod-term-789") {
			t.Error("expected pod ID in request body")
		}

		// podTerminate returns null on success.
		resp := `{
			"data": {
				"podTerminate": null
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(resp))
	}

	adapter, server := setupTestAdapter(handler)
	defer server.Close()

	err := adapter.Terminate(context.Background(), "pod-term-789")
	if err != nil {
		t.Fatalf("Terminate returned error: %v", err)
	}
}

func TestRetryOnServerError(t *testing.T) {
	var callCount atomic.Int32

	handler := func(w http.ResponseWriter, r *http.Request) {
		count := callCount.Add(1)

		if count == 1 {
			// First call: return 500.
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal server error"}`))
			return
		}

		// Second call: return success.
		resp := `{
			"data": {
				"podTerminate": null
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(resp))
	}

	adapter, server := setupTestAdapter(handler)
	defer server.Close()

	err := adapter.Terminate(context.Background(), "pod-retry-001")
	if err != nil {
		t.Fatalf("expected success after retry, got error: %v", err)
	}

	if callCount.Load() != 2 {
		t.Errorf("expected 2 calls (1 failure + 1 success), got %d", callCount.Load())
	}
}

func TestNormalizeGPUName(t *testing.T) {
	tests := []struct {
		input    string
		expected provider.GPUType
		ok       bool
	}{
		{"NVIDIA GeForce RTX 4090", provider.GPUTypeRTX4090, true},
		{"NVIDIA A100 80GB PCIe", provider.GPUTypeA10080GB, true},
		{"NVIDIA A100-SXM4-80GB", provider.GPUTypeA10080GB, true},
		{"NVIDIA H100 80GB HBM3", provider.GPUTypeH100SXM, true},
		{"NVIDIA H200", provider.GPUTypeH200SXM, true},
		{"Unknown GPU Model", "", false},
	}

	for _, tt := range tests {
		gpuType, ok := NormalizeGPUName(tt.input)
		if ok != tt.ok {
			t.Errorf("NormalizeGPUName(%q): ok = %v, want %v", tt.input, ok, tt.ok)
		}
		if gpuType != tt.expected {
			t.Errorf("NormalizeGPUName(%q) = %q, want %q", tt.input, gpuType, tt.expected)
		}
	}
}

func TestNormalizeRegion(t *testing.T) {
	tests := []struct {
		input              string
		expectedRegion     string
		expectedDatacenter string
	}{
		{"US-TX-3", "us-central", "US-TX-3"},
		{"US-CA-1", "us-west", "US-CA-1"},
		{"EU-RO-1", "eu-east", "EU-RO-1"},
		{"CA-MTL-1", "ca-central", "CA-MTL-1"},
		{"", "unknown", "Unknown"},
		{"JP-TK-1", "unknown", "JP-TK-1"},
	}

	for _, tt := range tests {
		region, dc := NormalizeRegion(tt.input)
		if region != tt.expectedRegion {
			t.Errorf("NormalizeRegion(%q): region = %q, want %q", tt.input, region, tt.expectedRegion)
		}
		if dc != tt.expectedDatacenter {
			t.Errorf("NormalizeRegion(%q): datacenter = %q, want %q", tt.input, dc, tt.expectedDatacenter)
		}
	}
}

func TestProvisionInvalidGPUType(t *testing.T) {
	// No server needed -- should fail before making API call.
	adapter := NewAdapter("test-key", WithBaseURL("http://localhost:1"))

	_, err := adapter.Provision(context.Background(), provider.ProvisionRequest{
		InstanceID:    "gpu-bad-001",
		GPUType:       "nonexistent_gpu",
		GPUCount:      1,
		Tier:          provider.TierOnDemand,
		InternalToken: "token-123",
		CallbackURL:   "https://api.gpu.ai/callbacks",
	})
	if err == nil {
		t.Fatal("expected error for invalid GPU type")
	}
	if !errors.Is(err, provider.ErrInvalidGPUType) {
		t.Errorf("expected ErrInvalidGPUType, got: %v", err)
	}
}

func TestReverseGPUNameMap(t *testing.T) {
	// Verify that all canonical types that appear in gpuNameMap have a reverse mapping.
	seenTypes := make(map[provider.GPUType]bool)
	for _, gpuType := range gpuNameMap {
		seenTypes[gpuType] = true
	}

	for gpuType := range seenTypes {
		name, ok := RunPodGPUName(gpuType)
		if !ok {
			t.Errorf("RunPodGPUName(%q) returned false, expected true", gpuType)
		}
		if name == "" {
			t.Errorf("RunPodGPUName(%q) returned empty name", gpuType)
		}
	}
}

func TestGraphQLRequestFormat(t *testing.T) {
	// Verify that the client sends properly formatted GraphQL requests.
	var receivedReq graphQLRequest
	handler := func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedReq)

		// Return valid terminate response.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"podTerminate": null}}`))
	}

	adapter, server := setupTestAdapter(handler)
	defer server.Close()

	adapter.Terminate(context.Background(), "pod-format-test")

	if receivedReq.Query == "" {
		t.Error("expected non-empty query in request")
	}
	if receivedReq.Variables == nil {
		t.Error("expected non-nil variables in request")
	}
	inputRaw, ok := receivedReq.Variables["input"]
	if !ok {
		t.Fatal("expected input variable in request")
	}
	inputMap, ok := inputRaw.(map[string]any)
	if !ok {
		t.Fatalf("expected input to be map, got %T", inputRaw)
	}
	if inputMap["podId"] != "pod-format-test" {
		t.Errorf("expected podId pod-format-test, got %v", inputMap["podId"])
	}
}
