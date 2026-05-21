package register

import (
	"context"
	"errors"
	"testing"

	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	autherrors "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/errors"
	authmocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories/mocks"
	servicemocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services/mocks"
)

// orgValidationError is a local fake for organization validation errors.
// It avoids a cross-module import of organization/domain/services.
type orgValidationError struct{ msg string }

func (e *orgValidationError) Error() string { return "invalid organization data: " + e.msg }

func newOrgNameErr(msg string) error   { return &orgValidationError{msg: msg} }
func newSubdomainErr(msg string) error { return &orgValidationError{msg: msg} }

func TestHandler_Handle(t *testing.T) {
	validReq := &Request{
		Email:                 "alice@example.com",
		Password:              "strongPassword1!",
		FirstName:             "Alice",
		LastName:              "Smith",
		OrganizationName:      "Acme Corp",
		OrganizationSubdomain: "acme-corp",
	}

	tests := []struct {
		name        string
		req         *Request
		setupMocks  func(userRepo *authmocks.MockUserRepository, prov *servicemocks.MockTrialProvisioner, orgDir *servicemocks.MockOrganizationDirectory)
		wantErr     bool
		wantErrCode string
	}{
		{
			name: "successful self-serve trial registration",
			req:  validReq,
			setupMocks: func(userRepo *authmocks.MockUserRepository, _ *servicemocks.MockTrialProvisioner, orgDir *servicemocks.MockOrganizationDirectory) {
				orgDir.CountBySubdomainResult = 0
				userRepo.ExistsByEmailResult = false
			},
			wantErr: false,
		},
		{
			name: "duplicate email returns USER_ALREADY_EXISTS",
			req:  validReq,
			setupMocks: func(userRepo *authmocks.MockUserRepository, _ *servicemocks.MockTrialProvisioner, orgDir *servicemocks.MockOrganizationDirectory) {
				orgDir.CountBySubdomainResult = 0
				userRepo.ExistsByEmailResult = true
			},
			wantErr:     true,
			wantErrCode: "USER_ALREADY_EXISTS",
		},
		{
			name: "subdomain already taken by approved org",
			req:  validReq,
			setupMocks: func(_ *authmocks.MockUserRepository, _ *servicemocks.MockTrialProvisioner, orgDir *servicemocks.MockOrganizationDirectory) {
				orgDir.CountBySubdomainResult = 1
			},
			wantErr:     true,
			wantErrCode: "SUBDOMAIN_TAKEN",
		},
		{
			name: "invalid organization name (empty)",
			req: &Request{
				Email:                 "bob@example.com",
				Password:              "strongPassword1!",
				FirstName:             "Bob",
				LastName:              "Jones",
				OrganizationName:      "",
				OrganizationSubdomain: "bobcorp",
			},
			setupMocks: func(_ *authmocks.MockUserRepository, _ *servicemocks.MockTrialProvisioner, orgDir *servicemocks.MockOrganizationDirectory) {
				orgDir.ValidateOrganizationNameErr = newOrgNameErr("name cannot be empty")
			},
			wantErr:     true,
			wantErrCode: "INVALID_ORG_NAME",
		},
		{
			name: "invalid subdomain (too short)",
			req: &Request{
				Email:                 "bob@example.com",
				Password:              "strongPassword1!",
				FirstName:             "Bob",
				LastName:              "Jones",
				OrganizationName:      "Bob Corp",
				OrganizationSubdomain: "ab",
			},
			setupMocks: func(_ *authmocks.MockUserRepository, _ *servicemocks.MockTrialProvisioner, orgDir *servicemocks.MockOrganizationDirectory) {
				orgDir.ValidateSubdomainErr = newSubdomainErr("subdomain must be at least 3 characters")
			},
			wantErr:     true,
			wantErrCode: "INVALID_SUBDOMAIN",
		},
		{
			name: "user repo Create fails propagates error",
			req:  validReq,
			setupMocks: func(userRepo *authmocks.MockUserRepository, _ *servicemocks.MockTrialProvisioner, orgDir *servicemocks.MockOrganizationDirectory) {
				orgDir.CountBySubdomainResult = 0
				userRepo.ExistsByEmailResult = false
				userRepo.CreateErr = errors.New("db connection lost")
			},
			wantErr: true,
		},
		{
			name: "trial provisioner fails propagates error",
			req:  validReq,
			setupMocks: func(userRepo *authmocks.MockUserRepository, prov *servicemocks.MockTrialProvisioner, orgDir *servicemocks.MockOrganizationDirectory) {
				orgDir.CountBySubdomainResult = 0
				userRepo.ExistsByEmailResult = false
				prov.ProvisionErr = errors.New("schema provisioning failed")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &authmocks.MockUserRepository{}
			prov := &servicemocks.MockTrialProvisioner{}
			orgDir := &servicemocks.MockOrganizationDirectory{}

			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, prov, orgDir)
			}

			handler := NewHandler(userRepo, prov, orgDir, 14)
			resp, err := handler.Handle(context.Background(), tt.req)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrCode != "" {
					var userErr autherrors.UserError
					if errors.As(err, &userErr) {
						if userErr.Code != tt.wantErrCode {
							t.Errorf("expected error code %q, got %q", tt.wantErrCode, userErr.Code)
						}
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp == nil {
				t.Fatal("expected non-nil response")
			}
			if resp.Email != tt.req.Email {
				t.Errorf("expected email %q, got %q", tt.req.Email, resp.Email)
			}
			if resp.Status != entities.UserStatusApproved {
				t.Errorf("expected status %q, got %q", entities.UserStatusApproved, resp.Status)
			}
			if resp.OrganizationSubdomain != tt.req.OrganizationSubdomain {
				t.Errorf("expected org_subdomain %q, got %q", tt.req.OrganizationSubdomain, resp.OrganizationSubdomain)
			}
			if resp.TrialEndsAt.IsZero() {
				t.Error("expected non-zero trial_ends_at on response")
			}
			if userRepo.CreateCalls != 1 {
				t.Errorf("expected 1 Create call on user repo, got %d", userRepo.CreateCalls)
			}
			if prov.ProvisionCalls != 1 {
				t.Errorf("expected 1 Provision call on trial provisioner, got %d", prov.ProvisionCalls)
			}
			if prov.LastInput.OrganizationSubdomain != tt.req.OrganizationSubdomain {
				t.Errorf("provisioner got subdomain %q, want %q", prov.LastInput.OrganizationSubdomain, tt.req.OrganizationSubdomain)
			}
			if prov.LastInput.TrialDays != 14 {
				t.Errorf("provisioner got trial_days %d, want 14", prov.LastInput.TrialDays)
			}
		})
	}
}
