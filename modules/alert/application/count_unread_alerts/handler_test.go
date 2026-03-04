package countunreadalerts

import (
	"context"
	"errors"
	"testing"

	"github.com/jcsoftdev/pulzifi-back/modules/alert/domain/repositories/mocks"
)

func TestCountUnreadAlertsHandler_Handle(t *testing.T) {
	tests := []struct {
		name      string
		count     int
		repoErr   error
		wantErr   bool
		wantHas   bool
		wantCount int
	}{
		{
			name:      "has unread alerts",
			count:     5,
			wantHas:   true,
			wantCount: 5,
		},
		{
			name:      "no unread alerts",
			count:     0,
			wantHas:   false,
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
				CountUnreadResult: tt.count,
				CountUnreadErr:    tt.repoErr,
			}

			handler := NewCountUnreadAlertsHandler(repo)
			resp, err := handler.Handle(context.Background())

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp.HasNotifications != tt.wantHas {
				t.Errorf("has_notifications: want %v, got %v", tt.wantHas, resp.HasNotifications)
			}
			if resp.NotificationCount != tt.wantCount {
				t.Errorf("notification_count: want %d, got %d", tt.wantCount, resp.NotificationCount)
			}
			if repo.CountUnreadCalls != 1 {
				t.Errorf("expected 1 CountUnread call, got %d", repo.CountUnreadCalls)
			}
		})
	}
}
