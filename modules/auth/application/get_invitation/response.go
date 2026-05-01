package getinvitation

import "time"

// Response is the public payload returned to anyone visiting an invitation
// acceptance link. It intentionally exposes only non-sensitive fields needed
// by the UI to greet the invitee.
type Response struct {
	Email         string    `json:"email"`
	InvitedByName string    `json:"invited_by_name"`
	ExpiresAt     time.Time `json:"expires_at"`
}
