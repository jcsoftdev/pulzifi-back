package deletecurrentuser

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// Handler handles the delete_current_user use case.
type Handler struct {
	userRepo repositories.UserRepository
	eventBus *eventbus.EventBus
}

// NewHandler creates a new delete current user handler.
func NewHandler(userRepo repositories.UserRepository, eventBus *eventbus.EventBus) *Handler {
	return &Handler{userRepo: userRepo, eventBus: eventBus}
}

// Handle deletes the user and publishes a user.deleted event.
func (h *Handler) Handle(ctx context.Context, req *Request) error {
	if err := h.userRepo.Delete(ctx, req.UserID); err != nil {
		logger.Error("delete_current_user: failed to delete user", zap.Error(err))
		return fmt.Errorf("failed to delete account")
	}

	if h.eventBus != nil {
		payload, _ := json.Marshal(map[string]string{"user_id": req.UserID.String()})
		if err := h.eventBus.Publish("user.deleted", req.UserID.String(), payload); err != nil {
			logger.Error("delete_current_user: failed to publish user.deleted event", zap.Error(err))
		}
	}

	return nil
}
