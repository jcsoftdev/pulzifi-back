package connectbyo_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	connectbyo "github.com/jcsoftdev/pulzifi-back/modules/integration/application/connect_byo"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
)

// ---------- stubs ----------

type stubFlags struct {
	on  bool
	err error
}

func (s *stubFlags) IsOn(context.Context, uuid.UUID, string) (bool, error) { return s.on, s.err }

type stubPlans struct {
	code string
	err  error
}

func (s *stubPlans) PlanCode(context.Context, uuid.UUID) (string, error) { return s.code, s.err }

type stubValidator struct {
	err   error
	calls int
}

func (s *stubValidator) Validate(context.Context, string, map[string]string) error {
	s.calls++
	return s.err
}

type stubRepo struct {
	created []*entities.Integration
	err     error
}

// Compile-time check that stubRepo satisfies the full IntegrationRepository interface.
var _ repositories.IntegrationRepository = (*stubRepo)(nil)

func (s *stubRepo) Create(_ context.Context, integ *entities.Integration) error {
	if s.err != nil {
		return s.err
	}
	s.created = append(s.created, integ)
	return nil
}

func (s *stubRepo) Update(_ context.Context, _ *entities.Integration) error { return nil }

func (s *stubRepo) GetByID(_ context.Context, _ uuid.UUID) (*entities.Integration, error) {
	return nil, nil
}

func (s *stubRepo) GetByOrgAndService(_ context.Context, _ uuid.UUID, _ string) (*entities.Integration, error) {
	return nil, nil
}

func (s *stubRepo) ListByOrg(_ context.Context, _ uuid.UUID) ([]*entities.Integration, error) {
	return nil, nil
}

func (s *stubRepo) SoftDelete(_ context.Context, _ uuid.UUID) error { return nil }

// ---------- helpers ----------

var (
	paidPlans  = []string{"pro", "business"}
	orgID      = uuid.New()
	userID     = uuid.New()
	validCreds = map[string]string{
		"account_sid":  "ACsid",
		"auth_token":   "tok",
		"from_number":  "+15555555555",
	}
)

func newHandler(flags *stubFlags, plans *stubPlans, val *stubValidator, repo *stubRepo) *connectbyo.Handler {
	return connectbyo.NewHandler(repo, val, flags, plans, paidPlans)
}

func validReq() connectbyo.Request {
	return connectbyo.Request{
		Provider:    "twilio",
		OrgID:       orgID,
		UserID:      userID,
		Credentials: validCreds,
	}
}

// ---------- tests ----------

func TestConnectBYO_HappyPath_Twilio(t *testing.T) {
	flags := &stubFlags{on: true}
	plans := &stubPlans{code: "enterprise"}
	val := &stubValidator{}
	repo := &stubRepo{}

	h := newHandler(flags, plans, val, repo)
	resp, err := h.Handle(context.Background(), validReq())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || resp.Integration == nil {
		t.Fatal("expected non-nil response and integration")
	}

	if len(repo.created) != 1 {
		t.Fatalf("expected repo.Create called once, got %d", len(repo.created))
	}
	integ := repo.created[0]
	if integ.ServiceType != "twilio" {
		t.Errorf("ServiceType: got %q, want %q", integ.ServiceType, "twilio")
	}
	if integ.Status != entities.IntegrationActive {
		t.Errorf("Status: got %q, want %q", integ.Status, entities.IntegrationActive)
	}
	if integ.AccessToken != "tok" {
		t.Errorf("AccessToken: got %q, want %q", integ.AccessToken, "tok")
	}
	if integ.RefreshToken != "ACsid" {
		t.Errorf("RefreshToken: got %q, want %q", integ.RefreshToken, "ACsid")
	}
	fromNum, _ := integ.ProviderMeta["from_number"].(string)
	if fromNum != "+15555555555" {
		t.Errorf("ProviderMeta[from_number]: got %q, want %q", fromNum, "+15555555555")
	}
	if integ.CreatedBy != userID {
		t.Errorf("CreatedBy: got %v, want %v", integ.CreatedBy, userID)
	}
	if integ.OrgID != orgID {
		t.Errorf("OrgID: got %v, want %v", integ.OrgID, orgID)
	}
	if integ.ID == uuid.Nil {
		t.Error("ID should not be uuid.Nil")
	}
}

func TestConnectBYO_FlagOff_Rejects(t *testing.T) {
	flags := &stubFlags{on: false}
	plans := &stubPlans{code: "enterprise"}
	val := &stubValidator{}
	repo := &stubRepo{}

	h := newHandler(flags, plans, val, repo)
	_, err := h.Handle(context.Background(), validReq())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not enabled") {
		t.Errorf("error %q should contain %q", err.Error(), "not enabled")
	}
	if len(repo.created) != 0 {
		t.Error("repo.Create should NOT have been called")
	}
	if val.calls != 0 {
		t.Error("validator should NOT have been called")
	}
}

