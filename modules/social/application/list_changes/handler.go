package listchanges

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories"
	valueobjects "github.com/jcsoftdev/pulzifi-back/modules/social/domain/value_objects"
)

// Handler handles the list-social-changes use case.
type Handler struct {
	changes repositories.ChangeRepository
}

// NewHandler creates a new Handler.
func NewHandler(changes repositories.ChangeRepository) *Handler {
	return &Handler{changes: changes}
}

// Handle returns a paginated list of social changes matching the request filter.
// Results are in descending created_at order (newest first).
func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	filter := repositories.ChangeFilter{
		Limit:  req.Limit,
		Offset: req.Offset,
	}
	// ProfileID takes precedence over WorkspaceID (REQ-FEED-01)
	if req.ProfileID != nil {
		filter.ProfileID = req.ProfileID
	} else if req.WorkspaceID != nil {
		filter.WorkspaceID = req.WorkspaceID
	}

	changes, err := h.changes.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]*ChangeSummaryItem, 0, len(changes))
	for _, c := range changes {
		items = append(items, toSummaryItem(c))
	}
	return &Response{Changes: items}, nil
}

func toSummaryItem(c *entities.SocialChange) *ChangeSummaryItem {
	types := make([]string, len(c.ChangeTypes))
	for i, ct := range c.ChangeTypes {
		types[i] = changeTypeStr(ct)
	}
	return &ChangeSummaryItem{
		ID:             c.ID,
		ProfileID:      c.ProfileID,
		FromSnapshotID: c.FromSnapshotID,
		ToSnapshotID:   c.ToSnapshotID,
		ChangeTypes:    types,
		CreatedAt:      c.CreatedAt,
	}
}

func changeTypeStr(ct valueobjects.ChangeType) string {
	return string(ct)
}
