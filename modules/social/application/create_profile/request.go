package createprofile

import "github.com/google/uuid"

// Request is the input DTO for creating a social profile.
type Request struct {
	WorkspaceID          uuid.UUID `json:"workspace_id"`
	Platform             string    `json:"platform"`
	Handle               string    `json:"handle"`
	CheckIntervalMinutes int       `json:"check_interval_minutes"`
}
