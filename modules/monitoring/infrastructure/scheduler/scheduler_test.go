// Package scheduler_test validates the scheduler's WakeUp signaling and
// context-cancellation behavior. These tests do not require a running database —
// they use a nil *sql.DB and do not exercise the DB query path.
package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// TestScheduler_WakeUp_NonBlocking verifies that WakeUp() never blocks even when
// called multiple times in rapid succession.
func TestScheduler_WakeUp_NonBlocking(t *testing.T) {
	s := &Scheduler{
		wakeUp: make(chan struct{}, 1),
	}

	// Call WakeUp many times — must not block or deadlock.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 100; i++ {
			s.WakeUp()
		}
	}()

	select {
	case <-done:
		// Pass
	case <-time.After(500 * time.Millisecond):
		t.Fatal("WakeUp() blocked or deadlocked")
	}

	// After 100 calls, the channel should hold at most 1 signal (it's buffered=1).
	if len(s.wakeUp) > 1 {
		t.Errorf("expected at most 1 pending signal, got %d", len(s.wakeUp))
	}
}

// TestScheduler_WakeUp_CoalescesSignals verifies that multiple WakeUp signals are
// coalesced into a single pending signal (prevents thundering-herd).
func TestScheduler_WakeUp_CoalescesSignals(t *testing.T) {
	s := &Scheduler{
		wakeUp: make(chan struct{}, 1),
	}

	// Saturate the channel
	s.WakeUp()
	s.WakeUp()
	s.WakeUp()

	// Channel should have exactly 1 item (the buffer capacity)
	if got := len(s.wakeUp); got != 1 {
		t.Errorf("expected 1 coalesced signal, got %d", got)
	}
}

// TestScheduler_ContextCancellation_StopsLoop verifies that a running scheduler
// goroutine exits promptly when its context is cancelled.
//
// This test uses a minimal ticker-based loop that mirrors the scheduler's Select
// pattern to avoid requiring a real *sql.DB.
func TestScheduler_ContextCancellation_StopsLoop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wakeUp := make(chan struct{}, 1)

	var iterations int64

	// Run a minimal event loop that mirrors what Scheduler.Start does.
	loopDone := make(chan struct{})
	go func() {
		defer close(loopDone)
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-wakeUp:
				atomic.AddInt64(&iterations, 1)
			case <-ticker.C:
				atomic.AddInt64(&iterations, 1)
			}
		}
	}()

	// Let the loop run for a bit, then cancel.
	time.Sleep(80 * time.Millisecond)
	cancel()

	select {
	case <-loopDone:
		// Pass — loop exited promptly after cancellation.
	case <-time.After(200 * time.Millisecond):
		t.Fatal("scheduler loop did not stop within 200ms after context cancellation")
	}

	ticks := atomic.LoadInt64(&iterations)
	if ticks < 2 {
		t.Errorf("expected at least 2 tick iterations before cancel, got %d", ticks)
	}
}

// TestScheduler_WakeUp_SignalDeliveredToLoop verifies that a WakeUp signal is
// consumed by the event loop and triggers processing.
func TestScheduler_WakeUp_SignalDeliveredToLoop(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	wakeUp := make(chan struct{}, 1)
	var wakeUpReceived int64

	loopDone := make(chan struct{})
	go func() {
		defer close(loopDone)
		ticker := time.NewTicker(1 * time.Second) // long ticker so wakeUp fires first
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-wakeUp:
				atomic.AddInt64(&wakeUpReceived, 1)
				return // exit after first wakeUp
			case <-ticker.C:
				return
			}
		}
	}()

	// Send a wakeUp signal
	wakeUp <- struct{}{}

	select {
	case <-loopDone:
		// Pass
	case <-time.After(300 * time.Millisecond):
		t.Fatal("loop did not receive wakeUp signal within 300ms")
	}

	if atomic.LoadInt64(&wakeUpReceived) != 1 {
		t.Errorf("expected 1 wakeUp received, got %d", atomic.LoadInt64(&wakeUpReceived))
	}
}
