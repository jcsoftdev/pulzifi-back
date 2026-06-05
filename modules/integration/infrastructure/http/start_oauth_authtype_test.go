package http

import (
	"testing"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
)

// Guards the contract that the OAuth start path must reject non-oauth providers.
// "email" is a no-auth adapter and must never be treated as OAuth-capable.
func TestEmailIsNotOAuth(t *testing.T) {
	if services.ProviderAuthType("email") == services.AuthTypeOAuth {
		t.Fatal("email must not be oauth — start_oauth must reject it with 400, not 500")
	}
	if services.ProviderAuthType("slack") != services.AuthTypeOAuth {
		t.Fatal("slack must be oauth")
	}
}
