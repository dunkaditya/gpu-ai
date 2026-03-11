package runpod

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gpuai/gpuctl/internal/provider"
)

// encodeScriptForDockerArgs base64-encodes a startup script so it can be
// safely embedded in a dockerArgs string without shell escaping issues.
func encodeScriptForDockerArgs(script string) string {
	return base64.StdEncoding.EncodeToString([]byte(script))
}

// Compile-time check: Adapter implements provider.Provider.
var _ provider.Provider = (*Adapter)(nil)

const (
	maxRetryAttempts = 3

	// Per-operation default timeouts from CONTEXT.md decisions.
	listTimeout      = 10 * time.Second
	provisionTimeout = 30 * time.Second
	statusTimeout    = 10 * time.Second
	terminateTimeout = 30 * time.Second

	defaultDockerImage    = "runpod/pytorch:2.4.0-py3.11-cuda12.4.1-devel-ubuntu22.04"
	defaultContainerDisk  = 40
	defaultEstReadySecond = 30
)

// Adapter implements the provider.Provider interface for RunPod GPU Cloud.
type Adapter struct {
	client *Client
}

// NewAdapter creates a new RunPod adapter with the given API key and options.
func NewAdapter(apiKey string, opts ...ClientOption) *Adapter {
	return &Adapter{client: NewClient(apiKey, opts...)}
}

// Name returns the provider identifier.
func (a *Adapter) Name() string {
	return "runpod"
}

// --- RunPod API response types ---

type gpuTypesResponse struct {
	GPUTypes []runpodGPUType `json:"gpuTypes"`
}

type runpodGPUType struct {
	ID               string       `json:"id"`
	DisplayName      string       `json:"displayName"`
	MemoryInGB       int          `json:"memoryInGb"`
	SecureCloud      bool         `json:"secureCloud"`
	CommunityCloud   bool         `json:"communityCloud"`
	SecurePrice      float64      `json:"securePrice"`
	CommunityPrice   float64      `json:"communityPrice"`
	SecureSpotPrice  float64      `json:"secureSpotPrice"`
	CommunitySpotPrc float64      `json:"communitySpotPrice"`
	LowestPrice      *lowestPrice `json:"secureLowestPrice"`
}

type lowestPrice struct {
	MinimumBidPrice        float64 `json:"minimumBidPrice"`
	UninterruptablePrice   float64 `json:"uninterruptablePrice"`
	MinVcpu                int     `json:"minVcpu"`
	MinMemory              int     `json:"minMemory"`
	StockStatus            string  `json:"stockStatus"`
	MaxUnreservedGPUCount  int     `json:"maxUnreservedGpuCount"`
}

type createPodResponse struct {
	PodFindAndDeployOnDemand *podResult `json:"podFindAndDeployOnDemand"`
}

type createSpotPodResponse struct {
	PodRentInterruptable *podResult `json:"podRentInterruptable"`
}

type podResult struct {
	ID               string   `json:"id"`
	CostPerHr        float64  `json:"costPerHr"`
	DesiredStatus    string   `json:"desiredStatus"`
	LastStatusChange string   `json:"lastStatusChange"`
	Machine          *machine `json:"machine"`
}

type getPodResponse struct {
	Pod *podDetail `json:"pod"`
}

type podDetail struct {
	ID               string   `json:"id"`
	DesiredStatus    string   `json:"desiredStatus"`
	LastStatusChange string   `json:"lastStatusChange"`
	CostPerHr        float64  `json:"costPerHr"`
	Runtime          *runtime `json:"runtime"`
	Machine          *machine `json:"machine"`
}

type runtime struct {
	UptimeInSeconds int        `json:"uptimeInSeconds"`
	Ports           []portInfo `json:"ports"`
}

type portInfo struct {
	IP          string `json:"ip"`
	IsIPPublic  bool   `json:"isIpPublic"`
	PrivatePort int    `json:"privatePort"`
	PublicPort  int    `json:"publicPort"`
	Type        string `json:"type"`
}

type machine struct {
	GPUDisplayName string `json:"gpuDisplayName"`
	Location       string `json:"location"`
}

type terminateResponse struct {
	PodTerminate any `json:"podTerminate"`
}

