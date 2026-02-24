package provider

import "time"

// GPUType enumerates supported GPU models.
type GPUType string

const (
	GPUTypeH100SXM  GPUType = "h100_sxm"
	GPUTypeH100PCIE GPUType = "h100_pcie"
	GPUTypeH200SXM  GPUType = "h200_sxm"
	GPUTypeA10080GB GPUType = "a100_80gb"
	GPUTypeA10040GB GPUType = "a100_40gb"
	GPUTypeL40S     GPUType = "l40s"
	GPUTypeA40      GPUType = "a40"
	GPUTypeL4       GPUType = "l4"
	GPUTypeRTX4090  GPUType = "rtx_4090"
	GPUTypeRTXA6000 GPUType = "rtx_a6000"
)

// InstanceTier represents the pricing/availability tier.
// Only two tiers for v1: on_demand and spot. No reserved tier.
type InstanceTier string

const (
	TierOnDemand InstanceTier = "on_demand"
	TierSpot     InstanceTier = "spot"
)

// GPUOffering represents an available GPU configuration from an upstream provider.
type GPUOffering struct {
	Provider           string       `json:"provider"`
	GPUType            GPUType      `json:"gpu_type"`
	GPUCount           int          `json:"gpu_count"`
	VRAMPerGPUGB       int          `json:"vram_per_gpu_gb"`
	PricePerHour       float64      `json:"price_per_hour"`
	Tier               InstanceTier `json:"tier"`
	Region             string       `json:"region"`
	DatacenterLocation string       `json:"datacenter_location"`
	StockStatus        string       `json:"stock_status"`
	AvailableCount     int          `json:"available_count"`
}

// ProvisionRequest contains the parameters for provisioning a new instance.
type ProvisionRequest struct {
	InstanceID       string       `json:"instance_id"`
	GPUType          GPUType      `json:"gpu_type"`
	GPUCount         int          `json:"gpu_count"`
	Tier             InstanceTier `json:"tier"`
	Region           string       `json:"region"`
	SSHPublicKeys    []string     `json:"ssh_public_keys"`
	DockerImage      string       `json:"docker_image,omitempty"`
	WireGuardAddress string       `json:"wireguard_address"`
	InternalToken    string       `json:"internal_token"`
	CallbackURL      string       `json:"callback_url"`
	StartupScript    string       `json:"startup_script,omitempty"`
}

// ProvisionResult is returned after a successful provisioning call.
type ProvisionResult struct {
	UpstreamID            string  `json:"upstream_id"`
	UpstreamIP            string  `json:"upstream_ip"`
	Provider              string  `json:"provider"`
	Status                string  `json:"status"`
	EstimatedReadySeconds int     `json:"estimated_ready_seconds"`
	CostPerHour           float64 `json:"cost_per_hour"`
	DatacenterLocation    string  `json:"datacenter_location"`
}

// PortMapping represents a port mapping on an upstream instance.
type PortMapping struct {
	IP          string `json:"ip"`
	PrivatePort int    `json:"private_port"`
	PublicPort  int    `json:"public_port"`
	Protocol    string `json:"protocol"`
}

// InstanceStatus represents the current state of an upstream instance.
type InstanceStatus struct {
	UpstreamID    string        `json:"upstream_id"`
	Status        string        `json:"status"` // creating, running, stopping, terminated, error
	IP            string        `json:"ip,omitempty"`
	CreatedAt     *time.Time    `json:"created_at,omitempty"`
	CostPerHour   float64       `json:"cost_per_hour,omitempty"`
	UptimeSeconds int           `json:"uptime_seconds,omitempty"`
	Ports         []PortMapping `json:"ports,omitempty"`
}
