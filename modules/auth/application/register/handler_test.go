package register

import (
	"context"
	"errors"
	"testing"

	autherrors "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/errors"
	adminmocks "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories/mocks"
	authmocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories/mocks"
	orgmocks "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/repositories/mocks"
	orgservices "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
)

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
		setupMocks  func(userRepo *authmocks.MockUserRepository, regReqRepo *adminmocks.MockRegistrationRequestRepository, orgRepo *orgmocks.MockOrganizationRepository)
		wantErr     bool
		wantErrCode string
	}{
		{
			name: "successful registration",
			req:  validReq,
			setupMocks: func(userRepo *authmocks.MockUserRepository, regReqRepo *adminmocks.MockRegistrationRequestRepository, orgRepo *orgmocks.MockOrganizationRepository) {
				orgRepo.CountBySubdomainResult = 0
				regReqRepo.ExistsPendingBySubdomainResult = false
				userRepo.ExistsByEmailResult = false
			},
			wantErr: false,
		},
		{
			name: "duplicate email returns USER_ALREADY_EXISTS",
			req:  validReq,
			setupMocks: func(userRepo *authmocks.MockUserRepository, regReqRepo *adminmocks.MockRegistrationRequestRepository, orgRepo *orgmocks.MockOrganizationRepository) {
				orgRepo.CountBySubdomainResult = 0
				regReqRepo.ExistsPendingBySubdomainResult = false
				userRepo.ExistsByEmailResult = true
			},
			wantErr:     true,
			wantErrCode: "USER_ALREADY_EXISTS",
		},
		{
			name: "subdomain already taken by approved org",
			req:  validReq,
			setupMocks: func(userRepo *authmocks.MockUserRepository, regReqRepo *adminmocks.MockRegistrationRequestRepository, orgRepo *orgmocks.MockOrganizationRepository) {
				orgRepo.CountBySubdomainResult = 1
			},
			wantErr:     true,
			wantErrCode: "SUBDOMAIN_TAKEN",
		},
		{
			name: "subdomain pending approval",
			req:  validReq,
			setupMocks: func(userRepo *authmocks.MockUserRepository, regReqRepo *adminmocks.MockRegistrationRequestRepository, orgRepo *orgmocks.MockOrganizationRepository) {
				orgRepo.CountBySubdomainResult = 0
				regReqRepo.ExistsPendingBySubdomainResult = true
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
			setupMocks: func(userRepo *authmocks.MockUserRepository, regReqRepo *adminmocks.MockRegistrationRequestRepository, orgRepo *orgmocks.MockOrganizationRepository) {
				// No mock setup needed; validation fails before any repo call
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
			setupMocks: func(userRepo *authmocks.MockUserRepository, regReqRepo *adminmocks.MockRegistrationRequestRepository, orgRepo *orgmocks.MockOrganizationRepository) {
			},
			wantErr:     true,
			wantErrCode: "INVALID_SUBDOMAIN",
		},
		{
			name: "user repo Create fails propagates error",
			req:  validReq,
			setupMocks: func(userRepo *authmocks.MockUserRepository, regReqRepo *adminmocks.MockRegistrationRequestRepository, orgRepo *orgmocks.MockOrganizationRepository) {
				orgRepo.CountBySubdomainResult = 0
				regReqRepo.ExistsPendingBySubdomainResult = false
				userRepo.ExistsByEmailResult = false
				userRepo.CreateErr = errors.New("db connection lost")
			},
			wantErr: true,
		},
		{
			name: "registration request Create fails propagates error",
			req:  validReq,
			setupMocks: func(userRepo *authmocks.MockUserRepository, regReqRepo *adminmocks.MockRegistrationRequestRepository, orgRepo *orgmocks.MockOrganizationRepository) {
				orgRepo.CountBySubdomainResult = 0
				regReqRepo.ExistsPendingBySubdomainResult = false
				userRepo.ExistsByEmailResult = false
				regReqRepo.CreateErr = errors.New("db connection lost")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &authmocks.MockUserRepository{}
			regReqRepo := &adminmocks.MockRegistrationRequestRepository{}
			orgRepo := &orgmocks.MockOrganizationRepository{}
			orgService := orgservices.NewOrganizationService()

			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, regReqRepo, orgRepo)
			}

			handler := NewHandler(userRepo, regReqRepo, orgRepo, orgService)
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
			if regReqRepo.CreateCalls != 1 {
				t.Errorf("expected 1 Create call on registration request repo, got %d", regReqRepo.CreateCalls)
			}
		})
	}
}
