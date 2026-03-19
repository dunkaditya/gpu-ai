package api

import (
	"net/http"

	"github.com/gpuai/gpuctl/internal/pricing"
)

// handlePricingComparison handles GET /api/v1/pricing/comparison.
// Public endpoint (no auth) — returns GPU pricing comparison across competitors.
func (s *Server) handlePricingComparison(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get GPU.ai offerings from availability cache.
	offerings, err := s.availCache.GetOfferings(ctx)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "cache-error",
			"Failed to retrieve GPU availability")
		return
	}

	// Read scraped competitor prices from Redis cache.
	competitorPrices := make(map[string]map[string]float64)
	if s.competitorCache != nil {
		cached, err := s.competitorCache.GetAll(ctx)
		if err == nil {
			for provider, prices := range cached {
				m := make(map[string]float64, len(prices))
				for k, v := range prices {
					m[k] = v
				}
				competitorPrices[provider] = m
			}
		}
	}

	// Add live RunPod competitor prices if provider is registered.
	if s.providers != nil {
		if runpodProvider, ok := s.providers.Get("runpod"); ok {
			if prices := pricing.FetchRunPodPrices(ctx, runpodProvider); prices != nil {
				competitorPrices["RunPod"] = prices
			}
		}
	}

	resp := pricing.BuildComparison(offerings, competitorPrices)

	w.Header().Set("Cache-Control", "public, max-age=60")
	writeJSON(w, http.StatusOK, resp)
}
