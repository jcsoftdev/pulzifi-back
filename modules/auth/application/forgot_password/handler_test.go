package forgotpassword_test

import (
	"context"
	"errors"
	"testing"
	"time"

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

// TestForgotPasswordHandler_UsesConfiguredTTL verifies that the handler stores
// the reset token with an expiry matching the provided TTL, not a hardcoded value.
func TestForgotPasswordHandler_UsesConfiguredTTL(t *testing.T) {
	existingUser := &entities.User{ID: uuid.New(), Email: "alice@example.com", FirstName: "Alice"}

	tests := []struct {
		name    string
		ttl     time.Duration
		wantTTL time.Duration
	}{
		{
			name:    "default 1h TTL",
			ttl:     1 * time.Hour,
			wantTTL: 1 * time.Hour,
		},
		{
			name:    "short 30m TTL",
			ttl:     30 * time.Minute,
			wantTTL: 30 * time.Minute,
		},
		{
			name:    "custom 2h TTL",
			ttl:     2 * time.Hour,
			wantTTL: 2 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &repomocks.MockUserRepository{GetByEmailUser: existingUser}
			passwordResetRepo := &repomocks.MockPasswordResetRepository{}
			notifier := &svcmocks.MockNotifier{}

			before := time.Now()
			h := forgotpassword.NewHandler(userRepo, passwordResetRepo, notifier).WithTTL(tt.ttl)
			_, err := h.Handle(context.Background(), &forgotpassword.Request{
				Email:       existingUser.Email,
				FrontendURL: "https://app.example.com",
			})
			after := time.Now()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			storedExpiry := passwordResetRepo.StoredExpiresAt
			if storedExpiry.IsZero() {
				t.Fatal("Store was not called — expires_at not captured")
			}

			minExpected := before.Add(tt.wantTTL)
			maxExpected := after.Add(tt.wantTTL)

			if storedExpiry.Before(minExpected) || storedExpiry.After(maxExpected) {
				t.Errorf("stored expiry %v not within expected range [%v, %v] for TTL %v",
					storedExpiry, minExpected, maxExpected, tt.wantTTL)
			}
		})
	}
}
