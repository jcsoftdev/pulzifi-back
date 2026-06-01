package config

import (
	"os"
	"testing"
	"time"
)

// requiredEnv returns the minimal env map needed to satisfy Load() without fataling.
func requiredEnv() map[string]string {
	return map[string]string{
		"DB_HOST":              "localhost",
		"DB_PORT":              "5432",
		"DB_NAME":              "test",
		"DB_USER":              "test",
		"DB_PASSWORD":          "test",
		"CORS_ALLOWED_ORIGINS": "http://localhost",
		"EXTRACTOR_URL":        "http://localhost:3005",
	}
}

// setEnvMap sets all keys from m for the duration of the test.
func setEnvMap(t *testing.T, m map[string]string) {
	t.Helper()
	for k, v := range m {
		t.Setenv(k, v)
	}
}

// TestLoad_JWTSecret_ProductionMinLength verifies that Load() fatals when
// JWT_SECRET is shorter than 32 bytes in production.
func TestLoad_JWTSecret_ProductionMinLength(t *testing.T) {
	env := requiredEnv()
	env["ENVIRONMENT"] = "production"
	env["EXTRACTOR_API_KEY"] = "some-key" // satisfy prod extractor check
	setEnvMap(t, env)

	tests := []struct {
		name      string
		jwtSecret string
		wantFatal bool
	}{
		{
			name:      "short secret causes fatal in production",
			jwtSecret: "tooshort",
			wantFatal: true,
		},
		{
			name:      "exactly 32 bytes is allowed",
			jwtSecret: "12345678901234567890123456789012", // 32 chars
			wantFatal: false,
		},
		{
			name:      "more than 32 bytes is allowed",
			jwtSecret: "a-very-long-secret-key-that-exceeds-the-minimum-requirement",
			wantFatal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("JWT_SECRET", tt.jwtSecret)

			if tt.wantFatal {
				assertFatals(t, func() { Load() })
			} else {
				assertNoFatal(t, func() { Load() })
			}
		})
	}
}

// TestLoad_ExtractorAPIKey_RequiredInProduction verifies that Load() fatals
// when EXTRACTOR_API_KEY is unset (or empty) in production.
func TestLoad_ExtractorAPIKey_RequiredInProduction(t *testing.T) {
	env := requiredEnv()
	env["ENVIRONMENT"] = "production"
	env["JWT_SECRET"] = "this-is-a-secure-secret-key-32ch"
	setEnvMap(t, env)

	tests := []struct {
		name      string
		apiKey    string
		unset     bool
		wantFatal bool
	}{
		{
			name:      "missing key causes fatal in production",
			unset:     true,
			wantFatal: true,
		},
		{
			name:      "empty string causes fatal in production",
			apiKey:    "",
			wantFatal: true,
		},
		{
			name:      "non-empty key is accepted",
			apiKey:    "some-api-key",
			wantFatal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.unset {
				os.Unsetenv("EXTRACTOR_API_KEY")
			} else {
				t.Setenv("EXTRACTOR_API_KEY", tt.apiKey)
			}

			if tt.wantFatal {
				assertFatals(t, func() { Load() })
			} else {
				assertNoFatal(t, func() { Load() })
			}
		})
	}
}

// TestLoad_ExtractorAPIKey_LenientInDev verifies that EXTRACTOR_API_KEY is
// optional in development (no fatal, just a warning).
func TestLoad_ExtractorAPIKey_LenientInDev(t *testing.T) {
	env := requiredEnv()
	env["ENVIRONMENT"] = "development"
	env["JWT_SECRET"] = "dev-secret"
	setEnvMap(t, env)
	os.Unsetenv("EXTRACTOR_API_KEY")

	assertNoFatal(t, func() { Load() })
}

// TestLoad_ResetTokenTTL_Defaults verifies that RESET_TOKEN_TTL defaults to 1h.
func TestLoad_ResetTokenTTL_Defaults(t *testing.T) {
	env := requiredEnv()
	env["JWT_SECRET"] = "dev-secret"
	setEnvMap(t, env)
	os.Unsetenv("RESET_TOKEN_TTL")

	cfg := Load()
	if cfg.ResetTokenTTL.Hours() != 1 {
		t.Errorf("ResetTokenTTL default = %v, want 1h", cfg.ResetTokenTTL)
	}
}

// TestLoad_ResetTokenTTL_FromEnv verifies that RESET_TOKEN_TTL can be overridden.
func TestLoad_ResetTokenTTL_FromEnv(t *testing.T) {
	env := requiredEnv()
	env["JWT_SECRET"] = "dev-secret"
	env["RESET_TOKEN_TTL"] = "30m"
	setEnvMap(t, env)

	cfg := Load()
	if cfg.ResetTokenTTL.Minutes() != 30 {
		t.Errorf("ResetTokenTTL = %v, want 30m", cfg.ResetTokenTTL)
	}
}

