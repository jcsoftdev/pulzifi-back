package promotesuperadmin

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
)

// Sentinel errors mapped to HTTP status codes by the HTTP layer.
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrAlreadySuperAdmin = errors.New("user is already a super admin")
)

// UserRepository is the narrow slice of repositories.UserRepository needed here.
type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
}

// RoleRepository is the narrow slice of repositories.RoleRepository needed here.
type RoleRepository interface {
	GetByName(ctx context.Context, name string) (*entities.Role, error)
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*entities.Role, error)
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error
}

// Response is the success payload returned by the handler.
type Response struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	IsSuperAdmin bool   `json:"isSuperAdmin"`
}

// Handler orchestrates the promote-to-super-admin use case.
type Handler struct {
	userRepo UserRepository
	roleRepo RoleRepository
}

// NewHandler creates a Handler with the required dependencies.
func NewHandler(userRepo UserRepository, roleRepo RoleRepository) *Handler {
	return &Handler{userRepo: userRepo, roleRepo: roleRepo}
}

// Handle promotes the target user to SUPER_ADMIN.
// Returns ErrUserNotFound if the user does not exist,
// ErrAlreadySuperAdmin if they already have the role.
func (h *Handler) Handle(ctx context.Context, targetID uuid.UUID) (*Response, error) {
	user, err := h.userRepo.GetByID(ctx, targetID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	userRoles, err := h.roleRepo.GetUserRoles(ctx, targetID)
	if err != nil {
		return nil, err
	}
	for _, r := range userRoles {
		if r.Name == "SUPER_ADMIN" {
			return nil, ErrAlreadySuperAdmin
		}
	}

	superAdminRole, err := h.roleRepo.GetByName(ctx, "SUPER_ADMIN")
	if err != nil {
		return nil, err
	}
	if superAdminRole == nil {
		return nil, errors.New("SUPER_ADMIN role not found in database")
	}

	if err := h.roleRepo.AssignRoleToUser(ctx, targetID, superAdminRole.ID); err != nil {
		return nil, err
	}

	return &Response{
		ID:           user.ID.String(),
		Email:        user.Email,
		IsSuperAdmin: true,
	}, nil
}
