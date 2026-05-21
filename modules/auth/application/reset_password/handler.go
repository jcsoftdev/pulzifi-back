package resetpassword

import (
	"context"
	"fmt"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// Handler handles the reset_password use case.
type Handler struct {
	passwordResetRepo repositories.PasswordResetRepository
	authService       services.AuthService
}

// NewHandler creates a new reset password handler.
func NewHandler(passwordResetRepo repositories.PasswordResetRepository, authService services.AuthService) *Handler {
	return &Handler{passwordResetRepo: passwordResetRepo, authService: authService}
}

// Handle resets the user's password atomically.
func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	if len(req.NewPassword) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}
	if h.passwordResetRepo == nil {
		return nil, fmt.Errorf("service unavailable")
	}

	pr, err := h.passwordResetRepo.Find(ctx, req.Token)
	if err != nil || pr == nil {
		return nil, fmt.Errorf("invalid or expired token")
	}
	if pr.Used || time.Now().After(pr.ExpiresAt) {
		return nil, fmt.Errorf("invalid or expired token")
	}

	hashedPassword, err := h.authService.HashPassword(req.NewPassword)
	if err != nil {
		logger.Error("reset_password: failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("failed to reset password")
	}

	if err := h.passwordResetRepo.Consume(ctx, req.Token, pr.UserID, hashedPassword); err != nil {
		logger.Error("reset_password: failed to consume reset token", zap.Error(err))
		return nil, fmt.Errorf("failed to reset password")
	}

	return &Response{Message: "password reset successfully"}, nil
}
