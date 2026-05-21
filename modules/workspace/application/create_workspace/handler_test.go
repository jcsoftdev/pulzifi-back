package createworkspace

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/repositories/mocks"
)

func TestCreateWorkspaceHandler_Handle(t *testing.T) {
	createdBy := uuid.New()

	tests := []struct {
		name         string
		req          *CreateWorkspaceRequest
		createErr    error
		addMemberErr error
		wantErr      bool
	}{
		{
			name: "happy path — workspace created",
			req: &CreateWorkspaceRequest{
				Name: "My Workspace",
				Type: "personal",
				Tags: []string{"tag1"},
			},
			wantErr: false,
		},
		{
			name: "nil tags — no panic",
			req: &CreateWorkspaceRequest{
				Name: "My Workspace",
				Type: "personal",
				Tags: nil,
			},
			wantErr: false,
		},
		{
			name: "repo create error propagated",
			req: &CreateWorkspaceRequest{
				Name: "My Workspace",
				Type: "personal",
			},
			createErr: errors.New("db error"),
			wantErr:   true,
		},
		{
			name: "AddMember error propagated",
			req: &CreateWorkspaceRequest{
				Name: "My Workspace",
				Type: "personal",
			},
			addMemberErr: errors.New("member add failed"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mocks.MockWorkspaceRepository{
				CreateErr:    tt.createErr,
				AddMemberErr: tt.addMemberErr,
			}
			handler := NewCreateWorkspaceHandler(repo)

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
			if resp.Name != tt.req.Name {
				t.Errorf("Name: want %q, got %q", tt.req.Name, resp.Name)
			}
			if resp.Type != tt.req.Type {
				t.Errorf("Type: want %q, got %q", tt.req.Type, resp.Type)
			}
			if resp.CreatedBy != createdBy {
				t.Errorf("CreatedBy: want %v, got %v", createdBy, resp.CreatedBy)
			}
			if resp.CreatedAt.IsZero() {
				t.Error("CreatedAt should not be zero")
			}
		})
	}
}
