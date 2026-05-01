package resendinvitation

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
)

type Emailer interface {
	Send(ctx context.Context, to, subject, html string) error
}

type Response struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	ExpiresAt     time.Time `json:"expires_at"`
	EmailDelivery string    `json:"email_delivery"`
}

type Handler struct {
	repo        repositories.InvitationRepository
	emailer     Emailer
	inviterName string
	frontendURL string
}

func New(repo repositories.InvitationRepository, emailer Emailer, inviterName, frontendURL string) *Handler {
	return &Handler{repo, emailer, inviterName, frontendURL}
}

func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (h *Handler) Handle(ctx context.Context, id uuid.UUID) (*Response, error) {
	tok, err := newToken()
	if err != nil {
		return nil, err
	}
	newExpiry := time.Now().Add(72 * time.Hour)
	if err := h.repo.RotateForResend(ctx, id, tok, newExpiry); err != nil {
		return nil, err
	}

	inv, err := h.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	sendCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	acceptURL := fmt.Sprintf("%s/invite/platform/%s", h.frontendURL, inv.Token)
	subject := "Reminder: invitation to Pulzifi"
	body := fmt.Sprintf(`<p>%s re-sent your invitation. <a href="%s">Accept invitation</a></p><p>Expires %s.</p>`,
		h.inviterName, acceptURL, inv.ExpiresAt.Format(time.RFC1123))

	delivery := "sent"
	now := time.Now().UTC()
	var sendErrPtr *string
	var sentAtPtr *time.Time
	if err := h.emailer.Send(sendCtx, inv.Email, subject, body); err != nil {
		delivery = "failed"
		errStr := err.Error()
		sendErrPtr = &errStr
		logger.Error("invitation resend email failed",
			zap.String("invite_id", inv.ID.String()),
			zap.String("email_hash", logger.HashEmail(inv.Email)),
			zap.Error(err))
	} else {
		sentAtPtr = &now
	}
	_ = h.repo.UpdateEmailDeliveryStatus(ctx, inv.ID, sentAtPtr, sendErrPtr)

	return &Response{
		ID:            inv.ID.String(),
		Email:         inv.Email,
		ExpiresAt:     inv.ExpiresAt,
		EmailDelivery: delivery,
	}, nil
}
