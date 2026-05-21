package resetpassword_test

import (
	"context"
	"testing"

	resetpassword "github.com/jcsoftdev/pulzifi-back/modules/auth/application/reset_password"
	repomocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories/mocks"
	svcmocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services/mocks"
)

func TestResetPasswordHandler_Handle(t *testing.T) {
	tests := []struct {
		name    string
		req     *resetpassword.Request
		setup   func(*repomocks.MockPasswordResetRepository, *svcmocks.MockAuthService)
		wantErr bool
	}{
		{
			name: "error when password too short",
			req:  &resetpassword.Request{Token: "sometoken", NewPassword: "short"},
			setup: func(p *repomocks.MockPasswordResetRepository, m *svcmocks.MockAuthService) {},
			wantErr: true,
		},
		{
			name: "error when repo is nil (service unavailable)",
			req:  &resetpassword.Request{Token: "sometoken", NewPassword: "validpassword"},
			setup: func(p *repomocks.MockPasswordResetRepository, m *svcmocks.MockAuthService) {},
			wantErr: true,
		},
		{
			name: "error when token not found",
			req:  &resetpassword.Request{Token: "badtoken", NewPassword: "validpassword"},
			setup: func(p *repomocks.MockPasswordResetRepository, m *svcmocks.MockAuthService) {
				p.FindResult = nil // not found
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passwordResetRepo := &repomocks.MockPasswordResetRepository{}
			authSvc := &svcmocks.MockAuthService{}
			tt.setup(passwordResetRepo, authSvc)

			var h *resetpassword.Handler
			if tt.name == "error when repo is nil (service unavailable)" {
				h = resetpassword.NewHandler(nil, authSvc)
			} else {
				h = resetpassword.NewHandler(passwordResetRepo, authSvc)
			}
			_, err := h.Handle(context.Background(), tt.req)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
