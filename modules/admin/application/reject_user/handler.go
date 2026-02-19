package rejectuser

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/entities"
	adminerrors "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories"
	authentities "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	authrepos "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// Handler handles user rejection
type Handler struct {
	db         *sql.DB
	regReqRepo repositories.RegistrationRequestRepository
	userRepo   authrepos.UserRepository
}

// NewHandler creates a new handler instance
func NewHandler(
	db *sql.DB,
	regReqRepo repositories.RegistrationRequestRepository,
	userRepo authrepos.UserRepository,
) *Handler {
	return &Handler{
		db:         db,
		regReqRepo: regReqRepo,
		userRepo:   userRepo,
	}
}

// Handle executes the reject user use case
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

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("Failed to begin transaction", zap.Error(err))
		return err
	}
	defer tx.Rollback()

	// Update user status to rejected
	_, err = tx.ExecContext(ctx,
		`UPDATE public.users SET status = $1, updated_at = NOW() WHERE id = $2`,
		authentities.UserStatusRejected, regReq.UserID,
	)
	if err != nil {
		logger.Error("Failed to update user status", zap.Error(err))
		return err
	}

	// Update registration request status to rejected
	_, err = tx.ExecContext(ctx,
		`UPDATE public.registration_requests SET status = $1, reviewed_by = $2, reviewed_at = NOW(), updated_at = NOW() WHERE id = $3`,
		entities.RegistrationStatusRejected, reviewerID, requestID,
	)
	if err != nil {
		logger.Error("Failed to update registration request status", zap.Error(err))
		return err
	}

	if err := tx.Commit(); err != nil {
		logger.Error("Failed to commit rejection transaction", zap.Error(err))
		return err
	}

	logger.Info("User rejected",
		zap.String("user_id", regReq.UserID.String()),
		zap.String("request_id", requestID.String()),
	)

	return nil
}
