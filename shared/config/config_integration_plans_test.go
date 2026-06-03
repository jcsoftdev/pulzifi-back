package config

import (
	"testing"
)

// TestLoad_IntegrationPaidPlans_Default asserts that when INTEGRATION_PAID_PLANS
// is not set, Load() returns the four expected default plan codes.
func TestLoad_IntegrationPaidPlans_Default(t *testing.T) {
	setEnvMap(t, requiredEnv())

	cfg := Load()

	want := []string{"trial", "starter", "pro", "enterprise"}
	if len(cfg.IntegrationPaidPlans) != len(want) {
		t.Fatalf("expected %d plans, got %d: %v", len(want), len(cfg.IntegrationPaidPlans), cfg.IntegrationPaidPlans)
	}
	for i, p := range want {
		if cfg.IntegrationPaidPlans[i] != p {
			t.Errorf("IntegrationPaidPlans[%d]: want %q, got %q", i, p, cfg.IntegrationPaidPlans[i])
		}
	}
}

// TestLoad_IntegrationPaidPlans_Override asserts that INTEGRATION_PAID_PLANS
// is parsed correctly from a comma-separated env var (trimming spaces).
func TestLoad_IntegrationPaidPlans_Override(t *testing.T) {
	env := requiredEnv()
	env["INTEGRATION_PAID_PLANS"] = " pro , enterprise "
	setEnvMap(t, env)

	cfg := Load()

	want := []string{"pro", "enterprise"}
	if len(cfg.IntegrationPaidPlans) != len(want) {
		t.Fatalf("expected %d plans, got %d: %v", len(want), len(cfg.IntegrationPaidPlans), cfg.IntegrationPaidPlans)
	}
	for i, p := range want {
		if cfg.IntegrationPaidPlans[i] != p {
			t.Errorf("IntegrationPaidPlans[%d]: want %q, got %q", i, p, cfg.IntegrationPaidPlans[i])
		}
	}
}
