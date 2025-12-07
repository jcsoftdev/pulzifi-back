package send_email

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/email/domain/entities"
	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/email/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/email/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/email/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// SendEmailRequest represents a request to send an email
type SendEmailRequest struct {
	To      string `json:"to" example:"user@example.com"`
	Subject string `json:"subject" example:"Welcome"`
	Body    string `json:"body" example:"Welcome to our platform!"`
}

// SendEmailResponse represents the response of sending an email
type SendEmailResponse struct {
	EmailID string `json:"email_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status  string `json:"status" example:"pending"`
	Message string `json:"message" example:"Email queued for sending"`
}

// SendEmailHandler handles the send email use case
type SendEmailHandler struct {
	emailRepo    repositories.EmailRepository
	emailService *services.EmailService
}

// NewSendEmailHandler creates a new send email handler
func NewSendEmailHandler(
	emailRepo repositories.EmailRepository,
	emailService *services.EmailService,
) *SendEmailHandler {
	return &SendEmailHandler{
		emailRepo:    emailRepo,
		emailService: emailService,
	}
}

// Handle processes the send email request
func (h *SendEmailHandler) Handle(ctx context.Context, req *SendEmailRequest) (*SendEmailResponse, error) {
	logger.Info("Processing send email request", zap.String("to", req.To))

	// Validate email address
	if err := h.emailService.ValidateEmail(ctx, req.To); err != nil {
		logger.Warn("Invalid email address", zap.Error(err))
		return nil, err
	}

	// Validate content
	if err := h.emailService.ValidateEmailContent(ctx, req.Subject, req.Body); err != nil {
		logger.Warn("Invalid email content", zap.Error(err))
		return nil, err
	}

	// Create email entity
	email := entities.NewEmail(req.To, req.Subject, req.Body)

	// Save to repository
	if err := h.emailRepo.Save(ctx, email); err != nil {
		logger.Error("Failed to save email", zap.Error(err))
		return nil, &domainerrors.SendingFailedError{Reason: "failed to save email"}
	}

	logger.Info("Email queued successfully", zap.String("email_id", email.ID.String()))

	return &SendEmailResponse{
		EmailID: email.ID.String(),
		Status:  string(email.Status),
		Message: "Email queued for sending",
	}, nil
}
