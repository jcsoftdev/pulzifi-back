package entities

import (
	"time"

	"github.com/google/uuid"
)

type NotificationPreference struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	WorkspaceID  *uuid.UUID // null if page_id is set
	PageID       *uuid.UUID // null if workspace_id is set
	EmailEnabled bool
	ChangeTypes  []string // ["page_change", "error", "performance_drop"]
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewNotificationPreference(userID uuid.UUID, emailEnabled bool) *NotificationPreference {
	return &NotificationPreference{
		ID:           uuid.New(),
		UserID:       userID,
		EmailEnabled: emailEnabled,
		ChangeTypes:  []string{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// SetWorkspace sets the workspace and clears page
func (np *NotificationPreference) SetWorkspace(workspaceID uuid.UUID) {
	np.WorkspaceID = &workspaceID
	np.PageID = nil
	np.UpdatedAt = time.Now()
}

// SetPage sets the page and clears workspace
func (np *NotificationPreference) SetPage(pageID uuid.UUID) {
	np.PageID = &pageID
	np.WorkspaceID = nil
	np.UpdatedAt = time.Now()
}
