package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
)

type fakeFlags struct{ on map[string]bool }

func (f fakeFlags) IsOn(_ context.Context, _ uuid.UUID, key string) (bool, error) {
	return f.on[key], nil
}

func TestProviderAuthType(t *testing.T) {
	cases := map[string]string{
		"slack": "oauth", "email": "none", "twilio": "byo", "sheets": "oauth", "unknown": "",
	}
	for key, want := range cases {
		if got := services.ProviderAuthType(key); got != want {
			t.Errorf("ProviderAuthType(%q) = %q, want %q", key, got, want)
		}
	}
}

func TestKnownProvidersHasEightCanonicalKeys(t *testing.T) {
	want := map[string]bool{
		"slack": true, "email": true, "gmail": true, "outlook": true,
		"discord": true, "twilio": true, "sheets": true, "teams": true,
	}
	if len(services.KnownProviders) != len(want) {
		t.Fatalf("KnownProviders len = %d, want %d", len(services.KnownProviders), len(want))
	}
	for _, d := range services.KnownProviders {
		if !want[d.Key] {
			t.Errorf("unexpected provider key %q (note: backend uses 'sheets', not 'google_sheets')", d.Key)
		}
	}
}

func TestProviderGateOpen(t *testing.T) {
	ctx := context.Background()
	org := uuid.New()

	if !services.ProviderGateOpen(ctx, fakeFlags{}, org, "slack") {
		t.Error("slack should be gate-open without flags")
	}
	if services.ProviderGateOpen(ctx, fakeFlags{on: map[string]bool{}}, org, "gmail") {
		t.Error("gmail should be gate-closed when flag off")
	}
	flags := fakeFlags{on: map[string]bool{"integrations.gmail": true}}
	if !services.ProviderGateOpen(ctx, flags, org, "gmail") {
		t.Error("gmail should be gate-open when flag on")
	}
	if !services.ProviderGateOpen(ctx, nil, org, "gmail") {
		t.Error("nil flags should be permissive")
	}
}
