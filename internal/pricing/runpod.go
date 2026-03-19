package pricing

import (
	"context"
	"log/slog"

	"github.com/gpuai/gpuctl/internal/provider"
)

// FetchRunPodPrices calls ListAvailable on the RunPod adapter and extracts
// the lowest on-demand per-GPU-hour price for each GPU model.
// Returns a map keyed by canonical GPU model (e.g. "h100_sxm") → price.
// These are raw provider prices (pre-markup).
func FetchRunPodPrices(ctx context.Context, p provider.Provider) map[string]float64 {
	offerings, err := p.ListAvailable(ctx)
	if err != nil {
		slog.Warn("failed to fetch RunPod prices for comparison",
			slog.String("error", err.Error()),
		)
		return nil
	}

	prices := make(map[string]float64)
	for _, o := range offerings {
		model := string(o.GPUType)
		if o.Tier != provider.TierOnDemand {
			continue
		}
		if existing, ok := prices[model]; !ok || o.PricePerHour < existing {
			prices[model] = o.PricePerHour
		}
	}

	return prices
}
