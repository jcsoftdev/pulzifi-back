package approveuser

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/entities"
	adminerrors "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories/mocks"
	adminservices "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/services"
)

// mockApprovalProvisioner is a local hand-rolled mock for adminservices.ApprovalProvisioner.
type mockApprovalProvisioner struct {
	ProvisionErr error
	ProvisionFn  func(ctx context.Context, input adminservices.ApprovalInput) error
	ProvisionCalls int
}

func (m *mockApprovalProvisioner) Provision(ctx context.Context, input adminservices.ApprovalInput) error {
	m.ProvisionCalls++
	if m.ProvisionFn != nil {
		return m.ProvisionFn(ctx, input)
	}
	return m.ProvisionErr
}

func TestApproveUserHandler_Handle(t *testing.T) {
	requestID := uuid.New()
	reviewerID := uuid.New()
	userID := uuid.New()

	pendingReq := &entities.RegistrationRequest{
		ID:                    requestID,
		UserID:                userID,
		OrganizationName:      "Acme Corp",
		OrganizationSubdomain: "acme",
		Status:                entities.RegistrationStatusPending,
	}

	tests := []struct {
		name         string
		repoResult   *entities.RegistrationRequest
		repoErr      error
		provisionErr error
		wantErr      bool
		wantErrIs    error
	}{
		{
			name:       "happy path — request approved",
			repoResult: pendingReq,
			wantErr:    false,
		},
		{
			name:      "repo GetByID error propagated",
			repoErr:   errors.New("db error"),
			wantErr:   true,
		},
		{
			name:      "request not found returns ErrRegistrationRequestNotFound",
			repoResult: nil,
			wantErr:   true,
			wantErrIs: adminerrors.ErrRegistrationRequestNotFound,
		},
		{
			name: "already reviewed returns ErrAlreadyReviewed",
			repoResult: &entities.RegistrationRequest{
				ID:     requestID,
				UserID: userID,
				Status: entities.RegistrationStatusApproved,
			},
			wantErr:   true,
			wantErrIs: adminerrors.ErrAlreadyReviewed,
		},
		{
			name:         "provisioner error propagated",
			repoResult:   pendingReq,
			provisionErr: errors.New("provision failed"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mocks.MockRegistrationRequestRepository{
				GetByIDResult: tt.repoResult,
				GetByIDErr:    tt.repoErr,
			}
			provisioner := &mockApprovalProvisioner{
				ProvisionErr: tt.provisionErr,
			}
			handler := NewHandler(repo, provisioner)

			// Act
			err := handler.Handle(context.Background(), requestID, reviewerID)

			// Assert
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("error = %v, want %v", err, tt.wantErrIs)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if provisioner.ProvisionCalls != 1 {
				t.Errorf("expected 1 Provision call, got %d", provisioner.ProvisionCalls)
			}
		})
	}
}
