package updateprofile

// Request is the input DTO for updating a social profile.
// All fields are pointers — only non-nil fields are applied.
type Request struct {
	// CheckIntervalMinutes, when non-nil, updates the check cadence.
	// Must be one of the allowed presets (720, 1440). Minimum cadence is 12h.
	CheckIntervalMinutes *int `json:"check_interval_minutes"`

	// IsActive, when non-nil, enables or disables the profile in the scheduler.
	// Setting to false sets next_check_at=nil; setting to true resets next_check_at=now().
	IsActive *bool `json:"is_active"`
}
