package entities

import (
	"time"

	"github.com/google/uuid"
)

type Alert struct {
	ID            uuid.UUID
	WorkspaceID   uuid.UUID
	PageID        uuid.UUID
	CheckID       uuid.UUID
	Type          string
	Title         string
	Description   string
	ChangeSummary string // Specific change description from Vision AI
	Metadata      Metadata
	ReadAt        *time.Time
	CreatedAt     time.Time
}

// Metadata is flexible JSON data
type Metadata map[string]interface{}

type AlertWithPage struct {
	Alert
	PageName string
	PageURL  string
}

func NewAlert(workspaceID, pageID, checkID uuid.UUID, alertType, title, description string) *Alert {
	return &Alert{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		PageID:      pageID,
		CheckID:     checkID,
		Type:        alertType,
		Title:       title,
		Description: description,
		CreatedAt:   time.Now(),
	}
}
