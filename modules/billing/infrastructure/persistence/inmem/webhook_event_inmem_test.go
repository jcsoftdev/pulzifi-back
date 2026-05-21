package inmem

import (
	"context"
	"testing"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
)

func TestWebhookEventRepo_SaveNewEvent(t *testing.T) {
	repo := NewWebhookEventRepo()
	event := &entities.WebhookEvent{
		EventID:    "evt_new_001",
		EventType:  "customer.subscription.updated",
		ReceivedAt: time.Now(),
		Status:     entities.WebhookEventReceived,
	}

	inserted, err := repo.Save(context.Background(), event)
	if err != nil {
		t.Fatalf("Save: unexpected error: %v", err)
	}
	if !inserted {
		t.Fatal("expected inserted=true for new event, got false")
	}
}

func TestWebhookEventRepo_SaveDuplicateIsIdempotent(t *testing.T) {
	repo := NewWebhookEventRepo()
	event := &entities.WebhookEvent{
		EventID:    "evt_dup_002",
		EventType:  "customer.subscription.deleted",
		ReceivedAt: time.Now(),
		Status:     entities.WebhookEventReceived,
	}

	first, err := repo.Save(context.Background(), event)
	if err != nil || !first {
		t.Fatalf("first Save: want (true, nil), got (%v, %v)", first, err)
	}

	second, err := repo.Save(context.Background(), event)
	if err != nil {
		t.Fatalf("second Save: unexpected error: %v", err)
	}
	if second {
		t.Fatal("second Save: expected inserted=false for duplicate, got true")
	}
}

func TestWebhookEventRepo_MarkProcessed(t *testing.T) {
	repo := NewWebhookEventRepo()
	event := &entities.WebhookEvent{
		EventID:    "evt_mark_003",
		EventType:  "invoice.paid",
		ReceivedAt: time.Now(),
		Status:     entities.WebhookEventReceived,
	}

	_, _ = repo.Save(context.Background(), event)

	if err := repo.MarkProcessed(context.Background(), event.EventID, entities.WebhookEventProcessed); err != nil {
		t.Fatalf("MarkProcessed: unexpected error: %v", err)
	}

	// Verify status was updated in the stored record.
	stored := repo.events[event.EventID]
	if stored.Status != entities.WebhookEventProcessed {
		t.Errorf("Status after MarkProcessed: want 'processed', got %q", stored.Status)
	}
	if stored.ProcessedAt == nil {
		t.Error("ProcessedAt should be non-nil after MarkProcessed")
	}
}

func TestWebhookEventRepo_MarkProcessed_UnknownEventIsNoOp(t *testing.T) {
	repo := NewWebhookEventRepo()

	// Should not panic or error for an event that was never saved.
	if err := repo.MarkProcessed(context.Background(), "evt_never_seen", entities.WebhookEventProcessed); err != nil {
		t.Errorf("expected no error for unknown eventID, got: %v", err)
	}
}
