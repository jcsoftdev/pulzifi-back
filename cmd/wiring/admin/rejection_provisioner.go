package adminwiring

import (
	"context"
	"database/sql"

	adminservices "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/services"
	authentities "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// rejectionProvisioner implements adminservices.RejectionProvisioner.
// Atomically updates the user status and registration request to rejected.
type rejectionProvisioner struct {
	db *sql.DB
}

// NewRejectionProvisioner creates a RejectionProvisioner backed by *sql.DB.
func NewRejectionProvisioner(db *sql.DB) adminservices.RejectionProvisioner {
	return &rejectionProvisioner{db: db}
}

func (p *rejectionProvisioner) Reject(ctx context.Context, input adminservices.RejectionInput) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("Failed to begin rejection transaction", zap.Error(err))
		return err
	}
	defer tx.Rollback()

	// Update user status to rejected
	_, err = tx.ExecContext(ctx,
		`UPDATE public.users SET status = $1, updated_at = NOW() WHERE id = $2`,
		authentities.UserStatusRejected, input.UserID,
	)
	if err != nil {
		logger.Error("Failed to update user status", zap.Error(err))
		return err
	}

	// Update registration request status to rejected
	_, err = tx.ExecContext(ctx,
		`UPDATE public.registration_requests SET status = $1, reviewed_by = $2, reviewed_at = NOW(), updated_at = NOW() WHERE id = $3`,
		"rejected", input.ReviewerID, input.RequestID,
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
		zap.String("user_id", input.UserID.String()),
		zap.String("request_id", input.RequestID.String()),
	)

	return nil
}
