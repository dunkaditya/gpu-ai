package pricing

// CompetitorEntry represents a single competitor's price for a GPU model.
type CompetitorEntry struct {
	Name  string   `json:"name"`
	Price *float64 `json:"price"` // nil if provider doesn't offer this GPU
}

// GPUComparison holds pricing data for a single GPU model across GPU.ai and competitors.
type GPUComparison struct {
	GPUModel       string            `json:"gpu_model"`
	DisplayName    string            `json:"display_name"`
	VRAMGB         int               `json:"vram_gb"`
	GPUAIPrice     *float64          `json:"gpuai_price"`
	AvailableCount int               `json:"available_count"`
	Competitors    []CompetitorEntry `json:"competitors"`
	SavingsPct     *int              `json:"savings_pct"`
}

// ComparisonResponse is the JSON body returned by GET /api/v1/pricing/comparison.
type ComparisonResponse struct {
	FeaturedModels  []string        `json:"featured_models"`
	GPUs            []GPUComparison `json:"gpus"`
	CompetitorNames []string        `json:"competitor_names"`
	UpdatedAt       string          `json:"updated_at"`
}
