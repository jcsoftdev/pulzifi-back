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
		setupMocks    func(orgDir *svcmocks.MockOrganizationDirectory)
		wantErr       bool
		wantAvailable bool
	}{
		{
			name:      "available subdomain returns available=true",
			subdomain: "newacme",
			setupMocks: func(orgDir *svcmocks.MockOrganizationDirectory) {
				orgDir.ValidateSubdomainErr = nil
				orgDir.CountBySubdomainResult = 0
			},
			wantErr:       false,
			wantAvailable: true,
		},
		{
			name:      "subdomain already in use returns available=false",
			subdomain: "existingacme",
			setupMocks: func(orgDir *svcmocks.MockOrganizationDirectory) {
				orgDir.CountBySubdomainResult = 1
			},
			wantErr:       false,
			wantAvailable: false,
		},
		{
			name:      "invalid subdomain format returns available=false (no error)",
			subdomain: "INVALID__FORMAT",
			setupMocks: func(orgDir *svcmocks.MockOrganizationDirectory) {
				orgDir.ValidateSubdomainErr = errors.New("subdomain: invalid characters")
			},
			wantErr:       false,
			wantAvailable: false,
		},
		{
			name:      "count repo error propagates",
			subdomain: "acme",
			setupMocks: func(orgDir *svcmocks.MockOrganizationDirectory) {
				orgDir.CountBySubdomainErr = errors.New("db error")
			},
			wantErr: true,
		},
		{
			name:      "subdomain is normalized to lowercase before lookup",
			subdomain: "ACME",
			setupMocks: func(orgDir *svcmocks.MockOrganizationDirectory) {
				orgDir.CountBySubdomainResult = 0
			},
			wantErr:       false,
			wantAvailable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orgDir := &svcmocks.MockOrganizationDirectory{}

			if tt.setupMocks != nil {
				tt.setupMocks(orgDir)
			}

			h := NewHandler(orgDir)
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
