package dispatchevent

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
	domainservices "github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

const emailServiceType = "email"

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
//
// Plan-tiered channel gate
// ========================
// When a ChannelEntitlement is provided (via NewHandlerWithEntitlement), the
// handler enforces the following rule at dispatch time:
//
//   - Paid org  (IsPaid == true)  → all matching destinations receive a Delivery.
//   - Free org  (IsPaid == false) → only destinations whose ServiceType is "email"
//     receive a Delivery; all other channels are silently skipped with an
//     info-level log entry to preserve observability.
//
// Error behaviour (documented design decision):
// If the entitlement check returns an error the handler FAILS OPEN TO EMAIL ONLY.
// Rationale: delivering a Slack/Discord alert to a free org on a transient DB
// error is worse than missing those channels for a paid org. Email is always
// delivered regardless. The error is logged at warn level.
//
// When entitlement is nil (NewHandler, legacy call sites) all destinations pass
// through unchanged — this preserves backward compatibility.
type Handler struct {
	factory      RepoFactory
	org          OrgGuard
	entitlement  domainservices.ChannelEntitlement // optional; nil = all channels allowed
}

// NewHandler constructs a Handler without a channel entitlement gate.
// All destinations are delivered regardless of the org's plan — use this only
// for backward-compatible call sites or tests that don't need plan gating.
func NewHandler(f RepoFactory, og OrgGuard) *Handler {
	return &Handler{factory: f, org: og}
}

// NewHandlerWithEntitlement constructs a Handler that enforces plan-tiered
// channel access via the supplied ChannelEntitlement port.
func NewHandlerWithEntitlement(f RepoFactory, og OrgGuard, e domainservices.ChannelEntitlement) *Handler {
	return &Handler{factory: f, org: og, entitlement: e}
}

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

	// Plan-tiered channel gate: determine whether this org may use non-email channels.
	emailOnly := false
	if h.entitlement != nil {
		paid, entErr := h.entitlement.IsPaid(ctx, ev.OrgID)
		if entErr != nil {
			// Fail open to email-only on transient error. Email alert still goes
			// out; paid channels are conservatively skipped. The error is logged
			// so ops can diagnose DB or network issues.
			logger.Warn("channel entitlement check failed — failing open to email-only",
				zap.String("org_id", ev.OrgID.String()),
				zap.Error(entErr),
			)
			emailOnly = true
		} else if !paid {
			emailOnly = true
		}
	}

	// Initialise to an empty map so the JSONB column is never SQL NULL.
	data := map[string]any{}
	if len(ev.Data) > 0 {
		_ = json.Unmarshal(ev.Data, &data)
	}

	skipped := 0
	for _, d := range dests {
		if emailOnly && !strings.EqualFold(strings.TrimSpace(d.ServiceType), emailServiceType) {
			skipped++
			continue
		}
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

	if skipped > 0 {
		logger.Info("channel entitlement: non-email destinations skipped for free org",
			zap.String("org_id", ev.OrgID.String()),
			zap.Int("skipped", skipped),
		)
	}

	return nil
}
