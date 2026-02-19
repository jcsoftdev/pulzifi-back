package entities

import "time"

type WorkspaceChanges struct {
	WorkspaceName   string
	DetectedChanges int
}

type RecentAlert struct {
	CheckedAt     time.Time
	WorkspaceName string
	ChangeType    string
	PageURL       string
}

type RecentInsight struct {
	CreatedAt     time.Time
	WorkspaceName string
	PageURL       string
	Title         string
	Content       string
}

type DashboardStats struct {
	WorkspacesCount     int
	PagesCount          int
	TodayChecksCount    int
	ChangesPerWorkspace []WorkspaceChanges
	RecentAlerts        []RecentAlert
	RecentInsights      []RecentInsight
}
