package provisionorganization

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	autherrors "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/errors"
	servicemocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services/mocks"
)

// validRequest returns a fully-populated Request for success-path tests.
func validRequest() Request {
	return Request{
		UserID:    uuid.New(),
		OrgName:   "Acme Corp",
		Subdomain: "acme",
	}
}

func TestHandler_Handle_HappyPath(t *testing.T) {
	req := validRequest()
	checker := &servicemocks.MockOrganizationMembershipChecker{
		HasAnyMembershipResult: false,
	}
	provisioner := &servicemocks.MockTrialProvisioner{}

	h := NewHandler(provisioner, checker)
	resp, err := h.Handle(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.OrganizationSubdomain != req.Subdomain {
		t.Errorf("expected subdomain %q, got %q", req.Subdomain, resp.OrganizationSubdomain)
	}
	if resp.TrialEndsAt.IsZero() {
		t.Error("expected non-zero trial_ends_at")
	}
	if checker.HasAnyMembershipCalls != 1 {
		t.Errorf("expected 1 membership check call, got %d", checker.HasAnyMembershipCalls)
	}
	if checker.LastUserID != req.UserID {
		t.Errorf("membership checker received user_id %v, want %v", checker.LastUserID, req.UserID)
	}
	if provisioner.ProvisionCalls != 1 {
		t.Errorf("expected 1 Provision call, got %d", provisioner.ProvisionCalls)
	}
	if provisioner.LastInput.TrialDays != defaultTrialDays {
		t.Errorf("expected TrialDays %d, got %d", defaultTrialDays, provisioner.LastInput.TrialDays)
	}
	if provisioner.LastInput.UserID != req.UserID {
		t.Errorf("provisioner received user_id %v, want %v", provisioner.LastInput.UserID, req.UserID)
	}
}

func TestHandler_Handle_AlreadyProvisioned(t *testing.T) {
	req := validRequest()
	checker := &servicemocks.MockOrganizationMembershipChecker{
		HasAnyMembershipResult: true, // user already has a membership
	}
	provisioner := &servicemocks.MockTrialProvisioner{}

	h := NewHandler(provisioner, checker)
	resp, err := h.Handle(context.Background(), req)

	if resp != nil {
		t.Errorf("expected nil response when already provisioned, got %+v", resp)
	}
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var userErr autherrors.UserError
	if !errors.As(err, &userErr) {
		t.Fatalf("expected autherrors.UserError, got %T: %v", err, err)
	}
	if userErr.Code != "ALREADY_PROVISIONED" {
		t.Errorf("expected error code ALREADY_PROVISIONED, got %q", userErr.Code)
	}

	// Provisioner MUST NOT be called when user already has a membership.
	if provisioner.ProvisionCalls != 0 {
		t.Errorf("provisioner should not be called; got %d calls", provisioner.ProvisionCalls)
	}
}

func TestHandler_Handle_ProvisionerError(t *testing.T) {
	req := validRequest()
	checker := &servicemocks.MockOrganizationMembershipChecker{
		HasAnyMembershipResult: false,
	}
	provisionErr := errors.New("schema provisioning failed")
	provisioner := &servicemocks.MockTrialProvisioner{
		ProvisionErr: provisionErr,
	}

	h := NewHandler(provisioner, checker)
	resp, err := h.Handle(context.Background(), req)

	if resp != nil {
		t.Errorf("expected nil response on provisioner error, got %+v", resp)
	}
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, provisionErr) {
		t.Errorf("expected provisioner error %v to be propagated, got %v", provisionErr, err)
	}
	if provisioner.ProvisionCalls != 1 {
		t.Errorf("expected 1 Provision call, got %d", provisioner.ProvisionCalls)
	}
}

func TestHandler_Handle_MembershipCheckerError(t *testing.T) {
	req := validRequest()
	checkerErr := errors.New("db connection lost")
	checker := &servicemocks.MockOrganizationMembershipChecker{
		HasAnyMembershipErr: checkerErr,
	}
	provisioner := &servicemocks.MockTrialProvisioner{}

	h := NewHandler(provisioner, checker)
	resp, err := h.Handle(context.Background(), req)

	if resp != nil {
		t.Errorf("expected nil response on checker error, got %+v", resp)
	}
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, checkerErr) {
		t.Errorf("expected checker error %v to be propagated, got %v", checkerErr, err)
	}
	// Provisioner must not be called when membership check itself fails.
	if provisioner.ProvisionCalls != 0 {
		t.Errorf("provisioner should not be called on checker error; got %d calls", provisioner.ProvisionCalls)
	}
}

