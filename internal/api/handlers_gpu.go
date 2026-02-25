package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gpuai/gpuctl/internal/availability"
)

// GPUAvailabilityResponse is the JSON body returned by GET /api/v1/gpu/available.
type GPUAvailabilityResponse struct {
	Available []availability.AvailableOffering `json:"available"`
}

// handleListGPUAvailability handles GET /api/v1/gpu/available.
// Returns aggregated GPU offerings from the Redis cache with optional server-side filtering.
// Query params: gpu_model, region, tier, min_price, max_price, min_vram.
func (s *Server) handleListGPUAvailability(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse optional filter query parameters.
	gpuModel := r.URL.Query().Get("gpu_model")
	region := r.URL.Query().Get("region")
	tier := r.URL.Query().Get("tier")

	var minPrice, maxPrice *float64
	var minVRAM *int

	if v := r.URL.Query().Get("min_price"); v != "" {
		p, err := strconv.ParseFloat(v, 64)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "invalid-filter",
				"Invalid min_price: must be a valid number")
			return
		}
		minPrice = &p
	}

	if v := r.URL.Query().Get("max_price"); v != "" {
		p, err := strconv.ParseFloat(v, 64)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "invalid-filter",
				"Invalid max_price: must be a valid number")
			return
		}
		maxPrice = &p
	}

	if v := r.URL.Query().Get("min_vram"); v != "" {
		vram, err := strconv.Atoi(v)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "invalid-filter",
				"Invalid min_vram: must be a valid integer")
			return
		}
		minVRAM = &vram
	}

	// Read offerings from Redis cache.
	offerings, err := s.availCache.GetOfferings(ctx)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "cache-error",
			"Failed to retrieve GPU availability")
		return
	}

	// If cache returns nil (no data cached yet), return 200 with empty array.
	if offerings == nil {
		writeJSON(w, http.StatusOK, GPUAvailabilityResponse{
			Available: []availability.AvailableOffering{},
		})
		return
	}

	// Apply filters.
	filtered := filterOfferings(offerings, gpuModel, region, tier, minPrice, maxPrice, minVRAM)

	writeJSON(w, http.StatusOK, GPUAvailabilityResponse{
		Available: filtered,
	})
}

// filterOfferings returns a new slice containing only offerings that match all filters.
func filterOfferings(offerings []availability.AvailableOffering, gpuModel, region, tier string, minPrice, maxPrice *float64, minVRAM *int) []availability.AvailableOffering {
	result := make([]availability.AvailableOffering, 0, len(offerings))
	for _, o := range offerings {
		if matchesFilters(o, gpuModel, region, tier, minPrice, maxPrice, minVRAM) {
			result = append(result, o)
		}
	}
	return result
}

// matchesFilters checks whether a single offering passes all active filters.
func matchesFilters(o availability.AvailableOffering, gpuModel, region, tier string, minPrice, maxPrice *float64, minVRAM *int) bool {
	if gpuModel != "" && !strings.EqualFold(o.GPUModel, gpuModel) {
		return false
	}
	if region != "" && o.Region != region {
		return false
	}
	if tier != "" && o.Tier != tier {
		return false
	}
	if minPrice != nil && o.PricePerHour < *minPrice {
		return false
	}
	if maxPrice != nil && o.PricePerHour > *maxPrice {
		return false
	}
	if minVRAM != nil && o.VRAMGB < *minVRAM {
		return false
	}
	return true
}
