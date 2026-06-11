package getunreadcount_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	getunreadcount "github.com/jcsoftdev/pulzifi-back/modules/social/application/get_unread_count"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/social/infrastructure/persistence/memory"
)

func TestGetUnreadCount(t *testing.T) {
	repo := memory.NewAlertRepository()
	ws := uuid.New()
	_ = repo.Save(context.Background(), entities.NewSocialSuspensionAlert(uuid.New(), ws, "suspended", "x"))

	h := getunreadcount.NewHandler(repo)
	resp, err := h.Handle(context.Background(), ws)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.UnreadCount != 1 {
		t.Errorf("expected 1, got %d", resp.UnreadCount)
	}
}
