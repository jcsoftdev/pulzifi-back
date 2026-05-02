package bulkretrydeliveries_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/application/bulk_retry_deliveries"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
)

type stubRepo struct {
	retryFn func(id uuid.UUID) error
}

func (s *stubRepo) Create(context.Context, *entities.Delivery) error                  { return nil }
func (s *stubRepo) GetByID(context.Context, uuid.UUID) (*entities.Delivery, error)    { return nil, nil }
func (s *stubRepo) ClaimPending(context.Context, int, time.Time) ([]*entities.Delivery, error) {
	return nil, nil
}
func (s *stubRepo) MarkDelivered(context.Context, uuid.UUID, int, string) error { return nil }
func (s *stubRepo) MarkFailed(context.Context, uuid.UUID, *int, string, string, time.Time, []entities.Attempt) error {
	return nil
}
func (s *stubRepo) MarkDead(context.Context, uuid.UUID, string, []entities.Attempt) error {
	return nil
}
func (s *stubRepo) ListByDestination(context.Context, uuid.UUID, int, int) ([]*entities.Delivery, error) {
	return nil, nil
}
func (s *stubRepo) Retry(_ context.Context, id uuid.UUID) error { return s.retryFn(id) }

var _ repositories.DeliveryRepository = (*stubRepo)(nil)

func TestBulk_AllRetried(t *testing.T) {
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	repo := &stubRepo{retryFn: func(uuid.UUID) error { return nil }}
	h := bulkretrydeliveries.NewHandler(repo)
	resp, _ := h.Handle(context.Background(), bulkretrydeliveries.Request{IDs: ids})
	if resp.Retried != 3 || resp.Skipped != 0 || resp.Failed != 0 {
		t.Errorf("got %+v want {3,0,0}", resp)
	}
}

func TestBulk_MixedSkipAndFail(t *testing.T) {
	id1, id2, id3, id4 := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	results := map[uuid.UUID]error{
		id1: nil,
		id2: nil,
		id3: repositories.ErrDeliveryNotInDeadState,
		id4: errors.New("boom"),
	}
	repo := &stubRepo{retryFn: func(id uuid.UUID) error { return results[id] }}
	h := bulkretrydeliveries.NewHandler(repo)
	resp, _ := h.Handle(context.Background(), bulkretrydeliveries.Request{IDs: []uuid.UUID{id1, id2, id3, id4}})
	if resp.Retried != 2 || resp.Skipped != 1 || resp.Failed != 1 {
		t.Errorf("got %+v want {2,1,1}", resp)
	}
}

func TestBulk_EmptyInput(t *testing.T) {
	h := bulkretrydeliveries.NewHandler(&stubRepo{})
	resp, _ := h.Handle(context.Background(), bulkretrydeliveries.Request{IDs: nil})
	if resp.Retried+resp.Skipped+resp.Failed != 0 {
		t.Errorf("expected zero counts, got %+v", resp)
	}
}

func TestBulk_NilInRequest(t *testing.T) {
	calls := 0
	repo := &stubRepo{retryFn: func(uuid.UUID) error { calls++; return nil }}
	h := bulkretrydeliveries.NewHandler(repo)
	resp, _ := h.Handle(context.Background(), bulkretrydeliveries.Request{IDs: []uuid.UUID{uuid.Nil, uuid.New()}})
	if resp.Retried != 1 || resp.Failed != 1 {
		t.Errorf("got %+v want Retried=1 Failed=1", resp)
	}
	if calls != 1 {
		t.Errorf("repo.Retry called %d times, want 1 (nil id should be skipped)", calls)
	}
}
