package services

import (
	"context"

	"github.com/google/uuid"
	valueobjects "github.com/jcsoftdev/pulzifi-back/modules/social/domain/value_objects"
)

// AlertPayload carries the information delivered to the alert system when a
// social change is detected or when a profile is suspended after repeated failures.
type AlertPayload struct {
	// ProfileID is the social profile that triggered the alert.
	ProfileID uuid.UUID

	// Handle is the @handle of the profile (e.g. "nasa").
	Handle string

	// Platform is the social network platform.
	Platform valueobjects.Platform

	// ChangeTypes lists every detected change category (empty for suspension alerts).
	ChangeTypes []valueobjects.ChangeType

	// Summary is a human-readable description of what changed
	// (e.g. "@nasa posted 2 new posts, bio changed").
	Summary string

	// Suspended is true when the alert is a monitoring-suspension notification
	// (consecutive failures reached the threshold).
	Suspended bool
}

// AlertCreator is the port for delivering social monitoring alerts.
// The concrete adapter lives in cmd/wiring/social/alert_creator.go and
// dispatches through the alert module (same pattern as cmd/wiring/snapshot).
type AlertCreator interface {
	// CreateAlert delivers an alert for a social change or profile suspension.
	// Implementations must be non-blocking with respect to the caller — i.e. they
	// may publish asynchronously but must not block run_check indefinitely.
	CreateAlert(ctx context.Context, payload AlertPayload) error
}
