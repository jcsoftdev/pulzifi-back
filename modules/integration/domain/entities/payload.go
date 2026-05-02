package entities

import "time"

type NotificationPayload struct {
	EventType string
	Title     string
	Body      string
	Severity  string         // info|warning|critical
	Links     []Link
	PageURL   string
	Raw       map[string]any
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
