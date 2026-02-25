package health

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/gpuai/gpuctl/internal/db"
	"github.com/gpuai/gpuctl/internal/provider"
)

// MonitorDeps holds the dependencies for the health monitor.
type MonitorDeps struct {
	DB       *db.Pool
	Registry *provider.Registry
	Logger   *slog.Logger
	Interval time.Duration // default 60s
	// OnEvent is called when a health event is detected (interruption, failure).
	// Used to push events to SSE broker. Nil means no notification.
	OnEvent func(event db.InstanceEvent)
}

// Monitor polls provider APIs to detect instance health issues.
type Monitor struct {
	db       *db.Pool
	registry *provider.Registry
	logger   *slog.Logger
	interval time.Duration
	onEvent  func(event db.InstanceEvent)
}

// NewMonitor creates a new health monitor.
func NewMonitor(deps MonitorDeps) *Monitor {
	interval := deps.Interval
	if interval == 0 {
		interval = 60 * time.Second
	}
	return &Monitor{
		db:       deps.DB,
		registry: deps.Registry,
		logger:   deps.Logger,
		interval: interval,
		onEvent:  deps.OnEvent,
	}
}

// SetOnEvent sets the callback invoked when a health event is detected.
// Allows post-construction wiring (same pattern as Engine.SetOnStatusChange).
func (m *Monitor) SetOnEvent(fn func(event db.InstanceEvent)) {
	m.onEvent = fn
}

// Start runs the health monitor loop. Checks immediately on startup,
// then on each tick. Blocks until ctx is cancelled.
func (m *Monitor) Start(ctx context.Context) {
	m.checkAll(ctx)

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			m.logger.Info("health monitor stopped")
			return
		case <-ticker.C:
			m.checkAll(ctx)
		}
	}
}

// maxConcurrentChecks limits parallel provider API calls.
const maxConcurrentChecks = 10

// checkAll queries all active instances and checks their health.
func (m *Monitor) checkAll(ctx context.Context) {
	instances, err := m.db.ListActiveInstances(ctx)
	if err != nil {
		m.logger.Error("health check: failed to list active instances",
			slog.String("error", err.Error()),
		)
		return
	}

	if len(instances) == 0 {
		return // Short-circuit: nothing to check.
	}

	// Bounded concurrency with semaphore channel.
	sem := make(chan struct{}, maxConcurrentChecks)
	var wg sync.WaitGroup

	for i := range instances {
		inst := instances[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}        // acquire
			defer func() { <-sem }() // release

			m.checkInstance(ctx, &inst)
		}()
	}
	wg.Wait()
}

// checkInstance checks a single instance's health via provider API.
func (m *Monitor) checkInstance(ctx context.Context, inst *db.Instance) {
	prov, ok := m.registry.Get(inst.UpstreamProvider)
	if !ok {
		m.logger.Warn("health check: provider not found",
			slog.String("instance_id", inst.InstanceID),
			slog.String("provider", inst.UpstreamProvider),
		)
		return
	}

	status, err := prov.GetStatus(ctx, inst.UpstreamID)
	if err != nil {
		m.logger.Warn("health check: provider status poll failed",
			slog.String("instance_id", inst.InstanceID),
			slog.String("provider", inst.UpstreamProvider),
			slog.String("error", err.Error()),
		)
		return
	}

	switch status.Status {
	case "running":
		// Healthy -- nothing to do.
		return

	case "terminated", "exited":
		// Check if this is a spot interruption or unexpected failure.
		if inst.Tier == "spot" {
			m.handleSpotInterruption(ctx, inst)
		} else {
			// Non-spot: retry before declaring failure.
			m.handleNonSpotFailure(ctx, inst, prov)
		}

	case "error":
		// Provider reports error state.
		if inst.Tier == "spot" {
			m.handleSpotInterruption(ctx, inst)
		} else {
			m.handleNonSpotFailure(ctx, inst, prov)
		}
	}
}

