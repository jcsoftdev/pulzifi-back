package acceptinvitation_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	domerrs "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories/mocks"
	acceptinvitation "github.com/jcsoftdev/pulzifi-back/modules/auth/application/accept_invitation"
	authmocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services/mocks"
)

func validRequest() acceptinvitation.Request {
	return acceptinvitation.Request{
		FirstName:             "Jane",
		LastName:              "Doe",
		Password:              "supersecret",
		OrganizationName:      "Acme Inc",
		OrganizationSubdomain: "ACME", // intentionally upper-case to verify normalization
	}
}

func TestAccept_HappyPath(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()
	var capturedInput repositories.AcceptInvitationInput

	repo := &mocks.InvitationRepositoryMock{
		AcceptInvitationFunc: func(ctx context.Context, in repositories.AcceptInvitationInput) (*repositories.AcceptInvitationOutput, error) {
			capturedInput = in
			return &repositories.AcceptInvitationOutput{
				UserID:        userID,
				OrgID:         orgID,
				OrgSubdomain:  in.OrgSubdomain,
				OrgSchemaName: "tenant_" + in.OrgSubdomain,
			}, nil
		},
	}
	auth := &authmocks.MockAuthService{HashResult: "hashed-pw"}

	provisionCalled := false
	provision := func(schema string) error {
		provisionCalled = true
		return nil
	}

	h := acceptinvitation.New(repo, auth, provision)
	resp, err := h.Handle(context.Background(), "tok-good", validRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.UserID != userID.String() {
		t.Fatalf("UserID mismatch: got %s want %s", resp.UserID, userID.String())
	}
	if resp.OrgID != orgID.String() {
		t.Fatalf("OrgID mismatch: got %s want %s", resp.OrgID, orgID.String())
	}
	if resp.OrgSubdomain != "acme" {
		t.Fatalf("expected normalized subdomain 'acme', got %s", resp.OrgSubdomain)
	}

	// Verify the handler propagated the right input to the repo.
	if capturedInput.Token != "tok-good" {
		t.Fatalf("token not forwarded: got %s", capturedInput.Token)
	}
	if capturedInput.PasswordHash != "hashed-pw" {
		t.Fatalf("expected hashed password, got %s", capturedInput.PasswordHash)
	}
	if capturedInput.OrgSubdomain != "acme" {
		t.Fatalf("expected lowercased subdomain forwarded, got %s", capturedInput.OrgSubdomain)
	}
	if capturedInput.ProvisionFunc == nil {
		t.Fatalf("ProvisionFunc must be forwarded to repo")
	}
	// Sanity: invoking the forwarded provision callback should reach our func.
	if err := capturedInput.ProvisionFunc("tenant_acme"); err != nil {
		t.Fatalf("provision returned err: %v", err)
	}
	if !provisionCalled {
		t.Fatalf("expected provision callback to be invoked through forwarded reference")
	}
}

func TestAccept_Expired_Errors(t *testing.T) {
	repo := &mocks.InvitationRepositoryMock{
		AcceptInvitationFunc: func(ctx context.Context, in repositories.AcceptInvitationInput) (*repositories.AcceptInvitationOutput, error) {
			return nil, domerrs.ErrInvitationNotFound
		},
	}
	auth := &authmocks.MockAuthService{HashResult: "hashed-pw"}
	h := acceptinvitation.New(repo, auth, func(string) error { return nil })

	_, err := h.Handle(context.Background(), "tok-expired", validRequest())
	if !errors.Is(err, domerrs.ErrInvitationNotFound) {
		t.Fatalf("expected ErrInvitationNotFound, got %v", err)
	}
}

func TestAccept_SubdomainTaken_Errors(t *testing.T) {
	repo := &mocks.InvitationRepositoryMock{
		AcceptInvitationFunc: func(ctx context.Context, in repositories.AcceptInvitationInput) (*repositories.AcceptInvitationOutput, error) {
			return nil, domerrs.ErrSubdomainTaken
		},
	}
	auth := &authmocks.MockAuthService{HashResult: "hashed-pw"}
	h := acceptinvitation.New(repo, auth, func(string) error { return nil })

	_, err := h.Handle(context.Background(), "tok", validRequest())
	if !errors.Is(err, domerrs.ErrSubdomainTaken) {
		t.Fatalf("expected ErrSubdomainTaken, got %v", err)
	}
}

func TestAccept_SchemaProvisioningFailed_Errors(t *testing.T) {
	repo := &mocks.InvitationRepositoryMock{
		AcceptInvitationFunc: func(ctx context.Context, in repositories.AcceptInvitationInput) (*repositories.AcceptInvitationOutput, error) {
			return nil, domerrs.ErrSchemaProvisioning
		},
	}
	auth := &authmocks.MockAuthService{HashResult: "hashed-pw"}
	h := acceptinvitation.New(repo, auth, func(string) error { return nil })

	_, err := h.Handle(context.Background(), "tok", validRequest())
	if !errors.Is(err, domerrs.ErrSchemaProvisioning) {
		t.Fatalf("expected ErrSchemaProvisioning, got %v", err)
	}
}
