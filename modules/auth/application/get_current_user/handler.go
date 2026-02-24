package get_current_user

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
)

// Handler handles the get current user use case
type Handler struct {
	userRepo repositories.UserRepository
}

// NewHandler creates a new get current user handler
func NewHandler(userRepo repositories.UserRepository) *Handler {
	return &Handler{
		userRepo: userRepo,
	}
}

// Response represents the current user response
type Response struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	Role      string  `json:"role"`
	Status    string  `json:"status"`
	Avatar    *string `json:"avatar,omitempty"`
	Tenant    *string `json:"tenant,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// Handle executes the get current user use case.
// roles should be the list already validated from the JWT token claims.
func (h *Handler) Handle(ctx context.Context, userID uuid.UUID, roles []string) (*Response, error) {
	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	resp := h.toResponse(user, highestRole(roles))

	tenant, err := h.userRepo.GetUserFirstOrganization(ctx, userID)
	if err == nil {
		resp.Tenant = tenant
	}

	return resp, nil
}

// highestRole returns the most privileged role name from a list of role name strings.
// Priority: SUPER_ADMIN > ADMIN > USER > VIEWER
func highestRole(roles []string) string {
	priority := map[string]int{
		"SUPER_ADMIN": 4,
		"ADMIN":       3,
		"USER":        2,
		"VIEWER":      1,
	}
	best := "USER"
	bestPrio := 0
	for _, r := range roles {
		if p, ok := priority[r]; ok && p > bestPrio {
			bestPrio = p
			best = r
		}
	}
	return best
}

func (h *Handler) toResponse(user *entities.User, role string) *Response {
	name := user.FirstName
	if user.LastName != "" {
		name = user.FirstName + " " + user.LastName
	}

	status := user.Status
	if status == "" {
		status = "approved"
	}

	return &Response{
		ID:        user.ID.String(),
		Name:      name,
		Email:     user.Email,
		Role:      role,
		Status:    status,
		Avatar:    user.AvatarURL,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
