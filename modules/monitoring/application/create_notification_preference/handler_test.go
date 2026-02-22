package create_notification_preference

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/repositories/mocks"
)

func TestCreateNotificationPreferenceHandler_Handle(t *testing.T) {
	userID := uuid.New()
	workspaceID := uuid.New()
	pageID := uuid.New()

	tests := []struct {
		name    string
		req     *CreateNotificationPreferenceRequest
		repoErr error
		wantErr bool
	}{
		{
			name: "create with workspace scope",
			req: &CreateNotificationPreferenceRequest{
				UserID:       userID,
				WorkspaceID:  &workspaceID,
				EmailEnabled: true,
				ChangeTypes:  []string{"page_change"},
			},
		},
		{
			name: "create with page scope",
			req: &CreateNotificationPreferenceRequest{
				UserID:       userID,
				PageID:       &pageID,
				EmailEnabled: false,
				ChangeTypes:  []string{"error", "performance_drop"},
			},
		},
		{
			name: "repo error",
			req: &CreateNotificationPreferenceRequest{
				UserID:       userID,
				WorkspaceID:  &workspaceID,
				EmailEnabled: true,
				ChangeTypes:  []string{"page_change"},
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockNotificationPreferenceRepository{
				CreateErr: tt.repoErr,
			}

			handler := NewCreateNotificationPreferenceHandler(repo)
			resp, err := handler.Handle(context.Background(), tt.req)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp == nil {
				t.Fatal("expected non-nil response")
			}
			if resp.UserID != userID {
				t.Errorf("user_id: want %v, got %v", userID, resp.UserID)
			}
			if resp.EmailEnabled != tt.req.EmailEnabled {
				t.Errorf("email_enabled: want %v, got %v", tt.req.EmailEnabled, resp.EmailEnabled)
			}
			if repo.CreateCalls != 1 {
				t.Errorf("expected 1 Create call, got %d", repo.CreateCalls)
			}
		})
	}
}
