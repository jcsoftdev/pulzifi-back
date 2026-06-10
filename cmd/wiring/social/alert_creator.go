// Package socialwiring provides cross-module adapters for the social module.
// All adapters implement domain ports defined in modules/social/domain/services,
// keeping the social module free of cross-module imports (REQ-WIRING-01).
package socialwiring

import (
	"context"
	"encoding/json"

	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/services"
	valueobjects "github.com/jcsoftdev/pulzifi-back/modules/social/domain/value_objects"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"go.uber.org/zap"
)

// alertCreatorAdapter implements social services.AlertCreator by publishing a
// TopicAlertCreated domain event through the event bus. This fires existing email
// and integration (Slack/Discord/etc.) notifications with zero changes in those
// modules (design §8).
//
// Social alerts are not written to the alerts DB table because that table has a
// NOT NULL FK on check_id referencing the checks table (URL monitoring). Phase 2
// will add a social_alerts table for in-app display.
type alertCreatorAdapter struct {
	bus    eventbus.MessageBus
	lookup *orgLookup
	tenant string
}

// Compile-time interface check.
var _ services.AlertCreator = (*alertCreatorAdapter)(nil)

// NewAlertCreator constructs a tenant-scoped AlertCreator adapter backed by the
// event bus. The tenant is fixed at construction time (scheduler path).
// The adapter publishes TopicAlertCreated events so the integration module's
// dispatch worker delivers them to Slack/Discord/etc. (REQ-WIRING-02, design §8).
func NewAlertCreator(bus eventbus.MessageBus, lookup *orgLookup, tenant string) services.AlertCreator {
	return &alertCreatorAdapter{
		bus:    bus,
		lookup: lookup,
		tenant: tenant,
	}
}

// NewHTTPAlertCreator constructs an AlertCreator that resolves the tenant from
// the request context at CreateAlert call time (HTTP module path).
func NewHTTPAlertCreator(bus eventbus.MessageBus, lookup *orgLookup) services.AlertCreator {
	return &contextAwareAlertCreatorAdapter{bus: bus, lookup: lookup}
}

// CreateAlert publishes a TopicAlertCreated event for the social change or
// suspension notification. All errors are logged and swallowed — alerts are
// best-effort so run_check is never blocked by notification failures.
func (a *alertCreatorAdapter) CreateAlert(ctx context.Context, payload services.AlertPayload) error {
	if a.bus == nil {
		return nil
	}

	orgID, err := a.lookup.lookupOrgID(ctx, a.tenant)
	if err != nil {
		logger.Warn("social alert creator: failed to resolve org ID, alert not published",
			zap.String("tenant", a.tenant),
			zap.Error(err),
		)
		return nil
	}

	alertType := "social_change"
	if payload.Suspended {
		alertType = "social_suspension"
	}

	data, err := json.Marshal(map[string]any{
		"profile_id":   payload.ProfileID.String(),
		"handle":       payload.Handle,
		"platform":     string(payload.Platform),
		"change_types": changeTypeStrings(payload.ChangeTypes),
		"summary":      payload.Summary,
		"suspended":    payload.Suspended,
		"alert_type":   alertType,
		"title":        socialAlertTitle(payload),
		"description":  payload.Summary,
	})
	if err != nil {
		logger.Warn("social alert creator: failed to marshal event data",
			zap.String("tenant", a.tenant),
			zap.Error(err),
		)
		return nil
	}

	ev := eventbus.DomainEvent{
		Type:   eventbus.TopicAlertCreated,
		OrgID:  orgID,
		Tenant: a.tenant,
		Data:   data,
	}

	if err := eventbus.PublishDomainEvent(a.bus, ev); err != nil {
		logger.Warn("social alert creator: failed to publish event",
			zap.String("tenant", a.tenant),
			zap.Error(err),
		)
		// best-effort: swallow so run_check is not blocked
	}
	return nil
}

// contextAwareAlertCreatorAdapter is an AlertCreator that reads the tenant from
// the request context at CreateAlert time rather than at construction time.
// Used by the HTTP module where the tenant is request-scoped (HTTP middleware).
type contextAwareAlertCreatorAdapter struct {
	bus    eventbus.MessageBus
	lookup *orgLookup
}

// Compile-time interface check.
var _ services.AlertCreator = (*contextAwareAlertCreatorAdapter)(nil)

// CreateAlert reads the tenant from context and publishes the alert event.
func (a *contextAwareAlertCreatorAdapter) CreateAlert(ctx context.Context, payload services.AlertPayload) error {
	if a.bus == nil {
		return nil
	}
	tenant := middleware.GetTenantFromContext(ctx)
	if tenant == "" {
		return nil // no tenant context — skip alert (e.g. test path)
	}
	inner := &alertCreatorAdapter{bus: a.bus, lookup: a.lookup, tenant: tenant}
	return inner.CreateAlert(ctx, payload)
}

// socialAlertTitle builds a human-readable alert title for integration delivery.
func socialAlertTitle(payload services.AlertPayload) string {
	if payload.Suspended {
		return "@" + payload.Handle + " monitoring suspended"
	}
	return "@" + payload.Handle + " social changes detected"
}

// changeTypeStrings converts a slice of ChangeType domain values to plain strings
// for JSON serialisation inside the alert event payload.
func changeTypeStrings(cts []valueobjects.ChangeType) []string {
	out := make([]string, len(cts))
	for i, ct := range cts {
		out[i] = string(ct)
	}
	return out
}
