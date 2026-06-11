package markalertread_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	markalertread "github.com/jcsoftdev/pulzifi-back/modules/social/application/mark_alert_read"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/social/infrastructure/persistence/memory"
)

func TestMarkAlertRead(t *testing.T) {
	repo := memory.NewAlertRepository()
	ws := uuid.New()
	alert := entities.NewSocialSuspensionAlert(uuid.New(), ws, "suspended", "x")
	_ = repo.Save(context.Background(), alert)

	h := markalertread.NewHandler(repo)
	if err := h.Handle(context.Background(), alert.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	count, _ := repo.UnreadCount(context.Background(), ws)
	if count != 0 {
		t.Errorf("expected 0 unread after mark read, got %d", count)
	}
}
