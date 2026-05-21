package listmembers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/team/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/team/domain/repositories/mocks"
)

func TestListMembersHandler_Handle(t *testing.T) {
	orgID := uuid.New()
	subdomain := "acme"
	now := time.Now()

	members := []*entities.TeamMember{
		{
			ID:               uuid.New(),
			OrganizationID:   orgID,
			UserID:           uuid.New(),
			Role:             "OWNER",
			FirstName:        "Alice",
			LastName:         "Smith",
			Email:            "alice@example.com",
			JoinedAt:         now,
			InvitationStatus: entities.InvitationStatusActive,
		},
		{
			ID:               uuid.New(),
			OrganizationID:   orgID,
			UserID:           uuid.New(),
			Role:             "MEMBER",
			FirstName:        "Bob",
			LastName:         "Jones",
			Email:            "bob@example.com",
			JoinedAt:         now,
			InvitationStatus: entities.InvitationStatusActive,
		},
	}

	tests := []struct {
		name              string
		orgIDResult       uuid.UUID
		orgIDErr          error
		listResult        []*entities.TeamMember
		listErr           error
		wantErr           bool
		wantCount         int
	}{
		{
			name:        "happy path returns member slice",
			orgIDResult: orgID,
			listResult:  members,
			wantErr:     false,
			wantCount:   2,
		},
		{
			name:        "empty list — no error",
			orgIDResult: orgID,
			listResult:  []*entities.TeamMember{},
			wantErr:     false,
			wantCount:   0,
		},
		{
			name:    "GetOrganizationIDBySubdomain error propagated",
			orgIDErr: errors.New("subdomain not found"),
			wantErr: true,
		},
		{
			name:        "ListByOrganization error propagated",
			orgIDResult: orgID,
			listErr:     errors.New("db error"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mocks.MockTeamMemberRepository{
				GetOrgIDBySubdomainResult: tt.orgIDResult,
				GetOrgIDBySubdomainErr:    tt.orgIDErr,
				ListByOrganizationResult:  tt.listResult,
				ListByOrganizationErr:     tt.listErr,
			}
			handler := NewListMembersHandler(repo)

			// Act
			resp, err := handler.Handle(context.Background(), subdomain)

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
			if len(resp.Members) != tt.wantCount {
				t.Errorf("expected %d members, got %d", tt.wantCount, len(resp.Members))
			}
		})
	}
}
