package createalert

import (
	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/alert/domain/entities"
)

type CreateAlertRequest struct {
	WorkspaceID uuid.UUID         `json:"workspace_id" binding:"required"`
	PageID      uuid.UUID         `json:"page_id" binding:"required"`
	CheckID     uuid.UUID         `json:"check_id" binding:"required"`
	Type        string            `json:"type" binding:"required"`
	Title       string            `json:"title" binding:"required"`
	Description string            `json:"description"`
	Metadata    entities.Metadata `json:"metadata"`
}
