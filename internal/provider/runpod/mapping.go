package runpod

import (
	"strings"

	"github.com/gpuai/gpuctl/internal/provider"
)

// gpuNameMap maps RunPod GPU display names to GPU.ai canonical GPU types.
// GPU types are pulled dynamically from RunPod's API; this table normalizes
// the display names into our canonical set. Unmapped types are logged and skipped.
var gpuNameMap = map[string]provider.GPUType{
	"NVIDIA GeForce RTX 4090": provider.GPUTypeRTX4090,
	"NVIDIA RTX A6000":        provider.GPUTypeRTXA6000,
	"NVIDIA A40":              provider.GPUTypeA40,
	"NVIDIA L40S":             provider.GPUTypeL40S,
	"NVIDIA L4":               provider.GPUTypeL4,
	"NVIDIA A100 80GB PCIe":   provider.GPUTypeA10080GB,
	"NVIDIA A100-SXM4-80GB":   provider.GPUTypeA10080GB,
	"NVIDIA H100 80GB HBM3":   provider.GPUTypeH100SXM,
	"NVIDIA H100 PCIe":        provider.GPUTypeH100PCIE,
	"NVIDIA H200":             provider.GPUTypeH200SXM,
}

// reverseGPUNameMap maps GPU.ai canonical types back to RunPod GPU IDs.
// For types with multiple RunPod names (e.g., A100 80GB has PCIe and SXM4),
// the first one encountered during map initialization is used.
var reverseGPUNameMap map[provider.GPUType]string

func init() {
	reverseGPUNameMap = make(map[provider.GPUType]string, len(gpuNameMap))
	for rpName, gpuType := range gpuNameMap {
		// Only set if not already present (first match wins).
		if _, exists := reverseGPUNameMap[gpuType]; !exists {
			reverseGPUNameMap[gpuType] = rpName
		}
	}
}

// NormalizeGPUName looks up a RunPod display name in the mapping table.
// Returns the canonical GPU type and true if found, or zero value and false if unmapped.
func NormalizeGPUName(runpodName string) (provider.GPUType, bool) {
	gpuType, ok := gpuNameMap[runpodName]
	return gpuType, ok
}

// RunPodGPUName returns the RunPod GPU ID for a canonical GPU type.
// Returns the name and true if found, or empty string and false if unmapped.
func RunPodGPUName(gpuType provider.GPUType) (string, bool) {
	name, ok := reverseGPUNameMap[gpuType]
	return name, ok
}

// regionMap maps RunPod datacenter location prefixes to GPU.ai region codes.
var regionMap = map[string]string{
	"US-CA": "us-west",
	"US-OR": "us-west",
	"US-WA": "us-west",
	"US-TX": "us-central",
	"US-GA": "us-east",
	"US-NJ": "us-east",
	"US-VA": "us-east",
	"US-NY": "us-east",
	"US-IL": "us-central",
	"US-KS": "us-central",
	"EU":    "eu-west",
	"CA":    "ca-central",
	"SE":    "eu-north",
	"NO":    "eu-north",
	"RO":    "eu-east",
	"CZ":    "eu-east",
	"BG":    "eu-east",
	"IS":    "eu-north",
}

// NormalizeRegion maps a RunPod location string to a GPU.ai region code
// and a human-readable datacenter location.
// RunPod locations are formatted like "US-TX-3", "EU-RO-1", etc.
// The raw location string is used as the datacenter location when no nicer
// mapping is available.
func NormalizeRegion(runpodLocation string) (regionCode string, datacenterLocation string) {
	datacenterLocation = runpodLocation
	if runpodLocation == "" {
		return "unknown", "Unknown"
	}

	// Try exact match on full location first (e.g., "US-TX-3").
	if region, ok := regionMap[runpodLocation]; ok {
		return region, datacenterLocation
	}

	// Try prefix matching: "US-TX-3" -> try "US-TX", then "US".
	parts := strings.Split(runpodLocation, "-")
	for i := len(parts) - 1; i >= 1; i-- {
		prefix := strings.Join(parts[:i], "-")
		if region, ok := regionMap[prefix]; ok {
			return region, datacenterLocation
		}
	}

	// Single-part fallback (e.g., "EU" or "CA").
	if region, ok := regionMap[parts[0]]; ok {
		return region, datacenterLocation
	}

	return "unknown", datacenterLocation
}