func TestHandler_Handle_SubdomainNormalized(t *testing.T) {
	// Subdomain with mixed case and spaces should be normalised to lowercase trimmed.
	req := Request{
		UserID:    uuid.New(),
		OrgName:   "My Org",
		Subdomain: "  MyOrg  ",
	}
	checker := &servicemocks.MockOrganizationMembershipChecker{HasAnyMembershipResult: false}
	provisioner := &servicemocks.MockTrialProvisioner{}

	h := NewHandler(provisioner, checker)
	_, err := h.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provisioner.LastInput.OrganizationSubdomain != "myorg" {
		t.Errorf("expected normalised subdomain %q, got %q", "myorg", provisioner.LastInput.OrganizationSubdomain)
	}
}

func TestHandler_Handle_TrialEndsAtInFuture(t *testing.T) {
	req := validRequest()
	expectedTrialEnd := time.Now().Add(14 * 24 * time.Hour)
	checker := &servicemocks.MockOrganizationMembershipChecker{HasAnyMembershipResult: false}
	// Let the mock produce a deterministic TrialEndsAt close to what the handler expects.
	provisioner := &servicemocks.MockTrialProvisioner{}

	h := NewHandler(provisioner, checker)
	resp, err := h.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// TrialEndsAt should be approximately 14 days from now (within a minute tolerance).
	diff := resp.TrialEndsAt.Sub(expectedTrialEnd)
	if diff < -time.Minute || diff > time.Minute {
		t.Errorf("trial_ends_at %v is not ~14 days from now (expected ~%v)", resp.TrialEndsAt, expectedTrialEnd)
	}
}

func TestHandler_Handle_WithOnboardingProfile_PersistsAnswers(t *testing.T) {
	orgID := uuid.New()
	req := Request{
		UserID:               uuid.New(),
		OrgName:              "Acme Corp",
		Subdomain:            "acme",
		CompanySize:          "11-50",
		BusinessType:         "SaaS / Software",
		CompetitorChallenges: []string{"I don't know what they're changing"},
		WebsiteURL:           "https://acme.io",
	}
	checker := &servicemocks.MockOrganizationMembershipChecker{HasAnyMembershipResult: false}
	provisioner := &servicemocks.MockTrialProvisioner{}
	writer := &servicemocks.MockOrganizationOnboardingWriter{}
	finder := &servicemocks.MockOrganizationOrgFinder{OrgID: &orgID}

	h := NewHandler(provisioner, checker).WithOnboardingProfile(writer, finder)
	resp, err := h.Handle(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if writer.SaveCalls != 1 {
		t.Errorf("expected 1 save call, got %d", writer.SaveCalls)
	}
	if writer.LastInput.CompanySize != req.CompanySize {
		t.Errorf("company_size: want %q, got %q", req.CompanySize, writer.LastInput.CompanySize)
	}
	if writer.LastInput.OrgID != orgID {
		t.Errorf("org_id: want %v, got %v", orgID, writer.LastInput.OrgID)
	}
	if writer.LastInput.OnboardingCompletedAt.IsZero() {
		t.Error("expected onboarding_completed_at to be set")
	}
}

func TestHandler_Handle_WithOnboardingProfile_NoAnswers_SkipsWrite(t *testing.T) {
	// When no answers are supplied, the profile step must be skipped.
	orgID := uuid.New()
	req := validRequest() // no onboarding fields
	checker := &servicemocks.MockOrganizationMembershipChecker{HasAnyMembershipResult: false}
	provisioner := &servicemocks.MockTrialProvisioner{}
	writer := &servicemocks.MockOrganizationOnboardingWriter{}
	finder := &servicemocks.MockOrganizationOrgFinder{OrgID: &orgID}

	h := NewHandler(provisioner, checker).WithOnboardingProfile(writer, finder)
	_, err := h.Handle(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if writer.SaveCalls != 0 {
		t.Errorf("writer should not be called when no answers provided; got %d calls", writer.SaveCalls)
	}
}

func TestHandler_Handle_WithOnboardingProfile_WriterError_NonFatal(t *testing.T) {
	// A writer error on the profile step must not fail the provision.
	orgID := uuid.New()
	req := Request{
		UserID:      uuid.New(),
		OrgName:     "Acme",
		Subdomain:   "acme",
		CompanySize: "1-10",
	}
	checker := &servicemocks.MockOrganizationMembershipChecker{HasAnyMembershipResult: false}
	provisioner := &servicemocks.MockTrialProvisioner{}
	writer := &servicemocks.MockOrganizationOnboardingWriter{SaveErr: errors.New("db error")}
	finder := &servicemocks.MockOrganizationOrgFinder{OrgID: &orgID}

	h := NewHandler(provisioner, checker).WithOnboardingProfile(writer, finder)
	resp, err := h.Handle(context.Background(), req)

	if err != nil {
		t.Fatalf("profile write error should be non-fatal, got: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response despite profile write error")
	}
}
