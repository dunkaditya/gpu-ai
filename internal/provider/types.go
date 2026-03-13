package provider

import (
	"fmt"
	"time"
)

// GPUType enumerates supported GPU models.
type GPUType string

const (
	// Blackwell
	GPUTypeB200 GPUType = "b200"
	GPUTypeB300 GPUType = "b300"

	// Hopper
	GPUTypeH200SXM  GPUType = "h200_sxm"
	GPUTypeH200NVL  GPUType = "h200_nvl"
	GPUTypeH100SXM  GPUType = "h100_sxm"
	GPUTypeH100NVL  GPUType = "h100_nvl"
	GPUTypeH100PCIE GPUType = "h100_pcie"

	// Ampere data center
	GPUTypeA10080GB GPUType = "a100_80gb"
	GPUTypeA10040GB GPUType = "a100_40gb"
	GPUTypeA40      GPUType = "a40"
	GPUTypeA30      GPUType = "a30"

	// Ada Lovelace / Ampere professional
	GPUTypeL40S        GPUType = "l40s"
	GPUTypeL40         GPUType = "l40"
	GPUTypeL4          GPUType = "l4"
	GPUTypeRTX6000Ada  GPUType = "rtx_6000_ada"
	GPUTypeRTX5000Ada  GPUType = "rtx_5000_ada"
	GPUTypeRTX4000Ada  GPUType = "rtx_4000_ada"
	GPUTypeRTX2000Ada  GPUType = "rtx_2000_ada"
	GPUTypeRTXA6000    GPUType = "rtx_a6000"
	GPUTypeRTXA5000    GPUType = "rtx_a5000"
	GPUTypeRTXA4500    GPUType = "rtx_a4500"
	GPUTypeRTXA4000    GPUType = "rtx_a4000"
	GPUTypeRTXPro6000  GPUType = "rtx_pro_6000"
	GPUTypeRTXPro4500  GPUType = "rtx_pro_4500"

	// Consumer
	GPUTypeRTX5090 GPUType = "rtx_5090"
	GPUTypeRTX5080 GPUType = "rtx_5080"
	GPUTypeRTX4090 GPUType = "rtx_4090"
	GPUTypeRTX4080 GPUType = "rtx_4080"
	GPUTypeRTX3090 GPUType = "rtx_3090"
	GPUTypeRTX3080 GPUType = "rtx_3080"

	// Legacy
	GPUTypeV100 GPUType = "v100"
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
	CPUCores           int          `json:"cpu_cores"`
	RAMGB              int          `json:"ram_gb"`
	StorageGB          int          `json:"storage_gb"`
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
	DockerImage   string `json:"docker_image,omitempty"`
	InternalToken string `json:"internal_token"`
	CallbackURL      string       `json:"callback_url"`
	StartupScript    string       `json:"startup_script,omitempty"`
	BidPricePerGPU   float64      `json:"bid_price_per_gpu,omitempty"` // spot bid price (provider's raw price, pre-markup)
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
	Region                string  `json:"region"`
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

// CustomerInstance is the customer-facing representation of an instance.
// It structurally excludes all upstream provider details (defense by omission).
// If a field does not exist here, it can never leak via json.Marshal.
type CustomerInstance struct {
	ID           string    `json:"id"`
	Hostname     string    `json:"hostname"`
	SSHCommand   string    `json:"ssh_command"`
	Status       string    `json:"status"`
	GPUType      string    `json:"gpu_type"`
	GPUCount     int       `json:"gpu_count"`
	Tier         string    `json:"tier"`
	Region       string    `json:"region"`
	PricePerHour float64   `json:"price_per_hour"`
	CreatedAt    time.Time `json:"created_at"`
}

// Instance represents the full internal instance with upstream details.
// Only used by provisioning engine and DB layer -- never serialized to customers.
type Instance struct {
	InstanceID       string
	OrgID            string
	UserID           string
	UpstreamProvider string
	UpstreamID       string
	UpstreamIP       string
	Hostname         string
	WGPublicKey      string
	WGAddress        string
	GPUType          string
	GPUCount         int
	Tier             string
	Region           string
	PricePerHour     float64
	Status           string
	CreatedAt        time.Time
}

// ToCustomer converts an internal Instance to a customer-safe representation.
// Upstream provider, ID, and IP are structurally absent from the result.
func (i *Instance) ToCustomer() CustomerInstance {
	return CustomerInstance{
		ID:           i.InstanceID,
		Hostname:     i.Hostname,
		SSHCommand:   fmt.Sprintf("ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@%s", i.Hostname),
		Status:       i.Status,
		GPUType:      i.GPUType,
		GPUCount:     i.GPUCount,
		Tier:         i.Tier,
		Region:       i.Region,
		PricePerHour: i.PricePerHour,
		CreatedAt:    i.CreatedAt,
	}
}
