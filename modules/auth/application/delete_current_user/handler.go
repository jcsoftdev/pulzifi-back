package deletecurrentuser

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// Handler handles the delete_current_user use case.
type Handler struct {
	userRepo    repositories.UserRepository
	eventBus    eventbus.MessageBus
	orgCascade  authservices.OrgCascade  // optional; nil → backward-compat (no org cascade)
	userCleanup authservices.UserCleanup // optional; nil → no membership pruning
}

// NewHandler creates a new delete current user handler.
func NewHandler(userRepo repositories.UserRepository, eventBus eventbus.MessageBus) *Handler {
	return &Handler{userRepo: userRepo, eventBus: eventBus}
}

// WithOrgCascade sets the OrgCascade port (injected from cmd/wiring/auth). Returns the
// handler for chaining. Both orgCascade and userCleanup are optional for backward compat.
func (h *Handler) WithOrgCascade(cascade authservices.OrgCascade) *Handler {
	h.orgCascade = cascade
	return h
}

// WithUserCleanup sets the UserCleanup port (injected from cmd/wiring/auth).
func (h *Handler) WithUserCleanup(cleanup authservices.UserCleanup) *Handler {
	h.userCleanup = cleanup
	return h
}

// Handle deletes the current user following the owner/member cascade sequence:
//
//  1. If orgCascade is set: cascade solely-owned orgs first.
//     On ErrBillingActive: abort, user is NOT deleted (caller returns 409).
//  2. If userCleanup is set: prune remaining non-owned memberships (best-effort).
//  3. Hard-delete the user row (FK cascades sessions, tokens, roles).
//  4. Publish user.deleted event (non-destructive listeners only).
//
// The ordering (orgs first, user second) ensures a billing block never leaves
// the user in a deleted state while orgs remain (design §5, DAO-13, DAO-14).
func (h *Handler) Handle(ctx context.Context, req *Request) error {
	// ── Step 1: cascade solely-owned orgs ────────────────────────────────────
	if h.orgCascade != nil {
		if err := h.orgCascade.CascadeSolelyOwnedOrgs(ctx, req.UserID); err != nil {
			if errors.Is(err, authservices.ErrBillingActive) {
				// Billing block: user survives. Return sentinel so the HTTP
				// layer can map to 409 (DAO-13, Scenario 6).
				return authservices.ErrBillingActive
			}
			return fmt.Errorf("delete_current_user: cascade owned orgs: %w", err)
		}
	}

	// ── Step 2: prune remaining memberships (best-effort) ───────────────────
	if h.userCleanup != nil {
		if err := h.userCleanup.PruneMemberships(ctx, req.UserID); err != nil {
			// Non-fatal: log and continue. User deletion must proceed (OQ-1).
			logger.Warn("delete_current_user: prune memberships failed (continuing)",
				zap.String("user_id", req.UserID.String()),
				zap.Error(err),
			)
		}
	}

	// ── Step 3: delete the user row ──────────────────────────────────────────
	if err := h.userRepo.Delete(ctx, req.UserID); err != nil {
		logger.Error("delete_current_user: failed to delete user", zap.Error(err))
		return fmt.Errorf("failed to delete account")
	}

	// ── Step 4: publish non-destructive event ───────────────────────────────
	if h.eventBus != nil {
		payload, _ := json.Marshal(map[string]string{"user_id": req.UserID.String()})
		if err := h.eventBus.Publish("user.deleted", req.UserID.String(), payload); err != nil {
			logger.Error("delete_current_user: failed to publish user.deleted event", zap.Error(err))
		}
	}

	return nil
}