// assertFatals temporarily replaces fatalFunc with one that panics,
// then runs f and asserts a panic (fatal) was triggered.
func assertFatals(t *testing.T, f func()) {
	t.Helper()
	orig := fatalFunc
	fatalFunc = func(msg string) { panic("fatal: " + msg) }
	defer func() {
		fatalFunc = orig
		if r := recover(); r == nil {
			t.Error("expected fatal, but Load() returned normally")
		}
	}()
	f()
}

// assertNoFatal temporarily replaces fatalFunc with one that panics on
// unexpected fatals, then runs f and asserts no panic occurred.
func assertNoFatal(t *testing.T, f func()) {
	t.Helper()
	orig := fatalFunc
	fatalFunc = func(msg string) { panic("unexpected fatal: " + msg) }
	defer func() {
		fatalFunc = orig
		if r := recover(); r != nil {
			t.Errorf("unexpected fatal/panic: %v", r)
		}
	}()
	f()
}

// TestLoad_TrialDefaults verifies that trial-related configuration loads with
// the documented defaults when no env vars are set.
func TestLoad_TrialDefaults(t *testing.T) {
	// Ensure required env vars exist (Load() fatals otherwise) but trial vars are unset.
	required := map[string]string{
		"DB_HOST":             "localhost",
		"DB_PORT":             "5432",
		"DB_NAME":             "test",
		"DB_USER":             "test",
		"DB_PASSWORD":         "test",
		"CORS_ALLOWED_ORIGINS": "http://localhost",
		"EXTRACTOR_URL":       "http://localhost:3005",
		"JWT_SECRET":          "test-secret",
	}
	prevReq := snapshotEnv(t, required)
	for k, v := range required {
		t.Setenv(k, v)
	}
	for _, k := range []string{"TRIAL_DAYS", "TRIAL_CHECKS_PER_MONTH", "TRIAL_EXPIRY_CRON"} {
		os.Unsetenv(k)
	}
	defer restoreEnv(prevReq)

	cfg := Load()

	if cfg.TrialDays != 14 {
		t.Errorf("TrialDays default = %d, want 14", cfg.TrialDays)
	}
	if cfg.TrialChecksPerMonth != 500 {
		t.Errorf("TrialChecksPerMonth default = %d, want 500", cfg.TrialChecksPerMonth)
	}
	if cfg.TrialExpiryCron != "0 0 * * *" {
		t.Errorf("TrialExpiryCron default = %q, want %q", cfg.TrialExpiryCron, "0 0 * * *")
	}
}

// TestLoad_TrialFromEnv verifies overriding the trial defaults via env vars.
func TestLoad_TrialFromEnv(t *testing.T) {
	required := map[string]string{
		"DB_HOST":             "localhost",
		"DB_PORT":             "5432",
		"DB_NAME":             "test",
		"DB_USER":             "test",
		"DB_PASSWORD":         "test",
		"CORS_ALLOWED_ORIGINS": "http://localhost",
		"EXTRACTOR_URL":       "http://localhost:3005",
		"JWT_SECRET":          "test-secret",
	}
	prevReq := snapshotEnv(t, required)
	for k, v := range required {
		t.Setenv(k, v)
	}
	t.Setenv("TRIAL_DAYS", "7")
	t.Setenv("TRIAL_CHECKS_PER_MONTH", "100")
	t.Setenv("TRIAL_EXPIRY_CRON", "30 1 * * *")
	defer restoreEnv(prevReq)

	cfg := Load()

	if cfg.TrialDays != 7 {
		t.Errorf("TrialDays = %d, want 7", cfg.TrialDays)
	}
	if cfg.TrialChecksPerMonth != 100 {
		t.Errorf("TrialChecksPerMonth = %d, want 100", cfg.TrialChecksPerMonth)
	}
	if cfg.TrialExpiryCron != "30 1 * * *" {
		t.Errorf("TrialExpiryCron = %q, want %q", cfg.TrialExpiryCron, "30 1 * * *")
	}
}

// TestLoad_SnapshotStorageDefaults verifies SNAPSHOT_BUCKET_PRIVATE and
// SNAPSHOT_PRESIGN_TTL default to false and 15m respectively.
func TestLoad_SnapshotStorageDefaults(t *testing.T) {
	env := requiredEnv()
	env["JWT_SECRET"] = "dev-secret"
	setEnvMap(t, env)
	os.Unsetenv("SNAPSHOT_BUCKET_PRIVATE")
	os.Unsetenv("SNAPSHOT_PRESIGN_TTL")

	cfg := Load()

	if cfg.SnapshotBucketPrivate {
		t.Error("SnapshotBucketPrivate default = true, want false")
	}
	if cfg.SnapshotPresignTTL != 15*time.Minute {
		t.Errorf("SnapshotPresignTTL default = %v, want 15m", cfg.SnapshotPresignTTL)
	}
}

