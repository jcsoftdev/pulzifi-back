package resendinvitation_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	resendinvitation "github.com/jcsoftdev/pulzifi-back/modules/admin/application/resend_invitation"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/entities"
	domerrs "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories/mocks"
)

type fakeEmailer struct {
	sent    bool
	sendErr error
}

func (f *fakeEmailer) Send(ctx context.Context, to, subject, html string) error {
	f.sent = true
	return f.sendErr
}

func TestResend_RotatesToken_OldInvalid(t *testing.T) {
	id := uuid.New()
	rotated := false
	var capturedToken string

	repo := &mocks.InvitationRepositoryMock{
		RotateForResendFunc: func(ctx context.Context, gotID uuid.UUID, newToken string, newExpiresAt time.Time) error {
			rotated = true
			capturedToken = newToken
			if gotID != id {
				t.Fatalf("ID mismatch")
			}
			if newToken == "" {
				t.Fatalf("expected non-empty new token")
			}
			if newExpiresAt.Before(time.Now()) {
				t.Fatalf("expected future expiry")
			}
			return nil
		},
		GetByIDFunc: func(ctx context.Context, gotID uuid.UUID) (*entities.Invitation, error) {
			return &entities.Invitation{
				ID:        gotID,
				Email:     "foo@example.com",
				Token:     capturedToken,
				Status:    entities.InvitationPending,
				ExpiresAt: time.Now().Add(72 * time.Hour),
			}, nil
		},
		UpdateEmailDeliveryStatusFunc: func(ctx context.Context, id uuid.UUID, sentAt *time.Time, sendErr *string) error {
			return nil
		},
	}

	emailer := &fakeEmailer{}
	h := resendinvitation.New(repo, emailer, "Inviter", "http://pulzifi.local:3000")

	resp, err := h.Handle(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if !rotated {
		t.Fatalf("expected RotateForResend to be called")
	}
	if !emailer.sent {
		t.Fatalf("expected emailer to be invoked")
	}
	if resp.EmailDelivery != "sent" {
		t.Fatalf("expected delivery=sent, got %s", resp.EmailDelivery)
	}
}

func TestResend_AlreadyAccepted_Errors(t *testing.T) {
	repo := &mocks.InvitationRepositoryMock{
		RotateForResendFunc: func(ctx context.Context, id uuid.UUID, newToken string, newExpiresAt time.Time) error {
			return domerrs.ErrInvitationAlreadyDecided
		},
	}

	h := resendinvitation.New(repo, &fakeEmailer{}, "Inviter", "http://pulzifi.local:3000")
	_, err := h.Handle(context.Background(), uuid.New())
	if !errors.Is(err, domerrs.ErrInvitationAlreadyDecided) {
		t.Fatalf("expected ErrInvitationAlreadyDecided, got %v", err)
	}
}

func TestResend_EmailFails_StatePersisted(t *testing.T) {
	id := uuid.New()
	var capturedSendErr *string
	var capturedSentAt *time.Time

	repo := &mocks.InvitationRepositoryMock{
		RotateForResendFunc: func(ctx context.Context, gotID uuid.UUID, newToken string, newExpiresAt time.Time) error {
			return nil
		},
		GetByIDFunc: func(ctx context.Context, gotID uuid.UUID) (*entities.Invitation, error) {
			return &entities.Invitation{
				ID:        gotID,
				Email:     "foo@example.com",
				Token:     "tok",
				Status:    entities.InvitationPending,
				ExpiresAt: time.Now().Add(72 * time.Hour),
			}, nil
		},
		UpdateEmailDeliveryStatusFunc: func(ctx context.Context, gotID uuid.UUID, sentAt *time.Time, sendErr *string) error {
			capturedSendErr = sendErr
			capturedSentAt = sentAt
			return nil
		},
	}

	emailer := &fakeEmailer{sendErr: errors.New("smtp boom")}
	h := resendinvitation.New(repo, emailer, "Inviter", "http://pulzifi.local:3000")

	resp, err := h.Handle(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if resp.EmailDelivery != "failed" {
		t.Fatalf("expected delivery=failed, got %s", resp.EmailDelivery)
	}
	if capturedSendErr == nil || *capturedSendErr != "smtp boom" {
		t.Fatalf("expected captured send error to be 'smtp boom', got %v", capturedSendErr)
	}
	if capturedSentAt != nil {
		t.Fatalf("expected sentAt to be nil on failure, got %v", capturedSentAt)
	}
}
