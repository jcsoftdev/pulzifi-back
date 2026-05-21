package listinsights

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/entities"
)

// mockInsightRepository is a hand-rolled mock for repositories.InsightRepository.
type mockInsightRepository struct {
	ListByPageIDResult  []*entities.Insight
	ListByPageIDErr     error
	ListByCheckIDResult []*entities.Insight
	ListByCheckIDErr    error

	ListByPageIDCalls  int
	ListByCheckIDCalls int
}

func (m *mockInsightRepository) Create(_ context.Context, _ *entities.Insight) error { return nil }

func (m *mockInsightRepository) ListByPageID(_ context.Context, _ uuid.UUID) ([]*entities.Insight, error) {
	m.ListByPageIDCalls++
	return m.ListByPageIDResult, m.ListByPageIDErr
}

func (m *mockInsightRepository) ListByCheckID(_ context.Context, _ uuid.UUID) ([]*entities.Insight, error) {
	m.ListByCheckIDCalls++
	return m.ListByCheckIDResult, m.ListByCheckIDErr
}

func (m *mockInsightRepository) GetByID(_ context.Context, _ uuid.UUID) (*entities.Insight, error) {
	return nil, nil
}

func TestListInsightsHandler_Handle(t *testing.T) {
	pageID := uuid.New()
	now := time.Now()

	insights := []*entities.Insight{
		{
			ID:          uuid.New(),
			PageID:      pageID,
			CheckID:     uuid.New(),
			InsightType: "marketing",
			Title:       "Market Analysis",
			Content:     "Some content here",
			CreatedAt:   now,
		},
	}

	tests := []struct {
		name        string
		repoResult  []*entities.Insight
		repoErr     error
		wantErr     bool
		wantCount   int
		wantPageCalls int
	}{
		{
			name:          "happy path returns insights for page",
			repoResult:    insights,
			wantErr:       false,
			wantCount:     1,
			wantPageCalls: 1,
		},
		{
			name:          "empty list — no error",
			repoResult:    []*entities.Insight{},
			wantErr:       false,
			wantCount:     0,
			wantPageCalls: 1,
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
			repo := &mockInsightRepository{
				ListByPageIDResult: tt.repoResult,
				ListByPageIDErr:    tt.repoErr,
			}
			handler := NewListInsightsHandler(repo)

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
			if resp == nil {
				t.Fatal("expected non-nil response")
			}
			if len(resp.Insights) != tt.wantCount {
				t.Errorf("expected %d insights, got %d", tt.wantCount, len(resp.Insights))
			}
			if repo.ListByPageIDCalls != tt.wantPageCalls {
				t.Errorf("ListByPageIDCalls: want %d, got %d", tt.wantPageCalls, repo.ListByPageIDCalls)
			}
		})
	}
}

func TestListInsightsHandler_HandleByCheckID(t *testing.T) {
	checkID := uuid.New()

	tests := []struct {
		name       string
		repoResult []*entities.Insight
		repoErr    error
		wantErr    bool
		wantCount  int
	}{
		{
			name:       "happy path by check ID",
			repoResult: []*entities.Insight{{ID: uuid.New(), CheckID: checkID, CreatedAt: time.Now()}},
			wantErr:    false,
			wantCount:  1,
		},
		{
			name:    "repo error propagated",
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockInsightRepository{
				ListByCheckIDResult: tt.repoResult,
				ListByCheckIDErr:    tt.repoErr,
			}
			handler := NewListInsightsHandler(repo)
			resp, err := handler.HandleByCheckID(context.Background(), checkID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(resp.Insights) != tt.wantCount {
				t.Errorf("expected %d insights, got %d", tt.wantCount, len(resp.Insights))
			}
			if repo.ListByCheckIDCalls != 1 {
				t.Errorf("expected 1 ListByCheckID call, got %d", repo.ListByCheckIDCalls)
			}
		})
	}
}
