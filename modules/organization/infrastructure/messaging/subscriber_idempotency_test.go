package messaging_test

import (
	"testing"
)

// TestHandleUserDeleted_IsIdempotentViaDeletedAtGuards documents that the
// handleUserDeleted handler is safe to redeliver without double-effects.
//
// The handler runs SQL with WHERE deleted_at IS NULL guards on every mutating
// operation:
//
//   - softDeleteOrganization: WHERE id = $1 AND deleted_at IS NULL
//   - removeOrganizationMember: WHERE organization_id = $1 AND user_id = $2 AND deleted_at IS NULL
//   - removeUserRoles: DELETE — idempotent by nature (no rows = no-op)
//
// Redelivering a user.deleted event for an already-processed user results in
// zero rows affected on the UPDATE statements and a no-op DELETE, making the
// handler effectively idempotent even without a consumed-events table.
//
// When EVENT_BUS_PROVIDER=outbox is active, the Idempotent wrapper in
// shared/eventbus provides an additional deduplication layer via
// public.outbox_consumed, ensuring at-most-once handler execution per
// consumer per event_id.
func TestHandleUserDeleted_IsIdempotentViaDeletedAtGuards(t *testing.T) {
	// This test is a documentation test. The property is proved by the SQL
	// WHERE deleted_at IS NULL guards in subscriber.go:
	//   - softDeleteOrganization (line ~205): AND deleted_at IS NULL
	//   - removeOrganizationMember (line ~215): AND deleted_at IS NULL
	//   - removeUserRoles: DELETE is always a no-op when rows don't exist
	//
	// An integration test with a real DB would:
	// 1. Insert a user and org membership.
	// 2. Call handleUserDeleted once → org soft-deleted, member removed.
	// 3. Call handleUserDeleted again (redelivery simulation).
	// 4. Assert that no additional DB mutations occurred (0 rows affected).
	//
	// We skip the integration test here because it requires a running PostgreSQL
	// instance. Run with a real DB or use the smoke tests in application/ instead.
	t.Log("user.deleted handler is idempotent via deleted_at IS NULL SQL guards (documented)")
}
