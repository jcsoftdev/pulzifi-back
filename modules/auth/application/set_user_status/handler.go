package setuserstatus

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
)

var (
	// ErrInvalidStatus is returned when the requested status is not "approved" or "suspended".
	ErrInvalidStatus = errors.New("invalid status")
	// ErrCannotSuspendSelf is returned when an admin tries to suspend their own account.
	ErrCannotSuspendSelf = errors.New("cannot suspend yourself")
)

// Writer is the port implemented by the persistence layer.
type Writer interface {
	SetStatus(ctx context.Context, id uuid.UUID, status string) error
}

// Request carries the inbound command.
type Request struct {
	TargetID uuid.UUID
	ActorID  uuid.UUID
	Status   string
}

// Handler executes the set_user_status use case.
type Handler struct{ writer Writer }

// NewHandler returns a new Handler backed by the given writer.
func NewHandler(writer Writer) *Handler { return &Handler{writer: writer} }

// Handle validates the request and delegates to the writer.
func (h *Handler) Handle(ctx context.Context, req Request) error {
	if req.Status != entities.UserStatusApproved && req.Status != entities.UserStatusSuspended {
		return ErrInvalidStatus
	}
	if req.Status == entities.UserStatusSuspended && req.TargetID == req.ActorID {
		return ErrCannotSuspendSelf
	}
	return h.writer.SetStatus(ctx, req.TargetID, req.Status)
}