// handleSpotInterruption handles a detected spot instance interruption.
// Immediately stops billing and logs interruption event.
func (m *Monitor) handleSpotInterruption(ctx context.Context, inst *db.Instance) {
	m.logger.Warn("SPOT_INTERRUPTION: detected spot instance interruption",
		slog.String("instance_id", inst.InstanceID),
		slog.String("org_id", inst.OrgID),
	)

	// Use optimistic locking: only act if we successfully transition the state.
	// This prevents duplicate events if the monitor runs while another process
	// is also transitioning (e.g., user-initiated termination).
	updated, err := m.db.UpdateInstanceStatus(ctx, inst.InstanceID, inst.Status, "error")
	if err != nil {
		m.logger.Error("health check: failed to set error state for spot interruption",
			slog.String("instance_id", inst.InstanceID),
			slog.String("error", err.Error()),
		)
		return
	}
	if !updated {
		// Status already changed concurrently -- skip event logging.
		m.logger.Debug("health check: instance status already changed, skipping event",
			slog.String("instance_id", inst.InstanceID),
		)
		return
	}

	// Set error reason.
	_ = m.db.SetInstanceError(ctx, inst.InstanceID, "spot instance interrupted by provider")

	// Close billing session immediately.
	if err := m.db.CloseBillingSession(ctx, inst.InstanceID, time.Now().UTC()); err != nil {
		m.logger.Error("health check: failed to close billing session for spot interruption",
			slog.String("instance_id", inst.InstanceID),
			slog.String("error", err.Error()),
		)
	}

	// Log event to instance_events.
	metadata, _ := json.Marshal(map[string]string{
		"reason":   "spot_interruption",
		"gpu_type": inst.GPUType,
		"region":   inst.Region,
		"tier":     inst.Tier,
	})
	event := &db.InstanceEvent{
		InstanceID: inst.InstanceID,
		OrgID:      inst.OrgID,
		EventType:  "interrupted",
		Metadata:   metadata,
	}
	if err := m.db.CreateInstanceEvent(ctx, event); err != nil {
		m.logger.Error("health check: failed to log interruption event",
			slog.String("instance_id", inst.InstanceID),
			slog.String("error", err.Error()),
		)
	}

	// Notify SSE subscribers.
	if m.onEvent != nil {
		m.onEvent(*event)
	}
}

// nonSpotRetryCount is the number of retries before declaring a non-spot failure.
const nonSpotRetryCount = 3

// nonSpotRetryInterval is the delay between non-spot failure retry checks.
const nonSpotRetryInterval = 10 * time.Second

// handleNonSpotFailure retries health checks for non-spot instances before declaring failure.
// Avoids false positives from transient network blips.
func (m *Monitor) handleNonSpotFailure(ctx context.Context, inst *db.Instance, prov provider.Provider) {
	// Retry 3 times over ~30 seconds.
	for retry := 0; retry < nonSpotRetryCount; retry++ {
		time.Sleep(nonSpotRetryInterval)

		// Re-check: instance may have been terminated by user during retry window.
		current, err := m.db.GetInstance(ctx, inst.InstanceID)
		if err != nil {
			m.logger.Warn("health check: failed to re-read instance during retry",
				slog.String("instance_id", inst.InstanceID),
				slog.String("error", err.Error()),
			)
			return
		}
		if current.Status != inst.Status {
			// Status changed (user terminated, etc.) -- stop retrying.
			return
		}

		status, err := prov.GetStatus(ctx, inst.UpstreamID)
		if err != nil {
			m.logger.Warn("health check: retry poll failed",
				slog.String("instance_id", inst.InstanceID),
				slog.Int("retry", retry+1),
				slog.String("error", err.Error()),
			)
			continue
		}

		if status.Status == "running" {
			// Instance recovered -- false alarm.
			m.logger.Info("health check: instance recovered after transient failure",
				slog.String("instance_id", inst.InstanceID),
				slog.Int("retries", retry+1),
			)
			return
		}
	}

	// All retries exhausted -- declare failure.
	m.logger.Error("INSTANCE_FAILURE: instance failed after retries",
		slog.String("instance_id", inst.InstanceID),
		slog.String("org_id", inst.OrgID),
	)

	updated, err := m.db.UpdateInstanceStatus(ctx, inst.InstanceID, inst.Status, "error")
	if err != nil {
		m.logger.Error("health check: failed to set error state",
			slog.String("instance_id", inst.InstanceID),
			slog.String("error", err.Error()),
		)
		return
	}
	if !updated {
		return // Status already changed concurrently.
	}

	_ = m.db.SetInstanceError(ctx, inst.InstanceID, "instance failed: provider reports terminated/error")

	// Close billing session.
	if err := m.db.CloseBillingSession(ctx, inst.InstanceID, time.Now().UTC()); err != nil {
		m.logger.Error("health check: failed to close billing session",
			slog.String("instance_id", inst.InstanceID),
			slog.String("error", err.Error()),
		)
	}

	// Log event.
	metadata, _ := json.Marshal(map[string]string{
		"reason":   "instance_failure",
		"gpu_type": inst.GPUType,
		"region":   inst.Region,
		"tier":     inst.Tier,
	})
	event := &db.InstanceEvent{
		InstanceID: inst.InstanceID,
		OrgID:      inst.OrgID,
		EventType:  "failed",
		Metadata:   metadata,
	}
	if err := m.db.CreateInstanceEvent(ctx, event); err != nil {
		m.logger.Error("health check: failed to log failure event",
			slog.String("instance_id", inst.InstanceID),
			slog.String("error", err.Error()),
		)
	}

	if m.onEvent != nil {
		m.onEvent(*event)
	}
}
