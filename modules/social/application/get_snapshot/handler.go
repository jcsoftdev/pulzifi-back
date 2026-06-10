package getsnapshot

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories"
)

type Handler struct {
	snapshots repositories.SnapshotRepository
}

func NewHandler(snapshots repositories.SnapshotRepository) *Handler {
	return &Handler{snapshots: snapshots}
}

func (h *Handler) Handle(ctx context.Context, snapshotID uuid.UUID) (*Response, error) {
	s, err := h.snapshots.GetByID(ctx, snapshotID)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, nil
	}
	return snapshotToResponse(s), nil
}