// envVar is a RunPod environment variable key-value pair for pod creation.
type envVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ListAvailable queries RunPod for current GPU inventory and pricing.
func (a *Adapter) ListAvailable(ctx context.Context) ([]provider.GPUOffering, error) {
	ctx, cancel := contextWithDefaultTimeout(ctx, listTimeout)
	defer cancel()

	var resp gpuTypesResponse
	if err := a.client.doWithRetry(ctx, "ListAvailable", maxRetryAttempts, graphQLRequest{
		Query: queryGPUTypes,
	}, &resp); err != nil {
		return nil, fmt.Errorf("list available GPUs: %w", err)
	}

	var offerings []provider.GPUOffering

	for _, gpu := range resp.GPUTypes {
		gpuType, ok := NormalizeGPUName(gpu.ID)
		if !ok {
			slog.Warn("unmapped GPU type",
				"runpod_name", gpu.DisplayName,
				"runpod_id", gpu.ID,
			)
			continue
		}

		if !gpu.SecureCloud || gpu.SecurePrice <= 0 {
			continue
		}

		var availableCount, cpuCores, ramGB int
		var stockStatus string
		if gpu.LowestPrice != nil {
			availableCount = gpu.LowestPrice.MaxUnreservedGPUCount
			stockStatus = gpu.LowestPrice.StockStatus
			cpuCores = gpu.LowestPrice.MinVcpu
			ramGB = gpu.LowestPrice.MinMemory
		}

		region, dcLocation := NormalizeRegion("")
		offerings = append(offerings, provider.GPUOffering{
			Provider:           "runpod",
			GPUType:            gpuType,
			GPUCount:           1,
			VRAMPerGPUGB:       gpu.MemoryInGB,
			CPUCores:           cpuCores,
			RAMGB:              ramGB,
			PricePerHour:       gpu.SecurePrice,
			Tier:               provider.TierOnDemand,
			Region:             region,
			DatacenterLocation: dcLocation,
			StockStatus:        stockStatus,
			AvailableCount:     availableCount,
		})
	}

	slog.Info("listed RunPod GPU offerings",
		"total_offerings", len(offerings),
	)

	return offerings, nil
}

// Provision creates a new GPU instance on RunPod.
func (a *Adapter) Provision(ctx context.Context, req provider.ProvisionRequest) (*provider.ProvisionResult, error) {
	ctx, cancel := contextWithDefaultTimeout(ctx, provisionTimeout)
	defer cancel()

	// Reverse-lookup the RunPod GPU name from our canonical type.
	rpGPUName, ok := RunPodGPUName(req.GPUType)
	if !ok {
		return nil, fmt.Errorf("GPU type %s: %w", req.GPUType, provider.ErrInvalidGPUType)
	}

	gpuCount := req.GPUCount
	if gpuCount == 0 {
		gpuCount = 1
	}

	dockerImage := req.DockerImage
	if dockerImage == "" {
		dockerImage = defaultDockerImage
	}

	// Build environment variables.
	envVars := []envVar{
		{Key: "GPUAI_INSTANCE_ID", Value: req.InstanceID},
		{Key: "GPUAI_CALLBACK_URL", Value: req.CallbackURL},
		{Key: "GPUAI_INTERNAL_TOKEN", Value: req.InternalToken},
	}
	if len(req.SSHPublicKeys) > 0 {
		envVars = append(envVars, envVar{
			Key:   "PUBLIC_KEY",
			Value: strings.Join(req.SSHPublicKeys, "\n"),
		})
	}

	switch req.Tier {
	case provider.TierOnDemand:
		return a.provisionOnDemand(ctx, rpGPUName, gpuCount, dockerImage, req.InstanceID, envVars, req.StartupScript)
	case provider.TierSpot:
		return a.provisionSpot(ctx, rpGPUName, gpuCount, dockerImage, req.InstanceID, envVars, req.StartupScript, req.BidPricePerGPU)
	default:
		return nil, fmt.Errorf("unsupported tier: %s", req.Tier)
	}
}

func (a *Adapter) provisionOnDemand(ctx context.Context, gpuTypeID string, gpuCount int, imageName, name string, env []envVar, startupScript string) (*provider.ProvisionResult, error) {
	input := map[string]any{
		"cloudType":         "SECURE",
		"gpuCount":          gpuCount,
		"gpuTypeId":         gpuTypeID,
		"name":              name,
		"imageName":         imageName,
		"containerDiskInGb": defaultContainerDisk,
		"volumeInGb":        0,
		"startSsh":          true,
		"ports":             "22/tcp",
		"env":               env,
	}
	if startupScript != "" {
		input["dockerArgs"] = "bash -c 'echo " + encodeScriptForDockerArgs(startupScript) + " | base64 -d | bash'"
	}

	var resp createPodResponse
	if err := a.client.doWithRetry(ctx, "Provision(on-demand)", maxRetryAttempts, graphQLRequest{
		Query:     mutationCreateOnDemandPod,
		Variables: map[string]any{"input": input},
	}, &resp); err != nil {
		return nil, fmt.Errorf("provision on-demand pod: %w", err)
	}

	if resp.PodFindAndDeployOnDemand == nil {
		return nil, fmt.Errorf("provision on-demand pod: empty response")
	}

	pod := resp.PodFindAndDeployOnDemand
	dcLocation := ""
	if pod.Machine != nil {
		dcLocation = pod.Machine.Location
	}

	slog.Info("provisioned on-demand RunPod pod",
		"pod_id", pod.ID,
		"cost_per_hr", pod.CostPerHr,
		"datacenter", dcLocation,
	)

	return &provider.ProvisionResult{
		UpstreamID:            pod.ID,
		Provider:              "runpod",
		Status:                "creating",
		CostPerHour:           pod.CostPerHr,
		DatacenterLocation:    dcLocation,
		EstimatedReadySeconds: defaultEstReadySecond,
	}, nil
}

