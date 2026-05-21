package entities

import (
	"testing"

	"github.com/google/uuid"
)

func TestScopeTypeConstants(t *testing.T) {
	// Verify scope type constants have expected string values
	tests := []struct {
		scope ScopeType
		want  string
	}{
		{ScopeOrg, "org"},
		{ScopeWorkspace, "workspace"},
		{ScopePage, "page"},
	}
	for _, tt := range tests {
		t.Run(string(tt.scope), func(t *testing.T) {
			if string(tt.scope) != tt.want {
				t.Errorf("ScopeType %q: want %q", tt.scope, tt.want)
			}
		})
	}
}

func TestDestination_Fields(t *testing.T) {
	integrationID := uuid.New()
	scopeID := uuid.New()

	dest := &Destination{
		ID:            uuid.New(),
		IntegrationID: &integrationID,
		ServiceType:   "slack",
		ScopeType:     ScopeWorkspace,
		ScopeID:       scopeID,
		Target:        map[string]any{"channel": "#alerts"},
		Events:        []string{"change_detected"},
		Enabled:       true,
	}

	if dest.ID == uuid.Nil {
		t.Error("ID should not be nil UUID")
	}
	if dest.IntegrationID == nil || *dest.IntegrationID != integrationID {
		t.Errorf("IntegrationID: want %v, got %v", integrationID, dest.IntegrationID)
	}
	if dest.ServiceType != "slack" {
		t.Errorf("ServiceType: want %q, got %q", "slack", dest.ServiceType)
	}
	if dest.ScopeType != ScopeWorkspace {
		t.Errorf("ScopeType: want %q, got %q", ScopeWorkspace, dest.ScopeType)
	}
	if !dest.Enabled {
		t.Error("Enabled should be true")
	}
	if len(dest.Events) != 1 || dest.Events[0] != "change_detected" {
		t.Errorf("Events: want [change_detected], got %v", dest.Events)
	}
	channel, ok := dest.Target["channel"]
	if !ok || channel != "#alerts" {
		t.Errorf("Target[channel]: want %q, got %v", "#alerts", channel)
	}
}

func TestIntegrationStatusConstants(t *testing.T) {
	tests := []struct {
		status IntegrationStatus
		want   string
	}{
		{IntegrationActive, "active"},
		{IntegrationDisconnected, "disconnected"},
		{IntegrationExpired, "expired"},
	}
	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("IntegrationStatus %q: want %q", tt.status, tt.want)
			}
		})
	}
}
