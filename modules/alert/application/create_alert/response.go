package createalert

import (
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/alert/domain/entities"
)

type CreateAlertResponse struct {
	ID          uuid.UUID         `json:"id"`
	WorkspaceID uuid.UUID         `json:"workspace_id"`
	PageID      uuid.UUID         `json:"page_id"`
	CheckID     uuid.UUID         `json:"check_id"`
	Type        string            `json:"type"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Metadata    entities.Metadata `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
}