func (a *Adapter) provisionSpot(ctx context.Context, gpuTypeID string, gpuCount int, imageName, name string, env []envVar, startupScript string, bidPerGPU float64) (*provider.ProvisionResult, error) {
	input := map[string]any{
		"cloudType":         "COMMUNITY",
		"gpuCount":          gpuCount,
		"gpuTypeId":         gpuTypeID,
		"name":              name,
		"imageName":         imageName,
		"containerDiskInGb": defaultContainerDisk,
		"volumeInGb":        0,
		"startSsh":          true,
		"ports":             "22/tcp",
		"bidPerGpu":         bidPerGPU,
		"env":               env,
	}
	if startupScript != "" {
		input["dockerArgs"] = "bash -c 'echo " + encodeScriptForDockerArgs(startupScript) + " | base64 -d | bash'"
	}

	var resp createSpotPodResponse
	if err := a.client.doWithRetry(ctx, "Provision(spot)", maxRetryAttempts, graphQLRequest{
		Query:     mutationCreateSpotPod,
		Variables: map[string]any{"input": input},
	}, &resp); err != nil {
		return nil, fmt.Errorf("provision spot pod: %w", err)
	}

	if resp.PodRentInterruptable == nil {
		return nil, fmt.Errorf("provision spot pod: empty response")
	}

	pod := resp.PodRentInterruptable
	dcLocation := ""
	if pod.Machine != nil {
		dcLocation = pod.Machine.Location
	}

	slog.Info("provisioned spot RunPod pod",
		"pod_id", pod.ID,
		"cost_per_hr", pod.CostPerHr,
		"datacenter", dcLocation,
	)

	return &provider.ProvisionResult{
		UpstreamID:            pod.ID,
		Provider:              "runpod",
		Status:                "creating",
		CostPerHour:           pod.CostPerHr,
		DatacenterLocation:    dcLocation,
		EstimatedReadySeconds: defaultEstReadySecond,
	}, nil
}

// statusMap maps RunPod desiredStatus values to GPU.ai status strings.
var statusMap = map[string]string{
	"CREATED": "creating",
	"RUNNING": "running",
	"EXITED":  "terminated",
}

// GetStatus returns the current status of a RunPod pod.
func (a *Adapter) GetStatus(ctx context.Context, upstreamID string) (*provider.InstanceStatus, error) {
	ctx, cancel := contextWithDefaultTimeout(ctx, statusTimeout)
	defer cancel()

	var resp getPodResponse
	if err := a.client.doWithRetry(ctx, "GetStatus", maxRetryAttempts, graphQLRequest{
		Query:     queryGetPod,
		Variables: map[string]any{"input": map[string]any{"podId": upstreamID}},
	}, &resp); err != nil {
		return nil, fmt.Errorf("get pod status: %w", err)
	}

	if resp.Pod == nil {
		return nil, fmt.Errorf("pod %s not found", upstreamID)
	}

	pod := resp.Pod

	// Map RunPod status to GPU.ai status.
	status, ok := statusMap[pod.DesiredStatus]
	if !ok {
		status = strings.ToLower(pod.DesiredStatus)
	}

	result := &provider.InstanceStatus{
		UpstreamID:  pod.ID,
		Status:      status,
		CostPerHour: pod.CostPerHr,
	}

	// Extract IP and ports from runtime.
	if pod.Runtime != nil {
		result.UptimeSeconds = pod.Runtime.UptimeInSeconds
		for _, p := range pod.Runtime.Ports {
			result.Ports = append(result.Ports, provider.PortMapping{
				IP:          p.IP,
				PrivatePort: p.PrivatePort,
				PublicPort:  p.PublicPort,
				Protocol:    p.Type,
			})
			// Extract public IP from SSH port (port 22 with public IP).
			if p.PrivatePort == 22 && p.IsIPPublic && p.IP != "" {
				result.IP = p.IP
			}
		}
	}

	return result, nil
}

// Terminate destroys a RunPod pod.
func (a *Adapter) Terminate(ctx context.Context, upstreamID string) error {
	ctx, cancel := contextWithDefaultTimeout(ctx, terminateTimeout)
	defer cancel()

	var resp terminateResponse
	if err := a.client.doWithRetry(ctx, "Terminate", maxRetryAttempts, graphQLRequest{
		Query:     mutationTerminatePod,
		Variables: map[string]any{"input": map[string]any{"podId": upstreamID}},
	}, &resp); err != nil {
		return fmt.Errorf("terminate pod: %w", err)
	}

	slog.Info("terminated RunPod pod", "pod_id", upstreamID)
	return nil
}

// contextWithDefaultTimeout applies a default timeout to a context only if
// the context does not already have a deadline set.
func contextWithDefaultTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		return ctx, func() {} // No-op cancel; caller's deadline takes precedence.
	}
	return context.WithTimeout(ctx, timeout)
}
