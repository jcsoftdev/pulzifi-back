package listinvitations

import "time"

type Response struct {
	Invitations []InvitationDTO `json:"invitations"`
	Total       int             `json:"total"`
}

type InvitationDTO struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	Status      string     `json:"status"`
	InvitedBy   string     `json:"invited_by"`
	ExpiresAt   time.Time  `json:"expires_at"`
	EmailSentAt *time.Time `json:"email_sent_at,omitempty"`
	EmailError  *string    `json:"email_error,omitempty"`
	ResentCount int        `json:"resent_count"`
	CreatedAt   time.Time  `json:"created_at"`
}
