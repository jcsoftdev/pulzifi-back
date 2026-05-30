package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

// --- fakes ---

// fakeStore is an in-memory Store used in relay tests.
type fakeStore struct {
	mu     sync.Mutex
	rows   map[uuid.UUID]*Row

	resetStaleCalled int
	failDispatch     bool // when true handlers return an error via panicHandler
}

func newFakeStore() *fakeStore {
	return &fakeStore{rows: make(map[uuid.UUID]*Row)}
}

func (f *fakeStore) Insert(_ context.Context, _ *sql.Tx, row Row) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	r := row
	f.rows[row.ID] = &r
	return nil
}

func (f *fakeStore) InsertStandalone(_ context.Context, row Row) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	r := row
	f.rows[row.ID] = &r
	return nil
}

func (f *fakeStore) ResetStale(_ context.Context, _ time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.resetStaleCalled++
	return nil
}

func (f *fakeStore) Claim(_ context.Context, batchSize int) ([]Row, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var pending []Row
	for _, r := range f.rows {
		if r.Status == "pending" {
			pending = append(pending, *r)
			if len(pending) >= batchSize {
				break
			}
		}
	}
	for _, r := range pending {
		f.rows[r.ID].Status = "publishing"
	}
	return pending, nil
}

func (f *fakeStore) MarkPublished(_ context.Context, id uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r, ok := f.rows[id]; ok {
		r.Status = "published"
		now := time.Now()
		r.PublishedAt = &now
	}
	return nil
}

func (f *fakeStore) MarkFailed(_ context.Context, id uuid.UUID, errMsg string, nextAttemptAt time.Time, dead bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r, ok := f.rows[id]; ok {
		r.Attempts++
		r.LastError = errMsg
		r.NextAttemptAt = nextAttemptAt
		if dead {
			r.Status = "dead"
		} else {
			r.Status = "pending"
		}
	}
	return nil
}

// fixedClock always returns t.
type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

// fixedRand always returns v (must be in [0,1)).
type fixedRand struct{ v float64 }

func (r fixedRand) Float64() float64 { return r.v }

// --- helpers ---

func pendingRow(topic string) Row {
	id := uuid.New()
	payload, _ := json.Marshal(map[string]string{"id": id.String(), "type": topic})
	return Row{
		ID:            id,
		Topic:         topic,
		Payload:       payload,
		Status:        "pending",
		NextAttemptAt: time.Now().Add(-1 * time.Second),
	}
}

func buildRelay(store *fakeStore, cfg Config) *Relay {
	r := NewRelay(store, cfg)
	return r
}

// --- tests ---

func TestRelay_ClaimedRowBecomesPublished(t *testing.T) {
	store := newFakeStore()
	row := pendingRow("test.topic")
	store.rows[row.ID] = &row

	cfg := DefaultConfig()
	cfg.BatchSize = 10
	relay := buildRelay(store, cfg)

	handlerCalled := false
	relay.Subscribe("test.topic", func(_ string, _ []byte) {
		handlerCalled = true
	})

	relay.tick(context.Background())

	if !handlerCalled {
		t.Error("handler was not called")
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	if store.rows[row.ID].Status != "published" {
		t.Errorf("status = %q, want published", store.rows[row.ID].Status)
	}
}

func TestRelay_NoHandlers_RowMarkedPublished(t *testing.T) {
	store := newFakeStore()
	row := pendingRow("orphan.topic")
	store.rows[row.ID] = &row

	cfg := DefaultConfig()
	relay := buildRelay(store, cfg)
	// No handler subscribed.

	relay.tick(context.Background())

	store.mu.Lock()
	defer store.mu.Unlock()
	if store.rows[row.ID].Status != "published" {
		t.Errorf("status = %q, want published (no handlers = skip)", store.rows[row.ID].Status)
	}
}

func TestRelay_StaleResetCalledEachTick(t *testing.T) {
	store := newFakeStore()
	cfg := DefaultConfig()
	relay := buildRelay(store, cfg)

	relay.tick(context.Background())
	relay.tick(context.Background())

	if store.resetStaleCalled != 2 {
		t.Errorf("ResetStale calls = %d, want 2", store.resetStaleCalled)
	}
}

func TestRelay_MaxAttemptsAtBoundary(t *testing.T) {
	// When attempts == maxAttempts-1 and the handler panics, the next MarkFailed
	// call should see dead=true (the relay computes nextAttempts = attempts+1 >= max).
	store := newFakeStore()
	row := pendingRow("panic.topic")
	row.Attempts = 7 // maxAttempts-1 = 7 when max=8
	store.rows[row.ID] = &row

	cfg := DefaultConfig()
	cfg.MaxAttempts = 8
	relay := buildRelay(store, cfg)

	// A panicking handler is a delivery failure: the relay recovers the panic,
	// surfaces it as dispatchErr, and routes to handleFailure. At attempts=7 with
	// max=8, the next failure crosses the threshold (attempts+1 >= max) → dead.
	panicFired := false
	relay.Subscribe("panic.topic", func(_ string, _ []byte) {
		panicFired = true
		panic("deliberate test panic")
	})

	relay.tick(context.Background())

	if !panicFired {
		t.Error("handler was not invoked")
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	if got := store.rows[row.ID].Status; got != "dead" {
		t.Errorf("status = %q, want dead (panic at attempts=7/max=8 is a poison failure)", got)
	}
	if store.rows[row.ID].Attempts != 8 {
		t.Errorf("attempts = %d, want 8", store.rows[row.ID].Attempts)
	}
}

func TestBackoff_GrowsExponentiallyWithJitterBounds(t *testing.T) {
	cfg := DefaultConfig()
	cfg.BackoffBase = 5 * time.Second
	cfg.BackoffCap = 1 * time.Hour

	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}

	tests := []struct {
		name     string
		attempts int
		wantMin  time.Duration
		wantMax  time.Duration
	}{
		{"0 attempts", 0, 4 * time.Second, 6 * time.Second},   // 5s ±20%
		{"1 attempt", 1, 8 * time.Second, 12 * time.Second},   // 10s ±20%
		{"2 attempts", 2, 16 * time.Second, 24 * time.Second}, // 20s ±20%
	}

	// Use mid-point rand so expected ≈ base * 2^n (jitter offset = 0 when v=0.5).
	rng := fixedRand{v: 0.5}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BackoffNext(cfg, tt.attempts, clock, rng)
			delay := got.Sub(now)
			if delay < tt.wantMin || delay > tt.wantMax {
				t.Errorf("attempts=%d: delay=%v, want [%v, %v]",
					tt.attempts, delay, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestBackoff_DoesNotExceedCap(t *testing.T) {
	cfg := DefaultConfig()
	cfg.BackoffBase = 5 * time.Second
	cfg.BackoffCap = 30 * time.Second

	now := time.Now()
	clock := fixedClock{t: now}
	rng := fixedRand{v: 0.5} // zero jitter

	got := BackoffNext(cfg, 20, clock, rng)
	delay := got.Sub(now)
	// With v=0.5 jitter offset = 0, delay should equal cap exactly.
	if delay > cfg.BackoffCap {
		t.Errorf("delay %v exceeds cap %v", delay, cfg.BackoffCap)
	}
}
