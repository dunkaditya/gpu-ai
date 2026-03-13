package runpod

import (
	"strings"

	"github.com/gpuai/gpuctl/internal/provider"
)

// gpuNameMap maps RunPod GPU IDs to GPU.ai canonical GPU types.
// Keys are the RunPod "id" field (e.g. "NVIDIA H100 80GB HBM3"), NOT the
// shorter displayName (e.g. "H100 SXM"). The adapter passes gpu.ID to
// NormalizeGPUName. The reverse map is used for gpuTypeId in provision requests.
var gpuNameMap = map[string]provider.GPUType{
	// Blackwell
	"NVIDIA B200":            provider.GPUTypeB200,
	"NVIDIA B300 SXM6 AC":   provider.GPUTypeB300,

	// Hopper
	"NVIDIA H200":           provider.GPUTypeH200SXM,
	"NVIDIA H200 NVL":       provider.GPUTypeH200NVL,
	"NVIDIA H100 80GB HBM3": provider.GPUTypeH100SXM,
	"NVIDIA H100 NVL":       provider.GPUTypeH100NVL,
	"NVIDIA H100 PCIe":      provider.GPUTypeH100PCIE,

	// Ampere data center
	"NVIDIA A100 80GB PCIe": provider.GPUTypeA10080GB,
	"NVIDIA A100-SXM4-80GB": provider.GPUTypeA10080GB,
	"NVIDIA A40":            provider.GPUTypeA40,
	"NVIDIA A30":            provider.GPUTypeA30,

	// Ada Lovelace / Ampere professional
	"NVIDIA L40S":                                            provider.GPUTypeL40S,
	"NVIDIA L40":                                             provider.GPUTypeL40,
	"NVIDIA L4":                                              provider.GPUTypeL4,
	"NVIDIA RTX 6000 Ada Generation":                         provider.GPUTypeRTX6000Ada,
	"NVIDIA RTX 5000 Ada Generation":                         provider.GPUTypeRTX5000Ada,
	"NVIDIA RTX 4000 Ada Generation":                         provider.GPUTypeRTX4000Ada,
	"NVIDIA RTX 4000 SFF Ada Generation":                     provider.GPUTypeRTX4000Ada,
	"NVIDIA RTX 2000 Ada Generation":                         provider.GPUTypeRTX2000Ada,
	"NVIDIA RTX A6000":                                       provider.GPUTypeRTXA6000,
	"NVIDIA RTX A5000":                                       provider.GPUTypeRTXA5000,
	"NVIDIA RTX A4500":                                       provider.GPUTypeRTXA4500,
	"NVIDIA RTX A4000":                                       provider.GPUTypeRTXA4000,
	"NVIDIA RTX PRO 6000 Blackwell Server Edition":           provider.GPUTypeRTXPro6000,
	"NVIDIA RTX PRO 6000 Blackwell Workstation Edition":      provider.GPUTypeRTXPro6000,
	"NVIDIA RTX PRO 6000 Blackwell Max-Q Workstation Edition": provider.GPUTypeRTXPro6000,
	"NVIDIA RTX PRO 4500 Blackwell":                          provider.GPUTypeRTXPro4500,

	// Consumer
	"NVIDIA GeForce RTX 5090":    provider.GPUTypeRTX5090,
	"NVIDIA GeForce RTX 5080":    provider.GPUTypeRTX5080,
	"NVIDIA GeForce RTX 4090":    provider.GPUTypeRTX4090,
	"NVIDIA GeForce RTX 4080 SUPER": provider.GPUTypeRTX4080,
	"NVIDIA GeForce RTX 4080":    provider.GPUTypeRTX4080,
	"NVIDIA GeForce RTX 4070 Ti": provider.GPUTypeRTX4080, // 4070 Ti ≈ 4080 tier
	"NVIDIA GeForce RTX 3090":    provider.GPUTypeRTX3090,
	"NVIDIA GeForce RTX 3090 Ti": provider.GPUTypeRTX3090,
	"NVIDIA GeForce RTX 3080":    provider.GPUTypeRTX3080,
	"NVIDIA GeForce RTX 3080 Ti": provider.GPUTypeRTX3080,
	"NVIDIA GeForce RTX 3070":    provider.GPUTypeRTX3080, // 3070 ≈ 3080 tier

	// Legacy
	"Tesla V100-PCIE-16GB":   provider.GPUTypeV100,
	"Tesla V100-SXM2-16GB":   provider.GPUTypeV100,
	"Tesla V100-SXM2-32GB":   provider.GPUTypeV100,
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
// RunPod uses formats like "US-TX-3", "EU-RO-1", "CA-MTL-1". We match
// longest prefix first (e.g., "EU-RO" before "EU") in the NormalizeRegion function.
var regionMap = map[string]string{
	// US locations.
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
	// EU locations (RunPod format: EU-{country}-{number}).
	"EU-RO": "eu-east",
	"EU-CZ": "eu-east",
	"EU-BG": "eu-east",
	"EU-SE": "eu-north",
	"EU-NO": "eu-north",
	"EU-IS": "eu-north",
	"EU":    "eu-west", // Fallback for unmapped EU locations.
	// Canada.
	"CA": "ca-central",
	// Standalone country codes (if RunPod uses them directly).
	"SE": "eu-north",
	"NO": "eu-north",
	"RO": "eu-east",
	"CZ": "eu-east",
	"BG": "eu-east",
	"IS": "eu-north",
	// Additional EU.
	"EU-NL": "eu-west",
	"EU-GB": "eu-west",
	"EU-DE": "eu-west",
	"EU-FR": "eu-west",
	"NL":    "eu-west",
	"GB":    "eu-west",
	"DE":    "eu-west",
	"FR":    "eu-west",
	"FI":    "eu-north",
	// APAC.
	"AU": "ap-southeast",
	"JP": "ap-northeast",
	"TH": "ap-southeast",
	"TW": "ap-northeast",
	"MY": "ap-southeast",
	"SG": "ap-southeast",
	"IN": "ap-south",
	"KR": "ap-northeast",
	"NZ": "ap-southeast",
	// Middle East.
	"AE": "me-central",
	"SA": "me-central",
	"QA": "me-central",
	// South America.
	"BR": "sa-east",
	// US fallback.
	"US": "us-east",
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
