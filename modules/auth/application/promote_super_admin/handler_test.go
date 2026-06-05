package promotesuperadmin_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	promote "github.com/jcsoftdev/pulzifi-back/modules/auth/application/promote_super_admin"
)

// ── In-memory fakes ───────────────────────────────────────────────────────────

type fakeUserRepo struct {
	users map[uuid.UUID]*entities.User
	err   error
}

func (f *fakeUserRepo) GetByID(_ context.Context, id uuid.UUID) (*entities.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	u, ok := f.users[id]
	if !ok {
		return nil, nil
	}
	return u, nil
}

type fakeRoleRepo struct {
	rolesByName     map[string]*entities.Role
	userRoles       map[uuid.UUID][]*entities.Role
	assignedRoles   []assignCall
	getByNameErr    error
	getUserRolesErr error
	assignErr       error
}

type assignCall struct {
	userID uuid.UUID
	roleID uuid.UUID
}

func (f *fakeRoleRepo) GetByName(_ context.Context, name string) (*entities.Role, error) {
	if f.getByNameErr != nil {
		return nil, f.getByNameErr
	}
	r, ok := f.rolesByName[name]
	if !ok {
		return nil, nil
	}
	return r, nil
}

func (f *fakeRoleRepo) GetUserRoles(_ context.Context, userID uuid.UUID) ([]*entities.Role, error) {
	if f.getUserRolesErr != nil {
		return nil, f.getUserRolesErr
	}
	return f.userRoles[userID], nil
}

func (f *fakeRoleRepo) AssignRoleToUser(_ context.Context, userID, roleID uuid.UUID) error {
	if f.assignErr != nil {
		return f.assignErr
	}
	f.assignedRoles = append(f.assignedRoles, assignCall{userID: userID, roleID: roleID})
	return nil
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestPromoteSuperAdminHandler_Handle(t *testing.T) {
	superAdminRole := &entities.Role{ID: uuid.New(), Name: "SUPER_ADMIN"}
	targetID := uuid.New()
	targetUser := &entities.User{
		ID:    targetID,
		Email: "target@example.com",
	}

	tests := []struct {
		name        string
		targetID    uuid.UUID
		userRepo    *fakeUserRepo
		roleRepo    *fakeRoleRepo
		wantSentinel error
		wantAnyErr  bool
		checkResult func(t *testing.T, resp *promote.Response)
	}{
		{
			name:     "promotes user successfully",
			targetID: targetID,
			userRepo: &fakeUserRepo{users: map[uuid.UUID]*entities.User{targetID: targetUser}},
			roleRepo: &fakeRoleRepo{
				rolesByName: map[string]*entities.Role{"SUPER_ADMIN": superAdminRole},
				userRoles:   map[uuid.UUID][]*entities.Role{targetID: {}},
			},
			checkResult: func(t *testing.T, resp *promote.Response) {
				t.Helper()
				if resp.ID != targetID.String() {
					t.Errorf("ID: want %q, got %q", targetID.String(), resp.ID)
				}
				if resp.Email != "target@example.com" {
					t.Errorf("Email: want %q, got %q", "target@example.com", resp.Email)
				}
				if !resp.IsSuperAdmin {
					t.Error("IsSuperAdmin: want true, got false")
				}
			},
		},
		{
			name:         "returns ErrUserNotFound when user does not exist",
			targetID:     uuid.New(),
			userRepo:     &fakeUserRepo{users: map[uuid.UUID]*entities.User{}},
			roleRepo:     &fakeRoleRepo{
				rolesByName: map[string]*entities.Role{"SUPER_ADMIN": superAdminRole},
				userRoles:   map[uuid.UUID][]*entities.Role{},
			},
			wantSentinel: promote.ErrUserNotFound,
		},
		{
			name:     "returns ErrAlreadySuperAdmin when user already has role",
			targetID: targetID,
			userRepo: &fakeUserRepo{users: map[uuid.UUID]*entities.User{targetID: targetUser}},
			roleRepo: &fakeRoleRepo{
				rolesByName: map[string]*entities.Role{"SUPER_ADMIN": superAdminRole},
				userRoles: map[uuid.UUID][]*entities.Role{
					targetID: {{ID: superAdminRole.ID, Name: "SUPER_ADMIN"}},
				},
			},
			wantSentinel: promote.ErrAlreadySuperAdmin,
		},
		{
			name:       "propagates user repo error",
			targetID:   targetID,
			userRepo:   &fakeUserRepo{err: errors.New("db error")},
			roleRepo:   &fakeRoleRepo{},
			wantAnyErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := promote.NewHandler(tt.userRepo, tt.roleRepo)
			resp, err := h.Handle(context.Background(), tt.targetID)

			if tt.wantSentinel != nil {
				if err == nil {
					t.Fatalf("expected sentinel error %v, got nil", tt.wantSentinel)
				}
				if !errors.Is(err, tt.wantSentinel) {
					t.Errorf("error: want sentinel %v, got %v", tt.wantSentinel, err)
				}
				return
			}

			if tt.wantAnyErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.checkResult != nil {
				tt.checkResult(t, resp)
			}
		})
	}
}
