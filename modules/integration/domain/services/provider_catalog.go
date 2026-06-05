package services

import (
	"context"

	"github.com/google/uuid"
)

// Auth-type constants for provider connection flows.
const (
	AuthTypeOAuth = "oauth"
	AuthTypeBYO   = "byo"
	AuthTypeNone  = "none"
)

// ProviderDescriptor is the canonical, compile-time description of a provider.
type ProviderDescriptor struct {
	Key      string
	AuthType string
}

// KnownProviders is the single source of truth for which providers exist.
// Keys MUST match the backend registry service types (e.g. "sheets", not "google_sheets").
var KnownProviders = []ProviderDescriptor{
	{Key: "slack", AuthType: AuthTypeOAuth},
	{Key: "email", AuthType: AuthTypeNone},
	{Key: "gmail", AuthType: AuthTypeOAuth},
	{Key: "outlook", AuthType: AuthTypeOAuth},
	{Key: "discord", AuthType: AuthTypeOAuth},
	{Key: "twilio", AuthType: AuthTypeBYO},
	{Key: "sheets", AuthType: AuthTypeOAuth},
	{Key: "teams", AuthType: AuthTypeOAuth},
}

// ProviderAuthType returns the auth type for a provider key, or "" if unknown.
func ProviderAuthType(key string) string {
	for _, d := range KnownProviders {
		if d.Key == key {
			return d.AuthType
		}
	}
	return ""
}

// alwaysVisibleProviders are never gated by a per-org feature flag.
var alwaysVisibleProviders = map[string]bool{"slack": true, "email": true}

// FlagReader is the narrow feature-flag dependency used by the catalog gate.
// *featureflags.Reader satisfies it.
type FlagReader interface {
	IsOn(ctx context.Context, orgID uuid.UUID, key string) (bool, error)
}

// ProviderGateOpen reports whether a provider passes the org-level feature gate.
// slack/email are always open. A nil flags reader is permissive (legacy dev behavior).
// Callers MUST pass a true-nil interface when no reader exists (guard the concrete
// pointer before calling) to avoid Go's typed-nil interface pitfall.
func ProviderGateOpen(ctx context.Context, flags FlagReader, orgID uuid.UUID, key string) bool {
	if alwaysVisibleProviders[key] {
		return true
	}
	if flags == nil {
		return true
	}
	on, _ := flags.IsOn(ctx, orgID, "integrations."+key)
	return on
}
