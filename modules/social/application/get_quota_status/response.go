package getquotastatus

import "time"

// Response is the output DTO for get_quota_status (REQ-QUOTA-STATUS-01 through REQ-QUOTA-STATUS-05).
type Response struct {
	// ChecksUsed is the number of social checks consumed today.
	ChecksUsed int `json:"checks_used"`

	// ChecksLimit is the plan's daily cap. nil means unlimited (REQ-QUOTA-STATUS-02).
	ChecksLimit *int `json:"checks_limit"`

	// ResetsAt is the ISO-8601 timestamp of the next UTC midnight,
	// when the quota pool resets (REQ-QUOTA-STATUS-02).
	ResetsAt time.Time `json:"resets_at"`
}
