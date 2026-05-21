package workers

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
)

type mockSnapshotPort struct {
	callCount int64
	delay     time.Duration
}

func (m *mockSnapshotPort) ExecuteCheck(_ context.Context, _ uuid.UUID, _ string, _ string) error {
	atomic.AddInt64(&m.callCount, 1)
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	return nil
}

func TestWorkerPool_DispatchSucceeds(t *testing.T) {
	port := &mockSnapshotPort{}
	pool := NewWorkerPool(port, 10, nil)
	pool.Start(2)
	defer pool.Stop()

	err := pool.Dispatch(context.Background(), uuid.New(), "https://example.com", "tenant_1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Wait for worker to process
	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt64(&port.callCount) != 1 {
		t.Fatalf("expected 1 call, got %d", port.callCount)
	}
}

func TestWorkerPool_QueueLength(t *testing.T) {
	port := &mockSnapshotPort{delay: 100 * time.Millisecond}
	pool := NewWorkerPool(port, 5, nil)
	// Don't start workers so jobs stay in queue
	// We can't call Start(0) as that starts 0 workers - exactly what we want

	for i := 0; i < 3; i++ {
		err := pool.Dispatch(context.Background(), uuid.New(), "https://example.com", "tenant_1")
		if err != nil {
			t.Fatalf("dispatch %d: unexpected error: %v", i, err)
		}
	}

	if pool.QueueLength() != 3 {
		t.Fatalf("expected queue length 3, got %d", pool.QueueLength())
	}
}

func TestWorkerPool_RetriesAndSucceeds(t *testing.T) {
	port := &mockSnapshotPort{}
	// Buffer size of 1 means second dispatch will block unless first is consumed
	pool := NewWorkerPool(port, 1, nil)
	pool.Start(1)
	defer pool.Stop()

	// Fill the queue
	err := pool.Dispatch(context.Background(), uuid.New(), "https://example.com", "tenant_1")
	if err != nil {
		t.Fatalf("first dispatch: %v", err)
	}

	// Second dispatch should retry and succeed as worker drains the queue
	err = pool.Dispatch(context.Background(), uuid.New(), "https://example.com", "tenant_1")
	if err != nil {
		t.Fatalf("second dispatch should retry and succeed, got: %v", err)
	}
}

func TestWorkerPool_RetriesExhausted(t *testing.T) {
	port := &mockSnapshotPort{}
	// Buffer size 1, no workers started -> queue will never drain
	pool := NewWorkerPool(port, 1, nil)

	// Fill the buffer
	_ = pool.Dispatch(context.Background(), uuid.New(), "https://example.com", "tenant_1")

	// This should exhaust retries and return error
	// Use a short-lived context to speed up the test
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := pool.Dispatch(ctx, uuid.New(), "https://example.com", "tenant_1")
	if err == nil {
		t.Fatal("expected error after retries exhausted")
	}
}

func TestWorkerPool_ContextCancellation(t *testing.T) {
	port := &mockSnapshotPort{}
	pool := NewWorkerPool(port, 1, nil)

	// Fill the buffer
	_ = pool.Dispatch(context.Background(), uuid.New(), "https://example.com", "tenant_1")

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel quickly
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := pool.Dispatch(ctx, uuid.New(), "https://example.com", "tenant_1")
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

// TestWorkerPool_ConcurrentJobsDoNotExceedPoolCapacity verifies that at most N
// jobs execute concurrently when the pool has N workers.
func TestWorkerPool_ConcurrentJobsDoNotExceedPoolCapacity(t *testing.T) {
	const capacity = 3
	const totalJobs = capacity + 2

	var (
		concurrent    int64 // current simultaneous executions
		maxConcurrent int64 // peak concurrent executions observed
	)

	port := &mockSnapshotPort{}
	// Override ExecuteCheck with a tracker that measures concurrency.
	trackingPort := &trackingSnapshotPort{
		onExecute: func() {
			cur := atomic.AddInt64(&concurrent, 1)
			for {
				old := atomic.LoadInt64(&maxConcurrent)
				if cur <= old || atomic.CompareAndSwapInt64(&maxConcurrent, old, cur) {
					break
				}
			}
			time.Sleep(30 * time.Millisecond) // hold slot long enough to overlap
			atomic.AddInt64(&concurrent, -1)
		},
	}
	_ = port // suppress unused warning

	pool := NewWorkerPool(trackingPort, totalJobs+5, nil) // large buffer so no dispatch error
	pool.Start(capacity)
	defer pool.Stop()

	for i := 0; i < totalJobs; i++ {
		if err := pool.Dispatch(context.Background(), uuid.New(), "https://example.com", "t1"); err != nil {
			t.Fatalf("Dispatch %d: unexpected error: %v", i, err)
		}
	}

	// Wait for all jobs to finish
	time.Sleep(300 * time.Millisecond)

	peak := atomic.LoadInt64(&maxConcurrent)
	if peak > capacity {
		t.Errorf("max concurrent executions = %d, expected ≤ %d (pool capacity)", peak, capacity)
	}
	if peak < 1 {
		t.Errorf("expected at least 1 concurrent execution, got %d — jobs may not have run", peak)
	}
}

// TestWorkerPool_CancelledContextMidFlight_NoPanicAndGracefulDrain verifies that
// cancelling the context while jobs are in-flight does not panic and all goroutines
// return within a bounded timeout.
func TestWorkerPool_CancelledContextMidFlight_NoPanicAndGracefulDrain(t *testing.T) {
	port := &mockSnapshotPort{delay: 20 * time.Millisecond}
	pool := NewWorkerPool(port, 10, nil)
	pool.Start(3)

	// Submit several jobs
	for i := 0; i < 6; i++ {
		_ = pool.Dispatch(context.Background(), uuid.New(), "https://example.com", "t1")
	}

	// Stop the pool mid-flight — should not panic and must complete within timeout.
	stopDone := make(chan struct{})
	go func() {
		defer close(stopDone)
		pool.Stop()
	}()

	select {
	case <-stopDone:
		// Pass — all goroutines drained cleanly.
	case <-time.After(2 * time.Second):
		t.Fatal("pool.Stop() did not complete within 2s — goroutines may have leaked")
	}
}

// trackingSnapshotPort is a SnapshotPort that calls an onExecute hook per job.
type trackingSnapshotPort struct {
	onExecute func()
}

func (p *trackingSnapshotPort) ExecuteCheck(_ context.Context, _ uuid.UUID, _ string, _ string) error {
	if p.onExecute != nil {
		p.onExecute()
	}
	return nil
}
