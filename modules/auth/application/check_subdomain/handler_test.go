package checksubdomain

import (
	"context"
	"errors"
	"testing"

	svcmocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services/mocks"
)

func TestCheckSubdomainHandler_Handle(t *testing.T) {
	tests := []struct {
		name          string
		subdomain     string
		setupMocks    func(orgDir *svcmocks.MockOrganizationDirectory, regReq *svcmocks.MockRegistrationRequestWriter)
		wantErr       bool
		wantAvailable bool
	}{
		{
			name:      "available subdomain returns available=true",
			subdomain: "newacme",
			setupMocks: func(orgDir *svcmocks.MockOrganizationDirectory, regReq *svcmocks.MockRegistrationRequestWriter) {
				orgDir.ValidateSubdomainErr = nil
				orgDir.CountBySubdomainResult = 0
				regReq.ExistsPendingBySubdomainResult = false
			},
			wantErr:       false,
			wantAvailable: true,
		},
		{
			name:      "subdomain already in use returns available=false",
			subdomain: "existingacme",
			setupMocks: func(orgDir *svcmocks.MockOrganizationDirectory, regReq *svcmocks.MockRegistrationRequestWriter) {
				orgDir.CountBySubdomainResult = 1
			},
			wantErr:       false,
			wantAvailable: false,
		},
		{
			name:      "subdomain pending approval returns available=false",
			subdomain: "pendingacme",
			setupMocks: func(orgDir *svcmocks.MockOrganizationDirectory, regReq *svcmocks.MockRegistrationRequestWriter) {
				orgDir.CountBySubdomainResult = 0
				regReq.ExistsPendingBySubdomainResult = true
			},
			wantErr:       false,
			wantAvailable: false,
		},
		{
			name:      "invalid subdomain format returns available=false (no error)",
			subdomain: "INVALID__FORMAT",
			setupMocks: func(orgDir *svcmocks.MockOrganizationDirectory, regReq *svcmocks.MockRegistrationRequestWriter) {
				orgDir.ValidateSubdomainErr = errors.New("subdomain: invalid characters")
			},
			wantErr:       false,
			wantAvailable: false,
		},
		{
			name:      "count repo error propagates",
			subdomain: "acme",
			setupMocks: func(orgDir *svcmocks.MockOrganizationDirectory, regReq *svcmocks.MockRegistrationRequestWriter) {
				orgDir.CountBySubdomainErr = errors.New("db error")
			},
			wantErr: true,
		},
		{
			name:      "pending check repo error propagates",
			subdomain: "acme",
			setupMocks: func(orgDir *svcmocks.MockOrganizationDirectory, regReq *svcmocks.MockRegistrationRequestWriter) {
				orgDir.CountBySubdomainResult = 0
				regReq.ExistsPendingBySubdomainErr = errors.New("db error")
			},
			wantErr: true,
		},
		{
			name:      "subdomain is normalized to lowercase before lookup",
			subdomain: "ACME",
			setupMocks: func(orgDir *svcmocks.MockOrganizationDirectory, regReq *svcmocks.MockRegistrationRequestWriter) {
				orgDir.CountBySubdomainResult = 0
				regReq.ExistsPendingBySubdomainResult = false
			},
			wantErr:       false,
			wantAvailable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orgDir := &svcmocks.MockOrganizationDirectory{}
			regReq := &svcmocks.MockRegistrationRequestWriter{}

			if tt.setupMocks != nil {
				tt.setupMocks(orgDir, regReq)
			}

			h := NewHandler(regReq, orgDir)
			resp, err := h.Handle(context.Background(), tt.subdomain)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp.Available != tt.wantAvailable {
				t.Errorf("available: want %v, got %v", tt.wantAvailable, resp.Available)
			}
		})
	}
}