// TestLoad_SnapshotStorageFromEnv verifies the env vars can be overridden.
func TestLoad_SnapshotStorageFromEnv(t *testing.T) {
	env := requiredEnv()
	env["JWT_SECRET"] = "dev-secret"
	env["SNAPSHOT_BUCKET_PRIVATE"] = "true"
	env["SNAPSHOT_PRESIGN_TTL"] = "30m"
	setEnvMap(t, env)

	cfg := Load()

	if !cfg.SnapshotBucketPrivate {
		t.Error("SnapshotBucketPrivate = false, want true")
	}
	if cfg.SnapshotPresignTTL != 30*time.Minute {
		t.Errorf("SnapshotPresignTTL = %v, want 30m", cfg.SnapshotPresignTTL)
	}
}

// TestLoad_EventBusDefaults verifies that EVENT_BUS_PROVIDER defaults to "memory"
// and all OUTBOX_* vars have correct defaults.
func TestLoad_EventBusDefaults(t *testing.T) {
	env := requiredEnv()
	env["JWT_SECRET"] = "dev-secret"
	setEnvMap(t, env)
	for _, k := range []string{
		"EVENT_BUS_PROVIDER", "OUTBOX_POLL_INTERVAL", "OUTBOX_BATCH_SIZE",
		"OUTBOX_MAX_ATTEMPTS", "OUTBOX_STALE_THRESHOLD",
	} {
		os.Unsetenv(k)
	}

	cfg := Load()

	if cfg.EventBusProvider != "memory" {
		t.Errorf("EventBusProvider default = %q, want %q", cfg.EventBusProvider, "memory")
	}
	if cfg.OutboxPollInterval != 1*time.Second {
		t.Errorf("OutboxPollInterval default = %v, want 1s", cfg.OutboxPollInterval)
	}
	if cfg.OutboxBatchSize != 100 {
		t.Errorf("OutboxBatchSize default = %d, want 100", cfg.OutboxBatchSize)
	}
	if cfg.OutboxMaxAttempts != 8 {
		t.Errorf("OutboxMaxAttempts default = %d, want 8", cfg.OutboxMaxAttempts)
	}
	if cfg.OutboxStaleThreshold != 30*time.Second {
		t.Errorf("OutboxStaleThreshold default = %v, want 30s", cfg.OutboxStaleThreshold)
	}
}

// TestLoad_EventBusFromEnv verifies that outbox config can be overridden via env vars.
func TestLoad_EventBusFromEnv(t *testing.T) {
	env := requiredEnv()
	env["JWT_SECRET"] = "dev-secret"
	env["EVENT_BUS_PROVIDER"] = "outbox"
	env["OUTBOX_POLL_INTERVAL"] = "500ms"
	env["OUTBOX_BATCH_SIZE"] = "50"
	env["OUTBOX_MAX_ATTEMPTS"] = "5"
	env["OUTBOX_STALE_THRESHOLD"] = "1m"
	setEnvMap(t, env)

	cfg := Load()

	if cfg.EventBusProvider != "outbox" {
		t.Errorf("EventBusProvider = %q, want %q", cfg.EventBusProvider, "outbox")
	}
	if cfg.OutboxPollInterval != 500*time.Millisecond {
		t.Errorf("OutboxPollInterval = %v, want 500ms", cfg.OutboxPollInterval)
	}
	if cfg.OutboxBatchSize != 50 {
		t.Errorf("OutboxBatchSize = %d, want 50", cfg.OutboxBatchSize)
	}
	if cfg.OutboxMaxAttempts != 5 {
		t.Errorf("OutboxMaxAttempts = %d, want 5", cfg.OutboxMaxAttempts)
	}
	if cfg.OutboxStaleThreshold != 1*time.Minute {
		t.Errorf("OutboxStaleThreshold = %v, want 1m", cfg.OutboxStaleThreshold)
	}
}

// snapshotEnv captures current values of keys so they can be restored later.
func snapshotEnv(_ *testing.T, keys map[string]string) map[string]string {
	prev := make(map[string]string, len(keys))
	for k := range keys {
		prev[k] = os.Getenv(k)
	}
	return prev
}

func restoreEnv(prev map[string]string) {
	for k, v := range prev {
		if v == "" {
			os.Unsetenv(k)
			continue
		}
		os.Setenv(k, v)
	}
}
