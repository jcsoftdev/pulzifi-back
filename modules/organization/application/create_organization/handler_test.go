package create_organization

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/domain/events"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/domain/repositories/mocks"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
)

// mockEventPublisher is a hand-rolled mock for EventPublisher.
type mockEventPublisher struct {
	PublishErr error
	PublishCalls int
}

func (m *mockEventPublisher) PublishOrganizationCreated(_ context.Context, _ *events.OrganizationCreated) error {
	m.PublishCalls++
	return m.PublishErr
}

func TestCreateOrganizationHandler_Handle(t *testing.T) {
	userID := uuid.New()
	svc := services.NewOrganizationService()

	tests := []struct {
		name            string
		req             *Request
		countBySubdomain int
		countErr        error
		createErr       error
		publishErr      error
		wantErr         bool
		wantErrContains string
	}{
		{
			name:             "happy path — org created",
			req:              &Request{Name: "Acme Corp", Subdomain: "acme"},
			countBySubdomain: 0,
			wantErr:          false,
		},
		{
			name:             "subdomain conflict — ErrAlreadyExists-like",
			req:              &Request{Name: "Acme Corp", Subdomain: "acme"},
			countBySubdomain: 1,
			wantErr:          true,
			wantErrContains:  "subdomain already exists",
		},
		{
			name:    "repo count error propagated",
			req:     &Request{Name: "Acme Corp", Subdomain: "acme"},
			countErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name:      "repo create error propagated",
			req:       &Request{Name: "Acme Corp", Subdomain: "acme"},
			createErr: errors.New("insert failed"),
			wantErr:   true,
		},
		{
			name:    "invalid name returns validation error",
			req:     &Request{Name: "A", Subdomain: "acme"},
			wantErr: true,
		},
		{
			name:    "invalid subdomain returns validation error",
			req:     &Request{Name: "Acme Corp", Subdomain: "ab"},
			wantErr: true,
		},
		{
			name:       "publisher error does not fail the request",
			req:        &Request{Name: "Acme Corp", Subdomain: "acme"},
			publishErr: errors.New("publish failed"),
			wantErr:    false, // per handler: publisher error is logged but not propagated
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mocks.MockOrganizationRepository{
				CountBySubdomainResult: tt.countBySubdomain,
				CountBySubdomainErr:    tt.countErr,
				CreateErr:              tt.createErr,
			}
			publisher := &mockEventPublisher{PublishErr: tt.publishErr}
			handler := NewCreateOrganizationHandler(repo, svc, publisher)

			// Act
			resp, err := handler.Handle(context.Background(), tt.req, userID)

			// Assert
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrContains != "" && !containsStr(err.Error(), tt.wantErrContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.wantErrContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp == nil {
				t.Fatal("expected non-nil response")
			}
			if resp.ID == uuid.Nil {
				t.Error("response ID should not be nil UUID")
			}
			if resp.Name != tt.req.Name {
				t.Errorf("Name: want %q, got %q", tt.req.Name, resp.Name)
			}
			if resp.Subdomain != tt.req.Subdomain {
				t.Errorf("Subdomain: want %q, got %q", tt.req.Subdomain, resp.Subdomain)
			}
		})
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
