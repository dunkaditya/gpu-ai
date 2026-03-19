package pricing

import (
	"math"
	"time"

	"github.com/gpuai/gpuctl/internal/availability"
)

// gpuMeta holds display metadata for a GPU model.
type gpuMeta struct {
	DisplayName string
	VRAMGB      int
}

// FeaturedGPUs defines which GPU models appear in the pricing comparison, in display order.
var FeaturedGPUs = []string{
	"h200_sxm",
	"h100_sxm",
	"b200",
	"a100_80gb",
}

// DisplayedCompetitors controls which competitors appear in the response.
var DisplayedCompetitors = []string{"Lambda", "CoreWeave", "AWS"}

// gpuMetadata maps canonical GPU model names to display metadata.
var gpuMetadata = map[string]gpuMeta{
	"h200_sxm":  {DisplayName: "H200 SXM", VRAMGB: 141},
	"h100_sxm":  {DisplayName: "H100 SXM", VRAMGB: 80},
	"b200":      {DisplayName: "B200", VRAMGB: 192},
	"a100_80gb": {DisplayName: "A100", VRAMGB: 80},
	"l40s":      {DisplayName: "L40S", VRAMGB: 48},
	"rtx_4090":  {DisplayName: "RTX 4090", VRAMGB: 24},
}

// fallbackPrices are used when the live competitor scrapers haven't populated
// Redis yet (cold boot). Overridden by scraped data from internal/competitor/.
var fallbackPrices = map[string]map[string]float64{
	"Lambda": {
		"h200_sxm":  4.99,
		"h100_sxm":  3.29,
		"a100_80gb": 2.06,
		"l40s":      1.10,
	},
	"CoreWeave": {
		"h200_sxm":  6.31, // $50.44/8 GPUs
		"h100_sxm":  6.16, // $49.24/8 GPUs
		"a100_80gb": 2.21,
		"l40s":      1.24,
	},
	"AWS": {
		"h200_sxm":  4.97, // p5en.48xlarge $39.80/8 GPUs
		"h100_sxm":  3.93, // p5.48xlarge $31.46/8 GPUs
		"a100_80gb": 5.12, // p4de.24xlarge $40.97/8 GPUs
		"l40s":      1.52,
	},
}

// BuildComparison assembles pricing comparison data from GPU.ai availability offerings
// and live competitor prices (e.g., from RunPod).
// liveCompetitorPrices is keyed by provider name → gpu_model → price.
// liveCompetitorPrices is keyed by provider name → gpu_model → price.
func BuildComparison(offerings []availability.AvailableOffering, liveCompetitorPrices map[string]map[string]float64) ComparisonResponse {
	// Group GPU.ai offerings by model: track lowest on-demand price and total availability.
	type gpuaiData struct {
		lowestPrice    float64
		hasPrice       bool
		availableCount int
		vramGB         int
	}
	gpuaiByModel := make(map[string]*gpuaiData)

	for _, o := range offerings {
		d, ok := gpuaiByModel[o.GPUModel]
		if !ok {
			d = &gpuaiData{}
			gpuaiByModel[o.GPUModel] = d
		}
		d.availableCount += o.AvailableCount
		if o.VRAMGB > d.vramGB {
			d.vramGB = o.VRAMGB
		}
		// Use on-demand prices for comparison (more meaningful than spot).
		if o.Tier == "on_demand" && (!d.hasPrice || o.PricePerHour < d.lowestPrice) {
			d.lowestPrice = o.PricePerHour
			d.hasPrice = true
		}
	}

	// Merge all competitor sources (static + live).
	allCompetitorPrices := make(map[string]map[string]float64)
	for provider, models := range fallbackPrices {
		allCompetitorPrices[provider] = make(map[string]float64)
		for model, price := range models {
			allCompetitorPrices[provider][model] = price
		}
	}
	for provider, models := range liveCompetitorPrices {
		if _, ok := allCompetitorPrices[provider]; !ok {
			allCompetitorPrices[provider] = make(map[string]float64)
		}
		for model, price := range models {
			allCompetitorPrices[provider][model] = price
		}
	}

	// Build displayed competitors set for filtering.
	displayedSet := make(map[string]bool)
	for _, name := range DisplayedCompetitors {
		displayedSet[name] = true
	}

	// Build comparison for each featured GPU.
	gpus := make([]GPUComparison, 0, len(FeaturedGPUs))

	for _, model := range FeaturedGPUs {
		meta, ok := gpuMetadata[model]
		if !ok {
			continue
		}

		comp := GPUComparison{
			GPUModel:    model,
			DisplayName: meta.DisplayName,
			VRAMGB:      meta.VRAMGB,
		}

		// GPU.ai price and availability from cache.
		if d, ok := gpuaiByModel[model]; ok {
			comp.AvailableCount = d.availableCount
			if d.hasPrice {
				price := d.lowestPrice
				comp.GPUAIPrice = &price
			}
			if d.vramGB > 0 {
				comp.VRAMGB = d.vramGB
			}
		}

		// Competitor prices (only displayed competitors).
		competitors := make([]CompetitorEntry, 0, len(DisplayedCompetitors))
		for _, compName := range DisplayedCompetitors {
			entry := CompetitorEntry{Name: compName}
			if models, ok := allCompetitorPrices[compName]; ok {
				if price, ok := models[model]; ok {
					p := price
					entry.Price = &p
				}
			}
			competitors = append(competitors, entry)
		}
		comp.Competitors = competitors

		// Compute savings percentage vs average competitor price.
		if comp.GPUAIPrice != nil {
			var compTotal float64
			var compCount int
			for _, c := range competitors {
				if c.Price != nil {
					compTotal += *c.Price
					compCount++
				}
			}
			if compCount > 0 {
				avgComp := compTotal / float64(compCount)
				if avgComp > 0 {
					savings := int(math.Round((1 - *comp.GPUAIPrice/avgComp) * 100))
					comp.SavingsPct = &savings
				}
			}
		}

		gpus = append(gpus, comp)
	}

	return ComparisonResponse{
		FeaturedModels:  FeaturedGPUs,
		GPUs:            gpus,
		CompetitorNames: DisplayedCompetitors,
		UpdatedAt:       time.Now().UTC().Format(time.RFC3339),
	}
}
