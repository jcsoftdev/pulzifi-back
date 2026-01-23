package createpage

import (
	"github.com/google/uuid"
)

type CreatePageRequest struct {
	WorkspaceID uuid.UUID `json:"workspace_id" binding:"required"`
	Name        string    `json:"name" binding:"required"`
	URL         string    `json:"url" binding:"required"`
	Tags        []string  `json:"tags"`
}
