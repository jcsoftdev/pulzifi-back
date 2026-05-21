package services_test

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
)

// stubUserRepo is a minimal in-memory UserRepository for use in domain tests.
type stubUserRepo struct {
	users map[uuid.UUID]*entities.User
}

func newStubUserRepo() *stubUserRepo {
	return &stubUserRepo{users: make(map[uuid.UUID]*entities.User)}
}

func (r *stubUserRepo) Create(_ context.Context, user *entities.User) error {
	r.users[user.ID] = user
	return nil
}

func (r *stubUserRepo) GetByID(_ context.Context, id uuid.UUID) (*entities.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, nil
	}
	return u, nil
}

func (r *stubUserRepo) GetByEmail(_ context.Context, email string) (*entities.User, error) {
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}

func (r *stubUserRepo) Update(_ context.Context, user *entities.User) error {
	if _, ok := r.users[user.ID]; !ok {
		return errors.New("user not found")
	}
	r.users[user.ID] = user
	return nil
}

func (r *stubUserRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.users, id)
	return nil
}

func (r *stubUserRepo) ExistsByEmail(_ context.Context, email string) (bool, error) {
	for _, u := range r.users {
		if u.Email == email {
			return true, nil
		}
	}
	return false, nil
}

func (r *stubUserRepo) GetUserFirstOrganization(_ context.Context, _ uuid.UUID) (*string, error) {
	return nil, nil
}

func (r *stubUserRepo) UpdateStatus(_ context.Context, id uuid.UUID, status string) error {
	u, ok := r.users[id]
	if !ok {
		return errors.New("user not found")
	}
	u.Status = status
	return nil
}

func (r *stubUserRepo) ListByStatus(_ context.Context, status string, _, _ int) ([]*entities.User, error) {
	var out []*entities.User
	for _, u := range r.users {
		if u.Status == status {
			out = append(out, u)
		}
	}
	return out, nil
}

// stubPermRepo is a minimal in-memory PermissionRepository for use in domain tests.
type stubPermRepo struct {
	perms map[uuid.UUID]map[string]bool // userID -> "resource:action" -> allowed
}

func newStubPermRepo() *stubPermRepo {
	return &stubPermRepo{perms: make(map[uuid.UUID]map[string]bool)}
}

func (r *stubPermRepo) GetByID(_ context.Context, _ uuid.UUID) (*entities.Permission, error) {
	return nil, nil
}

func (r *stubPermRepo) GetByName(_ context.Context, _ string) (*entities.Permission, error) {
	return nil, nil
}

func (r *stubPermRepo) GetUserPermissions(_ context.Context, _ uuid.UUID) ([]*entities.Permission, error) {
	return nil, nil
}

func (r *stubPermRepo) HasPermission(_ context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	m, ok := r.perms[userID]
	if !ok {
		return false, nil
	}
	return m[resource+":"+action], nil
}

// stubRoleRepo is a minimal in-memory RoleRepository for use in domain tests.
type stubRoleRepo struct {
	roles map[uuid.UUID][]*entities.Role
}

func newStubRoleRepo() *stubRoleRepo {
	return &stubRoleRepo{roles: make(map[uuid.UUID][]*entities.Role)}
}

func (r *stubRoleRepo) GetByID(_ context.Context, _ uuid.UUID) (*entities.Role, error) {
	return nil, nil
}

func (r *stubRoleRepo) GetByName(_ context.Context, _ string) (*entities.Role, error) {
	return nil, nil
}

func (r *stubRoleRepo) GetUserRoles(_ context.Context, userID uuid.UUID) ([]*entities.Role, error) {
	return r.roles[userID], nil
}

func (r *stubRoleRepo) GetRolePermissions(_ context.Context, _ uuid.UUID) ([]*entities.Permission, error) {
	return nil, nil
}

func (r *stubRoleRepo) AssignRoleToUser(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

// compile-time interface compliance checks
var _ repositories.UserRepository = (*stubUserRepo)(nil)
var _ repositories.PermissionRepository = (*stubPermRepo)(nil)
var _ repositories.RoleRepository = (*stubRoleRepo)(nil)
