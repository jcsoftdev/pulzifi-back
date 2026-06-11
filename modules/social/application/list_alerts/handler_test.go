package listalerts_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	listalerts "github.com/jcsoftdev/pulzifi-back/modules/social/application/list_alerts"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/social/infrastructure/persistence/memory"
	valueobjects "github.com/jcsoftdev/pulzifi-back/modules/social/domain/value_objects"
)

func TestListAlerts_ReturnsWorkspaceAlerts(t *testing.T) {
	repo := memory.NewAlertRepository()
	ws := uuid.New()
	profile := uuid.New()
	change := uuid.New()
	_ = repo.Save(context.Background(), entities.NewSocialChangeAlert(
		profile, ws, change, "@nasa changes detected", "followers +100",
		[]valueobjects.ChangeType{valueobjects.ChangeTypeFollowersChanged},
	))

	h := listalerts.NewHandler(repo)
	resp, err := h.Handle(context.Background(), ws, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(resp.Alerts))
	}
	if resp.Alerts[0].Title != "@nasa changes detected" {
		t.Errorf("unexpected title: %s", resp.Alerts[0].Title)
	}
	if len(resp.Alerts[0].ChangeTypes) != 1 || resp.Alerts[0].ChangeTypes[0] != "followers_changed" {
		t.Errorf("unexpected change types: %v", resp.Alerts[0].ChangeTypes)
	}
}
