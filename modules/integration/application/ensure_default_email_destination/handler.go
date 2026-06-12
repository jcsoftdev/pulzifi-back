package ensuredefaultemaildestination

import (
	"context"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
)

// defaultEvents are the events the seeded org-level email destination
// subscribes to. Matches the frontend's default selection.
var defaultEvents = []string{"change.detected", "alert.created"}

type Request struct {
	OrgID      uuid.UUID
	OwnerEmail string
}

type Response struct {
	// Created is true when a default destination was seeded by this call.
	Created bool
}

type Handler struct {
	repo repositories.DestinationRepository
}

func NewHandler(repo repositories.DestinationRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle seeds an org-scoped email destination pointing at the org owner's
// email when the org has no email destination yet. No-op (never an error)
// when the owner email is unknown or the org already has one, so callers can
// run it opportunistically before listing.
func (h *Handler) Handle(ctx context.Context, req Request) (*Response, error) {
	if req.OrgID == uuid.Nil || req.OwnerEmail == "" {
		return &Response{Created: false}, nil
	}

	existing, err := h.repo.ListByScope(ctx, entities.ScopeOrg, req.OrgID)
	if err != nil {
		return nil, err
	}
	for _, d := range existing {
		if d.ServiceType == "email" {
			return &Response{Created: false}, nil
		}
	}

	d := &entities.Destination{
		ID:          uuid.New(),
		ServiceType: "email",
		ScopeType:   entities.ScopeOrg,
		ScopeID:     req.OrgID,
		Target:      map[string]any{"emails": []any{req.OwnerEmail}},
		Events:      defaultEvents,
		Enabled:     true,
	}
	if err := h.repo.Create(ctx, d); err != nil {
		return nil, err
	}
	return &Response{Created: true}, nil
}
