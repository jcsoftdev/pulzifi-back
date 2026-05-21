package forgotpassword_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	forgotpassword "github.com/jcsoftdev/pulzifi-back/modules/auth/application/forgot_password"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	repomocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories/mocks"
	svcmocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services/mocks"
)

func TestForgotPasswordHandler_Handle(t *testing.T) {
	existingUser := &entities.User{ID: uuid.New(), Email: "alice@example.com", FirstName: "Alice"}

	tests := []struct {
		name    string
		req     *forgotpassword.Request
		setup   func(*repomocks.MockUserRepository, *repomocks.MockPasswordResetRepository, *svcmocks.MockNotifier)
		wantErr bool
	}{
		{
			name: "always returns success even when user not found (anti-enumeration)",
			req:  &forgotpassword.Request{Email: "unknown@example.com"},
			setup: func(u *repomocks.MockUserRepository, p *repomocks.MockPasswordResetRepository, n *svcmocks.MockNotifier) {
				u.GetByEmailUser = nil
			},
		},
		{
			name: "always returns success even when repo errors",
			req:  &forgotpassword.Request{Email: "alice@example.com"},
			setup: func(u *repomocks.MockUserRepository, p *repomocks.MockPasswordResetRepository, n *svcmocks.MockNotifier) {
				u.GetByEmailErr = errors.New("db error")
			},
		},
		{
			name: "returns success for existing user (token stored, email queued)",
			req:  &forgotpassword.Request{Email: "alice@example.com", FrontendURL: "https://app.example.com"},
			setup: func(u *repomocks.MockUserRepository, p *repomocks.MockPasswordResetRepository, n *svcmocks.MockNotifier) {
				u.GetByEmailUser = existingUser
				// StoreErr is nil by default — success path
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &repomocks.MockUserRepository{}
			passwordResetRepo := &repomocks.MockPasswordResetRepository{}
			notifier := &svcmocks.MockNotifier{}
			tt.setup(userRepo, passwordResetRepo, notifier)

			h := forgotpassword.NewHandler(userRepo, passwordResetRepo, notifier)
			resp, err := h.Handle(context.Background(), tt.req)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.Message == "" {
				t.Error("expected non-empty message")
			}
		})
	}
}
