package getchange

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/social/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories"
	valueobjects "github.com/jcsoftdev/pulzifi-back/modules/social/domain/value_objects"
)

// Handler handles the get-social-change use case.
type Handler struct {
	changes   repositories.ChangeRepository
	snapshots repositories.SnapshotRepository
}

// NewHandler creates a new Handler.
func NewHandler(changes repositories.ChangeRepository, snapshots repositories.SnapshotRepository) *Handler {
	return &Handler{changes: changes, snapshots: snapshots}
}

// Handle returns the full before/after payload for a social change (REQ-FEED-03).
func (h *Handler) Handle(ctx context.Context, id uuid.UUID) (*Response, error) {
	change, err := h.changes.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if change == nil {
		return nil, errors.New(domainerrors.ErrProfileNotFound.Error()) // reuse not-found semantics
	}

	// Load from/to snapshots to get the full before/after ProfileData
	fromSnapshot, err := h.snapshots.GetByID(ctx, change.FromSnapshotID)
	if err != nil {
		return nil, err
	}
	toSnapshot, err := h.snapshots.GetByID(ctx, change.ToSnapshotID)
	if err != nil {
		return nil, err
	}

	detail := &ChangeDetail{
		ID:             change.ID,
		ProfileID:      change.ProfileID,
		FromSnapshotID: change.FromSnapshotID,
		ToSnapshotID:   change.ToSnapshotID,
		ChangeTypes:    changeTypesToStrings(change.ChangeTypes),
		Summary:        change.Summary,
		CreatedAt:      change.CreatedAt,
	}

	resp := &Response{Change: detail}
	if fromSnapshot != nil {
		resp.BeforeData = fromSnapshot.Data
	}
	if toSnapshot != nil {
		resp.AfterData = toSnapshot.Data
	}

	return resp, nil
}

func changeTypesToStrings(cts []valueobjects.ChangeType) []string {
	result := make([]string, len(cts))
	for i, ct := range cts {
		result[i] = string(ct)
	}
	return result
}

// Ensure entities is used (it is, via fromSnapshot.Data which is *entities.ProfileData)
var _ = (*entities.SocialChange)(nil)
