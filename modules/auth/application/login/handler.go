package login

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Handler handles user login
type Handler struct {
	repo repositories.UserRepository
}

// NewHandler creates a new handler instance
func NewHandler(repo repositories.UserRepository) *Handler {
	return &Handler{
		repo: repo,
	}
}

// Handle executes the login use case
func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	// Get user by email
	user, err := h.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		logger.Error("Failed to get user by email", zap.Error(err))
		return nil, err
	}

	if user == nil {
		logger.Warn("User not found", zap.String("email", req.Email))
		return nil, errors.NewUserError("USER_NOT_FOUND", "user not found")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		logger.Warn("Invalid password", zap.String("email", req.Email))
		return nil, errors.NewUserError("INVALID_PASSWORD", "invalid password")
	}

	// TODO: Generate JWT tokens
	logger.Info("User logged in successfully", zap.String("email", user.Email), zap.String("id", user.ID.String()))

	return &Response{
		UserID:       user.ID,
		Email:        user.Email,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		AccessToken:  "placeholder_access_token",
		RefreshToken: "placeholder_refresh_token",
	}, nil
}
