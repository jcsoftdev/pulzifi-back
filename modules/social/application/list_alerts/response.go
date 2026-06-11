package listalerts

import (
	"time"

	"github.com/google/uuid"
)

// AlertResponse is the API representation of a SocialAlert.
type AlertResponse struct {
	ID          uuid.UUID  `json:"id"`
	ProfileID   uuid.UUID  `json:"profile_id"`
	WorkspaceID uuid.UUID  `json:"workspace_id"`
	ChangeID    *uuid.UUID `json:"change_id,omitempty"`
	AlertType   string     `json:"alert_type"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	ChangeTypes []string   `json:"change_types"`
	IsRead      bool       `json:"is_read"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Response is the paginated list payload.
type Response struct {
	Alerts []AlertResponse `json:"alerts"`
}
