package forgotpassword

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// Handler handles the forgot_password use case.
type Handler struct {
	userRepo          repositories.UserRepository
	passwordResetRepo repositories.PasswordResetRepository
	notifier          services.RegistrationNotifier
	ttl               time.Duration
}

// NewHandler creates a new forgot password handler with the default 1h TTL.
func NewHandler(
	userRepo repositories.UserRepository,
	passwordResetRepo repositories.PasswordResetRepository,
	notifier services.RegistrationNotifier,
) *Handler {
	return &Handler{
		userRepo:          userRepo,
		passwordResetRepo: passwordResetRepo,
		notifier:          notifier,
		ttl:               1 * time.Hour,
	}
}

// WithTTL returns a copy of the handler with a custom password-reset token TTL.
func (h *Handler) WithTTL(ttl time.Duration) *Handler {
	copy := *h
	copy.ttl = ttl
	return &copy
}

// Handle triggers a password reset email. Always returns success to prevent email enumeration.
func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	resp := &Response{Message: "if an account exists with that email, a password reset link has been sent"}

	if h.passwordResetRepo == nil {
		return resp, nil
	}

	user, err := h.userRepo.GetByEmail(ctx, req.Email)
	if err != nil || user == nil {
		return resp, nil
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		logger.Error("forgot_password: failed to generate reset token", zap.Error(err))
		return resp, nil
	}
	token := hex.EncodeToString(tokenBytes)
	expiresAt := time.Now().Add(h.ttl)

	if err := h.passwordResetRepo.Store(ctx, uuid.New(), user.ID, token, expiresAt); err != nil {
		logger.Error("forgot_password: failed to store reset token", zap.Error(err))
		return resp, nil
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", req.FrontendURL, token)
	go func() {
		if err := h.notifier.SendPasswordReset(context.Background(), user.Email, user.FirstName, resetURL); err != nil {
			logger.Error("forgot_password: failed to send email", zap.Error(err))
		}
	}()

	return resp, nil
}
