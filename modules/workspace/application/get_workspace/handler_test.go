package getworkspace

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/repositories/mocks"
)

func TestGetWorkspaceHandler_Handle(t *testing.T) {
	workspaceID := uuid.New()
	createdBy := uuid.New()
	now := time.Now()

	existingWorkspace := &entities.Workspace{
		ID:        workspaceID,
		Name:      "My Workspace",
		Type:      "personal",
		Tags:      []string{"tag1"},
		CreatedBy: createdBy,
		CreatedAt: now,
		UpdatedAt: now,
	}

	tests := []struct {
		name         string
		repoResult   *entities.Workspace
		repoErr      error
		wantErr      bool
		wantErrIs    error
		wantGetCalls int
	}{
		{
			name:         "happy path — found returns response",
			repoResult:   existingWorkspace,
			wantErr:      false,
			wantGetCalls: 1,
		},
		{
			name:         "not found returns ErrWorkspaceNotFound",
			repoResult:   nil,
			wantErr:      true,
			wantErrIs:    ErrWorkspaceNotFound,
			wantGetCalls: 1,
		},
		{
			name:         "repo error propagated",
			repoErr:      errors.New("db error"),
			wantErr:      true,
			wantGetCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mocks.MockWorkspaceRepository{
				GetByIDResult: tt.repoResult,
				GetByIDErr:    tt.repoErr,
			}
			handler := NewGetWorkspaceHandler(repo)

			// Act
			resp, err := handler.Handle(context.Background(), workspaceID)

			// Assert
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("error = %v, want %v", err, tt.wantErrIs)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp == nil {
				t.Fatal("expected non-nil response")
			}
			if resp.ID != workspaceID {
				t.Errorf("ID: want %v, got %v", workspaceID, resp.ID)
			}
			if resp.Name != existingWorkspace.Name {
				t.Errorf("Name: want %q, got %q", existingWorkspace.Name, resp.Name)
			}
			if repo.GetByIDCalls != tt.wantGetCalls {
				t.Errorf("GetByIDCalls: want %d, got %d", tt.wantGetCalls, repo.GetByIDCalls)
			}
		})
	}
}
