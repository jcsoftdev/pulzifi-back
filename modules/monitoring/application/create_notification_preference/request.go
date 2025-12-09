package create_notification_preference

import (
	"github.com/google/uuid"
)

type CreateNotificationPreferenceRequest struct {
	UserID       uuid.UUID  `json:"user_id"`
	WorkspaceID  *uuid.UUID `json:"workspace_id,omitempty"`
	PageID       *uuid.UUID `json:"page_id,omitempty"`
	EmailEnabled bool       `json:"email_enabled"`
	ChangeTypes  []string   `json:"change_types"`
}
