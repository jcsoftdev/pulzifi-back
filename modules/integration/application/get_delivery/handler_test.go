package getdelivery_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/application/get_delivery"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
)

type stubRepo struct {
	getByIDFn func(ctx context.Context, id uuid.UUID) (*entities.Delivery, error)
}

func (s *stubRepo) Create(context.Context, *entities.Delivery) error            { return nil }
func (s *stubRepo) GetByID(ctx context.Context, id uuid.UUID) (*entities.Delivery, error) {
	return s.getByIDFn(ctx, id)
}
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
func (s *stubRepo) Retry(context.Context, uuid.UUID) error { return nil }

var _ repositories.DeliveryRepository = (*stubRepo)(nil)

func TestGet_HappyPath(t *testing.T) {
	id := uuid.New()
	want := &entities.Delivery{ID: id, EventType: "change.detected"}
	repo := &stubRepo{getByIDFn: func(_ context.Context, gotID uuid.UUID) (*entities.Delivery, error) {
		if gotID != id {
			t.Errorf("repo got id %v, want %v", gotID, id)
		}
		return want, nil
	}}
	h := getdelivery.NewHandler(repo)
	resp, err := h.Handle(context.Background(), getdelivery.Request{ID: id})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Delivery != want {
		t.Errorf("response missing delivery")
	}
}

func TestGet_NotFound(t *testing.T) {
	repo := &stubRepo{getByIDFn: func(context.Context, uuid.UUID) (*entities.Delivery, error) {
		return nil, nil
	}}
	h := getdelivery.NewHandler(repo)
	resp, err := h.Handle(context.Background(), getdelivery.Request{ID: uuid.New()})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Delivery != nil {
		t.Errorf("expected nil delivery on not-found")
	}
}

func TestGet_MissingID(t *testing.T) {
	h := getdelivery.NewHandler(&stubRepo{})
	if _, err := h.Handle(context.Background(), getdelivery.Request{}); err == nil {
		t.Fatal("expected error for missing id")
	}
}

func TestGet_RepoError(t *testing.T) {
	boom := errors.New("boom")
	repo := &stubRepo{getByIDFn: func(context.Context, uuid.UUID) (*entities.Delivery, error) {
		return nil, boom
	}}
	h := getdelivery.NewHandler(repo)
	if _, err := h.Handle(context.Background(), getdelivery.Request{ID: uuid.New()}); !errors.Is(err, boom) {
		t.Fatalf("expected wrapped boom, got %v", err)
	}
}
