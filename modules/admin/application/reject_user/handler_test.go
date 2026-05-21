package rejectuser

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

// mockRejectionProvisioner is a local hand-rolled mock for adminservices.RejectionProvisioner.
type mockRejectionProvisioner struct {
	RejectErr   error
	RejectFn    func(ctx context.Context, input adminservices.RejectionInput) error
	RejectCalls int
}

func (m *mockRejectionProvisioner) Reject(ctx context.Context, input adminservices.RejectionInput) error {
	m.RejectCalls++
	if m.RejectFn != nil {
		return m.RejectFn(ctx, input)
	}
	return m.RejectErr
}

func TestRejectUserHandler_Handle(t *testing.T) {
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
		name        string
		repoResult  *entities.RegistrationRequest
		repoErr     error
		rejectErr   error
		wantErr     bool
		wantErrIs   error
		wantRejectCalls int
	}{
		{
			name:            "happy path — request rejected",
			repoResult:      pendingReq,
			wantErr:         false,
			wantRejectCalls: 1,
		},
		{
			name:    "repo GetByID error propagated",
			repoErr: errors.New("db error"),
			wantErr: true,
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
				Status: entities.RegistrationStatusRejected,
			},
			wantErr:   true,
			wantErrIs: adminerrors.ErrAlreadyReviewed,
		},
		{
			name:            "provisioner error propagated",
			repoResult:      pendingReq,
			rejectErr:       errors.New("reject failed"),
			wantErr:         true,
			wantRejectCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mocks.MockRegistrationRequestRepository{
				GetByIDResult: tt.repoResult,
				GetByIDErr:    tt.repoErr,
			}
			provisioner := &mockRejectionProvisioner{
				RejectErr: tt.rejectErr,
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
			if provisioner.RejectCalls != tt.wantRejectCalls {
				t.Errorf("expected %d Reject calls, got %d", tt.wantRejectCalls, provisioner.RejectCalls)
			}
		})
	}
}
