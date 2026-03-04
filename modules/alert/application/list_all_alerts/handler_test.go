package listallalerts

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/alert/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/alert/domain/repositories/mocks"
)

func TestListAllAlertsHandler_Handle(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		alerts    []*entities.AlertWithPage
		repoErr   error
		wantErr   bool
		wantCount int
	}{
		{
			name: "returns alerts with page info",
			alerts: []*entities.AlertWithPage{
				{
					Alert: entities.Alert{
						ID:        uuid.New(),
						Title:     "Content changed",
						Type:      "content_change",
						CreatedAt: now,
					},
					PageName: "Homepage",
					PageURL:  "https://example.com",
				},
			},
			wantCount: 1,
		},
		{
			name:      "empty list",
			alerts:    nil,
			wantCount: 0,
		},
		{
			name:    "repository error",
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockAlertRepository{
				ListAllResult: tt.alerts,
				ListAllErr:    tt.repoErr,
			}

			handler := NewListAllAlertsHandler(repo)
			resp, err := handler.Handle(context.Background(), 20)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(resp.Data) != tt.wantCount {
				t.Errorf("data count: want %d, got %d", tt.wantCount, len(resp.Data))
			}
			if repo.ListAllCalls != 1 {
				t.Errorf("expected 1 ListAll call, got %d", repo.ListAllCalls)
			}

			if tt.wantCount > 0 {
				item := resp.Data[0]
				if item.PageName != "Homepage" {
					t.Errorf("page_name: want %q, got %q", "Homepage", item.PageName)
				}
				if item.PageURL != "https://example.com" {
					t.Errorf("page_url: want %q, got %q", "https://example.com", item.PageURL)
				}
			}
		})
	}
}
