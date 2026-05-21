package jobs

import (
	"context"
	"testing"
	"time"

	emailservices "github.com/jcsoftdev/pulzifi-back/modules/email/domain/services"
)

// stubEmailProvider records every Send call without doing real network IO.
type stubEmailProvider struct {
	sent int
	err  error
}

func (s *stubEmailProvider) Send(_ context.Context, _, _, _ string) error {
	s.sent++
	return s.err
}

var _ emailservices.EmailProvider = (*stubEmailProvider)(nil)

func TestNewTrialExpirer_DefaultInterval(t *testing.T) {
	e := NewTrialExpirer(nil, &stubEmailProvider{}, "https://acme.test", 0)
	if e.interval != time.Hour {
		t.Errorf("expected default interval 1h when 0 passed, got %v", e.interval)
	}
}

func TestNewTrialExpirer_CustomInterval(t *testing.T) {
	e := NewTrialExpirer(nil, &stubEmailProvider{}, "https://acme.test", 30*time.Minute)
	if e.interval != 30*time.Minute {
		t.Errorf("expected 30m interval, got %v", e.interval)
	}
}

// TestTrialExpirer_SendExpiredEmail covers the email-dispatch helper without
// touching the database. It guards the FrontendURL fallback and template
// invocation.
func TestTrialExpirer_SendExpiredEmail(t *testing.T) {
	prov := &stubEmailProvider{}
	e := NewTrialExpirer(nil, prov, "https://acme.test", 0)

	e.sendExpiredEmail(context.Background(), expiredMember{
		Email:        "alice@example.com",
		FirstName:    "Alice",
		OrgSubdomain: "acme",
	})

	if prov.sent != 1 {
		t.Errorf("expected 1 email sent, got %d", prov.sent)
	}
}
