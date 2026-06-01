package config

import (
	"net"
	"os"
	"testing"
	"time"
)

// ── S0: InstanceMode, backend fields, ResolveBackend, CIDR parsing ──────────

func TestLoad_InstanceMode_Default(t *testing.T) {
	env := requiredEnv()
	env["JWT_SECRET"] = "dev-secret"
	setEnvMap(t, env)
	os.Unsetenv("INSTANCE_MODE")

	cfg := Load()
	if cfg.InstanceMode != "single" {
		t.Errorf("InstanceMode default = %q, want %q", cfg.InstanceMode, "single")
	}
}

func TestLoad_InstanceMode_HA(t *testing.T) {
	env := requiredEnv()
	env["JWT_SECRET"] = "dev-secret"
	env["INSTANCE_MODE"] = "ha"
	setEnvMap(t, env)

	cfg := Load()
	if cfg.InstanceMode != "ha" {
		t.Errorf("InstanceMode = %q, want %q", cfg.InstanceMode, "ha")
	}
}

func TestLoad_BackendDefaults(t *testing.T) {
	env := requiredEnv()
	env["JWT_SECRET"] = "dev-secret"
	setEnvMap(t, env)
	for _, k := range []string{"NONCE_BACKEND", "BROKER_BACKEND", "RATELIMIT_BACKEND", "SCHEDULER_MODE"} {
		os.Unsetenv(k)
	}

	cfg := Load()
	if cfg.NonceBackend != "" {
		t.Errorf("NonceBackend default = %q, want empty", cfg.NonceBackend)
	}
	if cfg.BrokerBackend != "" {
		t.Errorf("BrokerBackend default = %q, want empty", cfg.BrokerBackend)
	}
	if cfg.RateLimitBackend != "" {
		t.Errorf("RateLimitBackend default = %q, want empty", cfg.RateLimitBackend)
	}
	if cfg.SchedulerMode != "poll" {
		t.Errorf("SchedulerMode default = %q, want %q", cfg.SchedulerMode, "poll")
	}
}

func TestLoad_AuthRateLimitDefaults(t *testing.T) {
	env := requiredEnv()
	env["JWT_SECRET"] = "dev-secret"
	setEnvMap(t, env)
	os.Unsetenv("AUTH_RATE_LIMIT_REQUESTS")
	os.Unsetenv("AUTH_RATE_LIMIT_WINDOW")

	cfg := Load()
	if cfg.AuthRateLimitRequests != 5 {
		t.Errorf("AuthRateLimitRequests default = %d, want 5", cfg.AuthRateLimitRequests)
	}
	if cfg.AuthRateLimitWindow != 15*time.Minute {
		t.Errorf("AuthRateLimitWindow default = %v, want 15m", cfg.AuthRateLimitWindow)
	}
}

func TestLoad_AuthRateLimitFromEnv(t *testing.T) {
	env := requiredEnv()
	env["JWT_SECRET"] = "dev-secret"
	env["AUTH_RATE_LIMIT_REQUESTS"] = "10"
	env["AUTH_RATE_LIMIT_WINDOW"] = "30m"
	setEnvMap(t, env)

	cfg := Load()
	if cfg.AuthRateLimitRequests != 10 {
		t.Errorf("AuthRateLimitRequests = %d, want 10", cfg.AuthRateLimitRequests)
	}
	if cfg.AuthRateLimitWindow != 30*time.Minute {
		t.Errorf("AuthRateLimitWindow = %v, want 30m", cfg.AuthRateLimitWindow)
	}
}

// ── ResolveBackend ───────────────────────────────────────────────────────────

func TestResolveBackend_ExplicitRedis(t *testing.T) {
	cfg := &Config{InstanceMode: "single"}
	if got := cfg.ResolveBackend("redis"); got != "redis" {
		t.Errorf("explicit redis: got %q, want %q", got, "redis")
	}
}

func TestResolveBackend_ExplicitMemory(t *testing.T) {
	cfg := &Config{InstanceMode: "ha"}
	if got := cfg.ResolveBackend("memory"); got != "memory" {
		t.Errorf("explicit memory with ha: got %q, want %q", got, "memory")
	}
}

func TestResolveBackend_DeriveMatrix(t *testing.T) {
	tests := []struct {
		name         string
		instanceMode string
		explicit     string
		want         string
	}{
		{"single derive empty", "single", "", "memory"},
		{"ha derive empty", "ha", "", "redis"},
		{"ha derive explicit redis", "ha", "redis", "redis"},
		{"single derive explicit redis", "single", "redis", "redis"},
		{"ha explicit memory", "ha", "memory", "memory"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{InstanceMode: tt.instanceMode}
			got := cfg.ResolveBackend(tt.explicit)
			if got != tt.want {
				t.Errorf("ResolveBackend(%q) with InstanceMode=%q: got %q, want %q",
					tt.explicit, tt.instanceMode, got, tt.want)
			}
		})
	}
}

// ── TrustedProxyCIDRs ────────────────────────────────────────────────────────

func TestLoad_TrustedProxyCIDRs_Empty(t *testing.T) {
	env := requiredEnv()
	env["JWT_SECRET"] = "dev-secret"
	setEnvMap(t, env)
	os.Unsetenv("TRUSTED_PROXY_CIDRS")

	cfg := Load()
	if len(cfg.TrustedProxyCIDRs) != 0 {
		t.Errorf("TrustedProxyCIDRs empty env = %v, want nil/empty", cfg.TrustedProxyCIDRs)
	}
}

func TestLoad_TrustedProxyCIDRs_Valid(t *testing.T) {
	env := requiredEnv()
	env["JWT_SECRET"] = "dev-secret"
	env["TRUSTED_PROXY_CIDRS"] = "10.0.0.0/8,172.16.0.0/12"
	setEnvMap(t, env)

	cfg := Load()
	if len(cfg.TrustedProxyCIDRs) != 2 {
		t.Fatalf("TrustedProxyCIDRs len = %d, want 2", len(cfg.TrustedProxyCIDRs))
	}
	// verify the first CIDR contains 10.0.0.1
	testIP := net.ParseIP("10.0.0.1")
	if !cfg.TrustedProxyCIDRs[0].Contains(testIP) {
		t.Errorf("expected 10.0.0.0/8 to contain 10.0.0.1")
	}
}

func TestLoad_TrustedProxyCIDRs_OneGoodOneBad(t *testing.T) {
	env := requiredEnv()
	env["JWT_SECRET"] = "dev-secret"
	env["TRUSTED_PROXY_CIDRS"] = "10.0.0.0/8,not-a-cidr"
	setEnvMap(t, env)

	// Should not fatal; invalid entry is skipped
	assertNoFatal(t, func() {
		cfg := Load()
		// Only the valid one survives
		if len(cfg.TrustedProxyCIDRs) != 1 {
			t.Errorf("TrustedProxyCIDRs len = %d, want 1 (bad entry skipped)", len(cfg.TrustedProxyCIDRs))
		}
	})
}
