package register

import (
	"context"
	"errors"
	"testing"

	autherrors "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/errors"
	authmocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories/mocks"
	servicemocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services/mocks"
)

// orgValidationError is a local fake for organization validation errors.
// It avoids a cross-module import of organization/domain/services.
type orgValidationError struct{ msg string }

func (e *orgValidationError) Error() string { return "invalid organization data: " + e.msg }

func newOrgNameErr(msg string) error  { return &orgValidationError{msg: msg} }
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
		name          string
		req           *Request
		setupMocks    func(userRepo *authmocks.MockUserRepository, regReqWriter *servicemocks.MockRegistrationRequestWriter, orgDir *servicemocks.MockOrganizationDirectory)
		wantErr       bool
		wantErrCode   string
	}{
		{
			name: "successful registration",
			req:  validReq,
			setupMocks: func(userRepo *authmocks.MockUserRepository, regReqWriter *servicemocks.MockRegistrationRequestWriter, orgDir *servicemocks.MockOrganizationDirectory) {
				orgDir.CountBySubdomainResult = 0
				regReqWriter.ExistsPendingBySubdomainResult = false
				userRepo.ExistsByEmailResult = false
			},
			wantErr: false,
		},
		{
			name: "duplicate email returns USER_ALREADY_EXISTS",
			req:  validReq,
			setupMocks: func(userRepo *authmocks.MockUserRepository, regReqWriter *servicemocks.MockRegistrationRequestWriter, orgDir *servicemocks.MockOrganizationDirectory) {
				orgDir.CountBySubdomainResult = 0
				regReqWriter.ExistsPendingBySubdomainResult = false
				userRepo.ExistsByEmailResult = true
			},
			wantErr:     true,
			wantErrCode: "USER_ALREADY_EXISTS",
		},
		{
			name: "subdomain already taken by approved org",
			req:  validReq,
			setupMocks: func(userRepo *authmocks.MockUserRepository, regReqWriter *servicemocks.MockRegistrationRequestWriter, orgDir *servicemocks.MockOrganizationDirectory) {
				orgDir.CountBySubdomainResult = 1
			},
			wantErr:     true,
			wantErrCode: "SUBDOMAIN_TAKEN",
		},
		{
			name: "subdomain pending approval",
			req:  validReq,
			setupMocks: func(userRepo *authmocks.MockUserRepository, regReqWriter *servicemocks.MockRegistrationRequestWriter, orgDir *servicemocks.MockOrganizationDirectory) {
				orgDir.CountBySubdomainResult = 0
				regReqWriter.ExistsPendingBySubdomainResult = true
			},
			wantErr:     true,
			wantErrCode: "SUBDOMAIN_PENDING",
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
			setupMocks: func(userRepo *authmocks.MockUserRepository, regReqWriter *servicemocks.MockRegistrationRequestWriter, orgDir *servicemocks.MockOrganizationDirectory) {
				// Simulate real validation: empty org name fails
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
			setupMocks: func(userRepo *authmocks.MockUserRepository, regReqWriter *servicemocks.MockRegistrationRequestWriter, orgDir *servicemocks.MockOrganizationDirectory) {
				// Simulate real validation: subdomain "ab" is too short
				orgDir.ValidateSubdomainErr = newSubdomainErr("subdomain must be at least 3 characters")
			},
			wantErr:     true,
			wantErrCode: "INVALID_SUBDOMAIN",
		},
		{
			name: "user repo Create fails propagates error",
			req:  validReq,
			setupMocks: func(userRepo *authmocks.MockUserRepository, regReqWriter *servicemocks.MockRegistrationRequestWriter, orgDir *servicemocks.MockOrganizationDirectory) {
				orgDir.CountBySubdomainResult = 0
				regReqWriter.ExistsPendingBySubdomainResult = false
				userRepo.ExistsByEmailResult = false
				userRepo.CreateErr = errors.New("db connection lost")
			},
			wantErr: true,
		},
		{
			name: "registration request Create fails propagates error",
			req:  validReq,
			setupMocks: func(userRepo *authmocks.MockUserRepository, regReqWriter *servicemocks.MockRegistrationRequestWriter, orgDir *servicemocks.MockOrganizationDirectory) {
				orgDir.CountBySubdomainResult = 0
				regReqWriter.ExistsPendingBySubdomainResult = false
				userRepo.ExistsByEmailResult = false
				regReqWriter.CreateErr = errors.New("db connection lost")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &authmocks.MockUserRepository{}
			regReqWriter := &servicemocks.MockRegistrationRequestWriter{}
			orgDir := &servicemocks.MockOrganizationDirectory{}

			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, regReqWriter, orgDir)
			}

			handler := NewHandler(userRepo, regReqWriter, orgDir)
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
					// Some errors are not UserError (e.g. db errors); that's fine
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
			if resp.FirstName != tt.req.FirstName {
				t.Errorf("expected first name %q, got %q", tt.req.FirstName, resp.FirstName)
			}
			if resp.LastName != tt.req.LastName {
				t.Errorf("expected last name %q, got %q", tt.req.LastName, resp.LastName)
			}
			if resp.Status != "pending" {
				t.Errorf("expected status %q, got %q", "pending", resp.Status)
			}
			if resp.Message == "" {
				t.Error("expected non-empty message")
			}
			if userRepo.CreateCalls != 1 {
				t.Errorf("expected 1 Create call on user repo, got %d", userRepo.CreateCalls)
			}
			if regReqWriter.CreateCalls != 1 {
				t.Errorf("expected 1 Create call on registration request writer, got %d", regReqWriter.CreateCalls)
			}
		})
	}
}
