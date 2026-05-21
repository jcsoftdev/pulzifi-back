package createpage

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/page/domain/repositories/mocks"
)

func TestCreatePageHandler_Handle(t *testing.T) {
	workspaceID := uuid.New()
	createdBy := uuid.New()

	tests := []struct {
		name      string
		req       *CreatePageRequest
		createErr error
		wantErr   bool
	}{
		{
			name: "happy path — page created",
			req: &CreatePageRequest{
				WorkspaceID: workspaceID,
				Name:        "My Page",
				URL:         "https://example.com",
				Tags:        []string{"tag1"},
			},
			wantErr: false,
		},
		{
			name: "happy path — nil tags normalized to empty slice",
			req: &CreatePageRequest{
				WorkspaceID: workspaceID,
				Name:        "My Page",
				URL:         "https://example.com",
				Tags:        nil,
			},
			wantErr: false,
		},
		{
			name: "repo create error propagated",
			req: &CreatePageRequest{
				WorkspaceID: workspaceID,
				Name:        "My Page",
				URL:         "https://example.com",
			},
			createErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mocks.MockPageRepository{
				CreateErr: tt.createErr,
			}
			handler := NewCreatePageHandler(repo)

			// Act
			resp, err := handler.Handle(context.Background(), tt.req, createdBy)

			// Assert
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
			if resp.ID == uuid.Nil {
				t.Error("response ID should not be nil UUID")
			}
			if resp.WorkspaceID != tt.req.WorkspaceID {
				t.Errorf("WorkspaceID: want %v, got %v", tt.req.WorkspaceID, resp.WorkspaceID)
			}
			if resp.Name != tt.req.Name {
				t.Errorf("Name: want %q, got %q", tt.req.Name, resp.Name)
			}
			if resp.URL != tt.req.URL {
				t.Errorf("URL: want %q, got %q", tt.req.URL, resp.URL)
			}
			if resp.Tags == nil {
				t.Error("Tags should not be nil (normalized to [])")
			}
			if resp.CreatedAt.IsZero() {
				t.Error("CreatedAt should not be zero")
			}
			if repo.CreateCalls != 1 {
				t.Errorf("expected 1 Create call, got %d", repo.CreateCalls)
			}
		})
	}
}
