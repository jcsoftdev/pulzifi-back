package approveuser

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/entities"
	adminerrors "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories"
	adminservices "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// Handler handles user approval
type Handler struct {
	regReqRepo  repositories.RegistrationRequestRepository
	provisioner adminservices.ApprovalProvisioner
}

// NewHandler creates a new handler instance
func NewHandler(
	regReqRepo repositories.RegistrationRequestRepository,
	provisioner adminservices.ApprovalProvisioner,
) *Handler {
	return &Handler{
		regReqRepo:  regReqRepo,
		provisioner: provisioner,
	}
}

// Handle executes the approve user use case
func (h *Handler) Handle(ctx context.Context, requestID uuid.UUID, reviewerID uuid.UUID) error {
	// Get the registration request
	regReq, err := h.regReqRepo.GetByID(ctx, requestID)
	if err != nil {
		logger.Error("Failed to get registration request", zap.Error(err))
		return err
	}
	if regReq == nil {
		return adminerrors.ErrRegistrationRequestNotFound
	}

	if regReq.Status != entities.RegistrationStatusPending {
		return adminerrors.ErrAlreadyReviewed
	}

	return h.provisioner.Provision(ctx, adminservices.ApprovalInput{
		UserID:                regReq.UserID,
		OrganizationName:      regReq.OrganizationName,
		OrganizationSubdomain: regReq.OrganizationSubdomain,
		ReviewerID:            reviewerID,
		RequestID:             requestID,
	})
}

// GetRegistrationRequest retrieves a registration request by ID.
func (h *Handler) GetRegistrationRequest(ctx context.Context, requestID uuid.UUID) (*entities.RegistrationRequest, error) {
	return h.regReqRepo.GetByID(ctx, requestID)
}
