package outbox

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// Handler is the function signature for event consumers. It matches
// eventbus.EventHandler exactly so callers can pass values interchangeably
// without importing the parent package (which would create a cycle).
type Handler func(topic string, payload []byte)

// Clock is injectable so tests can control time without real sleeps.
type Clock interface {
	Now() time.Time
}

// RandSource is injectable so tests can assert jitter bounds deterministically.
type RandSource interface {
	// Float64 returns a value in [0.0, 1.0).
	Float64() float64
}

// realClock delegates to time.Now().
type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// realRand delegates to math/rand default source.
type realRand struct{}

func (realRand) Float64() float64 { return rand.Float64() } //nolint:gosec // jitter, not security

// Relay polls public.outbox_events and dispatches pending rows to the
// in-process handler registry.
type Relay struct {
	store    Store
	cfg      Config
	registry map[string][]Handler
	mu       sync.RWMutex
	clock    Clock
	rng      RandSource
	stop     chan struct{}
	done     chan struct{}
}

// NewRelay constructs a Relay using the given store and config.
// Call Start() to begin polling.
func NewRelay(store Store, cfg Config) *Relay {
	return &Relay{
		store:    store,
		cfg:      cfg,
		registry: make(map[string][]Handler),
		clock:    realClock{},
		rng:      realRand{},
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
}

// Subscribe registers a handler for a topic. Thread-safe.
func (r *Relay) Subscribe(topic string, h Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.registry[topic] = append(r.registry[topic], h)
}

// handlers returns a snapshot of registered handlers for topic.
func (r *Relay) handlers(topic string) []Handler {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]Handler(nil), r.registry[topic]...)
}

// Start begins the polling loop in a background goroutine.
func (r *Relay) Start(ctx context.Context) {
	go r.loop(ctx)
}

// Stop signals the relay to stop and waits until the loop exits.
func (r *Relay) Stop() {
	close(r.stop)
	<-r.done
}

func (r *Relay) loop(ctx context.Context) {
	defer close(r.done)

	ticker := time.NewTicker(r.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-r.stop:
			return
		case <-ticker.C:
			r.tick(ctx)
		}
	}
}

func (r *Relay) tick(ctx context.Context) {
	// 1. Reset stale rows that are stuck in 'publishing' (crash recovery).
	if err := r.store.ResetStale(ctx, r.cfg.StaleThreshold); err != nil {
		logger.Warn("outbox relay: reset stale failed", zap.Error(err))
	}

	// 2. Claim a batch of pending rows.
	rows, err := r.store.Claim(ctx, r.cfg.BatchSize)
	if err != nil {
		logger.Warn("outbox relay: claim failed", zap.Error(err))
		return
	}

	// 3. Dispatch each row synchronously.
	for _, row := range rows {
		r.dispatch(ctx, row)
	}
}

func (r *Relay) dispatch(ctx context.Context, row Row) {
	dispatchCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_ = dispatchCtx // handlers receive (topic, payload); context forwarding left for future

	handlers := r.handlers(row.Topic)
	var dispatchErr error

	if len(handlers) == 0 {
		// No handlers registered — mark as published to avoid blocking the queue.
		logger.Warn("outbox relay: no handlers for topic",
			zap.String("topic", row.Topic),
			zap.String("event_id", row.ID.String()),
		)
		_ = r.store.MarkPublished(ctx, row.ID)
		return
	}

	for _, h := range handlers {
		func() {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("outbox relay: handler panic",
						zap.Any("recover", rec),
						zap.String("event_id", row.ID.String()),
					)
					// A panicking handler is a delivery failure: surface it so the
					// row is retried / sent to the DLQ instead of being silently
					// marked published. EventHandler returns nothing, so a panic is
					// the only failure signal a handler can emit.
					dispatchErr = fmt.Errorf("handler panic: %v", rec)
				}
			}()
			h(row.Topic, row.Payload)
		}()
	}

	if dispatchErr != nil {
		r.handleFailure(ctx, row, dispatchErr.Error())
		return
	}
	if err := r.store.MarkPublished(ctx, row.ID); err != nil {
		logger.Error("outbox relay: mark published failed",
			zap.Error(err),
			zap.String("event_id", row.ID.String()),
		)
	}
}

func (r *Relay) handleFailure(ctx context.Context, row Row, errMsg string) {
	nextAttempts := row.Attempts + 1
	dead := nextAttempts >= r.cfg.MaxAttempts
	next := r.backoff(row.Attempts)

	if err := r.store.MarkFailed(ctx, row.ID, errMsg, next, dead); err != nil {
		logger.Error("outbox relay: mark failed error",
			zap.Error(err),
			zap.String("event_id", row.ID.String()),
		)
	}
}

// backoff computes: min(base * 2^attempts, cap) ± 20% jitter.
func (r *Relay) backoff(attempts int) time.Time {
	base := float64(r.cfg.BackoffBase)
	cap := float64(r.cfg.BackoffCap)

	delay := base * math.Pow(2, float64(attempts))
	if delay > cap {
		delay = cap
	}

	// Jitter: ±20% of the computed delay.
	jitter := delay * 0.2 * (2*r.rng.Float64() - 1)
	delay += jitter
	if delay < 0 {
		delay = 0
	}

	return r.clock.Now().Add(time.Duration(delay))
}

// BackoffNext is exported for relay state-transition tests. Returns when the
// next retry should be attempted given the current attempt count.
func BackoffNext(cfg Config, attempts int, clock Clock, rng RandSource) time.Time {
	rel := &Relay{cfg: cfg, clock: clock, rng: rng}
	return rel.backoff(attempts)
}

// DispatchResult captures what happened to a single relay row during a test tick.
// Exported for white-box relay tests.
type DispatchResult struct {
	ID      uuid.UUID
	Success bool
	Dead    bool
	ErrMsg  string
}
