package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/services"
	"golang.org/x/crypto/bcrypt"
)

type fakeUserRepo struct{ user *entities.User }

func (f *fakeUserRepo) GetByEmail(_ context.Context, _ string) (*entities.User, error) {
	return f.user, nil
}

// Unused UserRepository methods — satisfy the interface with no-ops.
func (f *fakeUserRepo) Create(_ context.Context, _ *entities.User) error        { return nil }
func (f *fakeUserRepo) GetByID(_ context.Context, _ uuid.UUID) (*entities.User, error) {
	return nil, nil
}
func (f *fakeUserRepo) Update(_ context.Context, _ *entities.User) error        { return nil }
func (f *fakeUserRepo) Delete(_ context.Context, _ uuid.UUID) error             { return nil }
func (f *fakeUserRepo) ExistsByEmail(_ context.Context, _ string) (bool, error) { return false, nil }
func (f *fakeUserRepo) GetUserFirstOrganization(_ context.Context, _ uuid.UUID) (*string, error) {
	return nil, nil
}
func (f *fakeUserRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ string) error { return nil }
func (f *fakeUserRepo) ListByStatus(_ context.Context, _ string, _, _ int) ([]*entities.User, error) {
	return nil, nil
}

func hashed(pw string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(h)
}

func TestAuthenticate_SuspendedUserRejected(t *testing.T) {
	user := &entities.User{
		ID:           uuid.New(),
		Email:        "banned@example.com",
		PasswordHash: hashed("secret"),
		Status:       entities.UserStatusSuspended,
	}
	svc := services.NewBcryptAuthService(&fakeUserRepo{user: user}, nil)

	_, err := svc.Authenticate(context.Background(), "banned@example.com", "secret")
	if err == nil {
		t.Fatal("expected suspended user to be rejected, got nil error")
	}
}

func TestAuthenticate_ApprovedUserPasses(t *testing.T) {
	user := &entities.User{
		ID:           uuid.New(),
		Email:        "ok@example.com",
		PasswordHash: hashed("secret"),
		Status:       entities.UserStatusApproved,
	}
	svc := services.NewBcryptAuthService(&fakeUserRepo{user: user}, nil)

	got, err := svc.Authenticate(context.Background(), "ok@example.com", "secret")
	if err != nil {
		t.Fatalf("expected approved user to pass, got %v", err)
	}
	if got.Email != "ok@example.com" {
		t.Errorf("returned wrong user: %s", got.Email)
	}
}
