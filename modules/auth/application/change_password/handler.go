package changepassword

import (
	"context"
	"fmt"

	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// Handler handles the change_password use case.
type Handler struct {
	userRepo    repositories.UserRepository
	authService services.AuthService
}

// NewHandler creates a new change password handler.
func NewHandler(userRepo repositories.UserRepository, authService services.AuthService) *Handler {
	return &Handler{userRepo: userRepo, authService: authService}
}

// Handle changes the authenticated user's password.
func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	if len(req.NewPassword) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}

	user, err := h.userRepo.GetByID(ctx, req.UserID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if err := h.authService.ValidateCredentials(ctx, user, req.CurrentPassword); err != nil {
		return nil, fmt.Errorf("current password is incorrect")
	}

	hashedPassword, err := h.authService.HashPassword(req.NewPassword)
	if err != nil {
		logger.Error("change_password: failed to hash new password", zap.Error(err))
		return nil, fmt.Errorf("failed to update password")
	}

	user.PasswordHash = hashedPassword
	if err := h.userRepo.Update(ctx, user); err != nil {
		logger.Error("change_password: failed to update user", zap.Error(err))
		return nil, fmt.Errorf("failed to update password")
	}

	return &Response{Message: "password updated successfully"}, nil
}
