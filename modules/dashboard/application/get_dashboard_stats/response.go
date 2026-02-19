package getdashboardstats

import (
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/dashboard/domain/entities"
)

type WorkspaceChangesResponse struct {
	WorkspaceName   string `json:"workspace_name"`
	DetectedChanges int    `json:"detected_changes"`
}

type RecentAlertResponse struct {
	CheckedAt     time.Time `json:"checked_at"`
	WorkspaceName string    `json:"workspace_name"`
	ChangeType    string    `json:"change_type"`
	PageURL       string    `json:"page_url"`
}

type RecentInsightResponse struct {
	CreatedAt     time.Time `json:"created_at"`
	WorkspaceName string    `json:"workspace_name"`
	PageURL       string    `json:"page_url"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
}

type GetDashboardStatsResponse struct {
	WorkspacesCount     int                        `json:"workspaces_count"`
	PagesCount          int                        `json:"pages_count"`
	TodayChecksCount    int                        `json:"today_checks_count"`
	ChangesPerWorkspace []WorkspaceChangesResponse `json:"changes_per_workspace"`
	RecentAlerts        []RecentAlertResponse      `json:"recent_alerts"`
	RecentInsights      []RecentInsightResponse    `json:"recent_insights"`
}

func buildResponse(stats *entities.DashboardStats) *GetDashboardStatsResponse {
	changesPerWorkspace := make([]WorkspaceChangesResponse, len(stats.ChangesPerWorkspace))
	for i, c := range stats.ChangesPerWorkspace {
		changesPerWorkspace[i] = WorkspaceChangesResponse{
			WorkspaceName:   c.WorkspaceName,
			DetectedChanges: c.DetectedChanges,
		}
	}

	recentAlerts := make([]RecentAlertResponse, len(stats.RecentAlerts))
	for i, a := range stats.RecentAlerts {
		recentAlerts[i] = RecentAlertResponse{
			CheckedAt:     a.CheckedAt,
			WorkspaceName: a.WorkspaceName,
			ChangeType:    a.ChangeType,
			PageURL:       a.PageURL,
		}
	}

	recentInsights := make([]RecentInsightResponse, len(stats.RecentInsights))
	for i, ins := range stats.RecentInsights {
		recentInsights[i] = RecentInsightResponse{
			CreatedAt:     ins.CreatedAt,
			WorkspaceName: ins.WorkspaceName,
			PageURL:       ins.PageURL,
			Title:         ins.Title,
			Content:       ins.Content,
		}
	}

	return &GetDashboardStatsResponse{
		WorkspacesCount:     stats.WorkspacesCount,
		PagesCount:          stats.PagesCount,
		TodayChecksCount:    stats.TodayChecksCount,
		ChangesPerWorkspace: changesPerWorkspace,
		RecentAlerts:        recentAlerts,
		RecentInsights:      recentInsights,
	}
}
