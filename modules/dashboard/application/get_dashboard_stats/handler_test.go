package getdashboardstats

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/dashboard/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/dashboard/domain/repositories/mocks"
)

func TestGetDashboardStatsHandler_Handle(t *testing.T) {
	now := time.Now()

	fullStats := &entities.DashboardStats{
		WorkspacesCount:  2,
		PagesCount:       5,
		TodayChecksCount: 10,
		ChangesPerWorkspace: []entities.WorkspaceChanges{
			{WorkspaceName: "Workspace A", DetectedChanges: 3},
		},
		RecentAlerts: []entities.RecentAlert{
			{CheckedAt: now, WorkspaceName: "Workspace A", ChangeType: "visual", PageURL: "https://example.com"},
		},
		RecentInsights: []entities.RecentInsight{
			{CreatedAt: now, WorkspaceName: "Workspace A", PageURL: "https://example.com", Title: "Insight 1", Content: "Content"},
		},
	}

	tests := []struct {
		name            string
		statsResult     *entities.DashboardStats
		statsErr        error
		wantErr         bool
		wantWorkspaces  int
		wantPages       int
		wantGetStatsCalls int
	}{
		{
			name:              "happy path returns aggregated stats DTO",
			statsResult:       fullStats,
			wantErr:           false,
			wantWorkspaces:    2,
			wantPages:         5,
			wantGetStatsCalls: 1,
		},
		{
			name:     "stats repo error propagated",
			statsErr: errors.New("db error"),
			wantErr:  true,
		},
		{
			name: "empty stats — no error",
			statsResult: &entities.DashboardStats{
				WorkspacesCount:     0,
				PagesCount:          0,
				TodayChecksCount:    0,
				ChangesPerWorkspace: []entities.WorkspaceChanges{},
				RecentAlerts:        []entities.RecentAlert{},
				RecentInsights:      []entities.RecentInsight{},
			},
			wantErr:           false,
			wantWorkspaces:    0,
			wantPages:         0,
			wantGetStatsCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mocks.MockDashboardRepository{
				GetStatsResult: tt.statsResult,
				GetStatsErr:    tt.statsErr,
			}
			handler := NewGetDashboardStatsHandler(repo)

			// Act
			resp, err := handler.Handle(context.Background())

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
			if resp.WorkspacesCount != tt.wantWorkspaces {
				t.Errorf("WorkspacesCount: want %d, got %d", tt.wantWorkspaces, resp.WorkspacesCount)
			}
			if resp.PagesCount != tt.wantPages {
				t.Errorf("PagesCount: want %d, got %d", tt.wantPages, resp.PagesCount)
			}
			if repo.GetStatsCalls != tt.wantGetStatsCalls {
				t.Errorf("GetStatsCalls: want %d, got %d", tt.wantGetStatsCalls, repo.GetStatsCalls)
			}
		})
	}
}
