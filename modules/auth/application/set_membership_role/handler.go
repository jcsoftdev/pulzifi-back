package setmembershiprole

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrInvalidRole       = errors.New("invalid role")
	ErrCannotModifyOwner = errors.New("cannot modify the organization owner")
)

var validRoles = map[string]bool{"ADMIN": true, "MEMBER": true}

// Writer is the port implemented by the persistence layer.
type Writer interface {
	IsOrgOwner(ctx context.Context, userID, orgID uuid.UUID) (bool, error)
	SetRole(ctx context.Context, userID, orgID uuid.UUID, role string) error
}

// Request carries the inbound command.
type Request struct {
	UserID uuid.UUID
	OrgID  uuid.UUID
	Role   string
}

// Handler orchestrates the set_membership_role use case.
type Handler struct{ writer Writer }

// NewHandler creates a Handler with the given persistence writer.
func NewHandler(writer Writer) *Handler { return &Handler{writer: writer} }

// Handle validates the role, guards against modifying the org owner, then delegates to the writer.
func (h *Handler) Handle(ctx context.Context, req Request) error {
	if !validRoles[req.Role] {
		return ErrInvalidRole
	}
	owner, err := h.writer.IsOrgOwner(ctx, req.UserID, req.OrgID)
	if err != nil {
		return err
	}
	if owner {
		return ErrCannotModifyOwner
	}
	return h.writer.SetRole(ctx, req.UserID, req.OrgID, req.Role)
}
