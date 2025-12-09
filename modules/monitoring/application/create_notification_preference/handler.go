package create_notification_preference

import (
	"context"
	"encoding/json"
	"net/http"

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

// HTTP Handler wrapper
func (h *CreateNotificationPreferenceHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	var req CreateNotificationPreferenceRequest

	// Parse JSON body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate: either workspace_id or page_id, not both
	if (req.WorkspaceID == nil && req.PageID == nil) || (req.WorkspaceID != nil && req.PageID != nil) {
		http.Error(w, "Either workspace_id or page_id must be provided, not both", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.UserID == [16]byte{} {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	// Execute use case
	resp, err := h.Handle(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
