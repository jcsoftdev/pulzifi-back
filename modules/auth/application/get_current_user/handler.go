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
	Avatar    *string `json:"avatar,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// Handle executes the get current user use case
func (h *Handler) Handle(ctx context.Context, userID uuid.UUID) (*Response, error) {
	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	return h.toResponse(user), nil
}

func (h *Handler) toResponse(user *entities.User) *Response {
	name := user.FirstName
	if user.LastName != "" {
		name = user.FirstName + " " + user.LastName
	}

	return &Response{
		ID:        user.ID.String(),
		Name:      name,
		Email:     user.Email,
		Role:      "ADMIN",
		Avatar:    user.AvatarURL,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
