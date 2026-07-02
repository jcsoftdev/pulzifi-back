package config

import (
	"os"
	"testing"
)

// TestLoad_EmailProviderDefault verifies EMAIL_PROVIDER defaults to "resend"
// when unset, preserving current production behavior.
func TestLoad_EmailProviderDefault(t *testing.T) {
	env := requiredEnv()
	env["JWT_SECRET"] = "dev-secret"
	setEnvMap(t, env)
	os.Unsetenv("EMAIL_PROVIDER")

	cfg := Load()

	if cfg.EmailProvider != "resend" {
		t.Errorf("EmailProvider default = %q, want %q", cfg.EmailProvider, "resend")
	}
}

// TestLoad_EmailProviderFromEnv verifies EMAIL_PROVIDER can be overridden to
// "log" so local/E2E runs never hit the real Resend API.
func TestLoad_EmailProviderFromEnv(t *testing.T) {
	env := requiredEnv()
	env["JWT_SECRET"] = "dev-secret"
	env["EMAIL_PROVIDER"] = "log"
	setEnvMap(t, env)

	cfg := Load()

	if cfg.EmailProvider != "log" {
		t.Errorf("EmailProvider = %q, want %q", cfg.EmailProvider, "log")
	}
}
