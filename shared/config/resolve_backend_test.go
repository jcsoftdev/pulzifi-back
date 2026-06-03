package config

// Additional ResolveBackend tests beyond those in config_ha_test.go.
// These focus on the explicit override always winning regardless of InstanceMode.

import "testing"

func TestResolveBackend_ExplicitAlwaysWins(t *testing.T) {
	tests := []struct {
		instanceMode string
		explicit     string
		want         string
	}{
		{"single", "redis", "redis"},
		{"ha", "memory", "memory"},
		{"single", "memory", "memory"},
		{"ha", "redis", "redis"},
	}
	for _, tt := range tests {
		cfg := &Config{InstanceMode: tt.instanceMode}
		got := cfg.ResolveBackend(tt.explicit)
		if got != tt.want {
			t.Errorf("InstanceMode=%q explicit=%q: got %q, want %q",
				tt.instanceMode, tt.explicit, got, tt.want)
		}
	}
}

func TestResolveBackend_UnknownModeDefaultsMemory(t *testing.T) {
	cfg := &Config{InstanceMode: "unknown-future-value"}
	got := cfg.ResolveBackend("")
	if got != "memory" {
		t.Errorf("unknown instance mode: got %q, want %q", got, "memory")
	}
}
