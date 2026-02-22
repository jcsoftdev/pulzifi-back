package invitemember

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/team/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/team/domain/repositories/mocks"
)

func TestInviteMemberHandler_Handle(t *testing.T) {
	orgID := uuid.New()
	inviterID := uuid.New()
	existingUserID := uuid.New()
	newUserID := uuid.New()
	memberID := uuid.New()

	existingUser := &entities.TeamMember{
		ID:     memberID,
		UserID: existingUserID,
		Email:  "existing@example.com",
	}

	addedMember := &entities.TeamMember{
		ID:       memberID,
		UserID:   existingUserID,
		Role:     "MEMBER",
		Email:    "existing@example.com",
		JoinedAt: time.Now(),
	}

	tests := []struct {
		name       string
		subdomain  string
		req        *InviteMemberRequest
		setupMocks func(repo *mocks.MockTeamMemberRepository)
		wantErr    error
	}{
		{
			name:      "invite existing user",
			subdomain: "acme",
			req:       &InviteMemberRequest{Email: "existing@example.com", Role: "member"},
			setupMocks: func(repo *mocks.MockTeamMemberRepository) {
				repo.GetOrgIDBySubdomainResult = orgID
				repo.FindUserByEmailResult = existingUser
				repo.GetByUserAndOrgErr = errors.New("not found")
				repo.AddMemberResult = addedMember
			},
			wantErr: nil,
		},
		{
			name:      "auto-create new user",
			subdomain: "acme",
			req:       &InviteMemberRequest{Email: "new@example.com", Role: "ADMIN"},
			setupMocks: func(repo *mocks.MockTeamMemberRepository) {
				repo.GetOrgIDBySubdomainResult = orgID
				repo.FindUserByEmailResult = nil // user not found
				repo.CreateUserResult = newUserID
				repo.GetByUserAndOrgErr = errors.New("not found")
				repo.AddMemberResult = &entities.TeamMember{
					ID:       uuid.New(),
					UserID:   newUserID,
					Role:     "ADMIN",
					Email:    "new@example.com",
					JoinedAt: time.Now(),
				}
			},
			wantErr: nil,
		},
		{
			name:      "already member",
			subdomain: "acme",
			req:       &InviteMemberRequest{Email: "existing@example.com", Role: "member"},
			setupMocks: func(repo *mocks.MockTeamMemberRepository) {
				repo.GetOrgIDBySubdomainResult = orgID
				repo.FindUserByEmailResult = existingUser
				repo.GetByUserAndOrgResult = &entities.TeamMember{ID: memberID}
			},
			wantErr: ErrAlreadyMember,
		},
		{
			name:      "organization not found",
			subdomain: "nonexistent",
			req:       &InviteMemberRequest{Email: "test@example.com", Role: "member"},
			setupMocks: func(repo *mocks.MockTeamMemberRepository) {
				repo.GetOrgIDBySubdomainErr = errors.New("not found")
			},
			wantErr: ErrOrganizationNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockTeamMemberRepository{}
			if tt.setupMocks != nil {
				tt.setupMocks(repo)
			}

			handler := NewInviteMemberHandler(repo, nil)
			resp, err := handler.Handle(context.Background(), tt.subdomain, inviterID, tt.req)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("want error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp == nil {
				t.Fatal("expected non-nil response")
			}
			if resp.UserID == uuid.Nil {
				t.Error("user ID should not be nil")
			}
		})
	}
}
