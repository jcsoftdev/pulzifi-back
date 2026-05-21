package trialstatus

import "time"

// Response is returned from GET /api/v1/usage/trial-status. It is the source
// of truth for the frontend trial banner + upgrade modal.
type Response struct {
	IsTrial       bool       `json:"is_trial"`
	IsExpired     bool       `json:"is_expired"`
	NeedsUpgrade  bool       `json:"needs_upgrade"`
	Converted     bool       `json:"converted"`
	DaysRemaining int        `json:"days_remaining"`
	TrialEndsAt   *time.Time `json:"trial_ends_at,omitempty"`
}
