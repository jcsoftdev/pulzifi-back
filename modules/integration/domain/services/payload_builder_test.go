package services

import (
	"encoding/json"
	"testing"

	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
)

func TestPayloadBuilder_Build(t *testing.T) {
	builder := NewPayloadBuilder()

	tests := []struct {
		name          string
		event         eventbus.DomainEvent
		wantSeverity  string
		wantTitle     string
		wantPageURL   string
		wantErr       bool
	}{
		{
			name: "change detected event maps to warning severity",
			event: eventbus.DomainEvent{
				Type: eventbus.TopicChangeDetected,
				Data: mustMarshal(t, map[string]any{
					"page_url": "https://example.com",
				}),
			},
			wantSeverity: "warning",
			wantTitle:    "Page change detected",
			wantPageURL:  "https://example.com",
		},
		{
			name: "alert created event maps to critical severity",
			event: eventbus.DomainEvent{
				Type: eventbus.TopicAlertCreated,
				Data: mustMarshal(t, map[string]any{
					"title":   "Alert Title",
					"message": "Alert body",
				}),
			},
			wantSeverity: "critical",
			wantTitle:    "Alert Title",
		},
		{
			name: "unknown event type maps to info severity",
			event: eventbus.DomainEvent{
				Type: "unknown.event",
				Data: mustMarshal(t, map[string]any{}),
			},
			wantSeverity: "info",
			wantTitle:    "unknown.event",
		},
		{
			name: "nil data — no panic",
			event: eventbus.DomainEvent{
				Type: "unknown.event",
				Data: nil,
			},
			wantSeverity: "info",
		},
		{
			name: "invalid JSON data returns error",
			event: eventbus.DomainEvent{
				Type: eventbus.TopicChangeDetected,
				Data: []byte("not-valid-json"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			payload, err := builder.Build(tt.event)

			// Assert
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if payload == nil {
				t.Fatal("expected non-nil payload")
			}
			if payload.Severity != tt.wantSeverity {
				t.Errorf("Severity: want %q, got %q", tt.wantSeverity, payload.Severity)
			}
			if tt.wantTitle != "" && payload.Title != tt.wantTitle {
				t.Errorf("Title: want %q, got %q", tt.wantTitle, payload.Title)
			}
			if tt.wantPageURL != "" && payload.PageURL != tt.wantPageURL {
				t.Errorf("PageURL: want %q, got %q", tt.wantPageURL, payload.PageURL)
			}
		})
	}
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("mustMarshal: %v", err)
	}
	return b
}
