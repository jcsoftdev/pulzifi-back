package deleteorganization_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	deleteorganization "github.com/jcsoftdev/pulzifi-back/modules/organization/application/delete_organization"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
)

// ── in-memory stubs ──────────────────────────────────────────────────────────

type stubOrgDeletionRepo struct {
	orgs      map[uuid.UUID]services.OrgForDeletion
	softDeleted  map[uuid.UUID]bool
	auditStatus  map[uuid.UUID]string
	auditStep    map[uuid.UUID]string
	auditErrMsg  map[uuid.UUID]string
	members      map[uuid.UUID][]uuid.UUID // orgID → memberUserIDs
	memberships  map[uuid.UUID][]uuid.UUID // userID → orgIDs they belong to
	deletedUsers []uuid.UUID

	lookupErr          error
	softDeleteErr      error
	markAuditErr       error
	cleanupErr         error
	solelyOwnedOrgs    []services.OrgForDeletion

	callOrder []string
}

func newStubRepo() *stubOrgDeletionRepo {
	return &stubOrgDeletionRepo{
		orgs:        make(map[uuid.UUID]services.OrgForDeletion),
		softDeleted: make(map[uuid.UUID]bool),
		auditStatus: make(map[uuid.UUID]string),
		auditStep:   make(map[uuid.UUID]string),
		auditErrMsg: make(map[uuid.UUID]string),
		members:     make(map[uuid.UUID][]uuid.UUID),
		memberships: make(map[uuid.UUID][]uuid.UUID),
	}
}

func (r *stubOrgDeletionRepo) LookupForDeletion(_ context.Context, orgID uuid.UUID) (services.OrgForDeletion, error) {
	r.callOrder = append(r.callOrder, "lookup")
	if r.lookupErr != nil {
		return services.OrgForDeletion{}, r.lookupErr
	}
	org, ok := r.orgs[orgID]
	if !ok {
		return services.OrgForDeletion{}, services.ErrOrgNotFound
	}
	return org, nil
}

func (r *stubOrgDeletionRepo) SoftDeleteAndOpenAudit(_ context.Context, in services.AuditOpenInput) (uuid.UUID, error) {
	r.callOrder = append(r.callOrder, "soft_delete")
	if r.softDeleteErr != nil {
		return uuid.Nil, r.softDeleteErr
	}
	r.softDeleted[in.OrgID] = true
	auditID := uuid.New()
	r.auditStatus[auditID] = "pending"
	return auditID, nil
}

func (r *stubOrgDeletionRepo) MarkAudit(_ context.Context, auditID uuid.UUID, status, failureStep, errMsg string) error {
	r.callOrder = append(r.callOrder, "mark_audit:"+status)
	if r.markAuditErr != nil {
		return r.markAuditErr
	}
	r.auditStatus[auditID] = status
	r.auditStep[auditID] = failureStep
	r.auditErrMsg[auditID] = errMsg
	return nil
}

func (r *stubOrgDeletionRepo) CleanupAndHardDelete(_ context.Context, orgID uuid.UUID, _ string) ([]uuid.UUID, error) {
	r.callOrder = append(r.callOrder, "hard_delete")
	if r.cleanupErr != nil {
		return nil, r.cleanupErr
	}
	// Delete solely-membered users
	var deleted []uuid.UUID
	for _, memberID := range r.members[orgID] {
		// Check remaining memberships after removing this org
		remaining := 0
		for _, oid := range r.memberships[memberID] {
			if oid != orgID {
				remaining++
			}
		}
		if remaining == 0 {
			deleted = append(deleted, memberID)
		}
	}
	r.deletedUsers = append(r.deletedUsers, deleted...)
	delete(r.orgs, orgID)
	return deleted, nil
}

func (r *stubOrgDeletionRepo) FindSolelyOwnedOrgs(_ context.Context, _ uuid.UUID) ([]services.OrgForDeletion, error) {
	return r.solelyOwnedOrgs, nil
}

type stubBillingCanceller struct {
	cancelErr   error
	cancelCalls int
}

func (b *stubBillingCanceller) CancelForOrg(_ context.Context, _ uuid.UUID) error {
	b.cancelCalls++
	return b.cancelErr
}

type stubStorageSweeper struct {
	sweepErr   error
	sweepCalls int
}

func (s *stubStorageSweeper) SweepTenant(_ context.Context, _ uuid.UUID, _ string) error {
	s.sweepCalls++
	return s.sweepErr
}

type stubSchemaDropper struct {
	dropErr   error
	dropCalls int
}

func (d *stubSchemaDropper) DropTenantSchema(_ context.Context, _ string) error {
	d.dropCalls++
	return d.dropErr
}

// ── helpers ──────────────────────────────────────────────────────────────────

func makeOrg() (uuid.UUID, services.OrgForDeletion) {
	id := uuid.New()
	return id, services.OrgForDeletion{ID: id, SchemaName: "testschema"}
}

// ── tests ────────────────────────────────────────────────────────────────────

