package get_current_organization

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/domain/repositories/mocks"
)

func TestGetCurrentOrganizationHandler_Handle(t *testing.T) {
	now := time.Now()
	orgID := uuid.New()
	ownerID := uuid.New()

	org := &entities.Organization{
		ID:          orgID,
		Name:        "Acme Corp",
		Subdomain:   "acme",
		SchemaName:  "acme",
		OwnerUserID: ownerID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	tests := []struct {
		name       string
		subdomain  string
		repoResult *entities.Organization
		repoErr    error
		wantErr    bool
		wantNil    bool
	}{
		{
			name:       "happy path returns org response",
			subdomain:  "acme",
			repoResult: org,
			wantErr:    false,
			wantNil:    false,
		},
		{
			name:      "org not found returns nil (no error)",
			subdomain: "unknown",
			wantErr:   false,
			wantNil:   true,
		},
		{
			name:    "repo error propagated",
			subdomain: "acme",
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mocks.MockOrganizationRepository{
				GetBySubdomainResult: tt.repoResult,
				GetBySubdomainErr:    tt.repoErr,
			}
			handler := NewHandler(repo)

			// Act
			resp, err := handler.Handle(context.Background(), tt.subdomain)

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
			if tt.wantNil {
				if resp != nil {
					t.Errorf("expected nil response, got %+v", resp)
				}
				return
			}
			if resp == nil {
				t.Fatal("expected non-nil response")
			}
			if resp.Name != org.Name {
				t.Errorf("Name: want %q, got %q", org.Name, resp.Name)
			}
		})
	}
}
