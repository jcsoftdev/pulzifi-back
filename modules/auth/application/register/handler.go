package register

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Handler handles user registration
type Handler struct {
	repo repositories.UserRepository
}

// NewHandler creates a new handler instance
func NewHandler(repo repositories.UserRepository) *Handler {
	return &Handler{
		repo: repo,
	}
}

// Handle executes the register use case
func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	// Check if user already exists
	exists, err := h.repo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		logger.Error("Failed to check if user exists", zap.Error(err))
		return nil, err
	}

	if exists {
		logger.Warn("User already exists", zap.String("email", req.Email))
		return nil, errors.NewUserError("USER_ALREADY_EXISTS", "user already exists with this email")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Failed to hash password", zap.Error(err))
		return nil, err
	}

	// Create user entity
	user := entities.NewUser(req.Email, string(hashedPassword), req.FirstName, req.LastName)

	// Persist user
	if err := h.repo.Create(ctx, user); err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		return nil, err
	}

	logger.Info("User registered successfully", zap.String("email", user.Email), zap.String("id", user.ID.String()))

	return &Response{
		UserID:    user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Message:   "User registered successfully",
	}, nil
}
