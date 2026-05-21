package refresh_token

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	repomocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories/mocks"
	svcmocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services/mocks"
)

func TestRefreshTokenHandler_Handle(t *testing.T) {
	userID := uuid.New()
	testUser := &entities.User{
		ID:    userID,
		Email: "user@example.com",
	}
	// Use unique token strings per test run to avoid Redis cache hits from previous runs.
	uniquePrefix := uuid.New().String()
	validToken := &entities.RefreshToken{
		UserID:    userID,
		Token:     uniquePrefix + "-valid",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		IsRevoked: false,
	}
	expiredToken := &entities.RefreshToken{
		UserID:    userID,
		Token:     uniquePrefix + "-expired",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // in the past
		IsRevoked: false,
	}
	revokedToken := &entities.RefreshToken{
		UserID:    userID,
		Token:     uniquePrefix + "-revoked",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		IsRevoked: true,
	}

	tests := []struct {
		name        string
		req         *Request
		setupMocks  func(rt *repomocks.MockRefreshTokenRepository, u *repomocks.MockUserRepository, ts *svcmocks.MockTokenService)
		wantErr     bool
		wantErrIs   error
		checkResult func(t *testing.T, resp *Response)
	}{
		{
			name: "valid token rotates and returns new access + refresh tokens",
			req:  &Request{RefreshToken: uniquePrefix + "-valid"},
			setupMocks: func(rt *repomocks.MockRefreshTokenRepository, u *repomocks.MockUserRepository, ts *svcmocks.MockTokenService) {
				rt.FindByTokenResult = validToken
				u.GetByIDUser = testUser
				ts.GenerateAccessTokenResult = "new-access-token"
				ts.GenerateRefreshTokenResult = "new-refresh-token"
			},
			wantErr: false,
			checkResult: func(t *testing.T, resp *Response) {
				t.Helper()
				if resp.AccessToken != "new-access-token" {
					t.Errorf("access token: want %q, got %q", "new-access-token", resp.AccessToken)
				}
				if resp.RefreshToken != "new-refresh-token" {
					t.Errorf("refresh token: want %q, got %q", "new-refresh-token", resp.RefreshToken)
				}
				if resp.TokenType != "Bearer" {
					t.Errorf("token type: want Bearer, got %q", resp.TokenType)
				}
			},
		},
		{
			name: "token not found in repo returns ErrInvalidRefreshToken",
			req:  &Request{RefreshToken: uniquePrefix + "-unknown"},
			setupMocks: func(rt *repomocks.MockRefreshTokenRepository, u *repomocks.MockUserRepository, ts *svcmocks.MockTokenService) {
				rt.FindByTokenErr = errors.New("not found")
			},
			wantErr:   true,
			wantErrIs: ErrInvalidRefreshToken,
		},
		{
			name: "revoked token returns ErrRevokedRefreshToken",
			req:  &Request{RefreshToken: uniquePrefix + "-revoked"},
			setupMocks: func(rt *repomocks.MockRefreshTokenRepository, u *repomocks.MockUserRepository, ts *svcmocks.MockTokenService) {
				rt.FindByTokenResult = revokedToken
			},
			wantErr:   true,
			wantErrIs: ErrRevokedRefreshToken,
		},
		{
			name: "expired token returns ErrExpiredRefreshToken",
			req:  &Request{RefreshToken: uniquePrefix + "-expired"},
			setupMocks: func(rt *repomocks.MockRefreshTokenRepository, u *repomocks.MockUserRepository, ts *svcmocks.MockTokenService) {
				rt.FindByTokenResult = expiredToken
			},
			wantErr:   true,
			wantErrIs: ErrExpiredRefreshToken,
		},
		{
			name: "access token generation failure propagates error",
			req:  &Request{RefreshToken: uniquePrefix + "-valid"},
			setupMocks: func(rt *repomocks.MockRefreshTokenRepository, u *repomocks.MockUserRepository, ts *svcmocks.MockTokenService) {
				rt.FindByTokenResult = validToken
				u.GetByIDUser = testUser
				ts.GenerateAccessTokenErr = errors.New("token gen failed")
			},
			wantErr: true,
		},
		{
			name: "refresh token storage failure propagates error",
			req:  &Request{RefreshToken: uniquePrefix + "-valid"},
			setupMocks: func(rt *repomocks.MockRefreshTokenRepository, u *repomocks.MockUserRepository, ts *svcmocks.MockTokenService) {
				rt.FindByTokenResult = validToken
				u.GetByIDUser = testUser
				ts.GenerateAccessTokenResult = "new-access-token"
				ts.GenerateRefreshTokenResult = "new-refresh-token"
				rt.CreateErr = errors.New("db error storing new token")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rtRepo := &repomocks.MockRefreshTokenRepository{}
			userRepo := &repomocks.MockUserRepository{}
			tokenSvc := &svcmocks.MockTokenService{}

			if tt.setupMocks != nil {
				tt.setupMocks(rtRepo, userRepo, tokenSvc)
			}

			h := NewHandler(rtRepo, userRepo, tokenSvc)
			resp, err := h.Handle(context.Background(), tt.req)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("expected error %v, got %v", tt.wantErrIs, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, resp)
			}
		})
	}
}
