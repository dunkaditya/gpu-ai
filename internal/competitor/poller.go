package competitor

import (
	"context"
	"log/slog"
	"time"
)

// scrapeFunc is a named scraper function.
type scrapeFunc struct {
	name   string
	scrape func(ctx context.Context) (Prices, error)
}

// Poller periodically scrapes competitor pricing pages and caches results.
type Poller struct {
	cache    *Cache
	interval time.Duration
	logger   *slog.Logger
	scrapers []scrapeFunc
}

// NewPoller creates a competitor pricing poller.
func NewPoller(cache *Cache, interval time.Duration, logger *slog.Logger) *Poller {
	return &Poller{
		cache:    cache,
		interval: interval,
		logger:   logger,
		scrapers: []scrapeFunc{
			{name: "Lambda", scrape: ScrapeLambda},
			{name: "CoreWeave", scrape: ScrapeCoreWeave},
			{name: "AWS", scrape: ScrapeAWS},
		},
	}
}

// Start runs the poller loop. Polls immediately on startup, then on each tick.
// Blocks until ctx is cancelled.
func (p *Poller) Start(ctx context.Context) {
	p.pollAll(ctx)

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("competitor pricing poller stopped")
			return
		case <-ticker.C:
			p.pollAll(ctx)
		}
	}
}

func (p *Poller) pollAll(ctx context.Context) {
	for _, s := range p.scrapers {
		prices, err := s.scrape(ctx)
		if err != nil {
			p.logger.Warn("competitor scrape failed",
				slog.String("provider", s.name),
				slog.String("error", err.Error()),
			)
			continue
		}

		if err := p.cache.Set(ctx, s.name, prices); err != nil {
			p.logger.Error("failed to cache competitor prices",
				slog.String("provider", s.name),
				slog.String("error", err.Error()),
			)
			continue
		}

		p.logger.Debug("scraped competitor prices",
			slog.String("provider", s.name),
			slog.Int("gpu_models", len(prices)),
		)
	}
}
