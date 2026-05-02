package retrydelivery_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	retrydelivery "github.com/jcsoftdev/pulzifi-back/modules/integration/application/retry_delivery"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
)

// stubDeliveryRepo is a test double for DeliveryRepository.
type stubDeliveryRepo struct {
	retryFn    func(ctx context.Context, id uuid.UUID) error
	retryCalls []uuid.UUID
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
func (r *stubDeliveryRepo) ListByDestination(_ context.Context, _ uuid.UUID, _, _ int) ([]*entities.Delivery, error) {
	return nil, nil
}
func (r *stubDeliveryRepo) Retry(ctx context.Context, id uuid.UUID) error {
	r.retryCalls = append(r.retryCalls, id)
	if r.retryFn != nil {
		return r.retryFn(ctx, id)
	}
	return nil
}

// --------------------------------------------------------------------------

func TestRetry_HappyPath(t *testing.T) {
	id := uuid.New()
	repo := &stubDeliveryRepo{
		retryFn: func(_ context.Context, _ uuid.UUID) error { return nil },
	}

	h := retrydelivery.NewHandler(repo)
	resp, err := h.Handle(context.Background(), retrydelivery.Request{ID: id})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(repo.retryCalls) != 1 || repo.retryCalls[0] != id {
		t.Errorf("repo.Retry not called with correct id: calls=%v", repo.retryCalls)
	}
}

func TestRetry_MissingID(t *testing.T) {
	repo := &stubDeliveryRepo{}

	h := retrydelivery.NewHandler(repo)
	_, err := h.Handle(context.Background(), retrydelivery.Request{ID: uuid.Nil})
	if err == nil {
		t.Fatal("expected error for missing delivery id, got nil")
	}
	if len(repo.retryCalls) != 0 {
		t.Error("repo.Retry should not be called when id is missing")
	}
}

func TestRetry_RepoError(t *testing.T) {
	id := uuid.New()
	repo := &stubDeliveryRepo{
		retryFn: func(_ context.Context, _ uuid.UUID) error {
			return errors.New("delivery repo: not found or not in dead state")
		},
	}

	h := retrydelivery.NewHandler(repo)
	_, err := h.Handle(context.Background(), retrydelivery.Request{ID: id})
	if err == nil {
		t.Fatal("expected error from repo, got nil")
	}
}
