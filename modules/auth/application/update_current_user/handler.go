package updatecurrentuser

import (
	"context"
	"fmt"

	"github.com/jcsoftdev/pulzifi-back/modules/auth/application/getcurrentuser"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
)

// Handler handles the update_current_user use case.
type Handler struct {
	userRepo              repositories.UserRepository
	getCurrentUserHandler *getcurrentuser.Handler
}

// NewHandler creates a new update current user handler.
func NewHandler(userRepo repositories.UserRepository, getCurrentUserHandler *getcurrentuser.Handler) *Handler {
	return &Handler{userRepo: userRepo, getCurrentUserHandler: getCurrentUserHandler}
}

// Handle updates a user's profile and returns the updated user.
func (h *Handler) Handle(ctx context.Context, req *Request) (*getcurrentuser.Response, error) {
	user, err := h.userRepo.GetByID(ctx, req.UserID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}

	if err := h.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update profile")
	}

	return h.getCurrentUserHandler.Handle(ctx, req.UserID, req.Roles)
}
