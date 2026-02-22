package login

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	repomocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories/mocks"
	svcmocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services/mocks"
)

func TestLoginHandler_Handle(t *testing.T) {
	testUserID := uuid.New()
	testUser := &entities.User{
		ID:    testUserID,
		Email: "alice@example.com",
	}
	tenant := "acme"

	tests := []struct {
		name       string
		req        *Request
		setupMocks func(
			authSvc *svcmocks.MockAuthService,
			tokenSvc *svcmocks.MockTokenService,
			userRepo *repomocks.MockUserRepository,
			rtRepo *repomocks.MockRefreshTokenRepository,
		)
		wantErr     bool
		wantToken   string
		wantRefresh string
	}{
		{
			name: "successful login",
			req:  &Request{Email: "alice@example.com", Password: "password"},
			setupMocks: func(authSvc *svcmocks.MockAuthService, tokenSvc *svcmocks.MockTokenService, userRepo *repomocks.MockUserRepository, rtRepo *repomocks.MockRefreshTokenRepository) {
				authSvc.AuthenticateResult = testUser
				tokenSvc.GenerateAccessTokenResult = "access-token-123"
				tokenSvc.GenerateRefreshTokenResult = "refresh-token-456"
				userRepo.GetUserFirstOrganizationResult = &tenant
			},
			wantErr:     false,
			wantToken:   "access-token-123",
			wantRefresh: "refresh-token-456",
		},
		{
			name: "authentication failure",
			req:  &Request{Email: "alice@example.com", Password: "wrong"},
			setupMocks: func(authSvc *svcmocks.MockAuthService, tokenSvc *svcmocks.MockTokenService, userRepo *repomocks.MockUserRepository, rtRepo *repomocks.MockRefreshTokenRepository) {
				authSvc.AuthenticateErr = errors.New("invalid credentials")
			},
			wantErr: true,
		},
		{
			name: "access token generation failure",
			req:  &Request{Email: "alice@example.com", Password: "password"},
			setupMocks: func(authSvc *svcmocks.MockAuthService, tokenSvc *svcmocks.MockTokenService, userRepo *repomocks.MockUserRepository, rtRepo *repomocks.MockRefreshTokenRepository) {
				authSvc.AuthenticateResult = testUser
				tokenSvc.GenerateAccessTokenErr = errors.New("token gen failed")
			},
			wantErr: true,
		},
		{
			name: "refresh token generation failure",
			req:  &Request{Email: "alice@example.com", Password: "password"},
			setupMocks: func(authSvc *svcmocks.MockAuthService, tokenSvc *svcmocks.MockTokenService, userRepo *repomocks.MockUserRepository, rtRepo *repomocks.MockRefreshTokenRepository) {
				authSvc.AuthenticateResult = testUser
				tokenSvc.GenerateAccessTokenResult = "access-token-123"
				tokenSvc.GenerateRefreshTokenErr = errors.New("refresh gen failed")
			},
			wantErr: true,
		},
		{
			name: "refresh token storage failure",
			req:  &Request{Email: "alice@example.com", Password: "password"},
			setupMocks: func(authSvc *svcmocks.MockAuthService, tokenSvc *svcmocks.MockTokenService, userRepo *repomocks.MockUserRepository, rtRepo *repomocks.MockRefreshTokenRepository) {
				authSvc.AuthenticateResult = testUser
				tokenSvc.GenerateAccessTokenResult = "access-token-123"
				tokenSvc.GenerateRefreshTokenResult = "refresh-token-456"
				rtRepo.CreateErr = errors.New("db error")
			},
			wantErr: true,
		},
		{
			name: "login succeeds even if org lookup fails",
			req:  &Request{Email: "alice@example.com", Password: "password"},
			setupMocks: func(authSvc *svcmocks.MockAuthService, tokenSvc *svcmocks.MockTokenService, userRepo *repomocks.MockUserRepository, rtRepo *repomocks.MockRefreshTokenRepository) {
				authSvc.AuthenticateResult = testUser
				tokenSvc.GenerateAccessTokenResult = "access-token-123"
				tokenSvc.GenerateRefreshTokenResult = "refresh-token-456"
				userRepo.GetUserFirstOrganizationErr = errors.New("no org")
			},
			wantErr:     false,
			wantToken:   "access-token-123",
			wantRefresh: "refresh-token-456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authSvc := &svcmocks.MockAuthService{}
			tokenSvc := &svcmocks.MockTokenService{}
			userRepo := &repomocks.MockUserRepository{}
			rtRepo := &repomocks.MockRefreshTokenRepository{}

			if tt.setupMocks != nil {
				tt.setupMocks(authSvc, tokenSvc, userRepo, rtRepo)
			}

			handler := NewHandler(authSvc, userRepo, rtRepo, tokenSvc)
			resp, err := handler.Handle(context.Background(), tt.req)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp.AccessToken != tt.wantToken {
				t.Errorf("access token: want %q, got %q", tt.wantToken, resp.AccessToken)
			}
			if resp.RefreshToken != tt.wantRefresh {
				t.Errorf("refresh token: want %q, got %q", tt.wantRefresh, resp.RefreshToken)
			}
			if resp.ExpiresIn <= 0 {
				t.Error("ExpiresIn should be positive")
			}
		})
	}
}
