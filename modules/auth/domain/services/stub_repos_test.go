// Package services_test — shared test stubs for domain service tests.
package services_test

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
)

// stubUserRepo is a no-op UserRepository for tests that don't need real user lookups.
type stubUserRepo struct{}

func (r *stubUserRepo) Create(_ context.Context, _ *entities.User) error               { return nil }
func (r *stubUserRepo) GetByID(_ context.Context, _ uuid.UUID) (*entities.User, error) { return nil, nil }
func (r *stubUserRepo) GetByEmail(_ context.Context, _ string) (*entities.User, error) { return nil, nil }
func (r *stubUserRepo) Update(_ context.Context, _ *entities.User) error               { return nil }
func (r *stubUserRepo) Delete(_ context.Context, _ uuid.UUID) error                    { return nil }
func (r *stubUserRepo) ExistsByEmail(_ context.Context, _ string) (bool, error)        { return false, nil }
func (r *stubUserRepo) GetUserFirstOrganization(_ context.Context, _ uuid.UUID) (*string, error) {
	return nil, nil
}
func (r *stubUserRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ string) error { return nil }
func (r *stubUserRepo) ListByStatus(_ context.Context, _ string, _, _ int) ([]*entities.User, error) {
	return nil, nil
}

// stubUserWithHash holds a bcrypt hash so ValidateCredentials can be called
// without a real database.
type stubUserWithHash struct {
	hash string
}

func (s *stubUserWithHash) asEntity() *entities.User {
	return &entities.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: s.hash,
	}
}
