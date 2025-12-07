package entities

import (
	"time"

	"github.com/google/uuid"
)

// Email represents an email entity in the domain
type Email struct {
	ID        uuid.UUID
	To        string
	Subject   string
	Body      string
	Status    EmailStatus
	CreatedAt time.Time
	SentAt    *time.Time
}

// EmailStatus represents the status of an email
type EmailStatus string

const (
	EmailStatusPending EmailStatus = "pending"
	EmailStatusSent    EmailStatus = "sent"
	EmailStatusFailed  EmailStatus = "failed"
	EmailStatusBounced EmailStatus = "bounced"
)

// NewEmail creates a new email entity
func NewEmail(to, subject, body string) *Email {
	return &Email{
		ID:        uuid.New(),
		To:        to,
		Subject:   subject,
		Body:      body,
		Status:    EmailStatusPending,
		CreatedAt: time.Now(),
	}
}

// MarkAsSent marks the email as sent
func (e *Email) MarkAsSent() {
	e.Status = EmailStatusSent
	now := time.Now()
	e.SentAt = &now
}

// MarkAsFailed marks the email as failed
func (e *Email) MarkAsFailed() {
	e.Status = EmailStatusFailed
}
