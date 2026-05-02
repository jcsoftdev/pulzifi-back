package listdeliveries_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	listdeliveries "github.com/jcsoftdev/pulzifi-back/modules/integration/application/list_deliveries"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
)

// stubDeliveryRepo is a test double for DeliveryRepository.
type stubDeliveryRepo struct {
	listFn  func(ctx context.Context, destinationID uuid.UUID, limit, offset int) ([]*entities.Delivery, error)
	retryFn func(ctx context.Context, id uuid.UUID) error

	capturedLimit  int
	capturedOffset int
}

// Compile-time interface guard.
var _ repositories.DeliveryRepository = (*stubDeliveryRepo)(nil)

func (r *stubDeliveryRepo) Create(_ context.Context, _ *entities.Delivery) error { return nil }
func (r *stubDeliveryRepo) ClaimPending(_ context.Context, _ int, _ time.Time) ([]*entities.Delivery, error) {
	return nil, nil
}
func (r *stubDeliveryRepo) MarkDelivered(_ context.Context, _ uuid.UUID, _ int, _ string) error {
	return nil
}
func (r *stubDeliveryRepo) MarkFailed(_ context.Context, _ uuid.UUID, _ *int, _, _ string, _ time.Time) error {
	return nil
}
func (r *stubDeliveryRepo) MarkDead(_ context.Context, _ uuid.UUID, _ string) error { return nil }
func (r *stubDeliveryRepo) ListByDestination(ctx context.Context, destinationID uuid.UUID, limit, offset int) ([]*entities.Delivery, error) {
	r.capturedLimit = limit
	r.capturedOffset = offset
	return r.listFn(ctx, destinationID, limit, offset)
}
func (r *stubDeliveryRepo) Retry(ctx context.Context, id uuid.UUID) error {
	if r.retryFn != nil {
		return r.retryFn(ctx, id)
	}
	return nil
}

// --------------------------------------------------------------------------

func TestList_HappyPath(t *testing.T) {
	destID := uuid.New()
	want := []*entities.Delivery{
		{ID: uuid.New(), DestinationID: destID},
		{ID: uuid.New(), DestinationID: destID},
		{ID: uuid.New(), DestinationID: destID},
	}

	repo := &stubDeliveryRepo{
		listFn: func(_ context.Context, _ uuid.UUID, _, _ int) ([]*entities.Delivery, error) {
			return want, nil
		},
	}

	h := listdeliveries.NewHandler(repo)
	resp, err := h.Handle(context.Background(), listdeliveries.Request{
		DestinationID: destID,
		Limit:         10,
		Offset:        0,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(resp.Deliveries) != 3 {
		t.Errorf("want 3 deliveries, got %d", len(resp.Deliveries))
	}
}

func TestList_MissingDestinationID(t *testing.T) {
	called := false
	repo := &stubDeliveryRepo{
		listFn: func(_ context.Context, _ uuid.UUID, _, _ int) ([]*entities.Delivery, error) {
			called = true
			return nil, nil
		},
	}

	h := listdeliveries.NewHandler(repo)
	_, err := h.Handle(context.Background(), listdeliveries.Request{
		DestinationID: uuid.Nil,
	})
	if err == nil {
		t.Fatal("expected error for missing destination_id, got nil")
	}
	if called {
		t.Error("repo should not be called when destination_id is missing")
	}
}

func TestList_DefaultsApplied(t *testing.T) {
	destID := uuid.New()
	repo := &stubDeliveryRepo{
		listFn: func(_ context.Context, _ uuid.UUID, _, _ int) ([]*entities.Delivery, error) {
			return nil, nil
		},
	}

	h := listdeliveries.NewHandler(repo)

	// Limit=0 → should default to 50, Offset=-1 → should default to 0
	_, err := h.Handle(context.Background(), listdeliveries.Request{
		DestinationID: destID,
		Limit:         0,
		Offset:        -1,
	})
	if err != nil {
		t.Fatalf("Handle (zero limit): %v", err)
	}
	if repo.capturedLimit != 50 {
		t.Errorf("expected default limit 50, got %d", repo.capturedLimit)
	}
	if repo.capturedOffset != 0 {
		t.Errorf("expected default offset 0, got %d", repo.capturedOffset)
	}

	// Limit=500 → should be capped at 200
	_, err = h.Handle(context.Background(), listdeliveries.Request{
		DestinationID: destID,
		Limit:         500,
		Offset:        0,
	})
	if err != nil {
		t.Fatalf("Handle (large limit): %v", err)
	}
	if repo.capturedLimit != 200 {
		t.Errorf("expected capped limit 200, got %d", repo.capturedLimit)
	}
}

func TestList_RepoError(t *testing.T) {
	destID := uuid.New()
	repo := &stubDeliveryRepo{
		listFn: func(_ context.Context, _ uuid.UUID, _, _ int) ([]*entities.Delivery, error) {
			return nil, errors.New("db connection lost")
		},
	}

	h := listdeliveries.NewHandler(repo)
	_, err := h.Handle(context.Background(), listdeliveries.Request{
		DestinationID: destID,
		Limit:         10,
	})
	if err == nil {
		t.Fatal("expected error from repo, got nil")
	}
}
