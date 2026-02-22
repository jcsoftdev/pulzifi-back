package listchecks

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/repositories/mocks"
)

func TestListChecksHandler_Handle(t *testing.T) {
	pageID := uuid.New()

	checks := []*entities.Check{
		{ID: uuid.New(), PageID: pageID, Status: "success", CheckedAt: time.Now()},
		{ID: uuid.New(), PageID: pageID, Status: "error", ErrorMessage: "timeout", CheckedAt: time.Now()},
	}

	tests := []struct {
		name      string
		result    []*entities.Check
		repoErr   error
		wantErr   bool
		wantCount int
	}{
		{
			name:      "returns checks",
			result:    checks,
			wantCount: 2,
		},
		{
			name:      "empty result",
			result:    nil,
			wantCount: 0,
		},
		{
			name:    "repo error",
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockCheckRepository{
				ListByPageResult: tt.result,
				ListByPageErr:    tt.repoErr,
			}

			handler := NewListChecksHandler(repo)
			resp, err := handler.Handle(context.Background(), pageID)

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
			if len(resp.Checks) != tt.wantCount {
				t.Errorf("checks count: want %d, got %d", tt.wantCount, len(resp.Checks))
			}
		})
	}
}
