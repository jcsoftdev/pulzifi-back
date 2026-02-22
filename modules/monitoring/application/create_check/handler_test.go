package createcheck

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/repositories/mocks"
)

func TestCreateCheckHandler_Handle(t *testing.T) {
	pageID := uuid.New()

	tests := []struct {
		name    string
		req     *CreateCheckRequest
		repoErr error
		wantErr bool
	}{
		{
			name: "successful creation",
			req: &CreateCheckRequest{
				PageID:         pageID,
				Status:         "success",
				ChangeDetected: true,
				ChangeType:     "content",
				ScreenshotURL:  "https://cdn.example.com/shot.png",
				DurationMs:     1234,
			},
			wantErr: false,
		},
		{
			name: "repository error",
			req: &CreateCheckRequest{
				PageID:         pageID,
				Status:         "error",
				ChangeDetected: false,
				ErrorMessage:   "timeout",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockCheckRepository{
				CreateErr: tt.repoErr,
			}

			handler := NewCreateCheckHandler(repo)
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
			if resp.PageID != tt.req.PageID {
				t.Errorf("page ID: want %v, got %v", tt.req.PageID, resp.PageID)
			}
			if resp.Status != tt.req.Status {
				t.Errorf("status: want %q, got %q", tt.req.Status, resp.Status)
			}
			if resp.ChangeDetected != tt.req.ChangeDetected {
				t.Errorf("change_detected: want %v, got %v", tt.req.ChangeDetected, resp.ChangeDetected)
			}
			if resp.CheckedAt.IsZero() {
				t.Error("checked_at should not be zero")
			}
			if repo.CreateCalls != 1 {
				t.Errorf("expected 1 Create call, got %d", repo.CreateCalls)
			}
		})
	}
}
