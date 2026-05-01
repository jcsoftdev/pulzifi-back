package getinvitation

import (
	"context"
	"fmt"
	"time"

	domerrs "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories"
	authrepos "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
)

// Handler resolves a public invitation token to a minimal preview payload.
// It does NOT consume the invitation; that is the responsibility of the
// accept_invitation use case.
type Handler struct {
	repo     repositories.InvitationRepository
	userRepo authrepos.UserRepository
}

// New wires a Handler with the invitation repository (admin module) and the
// user repository (auth module) used to fetch the inviter's display name.
func New(repo repositories.InvitationRepository, userRepo authrepos.UserRepository) *Handler {
	return &Handler{repo: repo, userRepo: userRepo}
}

// Handle returns a public Response for the invitation identified by token.
// Returns ErrInvitationNotFound when the invitation is missing, expired, or
// no longer pending — the HTTP layer maps this to 410 Gone.
func (h *Handler) Handle(ctx context.Context, token string) (*Response, error) {
	inv, err := h.repo.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if !inv.IsActive(time.Now()) {
		return nil, domerrs.ErrInvitationNotFound
	}

	inviter, err := h.userRepo.GetByID(ctx, inv.InvitedBy)
	if err != nil {
		// Inviter lookup is best-effort. If the row is missing we still want
		// the invitee to land on the acceptance page — fall back to a generic
		// label rather than failing the whole flow.
		return &Response{
			Email:         inv.Email,
			InvitedByName: "Pulzifi Admin",
			ExpiresAt:     inv.ExpiresAt,
		}, nil
	}

	name := fmt.Sprintf("%s %s", inviter.FirstName, inviter.LastName)
	return &Response{
		Email:         inv.Email,
		InvitedByName: name,
		ExpiresAt:     inv.ExpiresAt,
	}, nil
}
