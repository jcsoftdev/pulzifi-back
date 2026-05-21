package templates

import (
	"strings"
	"testing"
)

func TestApprovalNotification(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		subdomain string
		loginURL  string
		wantSubjectContains string
		wantHTMLContains    []string
	}{
		{
			name:      "generates subject and HTML with user info",
			firstName: "Alice",
			subdomain: "acme",
			loginURL:  "https://acme.pulzifi.com/login",
			wantSubjectContains: "approved",
			wantHTMLContains: []string{
				"Alice",
				"acme",
				"https://acme.pulzifi.com/login",
				"<!DOCTYPE html>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject, html := ApprovalNotification(tt.firstName, tt.subdomain, tt.loginURL)

			if !strings.Contains(strings.ToLower(subject), tt.wantSubjectContains) {
				t.Errorf("subject %q does not contain %q", subject, tt.wantSubjectContains)
			}
			for _, want := range tt.wantHTMLContains {
				if !strings.Contains(html, want) {
					t.Errorf("HTML does not contain %q", want)
				}
			}
		})
	}
}

func TestRejectionNotification(t *testing.T) {
	subject, html := RejectionNotification("Bob")

	if subject == "" {
		t.Error("subject should not be empty")
	}
	if !strings.Contains(html, "Bob") {
		t.Error("HTML should contain the first name")
	}
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("HTML should be valid HTML document")
	}
}

func TestTeamInvite(t *testing.T) {
	subject, html := TeamInvite("Alice", "Acme Corp", "https://acme.pulzifi.com/accept")

	if !strings.Contains(subject, "Acme Corp") {
		t.Errorf("subject %q should contain org name", subject)
	}
	if !strings.Contains(html, "Alice") {
		t.Error("HTML should contain inviter name")
	}
	if !strings.Contains(html, "https://acme.pulzifi.com/accept") {
		t.Error("HTML should contain the invite URL")
	}
}

func TestAlertNotification(t *testing.T) {
	pageURL := "https://example.com"
	changeType := "visual"
	dashboardURL := "https://acme.pulzifi.com/dashboard"

	subject, html := AlertNotification(pageURL, changeType, dashboardURL)

	if subject == "" {
		t.Error("subject should not be empty")
	}
	if !strings.Contains(html, pageURL) {
		t.Errorf("HTML should contain page URL %q", pageURL)
	}
	if !strings.Contains(html, changeType) {
		t.Errorf("HTML should contain change type %q", changeType)
	}
	if !strings.Contains(html, dashboardURL) {
		t.Errorf("HTML should contain dashboard URL %q", dashboardURL)
	}
}

func TestWrapNilSafe(t *testing.T) {
	// Ensure wrap doesn't panic with empty strings
	result := wrap("", "")
	if result == "" {
		t.Error("wrap should return non-empty string even with empty inputs")
	}
	if !strings.Contains(result, "<!DOCTYPE html>") {
		t.Error("wrap should return valid HTML structure")
	}
}

func TestWelcomeTrial(t *testing.T) {
	subject, html := WelcomeTrial("Alice", "Acme Co", "https://acme.pulzifi.com", 14)
	if !strings.Contains(subject, "trial") {
		t.Errorf("subject missing 'trial': %q", subject)
	}
	for _, want := range []string{"Alice", "Acme Co", "https://acme.pulzifi.com", "14-day"} {
		if !strings.Contains(html, want) {
			t.Errorf("HTML missing %q", want)
		}
	}
}

func TestTrialDay7(t *testing.T) {
	subject, html := TrialDay7("Alice", "https://acme.pulzifi.com")
	if !strings.Contains(subject, "week") {
		t.Errorf("subject missing 'week': %q", subject)
	}
	if !strings.Contains(html, "Alice") {
		t.Errorf("HTML missing first name")
	}
}

func TestTrialDay12(t *testing.T) {
	subject, html := TrialDay12("Alice", "https://acme.pulzifi.com/billing")
	if !strings.Contains(subject, "2 days") {
		t.Errorf("subject missing '2 days': %q", subject)
	}
	if !strings.Contains(html, "https://acme.pulzifi.com/billing") {
		t.Errorf("HTML missing upgrade URL")
	}
}

func TestTrialExpired(t *testing.T) {
	subject, html := TrialExpired("Alice", "https://acme.pulzifi.com/billing")
	if !strings.Contains(strings.ToLower(subject), "ended") {
		t.Errorf("subject missing 'ended': %q", subject)
	}
	for _, want := range []string{"Alice", "https://acme.pulzifi.com/billing"} {
		if !strings.Contains(html, want) {
			t.Errorf("HTML missing %q", want)
		}
	}
}

func TestTrialConverted(t *testing.T) {
	subject, html := TrialConverted("Alice", "Pro", "https://acme.pulzifi.com")
	if !strings.Contains(subject, "Pro") {
		t.Errorf("subject missing plan name: %q", subject)
	}
	for _, want := range []string{"Alice", "Pulzifi Pro", "https://acme.pulzifi.com"} {
		if !strings.Contains(html, want) {
			t.Errorf("HTML missing %q", want)
		}
	}
}
