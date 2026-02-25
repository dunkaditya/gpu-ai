package availability

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/gpuai/gpuctl/internal/provider"
)

// Poller queries all registered providers for GPU availability on a fixed interval
// and writes aggregated results to the Redis cache.
type Poller struct {
	registry  *provider.Registry
	cache     *Cache
	interval  time.Duration
	markupPct float64
	logger    *slog.Logger
}

// NewPoller creates a new availability poller.
// interval is the polling frequency (typically 30s).
// markupPct is the pricing markup percentage applied to provider prices.
func NewPoller(registry *provider.Registry, cache *Cache, interval time.Duration, markupPct float64, logger *slog.Logger) *Poller {
	return &Poller{
		registry:  registry,
		cache:     cache,
		interval:  interval,
		markupPct: markupPct,
		logger:    logger,
	}
}

// Start runs the poller loop. It polls immediately on startup, then on each tick.
// Blocks until ctx is cancelled. Intended to be run as a goroutine.
func (p *Poller) Start(ctx context.Context) {
	// Poll immediately on startup -- don't wait 30s for first data.
	p.poll(ctx)

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			p.logger.Info("availability poller stopped")
			return
		case <-ticker.C:
			p.poll(ctx)
		}
	}
}

// poll queries all providers concurrently and writes aggregated offerings to cache.
// Per-provider errors are logged and isolated -- one failing provider does not block others.
func (p *Poller) poll(ctx context.Context) {
	providers := p.registry.All()
	if len(providers) == 0 {
		p.logger.Debug("no providers registered, skipping poll")
		return
	}

	var mu sync.Mutex
	var allOfferings []AvailableOffering
	var wg sync.WaitGroup

	for _, prov := range providers {
		wg.Add(1)
		go func(prov provider.Provider) {
			defer wg.Done()

			offerings, err := prov.ListAvailable(ctx)
			if err != nil {
				p.logger.Error("availability poll failed",
					slog.String("provider", prov.Name()),
					slog.String("error", err.Error()),
				)
				return
			}

			// Convert to customer-facing offerings with markup pricing.
			converted := make([]AvailableOffering, 0, len(offerings))
			for _, o := range offerings {
				converted = append(converted, ToAvailableOffering(o, p.markupPct))
			}

			mu.Lock()
			allOfferings = append(allOfferings, converted...)
			mu.Unlock()

			p.logger.Debug("polled provider",
				slog.String("provider", prov.Name()),
				slog.Int("offerings", len(offerings)),
			)
		}(prov)
	}
	wg.Wait()

	// Write to cache even if empty (clears stale data from removed providers).
	if err := p.cache.SetOfferings(ctx, allOfferings); err != nil {
		p.logger.Error("failed to cache offerings", slog.String("error", err.Error()))
		return
	}

	p.logger.Debug("availability poll complete",
		slog.Int("total_offerings", len(allOfferings)),
	)
}
