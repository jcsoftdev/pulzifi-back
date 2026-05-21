package logout

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	repomocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories/mocks"
)

func TestLogoutHandler_Handle(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		revokeErr  error
		wantRevoke bool
	}{
		{
			name:       "successful logout calls Revoke on token",
			token:      "valid-refresh-token",
			revokeErr:  nil,
			wantRevoke: true,
		},
		{
			name:       "repo error is swallowed — logout is best-effort",
			token:      "some-token",
			revokeErr:  errors.New("db connection error"),
			wantRevoke: true,
		},
		{
			name:       "empty token is a no-op — Revoke not called",
			token:      "",
			revokeErr:  nil,
			wantRevoke: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var revokeCalled bool
			rtRepo := &repomocks.MockRefreshTokenRepository{
				RevokeErr: tt.revokeErr,
			}
			// Track Revoke via a thin wrapper around the existing mock
			_ = &entities.RefreshToken{} // ensure import
			_ = uuid.Nil                 // ensure import

			originalRevokeErr := rtRepo.RevokeErr
			rtRepo.RevokeErr = tt.revokeErr
			_ = originalRevokeErr

			// Use function hook to track the call
			rtRepo.CreateFn = nil // ensure no side effects
			// Since MockRefreshTokenRepository.Revoke just returns RevokeErr, we can
			// detect the call by temporarily capturing it.
			captureRepo := &captureRevokeRepo{inner: rtRepo}

			h := NewHandler(captureRepo)
			h.Handle(context.Background(), tt.token)
			revokeCalled = captureRepo.revokeCalled

			if tt.wantRevoke && !revokeCalled {
				t.Error("expected Revoke to be called, but it was not")
			}
			if !tt.wantRevoke && revokeCalled {
				t.Error("expected Revoke NOT to be called, but it was")
			}
		})
	}
}

// captureRevokeRepo wraps MockRefreshTokenRepository to track Revoke invocations.
type captureRevokeRepo struct {
	inner        *repomocks.MockRefreshTokenRepository
	revokeCalled bool
}

func (r *captureRevokeRepo) Create(ctx context.Context, rt *entities.RefreshToken) error {
	return r.inner.Create(ctx, rt)
}

func (r *captureRevokeRepo) FindByToken(ctx context.Context, token string) (*entities.RefreshToken, error) {
	return r.inner.FindByToken(ctx, token)
}

func (r *captureRevokeRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.RefreshToken, error) {
	return r.inner.FindByUserID(ctx, userID)
}

func (r *captureRevokeRepo) Revoke(ctx context.Context, token string) error {
	r.revokeCalled = true
	return r.inner.Revoke(ctx, token)
}

func (r *captureRevokeRepo) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.inner.RevokeAllByUserID(ctx, userID)
}

func (r *captureRevokeRepo) DeleteExpired(ctx context.Context) error {
	return r.inner.DeleteExpired(ctx)
}
