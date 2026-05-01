package invitetoplatform

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
)

type Emailer interface {
	Send(ctx context.Context, to, subject, html string) error
}

type Handler struct {
	repo               repositories.InvitationRepository
	emailer            Emailer
	inviterName        string
	frontendURL        string
	dailyCapPerInviter int
	dailyCapGlobal     int
}

func New(repo repositories.InvitationRepository, emailer Emailer, inviterName, frontendURL string, capPerInviter, capGlobal int) *Handler {
	return &Handler{repo, emailer, inviterName, frontendURL, capPerInviter, capGlobal}
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (h *Handler) Handle(ctx context.Context, req Request, inviterID uuid.UUID) (*Response, error) {
	token, err := generateToken()
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(72 * time.Hour)
	email := strings.ToLower(req.Email)

	inv, err := h.repo.Create(ctx, repositories.CreateInvitationInput{
		Email:              email,
		Token:              token,
		InvitedBy:          inviterID,
		ExpiresAt:          expiresAt,
		DailyCapPerInviter: h.dailyCapPerInviter,
		DailyCapGlobal:     h.dailyCapGlobal,
	})
	if err != nil {
		return nil, err
	}

	// Synchronous email with 5s timeout — outcome reported via response field.
	sendCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	acceptURL := fmt.Sprintf("%s/invite/platform/%s", h.frontendURL, token)
	subject := "You're invited to Pulzifi"
	body := fmt.Sprintf(`<p>%s invited you to create an organization on Pulzifi.</p><p><a href="%s">Accept invitation</a></p><p>Expires %s.</p>`,
		h.inviterName, acceptURL, expiresAt.Format(time.RFC1123))
	// NOTE: HTML inlined here for Phase 5; Phase 8 swaps to a real template.

	delivery := "sent"
	now := time.Now().UTC()
	var sendErrPtr *string
	var sentAtPtr *time.Time
	if err := h.emailer.Send(sendCtx, inv.Email, subject, body); err != nil {
		delivery = "failed"
		errStr := err.Error()
		sendErrPtr = &errStr
		logger.Error("invitation email send failed",
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