func TestDeleteOrganization_HappyPath(t *testing.T) {
	repo := newStubRepo()
	orgID, org := makeOrg()
	repo.orgs[orgID] = org
	userA := uuid.New()
	userB := uuid.New()
	repo.members[orgID] = []uuid.UUID{userA, userB}
	// userA belongs only to this org; userB belongs to another org too
	repo.memberships[userA] = []uuid.UUID{orgID}
	repo.memberships[userB] = []uuid.UUID{orgID, uuid.New()}

	billing := &stubBillingCanceller{}
	storage := &stubStorageSweeper{}
	schema := &stubSchemaDropper{}

	h := deleteorganization.NewHandler(repo, billing, storage, schema)
	resp, err := h.Handle(context.Background(), &deleteorganization.Request{
		OrgID:     orgID,
		ActorID:   uuid.New(),
		ActorType: "super_admin",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.OrgID != orgID {
		t.Errorf("resp.OrgID = %v, want %v", resp.OrgID, orgID)
	}
	if len(resp.DeletedUserIDs) != 1 || resp.DeletedUserIDs[0] != userA {
		t.Errorf("expected only userA deleted, got %v", resp.DeletedUserIDs)
	}
	if billing.cancelCalls != 1 {
		t.Errorf("billing.CancelForOrg calls = %d, want 1", billing.cancelCalls)
	}
	if storage.sweepCalls != 1 {
		t.Errorf("storage.SweepTenant calls = %d, want 1", storage.sweepCalls)
	}
	if schema.dropCalls != 1 {
		t.Errorf("schema.DropTenantSchema calls = %d, want 1", schema.dropCalls)
	}
}

func TestDeleteOrganization_BillingAbort(t *testing.T) {
	repo := newStubRepo()
	orgID, org := makeOrg()
	repo.orgs[orgID] = org

	billing := &stubBillingCanceller{cancelErr: services.ErrBillingActive}
	storage := &stubStorageSweeper{}
	schema := &stubSchemaDropper{}

	h := deleteorganization.NewHandler(repo, billing, storage, schema)
	_, err := h.Handle(context.Background(), &deleteorganization.Request{
		OrgID:     orgID,
		ActorID:   uuid.New(),
		ActorType: "super_admin",
	})

	if !errors.Is(err, services.ErrBillingActive) {
		t.Fatalf("expected ErrBillingActive, got %v", err)
	}
	// Org must be soft-deleted but not hard-deleted
	if !repo.softDeleted[orgID] {
		t.Error("org should be soft-deleted after billing abort")
	}
	if _, exists := repo.orgs[orgID]; !exists {
		t.Error("org should NOT be hard-deleted after billing abort")
	}
	// Storage and schema drop must NOT have run
	if storage.sweepCalls != 0 {
		t.Errorf("storage sweep must not run after billing abort, called %d times", storage.sweepCalls)
	}
	if schema.dropCalls != 0 {
		t.Errorf("schema drop must not run after billing abort, called %d times", schema.dropCalls)
	}
	// Audit must be marked failed with step=billing
	for auditID, status := range repo.auditStatus {
		if status == "failed" {
			step := repo.auditStep[auditID]
			if step != "billing" {
				t.Errorf("audit failure_step = %q, want billing", step)
			}
			return // found the right audit row
		}
	}
	t.Error("no audit row marked failed after billing abort")
}

func TestDeleteOrganization_BillingNoop(t *testing.T) {
	// BillingCanceller is a no-op (nil return) — cascade proceeds
	repo := newStubRepo()
	orgID, org := makeOrg()
	repo.orgs[orgID] = org

	billing := &stubBillingCanceller{} // cancelErr == nil
	storage := &stubStorageSweeper{}
	schema := &stubSchemaDropper{}

	h := deleteorganization.NewHandler(repo, billing, storage, schema)
	_, err := h.Handle(context.Background(), &deleteorganization.Request{
		OrgID:     orgID,
		ActorID:   uuid.New(),
		ActorType: "super_admin",
	})

	if err != nil {
		t.Fatalf("expected no error with noop billing, got %v", err)
	}
	if billing.cancelCalls != 1 {
		t.Errorf("billing.CancelForOrg calls = %d, want 1", billing.cancelCalls)
	}
	if schema.dropCalls != 1 {
		t.Errorf("expected schema drop to proceed after noop billing, dropCalls = %d", schema.dropCalls)
	}
}

func TestDeleteOrganization_StorageSweepFailureTolerated(t *testing.T) {
	repo := newStubRepo()
	orgID, org := makeOrg()
	repo.orgs[orgID] = org

	billing := &stubBillingCanceller{}
	storage := &stubStorageSweeper{sweepErr: errors.New("connection refused")}
	schema := &stubSchemaDropper{}

	h := deleteorganization.NewHandler(repo, billing, storage, schema)
	_, err := h.Handle(context.Background(), &deleteorganization.Request{
		OrgID:     orgID,
		ActorID:   uuid.New(),
		ActorType: "super_admin",
	})

	if err != nil {
		t.Fatalf("storage sweep failure must not abort cascade, got %v", err)
	}
	if schema.dropCalls != 1 {
		t.Errorf("schema drop must run after storage sweep failure, dropCalls = %d", schema.dropCalls)
	}
	// Audit must be completed
	for _, status := range repo.auditStatus {
		if status == "completed" {
			return
		}
	}
	t.Error("audit must be completed when storage sweep fails but cascade continues")
}

func TestDeleteOrganization_CleanupAndHardDeleteFailure(t *testing.T) {
	repo := newStubRepo()
	orgID, org := makeOrg()
	repo.orgs[orgID] = org
	repo.cleanupErr = errors.New("outbox delete failed")

	billing := &stubBillingCanceller{}
	storage := &stubStorageSweeper{}
	schema := &stubSchemaDropper{}

	h := deleteorganization.NewHandler(repo, billing, storage, schema)
	_, err := h.Handle(context.Background(), &deleteorganization.Request{
		OrgID:     orgID,
		ActorID:   uuid.New(),
		ActorType: "super_admin",
	})

	if err == nil {
		t.Fatal("expected error when CleanupAndHardDelete fails")
	}
	if schema.dropCalls != 0 {
		t.Errorf("schema drop must NOT run after hard_delete failure, dropCalls = %d", schema.dropCalls)
	}
	for auditID, status := range repo.auditStatus {
		if status == "failed" {
			step := repo.auditStep[auditID]
			if step != "hard_delete" {
				t.Errorf("audit failure_step = %q, want hard_delete", step)
			}
			return
		}
	}
	t.Error("no audit row marked failed after hard_delete failure")
}

func TestDeleteOrganization_SchemaDropFailure(t *testing.T) {
	repo := newStubRepo()
	orgID, org := makeOrg()
	repo.orgs[orgID] = org

	billing := &stubBillingCanceller{}
	storage := &stubStorageSweeper{}
	schema := &stubSchemaDropper{dropErr: errors.New("DROP SCHEMA failed")}

	h := deleteorganization.NewHandler(repo, billing, storage, schema)
	_, err := h.Handle(context.Background(), &deleteorganization.Request{
		OrgID:     orgID,
		ActorID:   uuid.New(),
		ActorType: "super_admin",
	})

	if err == nil {
		t.Fatal("expected error when DropTenantSchema fails")
	}
	for auditID, status := range repo.auditStatus {
		if status == "failed" {
			step := repo.auditStep[auditID]
			if step != "drop_schema" {
				t.Errorf("audit failure_step = %q, want drop_schema", step)
			}
			return
		}
	}
	t.Error("no audit row marked failed after drop_schema failure")
}

func TestDeleteOrganization_IdempotentResume(t *testing.T) {
	// Org already soft-deleted — use case must NOT return ErrOrgNotFound, must resume
	repo := newStubRepo()
	orgID, org := makeOrg()
	repo.orgs[orgID] = org
	// Mark as already soft-deleted — LookupForDeletion still returns the row
	repo.softDeleted[orgID] = true

	billing := &stubBillingCanceller{}
	storage := &stubStorageSweeper{}
	schema := &stubSchemaDropper{}

	h := deleteorganization.NewHandler(repo, billing, storage, schema)
	_, err := h.Handle(context.Background(), &deleteorganization.Request{
		OrgID:     orgID,
		ActorID:   uuid.New(),
		ActorType: "super_admin",
	})

	if errors.Is(err, services.ErrOrgNotFound) {
		t.Fatal("already-soft-deleted org must not return ErrOrgNotFound on re-run")
	}
	if err != nil {
		t.Fatalf("expected clean resume, got %v", err)
	}
}

func TestDeleteOrganization_OrgNotFound(t *testing.T) {
	repo := newStubRepo()
	billing := &stubBillingCanceller{}
	storage := &stubStorageSweeper{}
	schema := &stubSchemaDropper{}

	h := deleteorganization.NewHandler(repo, billing, storage, schema)
	_, err := h.Handle(context.Background(), &deleteorganization.Request{
		OrgID:     uuid.New(),
		ActorID:   uuid.New(),
		ActorType: "super_admin",
	})

	if !errors.Is(err, services.ErrOrgNotFound) {
		t.Fatalf("expected ErrOrgNotFound, got %v", err)
	}
}

func TestDeleteOrganization_UserPruning(t *testing.T) {
	// sole-org member deleted; multi-org member kept
	repo := newStubRepo()
	orgID, org := makeOrg()
	repo.orgs[orgID] = org

	soleUser := uuid.New()
	multiUser := uuid.New()
	repo.members[orgID] = []uuid.UUID{soleUser, multiUser}
	repo.memberships[soleUser] = []uuid.UUID{orgID}
	repo.memberships[multiUser] = []uuid.UUID{orgID, uuid.New()}

	h := deleteorganization.NewHandler(repo,
		&stubBillingCanceller{},
		&stubStorageSweeper{},
		&stubSchemaDropper{},
	)
	resp, err := h.Handle(context.Background(), &deleteorganization.Request{
		OrgID:     orgID,
		ActorID:   uuid.New(),
		ActorType: "super_admin",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(resp.DeletedUserIDs) != 1 || resp.DeletedUserIDs[0] != soleUser {
		t.Errorf("DeletedUserIDs = %v, want [%v]", resp.DeletedUserIDs, soleUser)
	}
}
