package availability

import (
	"math"

	"github.com/gpuai/gpuctl/internal/provider"
)

// AvailableOffering is the customer-facing GPU offering representation.
// Uses defense-by-omission: no Provider field exists, so provider identity
// can never leak via json.Marshal.
type AvailableOffering struct {
	GPUModel       string  `json:"gpu_model"`
	VRAMGB         int     `json:"vram_gb"`
	CPUCores       int     `json:"cpu_cores"`
	RAMGB          int     `json:"ram_gb"`
	StorageGB      int     `json:"storage_gb"`
	PricePerHour   float64 `json:"price_per_hour"`
	Region         string  `json:"region"`
	Tier           string  `json:"tier"`
	AvailableCount int     `json:"available_count"`
	AvgUptimePct   float64 `json:"avg_uptime_pct"`
}

// DefaultMarkupPct is the default markup percentage applied to provider prices.
// GPU.ai retail price = provider price * (1 + markup/100).
const DefaultMarkupPct = 6.0

// defaultAvgUptime returns a static uptime percentage based on tier.
// On-demand has higher reliability than spot.
const (
	avgUptimeOnDemand = 99.5
	avgUptimeSpot     = 95.0
)

// defaultStorageGB is the default storage assumption when provider doesn't specify.
const defaultStorageGB = 40

// ToAvailableOffering converts a provider-internal GPUOffering to a customer-safe
// AvailableOffering with markup pricing applied.
// markupPct is the percentage markup (e.g., 15.0 for 15% markup).
func ToAvailableOffering(o provider.GPUOffering, markupPct float64) AvailableOffering {
	retailPrice := math.Round(o.PricePerHour*(1.0+markupPct/100.0)*100) / 100

	storageGB := o.StorageGB
	if storageGB == 0 {
		storageGB = defaultStorageGB
	}

	avgUptime := avgUptimeOnDemand
	if o.Tier == provider.TierSpot {
		avgUptime = avgUptimeSpot
	}

	return AvailableOffering{
		GPUModel:       string(o.GPUType),
		VRAMGB:         o.VRAMPerGPUGB,
		CPUCores:       o.CPUCores,
		RAMGB:          o.RAMGB,
		StorageGB:      storageGB,
		PricePerHour:   retailPrice,
		Region:         o.Region,
		Tier:           string(o.Tier),
		AvailableCount: o.AvailableCount,
		AvgUptimePct:   avgUptime,
	}
}
