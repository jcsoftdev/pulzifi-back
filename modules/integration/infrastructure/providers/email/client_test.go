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
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
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
