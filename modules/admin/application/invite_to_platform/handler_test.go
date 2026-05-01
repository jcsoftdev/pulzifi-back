package invitetoplatform_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	invitetoplatform "github.com/jcsoftdev/pulzifi-back/modules/admin/application/invite_to_platform"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/entities"
	domerrs "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories"
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

func newRepoMock(create func(context.Context, repositories.CreateInvitationInput) (*entities.Invitation, error)) *mocks.InvitationRepositoryMock {
	return &mocks.InvitationRepositoryMock{
		CreateFunc: create,
		UpdateEmailDeliveryStatusFunc: func(ctx context.Context, id uuid.UUID, sentAt *time.Time, sendErr *string) error {
			return nil
		},
	}
}

func TestHandle_HappyPath(t *testing.T) {
	repo := newRepoMock(func(ctx context.Context, in repositories.CreateInvitationInput) (*entities.Invitation, error) {
		return &entities.Invitation{ID: uuid.New(), Email: in.Email, Token: in.Token, ExpiresAt: in.ExpiresAt}, nil
	})
	h := invitetoplatform.New(repo, &fakeEmailer{}, "Inviter", "http://pulzifi.local:3000", 50, 200)

	resp, err := h.Handle(context.Background(), invitetoplatform.Request{Email: "Foo@Example.com"}, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if resp.EmailDelivery != "sent" {
		t.Fatalf("got %s", resp.EmailDelivery)
	}
	if resp.Email != "foo@example.com" {
		t.Fatalf("expected lowercased email, got %s", resp.Email)
	}
}

func TestHandle_DuplicateEmail(t *testing.T) {
	repo := newRepoMock(func(ctx context.Context, in repositories.CreateInvitationInput) (*entities.Invitation, error) {
		return nil, domerrs.ErrCannotInviteEmail
	})
	h := invitetoplatform.New(repo, &fakeEmailer{}, "Inviter", "http://pulzifi.local:3000", 50, 200)
	_, err := h.Handle(context.Background(), invitetoplatform.Request{Email: "x@y.com"}, uuid.New())
	if !errors.Is(err, domerrs.ErrCannotInviteEmail) {
		t.Fatalf("got %v", err)
	}
}

func TestHandle_EmailSendTimeout_Returns201Failed(t *testing.T) {
	repo := newRepoMock(func(ctx context.Context, in repositories.CreateInvitationInput) (*entities.Invitation, error) {
		return &entities.Invitation{ID: uuid.New(), Email: in.Email, Token: in.Token, ExpiresAt: in.ExpiresAt}, nil
	})
	emailer := &fakeEmailer{sendErr: context.DeadlineExceeded}
	h := invitetoplatform.New(repo, emailer, "Inviter", "http://pulzifi.local:3000", 50, 200)

	resp, err := h.Handle(context.Background(), invitetoplatform.Request{Email: "x@y.com"}, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if resp.EmailDelivery != "failed" {
		t.Fatalf("got %s", resp.EmailDelivery)
	}
}