func TestConnectBYO_FlagReaderError_PropagatesAndDoesNotPersist(t *testing.T) {
	flagErr := errors.New("redis unavailable")
	flags := &stubFlags{err: flagErr}
	plans := &stubPlans{code: "enterprise"}
	val := &stubValidator{}
	repo := &stubRepo{}

	h := newHandler(flags, plans, val, repo)
	_, err := h.Handle(context.Background(), validReq())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "flag check") {
		t.Errorf("error %q should contain %q", err.Error(), "flag check")
	}
	if !errors.Is(err, flagErr) {
		t.Error("error should wrap original flag error")
	}
	if len(repo.created) != 0 {
		t.Error("repo.Create should NOT have been called")
	}
}

func TestConnectBYO_PaidTier_Rejects(t *testing.T) {
	flags := &stubFlags{on: true}
	plans := &stubPlans{code: "pro"} // in paidPlans
	val := &stubValidator{}
	repo := &stubRepo{}

	h := newHandler(flags, plans, val, repo)
	_, err := h.Handle(context.Background(), validReq())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "BYO only available on enterprise") {
		t.Errorf("error %q should contain %q", err.Error(), "BYO only available on enterprise")
	}
	if len(repo.created) != 0 {
		t.Error("repo.Create should NOT have been called")
	}
}

func TestConnectBYO_FreeTier_Rejects(t *testing.T) {
	flags := &stubFlags{on: true}
	plans := &stubPlans{code: ""} // empty = free tier
	val := &stubValidator{}
	repo := &stubRepo{}

	h := newHandler(flags, plans, val, repo)
	_, err := h.Handle(context.Background(), validReq())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "BYO only available on enterprise") {
		t.Errorf("error %q should contain %q", err.Error(), "BYO only available on enterprise")
	}
	if len(repo.created) != 0 {
		t.Error("repo.Create should NOT have been called")
	}
}

func TestConnectBYO_PlanLookupError_DoesNotPersist(t *testing.T) {
	planErr := errors.New("db timeout")
	flags := &stubFlags{on: true}
	plans := &stubPlans{err: planErr}
	val := &stubValidator{}
	repo := &stubRepo{}

	h := newHandler(flags, plans, val, repo)
	_, err := h.Handle(context.Background(), validReq())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "plan lookup") {
		t.Errorf("error %q should contain %q", err.Error(), "plan lookup")
	}
	if !errors.Is(err, planErr) {
		t.Error("error should wrap original plan lookup error")
	}
	if len(repo.created) != 0 {
		t.Error("repo.Create should NOT have been called")
	}
}

func TestConnectBYO_ValidationFails_NoRepoWrite(t *testing.T) {
	valErr := errors.New("invalid credentials")
	flags := &stubFlags{on: true}
	plans := &stubPlans{code: "enterprise"}
	val := &stubValidator{err: valErr}
	repo := &stubRepo{}

	h := newHandler(flags, plans, val, repo)
	_, err := h.Handle(context.Background(), validReq())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid credentials") {
		t.Errorf("error %q should contain %q", err.Error(), "invalid credentials")
	}
	if len(repo.created) != 0 {
		t.Error("repo.Create should NOT have been called")
	}
}

func TestConnectBYO_UnsupportedProvider(t *testing.T) {
	flags := &stubFlags{}
	plans := &stubPlans{}
	val := &stubValidator{}
	repo := &stubRepo{}

	h := newHandler(flags, plans, val, repo)
	req := validReq()
	req.Provider = "slack"
	_, err := h.Handle(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "provider not supported") {
		t.Errorf("error %q should contain %q", err.Error(), "provider not supported")
	}
	// None of the collaborators should have been called.
	if flags.on {
		t.Error("flag reader should NOT have been called (stubFlags.on starts false)")
	}
	if val.calls != 0 {
		t.Error("validator should NOT have been called")
	}
	if len(repo.created) != 0 {
		t.Error("repo.Create should NOT have been called")
	}
}

func TestConnectBYO_RepoCreateError_Propagates(t *testing.T) {
	repoErr := errors.New("constraint violation")
	flags := &stubFlags{on: true}
	plans := &stubPlans{code: "enterprise"}
	val := &stubValidator{}
	repo := &stubRepo{err: repoErr}

	h := newHandler(flags, plans, val, repo)
	_, err := h.Handle(context.Background(), validReq())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "persist") {
		t.Errorf("error %q should contain %q", err.Error(), "persist")
	}
	if !errors.Is(err, repoErr) {
		t.Error("error should wrap original repo error")
	}
}
