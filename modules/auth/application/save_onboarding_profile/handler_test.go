package saveonboardingprofile

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
)

// --- in-memory stubs ---

type stubOnboardingWriter struct {
	savedInput *authservices.OnboardingProfileInput
	saveErr    error
	savedCalls int
}

func (s *stubOnboardingWriter) SaveOnboardingProfile(ctx context.Context, input authservices.OnboardingProfileInput) error {
	s.savedCalls++
	s.savedInput = &input
	return s.saveErr
}

type stubOrgFinder struct {
	orgID    *uuid.UUID
	findErr  error
	findCalls int
}

func (s *stubOrgFinder) GetUserOrgID(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error) {
	s.findCalls++
	return s.orgID, s.findErr
}

// --- tests ---

func TestHandler_Handle_Success(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()

	writer := &stubOnboardingWriter{}
	finder := &stubOrgFinder{orgID: &orgID}
	h := NewHandler(finder, writer)

	req := Request{
		UserID:                userID,
		CompanySize:           "11-50",
		BusinessType:          "SaaS / Software",
		CompetitorChallenges:  []string{"I don't know what they're changing", "I react too late to their moves"},
		WebsiteURL:            "https://example.com",
	}

	resp, err := h.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	if writer.savedCalls != 1 {
		t.Errorf("expected 1 save call, got %d", writer.savedCalls)
	}
	if writer.savedInput == nil {
		t.Fatal("expected savedInput to be set")
	}
	if writer.savedInput.OrgID != orgID {
		t.Errorf("expected org ID %v, got %v", orgID, writer.savedInput.OrgID)
	}
	if writer.savedInput.CompanySize != req.CompanySize {
		t.Errorf("company_size: want %q, got %q", req.CompanySize, writer.savedInput.CompanySize)
	}
	if writer.savedInput.BusinessType != req.BusinessType {
		t.Errorf("business_type: want %q, got %q", req.BusinessType, writer.savedInput.BusinessType)
	}
	if len(writer.savedInput.CompetitorChallenges) != 2 {
		t.Errorf("competitor_challenges: want 2 items, got %d", len(writer.savedInput.CompetitorChallenges))
	}
	if writer.savedInput.WebsiteURL != req.WebsiteURL {
		t.Errorf("website_url: want %q, got %q", req.WebsiteURL, writer.savedInput.WebsiteURL)
	}
	// OnboardingCompletedAt must be set and recent.
	if writer.savedInput.OnboardingCompletedAt.IsZero() {
		t.Error("expected onboarding_completed_at to be set")
	}
	if time.Since(writer.savedInput.OnboardingCompletedAt) > time.Minute {
		t.Error("onboarding_completed_at should be recent (within 1 minute)")
	}
}

func TestHandler_Handle_EmptyWebsiteURL_Allowed(t *testing.T) {
	orgID := uuid.New()
	writer := &stubOnboardingWriter{}
	finder := &stubOrgFinder{orgID: &orgID}
	h := NewHandler(finder, writer)

	req := Request{
		UserID:      uuid.New(),
		CompanySize: "1-10",
	}

	_, err := h.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("empty website_url should be allowed, got: %v", err)
	}
	if writer.savedInput.WebsiteURL != "" {
		t.Errorf("expected empty website_url, got %q", writer.savedInput.WebsiteURL)
	}
}

func TestHandler_Handle_Idempotent_ReSubmit(t *testing.T) {
	// Re-submitting should succeed (overwrite) without error.
	orgID := uuid.New()
	writer := &stubOnboardingWriter{}
	finder := &stubOrgFinder{orgID: &orgID}
	h := NewHandler(finder, writer)

	req := Request{UserID: uuid.New(), CompanySize: "1-10"}

	if _, err := h.Handle(context.Background(), req); err != nil {
		t.Fatalf("first submit error: %v", err)
	}
	if _, err := h.Handle(context.Background(), req); err != nil {
		t.Fatalf("second (idempotent) submit error: %v", err)
	}
	if writer.savedCalls != 2 {
		t.Errorf("expected 2 save calls, got %d", writer.savedCalls)
	}
}

func TestHandler_Handle_OrgNotFound(t *testing.T) {
	// User has no org — should propagate an error.
	writer := &stubOnboardingWriter{}
	finder := &stubOrgFinder{orgID: nil}
	h := NewHandler(finder, writer)

	req := Request{UserID: uuid.New()}

	_, err := h.Handle(context.Background(), req)
	if err == nil {
		t.Fatal("expected error when user has no org, got nil")
	}
	if writer.savedCalls != 0 {
		t.Errorf("writer should not be called when org not found; got %d calls", writer.savedCalls)
	}
}

func TestHandler_Handle_OrgFinderError(t *testing.T) {
	finderErr := errors.New("db unavailable")
	writer := &stubOnboardingWriter{}
	finder := &stubOrgFinder{findErr: finderErr}
	h := NewHandler(finder, writer)

	_, err := h.Handle(context.Background(), Request{UserID: uuid.New()})
	if err == nil {
		t.Fatal("expected error from org finder, got nil")
	}
	if !errors.Is(err, finderErr) {
		t.Errorf("expected finder error %v to be propagated, got %v", finderErr, err)
	}
}

func TestHandler_Handle_WriterError(t *testing.T) {
	orgID := uuid.New()
	writeErr := errors.New("write failed")
	writer := &stubOnboardingWriter{saveErr: writeErr}
	finder := &stubOrgFinder{orgID: &orgID}
	h := NewHandler(finder, writer)

	_, err := h.Handle(context.Background(), Request{UserID: uuid.New()})
	if err == nil {
		t.Fatal("expected error from writer, got nil")
	}
	if !errors.Is(err, writeErr) {
		t.Errorf("expected writer error %v to be propagated, got %v", writeErr, err)
	}
}
