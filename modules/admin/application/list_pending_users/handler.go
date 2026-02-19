package listpendingusers

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories"
	authrepos "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// Handler handles listing pending user registrations
type Handler struct {
	regReqRepo repositories.RegistrationRequestRepository
	userRepo   authrepos.UserRepository
}

// NewHandler creates a new handler instance
func NewHandler(regReqRepo repositories.RegistrationRequestRepository, userRepo authrepos.UserRepository) *Handler {
	return &Handler{
		regReqRepo: regReqRepo,
		userRepo:   userRepo,
	}
}

// Handle executes the list pending users use case
func (h *Handler) Handle(ctx context.Context, limit, offset int) (*Response, error) {
	requests, err := h.regReqRepo.ListPending(ctx, limit, offset)
	if err != nil {
		logger.Error("Failed to list pending registration requests", zap.Error(err))
		return nil, err
	}

	var pendingUsers []PendingUserResponse
	for _, req := range requests {
		user, err := h.userRepo.GetByID(ctx, req.UserID)
		if err != nil {
			logger.Error("Failed to get user for pending request", zap.Error(err), zap.String("user_id", req.UserID.String()))
			continue
		}
		if user == nil {
			continue
		}

		pendingUsers = append(pendingUsers, PendingUserResponse{
			RequestID:             req.ID,
			UserID:                req.UserID,
			Email:                 user.Email,
			FirstName:             user.FirstName,
			LastName:              user.LastName,
			OrganizationName:      req.OrganizationName,
			OrganizationSubdomain: req.OrganizationSubdomain,
			CreatedAt:             req.CreatedAt,
		})
	}

	if pendingUsers == nil {
		pendingUsers = []PendingUserResponse{}
	}

	return &Response{PendingUsers: pendingUsers}, nil
}
