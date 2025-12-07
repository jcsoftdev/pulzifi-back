package services

import (
	"context"
	"regexp"

	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/email/domain/errors"
)

// EmailService provides domain logic for email operations
type EmailService struct{}

// NewEmailService creates a new email service
func NewEmailService() *EmailService {
	return &EmailService{}
}

// ValidateEmail validates email address format
func (s *EmailService) ValidateEmail(ctx context.Context, email string) error {
	// Simple regex for email validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return &domainerrors.InvalidEmailError{Message: "invalid email format"}
	}
	return nil
}

// ValidateEmailContent validates email content
func (s *EmailService) ValidateEmailContent(ctx context.Context, subject, body string) error {
	if subject == "" {
		return &domainerrors.InvalidEmailError{Message: "subject cannot be empty"}
	}
	if body == "" {
		return &domainerrors.InvalidEmailError{Message: "body cannot be empty"}
	}
	if len(subject) > 255 {
		return &domainerrors.InvalidEmailError{Message: "subject too long"}
	}
	if len(body) > 10000 {
		return &domainerrors.InvalidEmailError{Message: "body too long"}
	}
	return nil
}
