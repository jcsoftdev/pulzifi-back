package get_organization

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/domain/repositories/mocks"
)

func TestGetOrganizationHandler_Handle(t *testing.T) {
	orgID := uuid.New()
	ownerID := uuid.New()
	now := time.Now()

	activeOrg := &entities.Organization{
		ID:          orgID,
		Name:        "Acme Corp",
		Subdomain:   "acme",
		SchemaName:  "acme",
		OwnerUserID: ownerID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	deletedOrg := &entities.Organization{
		ID:          orgID,
		Name:        "Acme Corp",
		Subdomain:   "acme",
		SchemaName:  "acme",
		OwnerUserID: ownerID,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   &now,
	}

	tests := []struct {
		name            string
		repoResult      *entities.Organization
		repoErr         error
		wantErr         bool
		wantErrContains string
	}{
		{
			name:       "happy path — found returns DTO",
			repoResult: activeOrg,
			wantErr:    false,
		},
		{
			name:            "not found returns ErrOrganizationNotFound",
			repoResult:      nil,
			wantErr:         true,
			wantErrContains: "organization not found",
		},
		{
			name:            "deleted org returns ErrOrganizationAlreadyDeleted",
			repoResult:      deletedOrg,
			wantErr:         true,
			wantErrContains: "organization already deleted",
		},
		{
			name:    "repo error propagated",
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mocks.MockOrganizationRepository{
				GetByIDResult: tt.repoResult,
				GetByIDErr:    tt.repoErr,
			}
			handler := NewGetOrganizationHandler(repo)

			// Act
			resp, err := handler.Handle(context.Background(), orgID)

			// Assert
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrContains != "" {
					errStr := err.Error()
					found := false
					for i := 0; i <= len(errStr)-len(tt.wantErrContains); i++ {
						if errStr[i:i+len(tt.wantErrContains)] == tt.wantErrContains {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("error %q does not contain %q", errStr, tt.wantErrContains)
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
			if resp.ID != orgID {
				t.Errorf("ID: want %v, got %v", orgID, resp.ID)
			}
			if resp.Name != activeOrg.Name {
				t.Errorf("Name: want %q, got %q", activeOrg.Name, resp.Name)
			}
		})
	}
}
