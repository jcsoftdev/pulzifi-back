package removemembership

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// ErrCannotRemoveOwner is returned when the target user is the org owner.
var ErrCannotRemoveOwner = errors.New("cannot remove the organization owner")

// Remover is the port implemented by the persistence layer.
type Remover interface {
	IsOrgOwner(ctx context.Context, userID, orgID uuid.UUID) (bool, error)
	RemoveMembership(ctx context.Context, userID, orgID uuid.UUID) error
}

// Request carries the inbound command.
type Request struct {
	UserID uuid.UUID
	OrgID  uuid.UUID
}

// Handler orchestrates the remove_membership use case.
type Handler struct{ remover Remover }

// NewHandler creates a Handler backed by the given Remover port.
func NewHandler(remover Remover) *Handler { return &Handler{remover: remover} }

// Handle executes the use case: guard against owner removal, then soft-delete the membership.
func (h *Handler) Handle(ctx context.Context, req Request) error {
	owner, err := h.remover.IsOrgOwner(ctx, req.UserID, req.OrgID)
	if err != nil {
		return err
	}
	if owner {
		return ErrCannotRemoveOwner
	}
	return h.remover.RemoveMembership(ctx, req.UserID, req.OrgID)
}
