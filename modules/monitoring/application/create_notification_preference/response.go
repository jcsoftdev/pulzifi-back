package create_notification_preference

import (
	"time"

	"github.com/google/uuid"
)

type CreateNotificationPreferenceResponse struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	WorkspaceID  *uuid.UUID `json:"workspace_id"`
	PageID       *uuid.UUID `json:"page_id"`
	EmailEnabled bool `json:"email_enabled"`
	ChangeTypes  []string `json:"change_types"`
	CreatedAt    time.Time `json:"created_at"`
}
