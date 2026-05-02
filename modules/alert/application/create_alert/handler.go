package createalert

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/alert/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/alert/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type CreateAlertHandler struct {
	repo   repositories.AlertRepository
	bus    eventbus.MessageBus // optional; nil disables publishing
	tenant string
	orgID  uuid.UUID
}

func NewCreateAlertHandler(repo repositories.AlertRepository) *CreateAlertHandler {
	return &CreateAlertHandler{repo: repo}
}

// SetEventContext enables alert.created publishing by attaching tenant + org context + bus.
func (h *CreateAlertHandler) SetEventContext(bus eventbus.MessageBus, tenant string, orgID uuid.UUID) {
	h.bus = bus
	h.tenant = tenant
	h.orgID = orgID
}

func (h *CreateAlertHandler) Handle(ctx context.Context, req *CreateAlertRequest) (*CreateAlertResponse, error) {
	// Create alert entity
	alert := &entities.Alert{
		ID:          uuid.New(),
		WorkspaceID: req.WorkspaceID,
		PageID:      req.PageID,
		CheckID:     req.CheckID,
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
	}

	// Save to database
	if err := h.repo.Create(ctx, alert); err != nil {
		return nil, err
	}

	// Publish alert.created domain event for cross-module integrations.
	if h.bus != nil && h.orgID != uuid.Nil {
		data, err := json.Marshal(map[string]any{
			"alert_id":    alert.ID.String(),
			"title":       alert.Title,
			"description": alert.Description,
			"type":        alert.Type,
			"metadata":    alert.Metadata,
		})
		if err == nil {
			pid := alert.PageID
			wsid := alert.WorkspaceID
			ev := eventbus.DomainEvent{
				Type:        eventbus.TopicAlertCreated,
				OrgID:       h.orgID,
				WorkspaceID: &wsid,
				PageID:      &pid,
				Tenant:      h.tenant,
				Data:        data,
			}
			if perr := eventbus.PublishDomainEvent(h.bus, ev); perr != nil {
				logger.Error("publish alert.created", zap.Error(perr))
			}
		}
	}

	// Return response
	return &CreateAlertResponse{
		ID:          alert.ID,
		WorkspaceID: alert.WorkspaceID,
		PageID:      alert.PageID,
		CheckID:     alert.CheckID,
		Type:        alert.Type,
		Title:       alert.Title,
		Description: alert.Description,
		Metadata:    alert.Metadata,
		CreatedAt:   alert.CreatedAt,
	}, nil
}

// HTTP Handler wrapper
func (h *CreateAlertHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	var req CreateAlertRequest

	// Parse JSON body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.WorkspaceID == uuid.Nil || req.PageID == uuid.Nil || req.CheckID == uuid.Nil {
		http.Error(w, "workspace_id, page_id, and check_id are required", http.StatusBadRequest)
		return
	}
	if req.Type == "" || req.Title == "" {
		http.Error(w, "type and title are required", http.StatusBadRequest)
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
