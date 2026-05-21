package changepassword_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	changepassword "github.com/jcsoftdev/pulzifi-back/modules/auth/application/change_password"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	repomocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories/mocks"
	svcmocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services/mocks"
)

func TestChangePasswordHandler_Handle(t *testing.T) {
	userID := uuid.New()
	existingUser := &entities.User{ID: userID, Email: "alice@example.com"}

	tests := []struct {
		name    string
		req     *changepassword.Request
		setup   func(*repomocks.MockUserRepository, *svcmocks.MockAuthService)
		wantErr bool
	}{
		{
			name: "successfully changes password",
			req:  &changepassword.Request{UserID: userID, CurrentPassword: "oldpass", NewPassword: "newpassword1"},
			setup: func(u *repomocks.MockUserRepository, a *svcmocks.MockAuthService) {
				u.GetByIDUser = existingUser
				a.ValidateErr = nil
				a.HashResult = "hashed-new-password"
			},
		},
		{
			name: "error when new password too short",
			req:  &changepassword.Request{UserID: userID, CurrentPassword: "oldpass", NewPassword: "short"},
			setup: func(u *repomocks.MockUserRepository, a *svcmocks.MockAuthService) {},
			wantErr: true,
		},
		{
			name: "error when user not found",
			req:  &changepassword.Request{UserID: uuid.New(), CurrentPassword: "oldpass", NewPassword: "newpassword1"},
			setup: func(u *repomocks.MockUserRepository, a *svcmocks.MockAuthService) {
				u.GetByIDUser = nil
				u.GetByIDErr = errors.New("not found")
			},
			wantErr: true,
		},
		{
			name: "error when current password is wrong",
			req:  &changepassword.Request{UserID: userID, CurrentPassword: "wrongpass", NewPassword: "newpassword1"},
			setup: func(u *repomocks.MockUserRepository, a *svcmocks.MockAuthService) {
				u.GetByIDUser = existingUser
				a.ValidateErr = errors.New("invalid credentials")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &repomocks.MockUserRepository{}
			authSvc := &svcmocks.MockAuthService{}
			tt.setup(userRepo, authSvc)

			h := changepassword.NewHandler(userRepo, authSvc)
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
				t.Error("expected non-empty success message")
			}
		})
	}
}
