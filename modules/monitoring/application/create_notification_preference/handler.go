package create_notification_preference

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/repositories"
)

type CreateNotificationPreferenceHandler struct {
	repo repositories.NotificationPreferenceRepository
}

func NewCreateNotificationPreferenceHandler(repo repositories.NotificationPreferenceRepository) *CreateNotificationPreferenceHandler {
	return &CreateNotificationPreferenceHandler{repo: repo}
}

func (h *CreateNotificationPreferenceHandler) Handle(ctx context.Context, req *CreateNotificationPreferenceRequest) (*CreateNotificationPreferenceResponse, error) {
	pref := entities.NewNotificationPreference(req.UserID, req.EmailEnabled)
	pref.ChangeTypes = req.ChangeTypes

	if req.WorkspaceID != nil {
		pref.SetWorkspace(*req.WorkspaceID)
	} else if req.PageID != nil {
		pref.SetPage(*req.PageID)
	}

	if err := h.repo.Create(ctx, pref); err != nil {
		return nil, err
	}

	return &CreateNotificationPreferenceResponse{
		ID:           pref.ID,
		UserID:       pref.UserID,
		WorkspaceID:  pref.WorkspaceID,
		PageID:       pref.PageID,
		EmailEnabled: pref.EmailEnabled,
		ChangeTypes:  pref.ChangeTypes,
		CreatedAt:    pref.CreatedAt,
	}, nil
}
