package markallalerts

import (
	"context"
	"errors"
	"testing"

	"github.com/jcsoftdev/pulzifi-back/modules/alert/domain/repositories/mocks"
)

func TestMarkAllAlertsReadHandler_Handle(t *testing.T) {
	tests := []struct {
		name    string
		repoErr error
		wantErr bool
	}{
		{
			name: "successful mark all as read",
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
				MarkAllAsReadErr: tt.repoErr,
			}

			handler := NewMarkAllAlertsReadHandler(repo)
			err := handler.Handle(context.Background())

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if repo.MarkAllAsReadCalls != 1 {
				t.Errorf("expected 1 MarkAllAsRead call, got %d", repo.MarkAllAsReadCalls)
			}
		})
	}
}
