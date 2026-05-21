package removemember

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/team/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/team/domain/repositories/mocks"
)

func TestRemoveMemberHandler_Handle(t *testing.T) {
	memberID := uuid.New()
	requesterID := uuid.New()
	ownerID := uuid.New()

	regularMember := &entities.TeamMember{
		ID:     memberID,
		UserID: uuid.New(), // different from requesterID
		Role:   "MEMBER",
	}

	ownerMember := &entities.TeamMember{
		ID:     memberID,
		UserID: ownerID,
		Role:   "OWNER",
	}

	selfMember := &entities.TeamMember{
		ID:     memberID,
		UserID: requesterID, // same as requesterID
		Role:   "MEMBER",
	}

	tests := []struct {
		name        string
		getResult   *entities.TeamMember
		getErr      error
		removeErr   error
		wantErr     bool
		wantErrIs   error
	}{
		{
			name:      "happy path — member removed",
			getResult: regularMember,
			wantErr:   false,
		},
		{
			name:      "member not found returns ErrMemberNotFound",
			getResult: nil,
			wantErr:   true,
			wantErrIs: ErrMemberNotFound,
		},
		{
			name:    "GetByID error returns ErrMemberNotFound",
			getErr:  errors.New("db error"),
			wantErr: true,
			wantErrIs: ErrMemberNotFound,
		},
		{
			name:      "cannot remove owner",
			getResult: ownerMember,
			wantErr:   true,
			wantErrIs: ErrCannotRemoveOwner,
		},
		{
			name:      "cannot remove self",
			getResult: selfMember,
			wantErr:   true,
			wantErrIs: ErrCannotRemoveSelf,
		},
		{
			name:      "repo remove error propagated",
			getResult: regularMember,
			removeErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mocks.MockTeamMemberRepository{
				GetByIDResult: tt.getResult,
				GetByIDErr:    tt.getErr,
				RemoveErr:     tt.removeErr,
			}
			handler := NewRemoveMemberHandler(repo)

			// Act
			err := handler.Handle(context.Background(), memberID, requesterID)

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
		})
	}
}
