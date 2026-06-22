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

// FallbackRecipientResolver supplies the default recipient emails (a workspace's
// active members) used to seed an email destination when none is configured yet.
type FallbackRecipientResolver interface {
	WorkspaceMemberEmails(ctx context.Context, tenant string, workspaceID uuid.UUID) ([]string, error)
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
	factory     RepoFactory
	org         OrgGuard
	entitlement domainservices.ChannelEntitlement // optional; nil = all channels allowed
	fallback    FallbackRecipientResolver         // optional; nil = no default-email seeding
}

// SetFallbackResolver enables the default-email fallback: when an event resolves no
// email destination, a workspace-scoped email destination is seeded from the
// workspace's active members so change alerts reach the team without manual setup.
// Optional — safe to leave unset (no seeding occurs).
func (h *Handler) SetFallbackResolver(r FallbackRecipientResolver) { h.fallback = r }

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

// maybeSeedDefaultEmailDestination returns a newly-persisted workspace-scoped email
// destination when the event resolved no email destination and the workspace has
// active members. Returns nil (and never an error) otherwise — fallback is
// best-effort and must never block other destinations from being delivered.
func (h *Handler) maybeSeedDefaultEmailDestination(
	ctx context.Context,
	destRepo repositories.DestinationRepository,
	ev eventbus.DomainEvent,
	resolved []*entities.Destination,
) *entities.Destination {
	if h.fallback == nil || ev.WorkspaceID == nil {
		return nil
	}
	for _, d := range resolved {
		if strings.EqualFold(strings.TrimSpace(d.ServiceType), emailServiceType) {
			return nil // a configured email destination already covers this event
		}
	}

	// Re-check the workspace scope directly before seeding. ResolveForEvent already
	// told us no email destination matched, but a concurrent dispatch for a sibling
	// page in the same workspace may have seeded one in between — this narrows that
	// race so we don't create a duplicate default destination.
	if existing, lerr := destRepo.ListByScope(ctx, entities.ScopeWorkspace, *ev.WorkspaceID); lerr == nil {
		for _, d := range existing {
			if strings.EqualFold(strings.TrimSpace(d.ServiceType), emailServiceType) {
				return nil
			}
		}
	}

	emails, err := h.fallback.WorkspaceMemberEmails(ctx, ev.Tenant, *ev.WorkspaceID)
	if err != nil {
		logger.Warn("dispatch: default-email fallback lookup failed",
			zap.String("tenant", ev.Tenant), zap.Error(err))
		return nil
	}
	if len(emails) == 0 {
		return nil
	}

	dest := &entities.Destination{
		ID:          uuid.New(),
		ServiceType: emailServiceType,
		ScopeType:   entities.ScopeWorkspace,
		ScopeID:     *ev.WorkspaceID,
		Target:      map[string]any{"emails": emails},
		Events:      []string{eventbus.TopicChangeDetected, eventbus.TopicAlertCreated},
		Enabled:     true,
	}
	if err := destRepo.Create(ctx, dest); err != nil {
		logger.Warn("dispatch: failed to seed default email destination",
			zap.String("tenant", ev.Tenant), zap.Error(err))
		return nil
	}
	logger.Info("dispatch: seeded default email destination for workspace",
		zap.String("tenant", ev.Tenant),
		zap.String("workspace_id", ev.WorkspaceID.String()),
		zap.Int("recipients", len(emails)))
	return dest
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

	// Default-on fallback: when no email destination is configured for this event,
	// seed a workspace-scoped email destination targeting the workspace's active
	// members so change alerts reach the team without manual setup. Once seeded — or
	// once the user configures emails in settings — this branch stops firing, so it
	// never duplicates a configured destination.
	if seeded := h.maybeSeedDefaultEmailDestination(ctx, destRepo, ev, dests); seeded != nil {
		dests = append(dests, seeded)
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
