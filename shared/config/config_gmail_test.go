package config

import (
	"os"
	"testing"
)

func TestLoad_GmailIntegrationEnabled_ParsesBool(t *testing.T) {
	os.Setenv("GMAIL_INTEGRATION_ENABLED", "true")
	os.Setenv("DB_HOST", "h")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "n")
	os.Setenv("DB_USER", "u")
	os.Setenv("DB_PASSWORD", "p")
	os.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")
	os.Setenv("EXTRACTOR_URL", "http://extractor:3000")
	t.Cleanup(func() {
		os.Unsetenv("GMAIL_INTEGRATION_ENABLED")
		os.Unsetenv("DB_HOST"); os.Unsetenv("DB_PORT"); os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_USER"); os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("CORS_ALLOWED_ORIGINS"); os.Unsetenv("EXTRACTOR_URL")
	})

	cfg := Load()
	if !cfg.GmailIntegrationEnabled {
		t.Error("GmailIntegrationEnabled should be true when GMAIL_INTEGRATION_ENABLED=true")
	}
}

func TestLoad_MicrosoftCredentials_Loaded(t *testing.T) {
	os.Setenv("MICROSOFT_CLIENT_ID", "ms-id")
	os.Setenv("MICROSOFT_CLIENT_SECRET", "ms-secret")
	os.Setenv("DB_HOST", "h"); os.Setenv("DB_PORT", "5432"); os.Setenv("DB_NAME", "n")
	os.Setenv("DB_USER", "u"); os.Setenv("DB_PASSWORD", "p")
	os.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")
	os.Setenv("EXTRACTOR_URL", "http://extractor:3000")
	t.Cleanup(func() {
		os.Unsetenv("MICROSOFT_CLIENT_ID"); os.Unsetenv("MICROSOFT_CLIENT_SECRET")
		os.Unsetenv("DB_HOST"); os.Unsetenv("DB_PORT"); os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_USER"); os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("CORS_ALLOWED_ORIGINS"); os.Unsetenv("EXTRACTOR_URL")
	})

	cfg := Load()
	if cfg.MicrosoftClientID != "ms-id" {
		t.Errorf("MicrosoftClientID = %q, want ms-id", cfg.MicrosoftClientID)
	}
	if cfg.MicrosoftClientSecret != "ms-secret" {
		t.Errorf("MicrosoftClientSecret = %q, want ms-secret", cfg.MicrosoftClientSecret)
	}
}
