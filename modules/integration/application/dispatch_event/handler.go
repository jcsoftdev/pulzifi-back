package dispatchevent

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
)

// RepoFactory returns tenant-scoped repository instances.
type RepoFactory interface {
	DestinationRepoForTenant(tenant string) repositories.DestinationRepository
	DeliveryRepoForTenant(tenant string) repositories.DeliveryRepository
}

// OrgGuard checks whether an organisation is active and may receive deliveries.
type OrgGuard interface {
	IsActive(ctx context.Context, orgID uuid.UUID) (bool, error)
}

// Handler dispatches a DomainEvent: it resolves matching destinations and
// enqueues one Delivery row per destination (status=pending).
type Handler struct {
	factory RepoFactory
	org     OrgGuard
}

// NewHandler constructs a Handler.
func NewHandler(f RepoFactory, og OrgGuard) *Handler { return &Handler{factory: f, org: og} }

// Handle is the entry point. It is safe to call concurrently.
func (h *Handler) Handle(ctx context.Context, ev eventbus.DomainEvent) error {
	active, err := h.org.IsActive(ctx, ev.OrgID)
	if err != nil || !active {
		// inactive org → no-op (err is nil when org is merely inactive)
		return err
	}

	destRepo := h.factory.DestinationRepoForTenant(ev.Tenant)
	delRepo := h.factory.DeliveryRepoForTenant(ev.Tenant)

	dests, err := destRepo.ResolveForEvent(ctx, ev.Type, ev.OrgID, ev.WorkspaceID, ev.PageID)
	if err != nil {
		return err
	}

	// Initialise to an empty map so the JSONB column is never SQL NULL.
	data := map[string]any{}
	if len(ev.Data) > 0 {
		_ = json.Unmarshal(ev.Data, &data)
	}

	for _, d := range dests {
		del := &entities.Delivery{
			ID:            uuid.New(),
			DestinationID: d.ID,
			EventType:     ev.Type,
			EventPayload:  data,
			Status:        entities.DeliveryPending,
		}
		if err := delRepo.Create(ctx, del); err != nil {
			// stop on first persistence failure; partial enqueue is acceptable
			// for retry-from-source semantics
			return err
		}
	}
	return nil
}
