package listpages

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/page/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/page/domain/repositories/mocks"
)

func TestListPagesHandler_Handle(t *testing.T) {
	workspaceID := uuid.New()
	now := time.Now()

	pages := []*entities.Page{
		{
			ID:              uuid.New(),
			WorkspaceID:     workspaceID,
			Name:            "Page 1",
			URL:             "https://example1.com",
			Tags:            []string{"tag1"},
			CheckFrequency:  "Every day",
			DetectedChanges: 0,
			CreatedBy:       uuid.New(),
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              uuid.New(),
			WorkspaceID:     workspaceID,
			Name:            "Page 2",
			URL:             "https://example2.com",
			Tags:            []string{},
			CheckFrequency:  "Off",
			DetectedChanges: 3,
			CreatedBy:       uuid.New(),
			CreatedAt:       now,
			UpdatedAt:       now,
		},
	}

	tests := []struct {
		name         string
		repoResult   []*entities.Page
		repoErr      error
		wantErr      bool
		wantCount    int
		wantListCalls int
	}{
		{
			name:          "happy path returns page slice",
			repoResult:    pages,
			wantErr:       false,
			wantCount:     2,
			wantListCalls: 1,
		},
		{
			name:          "empty list — no error",
			repoResult:    []*entities.Page{},
			wantErr:       false,
			wantCount:     0,
			wantListCalls: 1,
		},
		{
			name:          "nil result — treated as empty",
			repoResult:    nil,
			wantErr:       false,
			wantCount:     0,
			wantListCalls: 1,
		},
		{
			name:    "repo error propagated",
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var capturedWorkspaceID uuid.UUID
			repo := &mocks.MockPageRepository{
				ListByWorkResult: tt.repoResult,
				ListByWorkErr:    tt.repoErr,
				ListByWorkFn: func(ctx context.Context, wsID uuid.UUID) ([]*entities.Page, error) {
					capturedWorkspaceID = wsID
					return tt.repoResult, tt.repoErr
				},
			}
			handler := NewListPagesHandler(repo)

			// Act
			resp, err := handler.Handle(context.Background(), workspaceID)

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
			if len(resp.Pages) != tt.wantCount {
				t.Errorf("expected %d pages, got %d", tt.wantCount, len(resp.Pages))
			}
			// Verify workspace filter is applied correctly
			if tt.wantListCalls > 0 && capturedWorkspaceID != workspaceID {
				t.Errorf("workspace filter: want %v, got %v", workspaceID, capturedWorkspaceID)
			}
		})
	}
}
