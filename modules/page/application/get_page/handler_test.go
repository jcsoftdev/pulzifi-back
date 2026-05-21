package getpage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/page/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/page/domain/repositories/mocks"
)

func TestGetPageHandler_Handle(t *testing.T) {
	pageID := uuid.New()
	workspaceID := uuid.New()
	now := time.Now()

	existingPage := &entities.Page{
		ID:              pageID,
		WorkspaceID:     workspaceID,
		Name:            "My Page",
		URL:             "https://example.com",
		Tags:            []string{"tag1"},
		CheckFrequency:  "Every day",
		DetectedChanges: 2,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	tests := []struct {
		name         string
		repoResult   *entities.Page
		repoErr      error
		wantErr      bool
		wantNil      bool
		wantGetCalls int
	}{
		{
			name:         "happy path — found returns DTO",
			repoResult:   existingPage,
			wantErr:      false,
			wantNil:      false,
			wantGetCalls: 1,
		},
		{
			name:         "not found returns nil (no error)",
			repoResult:   nil,
			wantErr:      false,
			wantNil:      true,
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
			repo := &mocks.MockPageRepository{
				GetByIDResult: tt.repoResult,
				GetByIDErr:    tt.repoErr,
			}
			handler := NewGetPageHandler(repo)

			// Act
			resp, err := handler.Handle(context.Background(), pageID)

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
			if tt.wantNil {
				if resp != nil {
					t.Errorf("expected nil response, got %+v", resp)
				}
				return
			}
			if resp == nil {
				t.Fatal("expected non-nil response")
			}
			if resp.ID != pageID {
				t.Errorf("ID: want %v, got %v", pageID, resp.ID)
			}
			if resp.Name != existingPage.Name {
				t.Errorf("Name: want %q, got %q", existingPage.Name, resp.Name)
			}
			if repo.GetByIDCalls != tt.wantGetCalls {
				t.Errorf("GetByIDCalls: want %d, got %d", tt.wantGetCalls, repo.GetByIDCalls)
			}
		})
	}
}
