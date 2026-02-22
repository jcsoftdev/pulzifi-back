package createalert

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/alert/domain/repositories/mocks"
)

func TestCreateAlertHandler_Handle(t *testing.T) {
	workspaceID := uuid.New()
	pageID := uuid.New()
	checkID := uuid.New()

	tests := []struct {
		name    string
		req     *CreateAlertRequest
		repoErr error
		wantErr bool
	}{
		{
			name: "successful creation",
			req: &CreateAlertRequest{
				WorkspaceID: workspaceID,
				PageID:      pageID,
				CheckID:     checkID,
				Type:        "change_detected",
				Title:       "Content changed",
				Description: "Major changes detected",
			},
			wantErr: false,
		},
		{
			name: "repository error",
			req: &CreateAlertRequest{
				WorkspaceID: workspaceID,
				PageID:      pageID,
				CheckID:     checkID,
				Type:        "change_detected",
				Title:       "Content changed",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockAlertRepository{
				CreateErr: tt.repoErr,
			}

			handler := NewCreateAlertHandler(repo)
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

			if resp.ID == uuid.Nil {
				t.Error("response ID should not be nil")
			}
			if resp.WorkspaceID != tt.req.WorkspaceID {
				t.Errorf("workspace ID: want %v, got %v", tt.req.WorkspaceID, resp.WorkspaceID)
			}
			if resp.PageID != tt.req.PageID {
				t.Errorf("page ID: want %v, got %v", tt.req.PageID, resp.PageID)
			}
			if resp.CheckID != tt.req.CheckID {
				t.Errorf("check ID: want %v, got %v", tt.req.CheckID, resp.CheckID)
			}
			if resp.Type != tt.req.Type {
				t.Errorf("type: want %q, got %q", tt.req.Type, resp.Type)
			}
			if resp.Title != tt.req.Title {
				t.Errorf("title: want %q, got %q", tt.req.Title, resp.Title)
			}
			if resp.CreatedAt.IsZero() {
				t.Error("created_at should not be zero")
			}
			if repo.CreateCalls != 1 {
				t.Errorf("expected 1 Create call, got %d", repo.CreateCalls)
			}
		})
	}
}
