package config

import (
	"os"
	"testing"
)

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
