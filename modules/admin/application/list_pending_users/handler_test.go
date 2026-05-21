package listpendingusers

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories/mocks"
	adminservices "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/services"
)

// mockPendingUserReader is a local hand-rolled mock for adminservices.PendingUserReader.
type mockPendingUserReader struct {
	GetByIDResult *adminservices.PendingUser
	GetByIDErr    error
}

func (m *mockPendingUserReader) GetByID(_ context.Context, _ uuid.UUID) (*adminservices.PendingUser, error) {
	return m.GetByIDResult, m.GetByIDErr
}

func TestListPendingUsersHandler_Handle(t *testing.T) {
	userID := uuid.New()
	requestID := uuid.New()

	pendingUser := &adminservices.PendingUser{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	pendingRequests := []*entities.RegistrationRequest{
		{
			ID:                    requestID,
			UserID:                userID,
			OrganizationName:      "Acme Corp",
			OrganizationSubdomain: "acme",
			Status:                entities.RegistrationStatusPending,
		},
	}

	tests := []struct {
		name             string
		listPendingResult []*entities.RegistrationRequest
		listPendingErr   error
		userReaderResult *adminservices.PendingUser
		userReaderErr    error
		wantErr          bool
		wantCount        int
	}{
		{
			name:             "happy path — returns slice of pending DTOs",
			listPendingResult: pendingRequests,
			userReaderResult: pendingUser,
			wantErr:          false,
			wantCount:        1,
		},
		{
			name:             "empty list — no error",
			listPendingResult: []*entities.RegistrationRequest{},
			userReaderResult: pendingUser,
			wantErr:          false,
			wantCount:        0,
		},
		{
			name:           "repo error propagated",
			listPendingErr: errors.New("db error"),
			wantErr:        true,
		},
		{
			name:             "user reader returns nil — skipped (no panic)",
			listPendingResult: pendingRequests,
			userReaderResult: nil,
			wantErr:          false,
			wantCount:        0, // skipped because user is nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mocks.MockRegistrationRequestRepository{
				ListPendingResult: tt.listPendingResult,
				ListPendingErr:    tt.listPendingErr,
			}
			userReader := &mockPendingUserReader{
				GetByIDResult: tt.userReaderResult,
				GetByIDErr:    tt.userReaderErr,
			}
			handler := NewHandler(repo, userReader)

			// Act
			resp, err := handler.Handle(context.Background(), 10, 0)

			// Assert
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp == nil {
				t.Fatal("expected non-nil response")
			}
			if len(resp.PendingUsers) != tt.wantCount {
				t.Errorf("expected %d pending users, got %d", tt.wantCount, len(resp.PendingUsers))
			}
		})
	}
}
