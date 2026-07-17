package emailprovider_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	emailprovider "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/providers/email"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
)

// stubSender records every Send call made against it.
type stubSender struct {
	calls []sendCall
	err   error // if non-nil, returned by Send
}

type sendCall struct {
	to      []string
	subject string
	body    string
}

func (s *stubSender) Send(_ context.Context, to []string, subject, htmlBody string) error {
	s.calls = append(s.calls, sendCall{to: to, subject: subject, body: htmlBody})
	return s.err
}

func makePayload() *entities.NotificationPayload {
	return &entities.NotificationPayload{
		Title:   "Page changed",
		Body:    "Something is different",
		PageURL: "https://example.com/page",
	}
}

func makeIntegration() *entities.Integration {
	return &entities.Integration{ServiceType: "email"}
}

func makeDest(target map[string]any) *entities.Destination {
	return &entities.Destination{Target: target}
}

func TestEmailClient(t *testing.T) {
	tests := []struct {
		name        string
		target      map[string]any
		senderErr   error
		wantErr     string
		wantSendTo  []string
		wantCode    int
		wantNoSend  bool
	}{
		{
			name:       "valid emails",
			target:     map[string]any{"emails": []any{"a@b.com", "c@d.com"}},
			wantSendTo: []string{"a@b.com", "c@d.com"},
			wantCode:   200,
		},
		{
			name:       "mixed valid and invalid",
			target:     map[string]any{"emails": []any{"good@x.com", "noatsign", 42, "another@y.com"}},
			wantSendTo: []string{"good@x.com", "another@y.com"},
			wantCode:   200,
		},
		{
			name:       "deduplicates exact duplicate recipients",
			target:     map[string]any{"emails": []any{"a@b.com", "a@b.com", "c@d.com"}},
			wantSendTo: []string{"a@b.com", "c@d.com"},
			wantCode:   200,
		},
		{
			name:       "deduplicates case-insensitively and trims, keeps first-seen",
			target:     map[string]any{"emails": []any{"  Member@X.com ", "member@x.com", "manual@x.com"}},
			wantSendTo: []string{"Member@X.com", "manual@x.com"},
			wantCode:   200,
		},
		{
			name:       "empty list",
			target:     map[string]any{"emails": []any{}},
			wantErr:    "email: no recipients",
			wantNoSend: true,
		},
		{
			name:       "nil target / missing key",
			target:     map[string]any{},
			wantErr:    "email: no recipients",
			wantNoSend: true,
		},
		{
			name:      "sender returns error",
			target:    map[string]any{"emails": []any{"ok@test.com"}},
			senderErr: errors.New("smtp down"),
			wantErr:   "smtp down",
		},
		{
			name:       "body contains title, body text and pageURL",
			target:     map[string]any{"emails": []any{"x@y.com"}},
			wantCode:   200,
			wantSendTo: []string{"x@y.com"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stub := &stubSender{err: tc.senderErr}
			client := emailprovider.New(stub)

			p := makePayload()
			result, err := client.Send(context.Background(), makeIntegration(), makeDest(tc.target), p)

			// error assertions
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error %q, got %q", tc.wantErr, err.Error())
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Send not called assertions
			if tc.wantNoSend && len(stub.calls) > 0 {
				t.Fatalf("expected Send not to be called, but got %d calls", len(stub.calls))
			}

			// result code assertions
			if tc.wantCode != 0 {
				if result == nil {
					t.Fatal("expected non-nil DeliveryResult")
				}
				if result.Code != tc.wantCode {
					t.Fatalf("expected Code=%d, got %d", tc.wantCode, result.Code)
				}
			}

			// recipient assertions
			if len(tc.wantSendTo) > 0 {
				if len(stub.calls) == 0 {
					t.Fatal("expected Send to be called")
				}
				got := stub.calls[0].to
				if len(got) != len(tc.wantSendTo) {
					t.Fatalf("expected recipients %v, got %v", tc.wantSendTo, got)
				}
				for i, addr := range tc.wantSendTo {
					if got[i] != addr {
						t.Fatalf("recipient[%d]: expected %q, got %q", i, addr, got[i])
					}
				}
			}

			// body content assertion for the relevant test case
			if tc.name == "body contains title, body text and pageURL" {
				if len(stub.calls) == 0 {
					t.Fatal("expected Send to be called")
				}
				call := stub.calls[0]
				if !strings.Contains(call.body, p.Title) {
					t.Errorf("body missing title %q", p.Title)
				}
				if !strings.Contains(call.body, p.Body) {
					t.Errorf("body missing body text %q", p.Body)
				}
				if !strings.Contains(call.body, p.PageURL) {
					t.Errorf("body missing pageURL %q", p.PageURL)
				}
				if call.subject != p.Title {
					t.Errorf("subject: expected %q, got %q", p.Title, call.subject)
				}
			}
		})
	}
}

