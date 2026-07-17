package entities

import "time"

type NotificationPayload struct {
	EventType string
	Title     string
	Body      string
	Severity  string // info|warning|critical
	Links     []Link
	PageURL   string
	Raw       map[string]any

	// Enriched fields for change.detected — populated by PayloadBuilder from the
	// event payload. Optional: empty when the source event does not carry them
	// (e.g. alert.created), so renderers must degrade gracefully.
	PageTitle    string // human-readable monitored page name
	ChangeType   string // "content" | "visual"
	DiffSummary  string // AI/structural summary of what changed
	DiffImageURL string // public URL of the visual diff image, when available
	DashboardURL string // deep link to the page's changes view in the tenant dashboard
	ChangedAt    string // RFC3339 timestamp of the detecting check
}

type Link struct {
	Label, URL string
}

type Target struct {
	ID   string
	Name string
	Meta map[string]any
}

type DeliveryResult struct {
	Code     int
	BodySnip string
}

type OAuthResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    *time.Time   // nil if no expiry
	ProviderMeta map[string]any
}
