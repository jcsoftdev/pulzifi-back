package getinvitation_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/entities"
	domerrs "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories/mocks"
	getinvitation "github.com/jcsoftdev/pulzifi-back/modules/auth/application/get_invitation"
	authentities "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	authrepos "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
)

// fakeUserRepo is a minimal stub that only implements GetByID — the only
// method the handler relies on. Other methods panic to make accidental usage
// loud in tests.
type fakeUserRepo struct {
	getByID func(ctx context.Context, id uuid.UUID) (*authentities.User, error)
}

func (f *fakeUserRepo) Create(ctx context.Context, user *authentities.User) error {
	panic("not implemented")
}
func (f *fakeUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*authentities.User, error) {
	return f.getByID(ctx, id)
}
func (f *fakeUserRepo) GetByEmail(ctx context.Context, email string) (*authentities.User, error) {
	panic("not implemented")
}
func (f *fakeUserRepo) Update(ctx context.Context, user *authentities.User) error {
	panic("not implemented")
}
func (f *fakeUserRepo) Delete(ctx context.Context, id uuid.UUID) error { panic("not implemented") }
func (f *fakeUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	panic("not implemented")
}
func (f *fakeUserRepo) GetUserFirstOrganization(ctx context.Context, userID uuid.UUID) (*string, error) {
	panic("not implemented")
}
func (f *fakeUserRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	panic("not implemented")
}
func (f *fakeUserRepo) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*authentities.User, error) {
	panic("not implemented")
}

// Compile-time check.
var _ authrepos.UserRepository = (*fakeUserRepo)(nil)

func TestGet_Pending_ReturnsEmail(t *testing.T) {
	inviterID := uuid.New()
	repo := &mocks.InvitationRepositoryMock{
		GetByTokenFunc: func(ctx context.Context, token string) (*entities.Invitation, error) {
			return &entities.Invitation{
				ID:        uuid.New(),
				Email:     "newhire@example.com",
				Token:     token,
				Status:    entities.InvitationPending,
				InvitedBy: inviterID,
				ExpiresAt: time.Now().Add(72 * time.Hour),
			}, nil
		},
	}
	users := &fakeUserRepo{
		getByID: func(ctx context.Context, id uuid.UUID) (*authentities.User, error) {
			if id != inviterID {
				t.Fatalf("unexpected inviter id: got %s want %s", id, inviterID)
			}
			return &authentities.User{ID: id, FirstName: "Carlos", LastName: "Admin"}, nil
		},
	}

	h := getinvitation.New(repo, users)
	resp, err := h.Handle(context.Background(), "tok-pending")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Email != "newhire@example.com" {
		t.Fatalf("expected email newhire@example.com, got %s", resp.Email)
	}
	if resp.InvitedByName != "Carlos Admin" {
		t.Fatalf("expected inviter 'Carlos Admin', got %s", resp.InvitedByName)
	}
	if resp.ExpiresAt.IsZero() {
		t.Fatalf("expected non-zero ExpiresAt")
	}
}

func TestGet_Expired_Returns410(t *testing.T) {
	repo := &mocks.InvitationRepositoryMock{
		GetByTokenFunc: func(ctx context.Context, token string) (*entities.Invitation, error) {
			return &entities.Invitation{
				ID:        uuid.New(),
				Email:     "stale@example.com",
				Token:     token,
				Status:    entities.InvitationPending,
				InvitedBy: uuid.New(),
				ExpiresAt: time.Now().Add(-1 * time.Hour), // already expired
			}, nil
		},
	}
	users := &fakeUserRepo{
		getByID: func(ctx context.Context, id uuid.UUID) (*authentities.User, error) {
			t.Fatalf("user repo should not be called for expired invitation")
			return nil, nil
		},
	}

	h := getinvitation.New(repo, users)
	_, err := h.Handle(context.Background(), "tok-expired")
	if !errors.Is(err, domerrs.ErrInvitationNotFound) {
		t.Fatalf("expected ErrInvitationNotFound, got %v", err)
	}
}