func TestEmailClient_RichChangeBody(t *testing.T) {
	stub := &stubSender{}
	client := emailprovider.New(stub)

	p := &entities.NotificationPayload{
		Title:        "Change detected on Pricing Page",
		Body:         "Pro plan price rose from $29 to $39",
		PageURL:      "https://example.com/pricing",
		PageTitle:    "Pricing Page",
		ChangeType:   "content",
		DiffImageURL: "https://storage.pulzifi.com/diffs/abc.png",
		DashboardURL: "https://acme.pulzifi.com/workspaces/ws-1/pages/pg-1/changes",
		ChangedAt:    "2026-07-17T19:18:00Z",
	}
	dest := makeDest(map[string]any{"emails": []any{"x@y.com"}})

	if _, err := client.Send(context.Background(), makeIntegration(), dest, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stub.calls) == 0 {
		t.Fatal("expected Send to be called")
	}
	body := stub.calls[0].body

	wantSubstrings := map[string]string{
		"summary":       p.Body,
		"page name":     p.PageTitle,
		"diff image":    p.DiffImageURL,
		"dashboard CTA": p.DashboardURL,
		"CTA label":     "View changes",
		"badge":         "content",
	}
	for label, want := range wantSubstrings {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %s (%q)", label, want)
		}
	}
	// When a dashboard link exists, the CTA must NOT use the raw-page label.
	if strings.Contains(body, "View page") {
		t.Errorf("expected dashboard CTA, but body used raw-page label")
	}
}

func TestEmailClient_SanitizesControlCharsAndSubject(t *testing.T) {
	stub := &stubSender{}
	client := emailprovider.New(stub)

	// Title carries a CRLF (header-injection vector) + a control char; summary
	// carries a NUL. Accented text must be preserved.
	p := &entities.NotificationPayload{
		Title:     "Change detected on Página\r\nBcc: evil@x.com",
		Body:      "Precio\x00 subió de $29 a $39",
		PageURL:   "https://example.com",
		PageTitle: "Página\x07",
	}
	dest := makeDest(map[string]any{"emails": []any{"x@y.com"}})

	if _, err := client.Send(context.Background(), makeIntegration(), dest, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	call := stub.calls[0]

	if strings.ContainsAny(call.subject, "\r\n") {
		t.Errorf("subject still contains CR/LF (header injection): %q", call.subject)
	}
	if strings.ContainsRune(call.body, '\x00') || strings.ContainsRune(call.body, '\x07') {
		t.Errorf("body still contains control characters")
	}
	// Legitimate accented text survives.
	if !strings.Contains(call.body, "Precio") || !strings.Contains(call.subject, "Página") {
		t.Errorf("expected accented text preserved; subject=%q", call.subject)
	}
}

func TestEmailClient_NoDeadLinkWhenNoURLs(t *testing.T) {
	// alert.created payloads carry only Title+Body (no page/dashboard URL). The
	// CTA must be omitted entirely rather than render a dead <a href="">.
	stub := &stubSender{}
	client := emailprovider.New(stub)

	p := &entities.NotificationPayload{
		Title: "Quota threshold reached",
		Body:  "You have used 90% of your monthly checks.",
	}
	dest := makeDest(map[string]any{"emails": []any{"x@y.com"}})

	if _, err := client.Send(context.Background(), makeIntegration(), dest, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body := stub.calls[0].body
	if strings.Contains(body, `href=""`) {
		t.Errorf("rendered a dead empty-href link:\n%s", body)
	}
	if strings.Contains(body, "View page") || strings.Contains(body, "View changes") {
		t.Errorf("expected no CTA button when no URL is present")
	}
	if !strings.Contains(body, p.Title) || !strings.Contains(body, p.Body) {
		t.Errorf("expected title and body still rendered")
	}
}

func TestEmailClient_FallsBackToPageURLWithoutDashboard(t *testing.T) {
	stub := &stubSender{}
	client := emailprovider.New(stub)

	p := &entities.NotificationPayload{
		Title:   "Change detected on https://example.com",
		Body:    "A change was detected on https://example.com",
		PageURL: "https://example.com",
	}
	dest := makeDest(map[string]any{"emails": []any{"x@y.com"}})

	if _, err := client.Send(context.Background(), makeIntegration(), dest, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body := stub.calls[0].body
	if !strings.Contains(body, "View page") {
		t.Errorf("expected 'View page' CTA fallback, body: %s", body)
	}
	if !strings.Contains(body, p.PageURL) {
		t.Errorf("expected page URL in CTA fallback")
	}
}
