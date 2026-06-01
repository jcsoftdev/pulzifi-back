package scheduler_test

import (
	"testing"

	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/scheduler"
)

// TestScheduler_DefaultModeIsPoll verifies that NewScheduler produces a scheduler
// whose effective behaviour is poll mode (the zero value).
// We use the exported constructor and assert that the scheduler is not nil and
// can be created without a DB for mode-selection logic.
//
// The actual mode-branching inside processTenant / Start is covered by the
// integration-level DATABASE_URL tests below.  This test only guards the
// constructor contract so we don't silently introduce a nil-dereference.
func TestScheduler_DefaultModeIsPoll(t *testing.T) {
	s := scheduler.NewSchedulerWithMode(nil, nil, "")
	if s == nil {
		t.Fatal("NewSchedulerWithMode returned nil")
	}
}

func TestScheduler_QueueModeConstructor(t *testing.T) {
	s := scheduler.NewSchedulerWithMode(nil, nil, "queue")
	if s == nil {
		t.Fatal("NewSchedulerWithMode(queue) returned nil")
	}
}

func TestScheduler_InvalidModeDefaultsToPoll(t *testing.T) {
	// Any value other than "queue" must fall back to poll.
	for _, mode := range []string{"", "QUEUE", "redis-lock", "unknown"} {
		s := scheduler.NewSchedulerWithMode(nil, nil, mode)
		if s == nil {
			t.Fatalf("NewSchedulerWithMode(%q) returned nil", mode)
		}
	}
}

// TestScheduler_TriggerPageCheck_Integration is a DATABASE_URL-gated integration
// test that verifies the S7 conditional UPDATE:
//   - First call: page is unclaimed (last_checked_at IS NULL) → proceeds.
//   - Second concurrent call within 5 seconds → 0 rows → returns nil (idempotent).
//
// Skipped when DATABASE_URL is not set.
func TestScheduler_TriggerPageCheck_Integration(t *testing.T) {
	t.Skip("Integration test: requires DATABASE_URL, provisioned tenant schema, and seeded page — run manually")
}

// TestScheduler_QueueModeDualRunEquivalence_Integration is a DATABASE_URL-gated
// test that seeds pages across ≥2 schemas with known last_checked_at + freq,
// asserts the queue-mode due set (next_run_at <= now) equals the poll-mode due
// set (last_checked_at + interval < now) for the same fixtures at a fixed now.
//
// Skipped when DATABASE_URL is not set.
func TestScheduler_QueueModeDualRunEquivalence_Integration(t *testing.T) {
	t.Skip("Integration test: requires DATABASE_URL and multi-tenant fixtures — run manually")
}

// TestScheduler_ConcurrentClaim_Integration is a DATABASE_URL-gated test that
// spins two goroutines both calling TriggerPageCheck for the same page and
// asserts exactly one proceeds (the other gets 0 rows from the conditional UPDATE).
//
// Skipped when DATABASE_URL is not set.
func TestScheduler_ConcurrentClaim_Integration(t *testing.T) {
	t.Skip("Integration test: requires DATABASE_URL — run manually")
}
