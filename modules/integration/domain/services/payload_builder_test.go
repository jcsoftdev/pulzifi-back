package services

import (
	"encoding/json"
	"testing"

	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
)

func TestPayloadBuilder_Build(t *testing.T) {
	builder := NewPayloadBuilder("pulzifi.com")

	tests := []struct {
		name             string
		event            eventbus.DomainEvent
		wantSeverity     string
		wantTitle        string
		wantPageURL      string
		wantBody         string
		wantDashboardURL string
		wantErr          bool
	}{
		{
			name: "change detected falls back to page url when only url present",
			event: eventbus.DomainEvent{
				Type: eventbus.TopicChangeDetected,
				Data: mustMarshal(t, map[string]any{
					"page_url": "https://example.com",
				}),
			},
			wantSeverity: "warning",
			wantTitle:    "Change detected on https://example.com",
			wantPageURL:  "https://example.com",
			wantBody:     "A change was detected on https://example.com",
		},
		{
			name: "change detected enriches title, summary and dashboard deep link",
			event: eventbus.DomainEvent{
				Type: eventbus.TopicChangeDetected,
				Data: mustMarshal(t, map[string]any{
					"page_url":     "https://example.com/pricing",
					"page_title":   "Pricing Page",
					"change_type":  "content",
					"diff_summary": "Pro plan price rose from $29 to $39",
					"tenant":       "acme",
					"workspace_id": "ws-1",
					"page_id":      "pg-1",
				}),
			},
			wantSeverity:     "warning",
			wantTitle:        "Change detected on Pricing Page",
			wantPageURL:      "https://example.com/pricing",
			wantBody:         "Pro plan price rose from $29 to $39",
			wantDashboardURL: "https://acme.pulzifi.com/workspaces/ws-1/pages/pg-1/changes",
		},
		{
			name: "change detected omits deep link when identifiers missing",
			event: eventbus.DomainEvent{
				Type: eventbus.TopicChangeDetected,
				Data: mustMarshal(t, map[string]any{
					"page_url":     "https://example.com",
					"page_title":   "Home",
					"diff_summary": "Header changed",
					// no tenant/workspace_id/page_id
				}),
			},
			wantSeverity:     "warning",
			wantTitle:        "Change detected on Home",
			wantPageURL:      "https://example.com",
			wantBody:         "Header changed",
			wantDashboardURL: "",
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
			if tt.wantBody != "" && payload.Body != tt.wantBody {
				t.Errorf("Body: want %q, got %q", tt.wantBody, payload.Body)
			}
			if payload.DashboardURL != tt.wantDashboardURL {
				t.Errorf("DashboardURL: want %q, got %q", tt.wantDashboardURL, payload.DashboardURL)
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
