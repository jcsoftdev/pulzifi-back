package get_email_status

import (
	"context"

	"github.com/google/uuid"
	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/email/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/email/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// GetEmailStatusRequest represents a request to get email status
type GetEmailStatusRequest struct {
	EmailID string `json:"email_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// GetEmailStatusResponse represents the response of getting email status
type GetEmailStatusResponse struct {
	EmailID string `json:"email_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	To      string `json:"to" example:"user@example.com"`
	Subject string `json:"subject" example:"Welcome"`
	Status  string `json:"status" example:"sent"`
	SentAt  string `json:"sent_at" example:"2025-11-11T12:00:00Z"`
}

// GetEmailStatusHandler handles the get email status use case
type GetEmailStatusHandler struct {
	emailRepo repositories.EmailRepository
}

// NewGetEmailStatusHandler creates a new get email status handler
func NewGetEmailStatusHandler(emailRepo repositories.EmailRepository) *GetEmailStatusHandler {
	return &GetEmailStatusHandler{
		emailRepo: emailRepo,
	}
}

// Handle processes the get email status request
func (h *GetEmailStatusHandler) Handle(ctx context.Context, req *GetEmailStatusRequest) (*GetEmailStatusResponse, error) {
	logger.Info("Getting email status", zap.String("email_id", req.EmailID))

	emailID, err := uuid.Parse(req.EmailID)
	if err != nil {
		logger.Warn("Invalid email ID", zap.Error(err))
		return nil, &domainerrors.InvalidEmailError{Message: "invalid email id format"}
	}

	email, err := h.emailRepo.GetByID(ctx, emailID)
	if err != nil {
		logger.Warn("Email not found", zap.Error(err))
		return nil, &domainerrors.EmailNotFoundError{EmailID: req.EmailID}
	}

	sentAt := ""
	if email.SentAt != nil {
		sentAt = email.SentAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return &GetEmailStatusResponse{
		EmailID: email.ID.String(),
		To:      email.To,
		Subject: email.Subject,
		Status:  string(email.Status),
		SentAt:  sentAt,
	}, nil
}
