package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories"
)

// InvitationRepositoryMock is a func-field mock implementation of
// repositories.InvitationRepository for use in unit tests. Each *Func is
// optional; when unset, the corresponding method panics so missing
// expectations are surfaced loudly during test development.
type InvitationRepositoryMock struct {
	CreateFunc                    func(ctx context.Context, in repositories.CreateInvitationInput) (*entities.Invitation, error)
	GetByTokenFunc                func(ctx context.Context, token string) (*entities.Invitation, error)
	GetByIDFunc                   func(ctx context.Context, id uuid.UUID) (*entities.Invitation, error)
	ListFunc                      func(ctx context.Context, f repositories.ListInvitationsFilter) ([]*entities.Invitation, int, error)
	AcceptInvitationFunc          func(ctx context.Context, in repositories.AcceptInvitationInput) (*repositories.AcceptInvitationOutput, error)
	RevokeFunc                    func(ctx context.Context, id uuid.UUID, revokedBy uuid.UUID) error
	UpdateEmailDeliveryStatusFunc func(ctx context.Context, id uuid.UUID, sentAt *time.Time, sendErr *string) error
	RotateForResendFunc           func(ctx context.Context, id uuid.UUID, newToken string, newExpiresAt time.Time) error
}

func (m *InvitationRepositoryMock) Create(ctx context.Context, in repositories.CreateInvitationInput) (*entities.Invitation, error) {
	if m.CreateFunc == nil {
		panic("InvitationRepositoryMock.Create not set")
	}
	return m.CreateFunc(ctx, in)
}

func (m *InvitationRepositoryMock) GetByToken(ctx context.Context, token string) (*entities.Invitation, error) {
	if m.GetByTokenFunc == nil {
		panic("InvitationRepositoryMock.GetByToken not set")
	}
	return m.GetByTokenFunc(ctx, token)
}

func (m *InvitationRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*entities.Invitation, error) {
	if m.GetByIDFunc == nil {
		panic("InvitationRepositoryMock.GetByID not set")
	}
	return m.GetByIDFunc(ctx, id)
}

func (m *InvitationRepositoryMock) List(ctx context.Context, f repositories.ListInvitationsFilter) ([]*entities.Invitation, int, error) {
	if m.ListFunc == nil {
		panic("InvitationRepositoryMock.List not set")
	}
	return m.ListFunc(ctx, f)
}

func (m *InvitationRepositoryMock) AcceptInvitation(ctx context.Context, in repositories.AcceptInvitationInput) (*repositories.AcceptInvitationOutput, error) {
	if m.AcceptInvitationFunc == nil {
		panic("InvitationRepositoryMock.AcceptInvitation not set")
	}
	return m.AcceptInvitationFunc(ctx, in)
}

func (m *InvitationRepositoryMock) Revoke(ctx context.Context, id uuid.UUID, revokedBy uuid.UUID) error {
	if m.RevokeFunc == nil {
		panic("InvitationRepositoryMock.Revoke not set")
	}
	return m.RevokeFunc(ctx, id, revokedBy)
}

func (m *InvitationRepositoryMock) UpdateEmailDeliveryStatus(ctx context.Context, id uuid.UUID, sentAt *time.Time, sendErr *string) error {
	if m.UpdateEmailDeliveryStatusFunc == nil {
		panic("InvitationRepositoryMock.UpdateEmailDeliveryStatus not set")
	}
	return m.UpdateEmailDeliveryStatusFunc(ctx, id, sentAt, sendErr)
}

func (m *InvitationRepositoryMock) RotateForResend(ctx context.Context, id uuid.UUID, newToken string, newExpiresAt time.Time) error {
	if m.RotateForResendFunc == nil {
		panic("InvitationRepositoryMock.RotateForResend not set")
	}
	return m.RotateForResendFunc(ctx, id, newToken, newExpiresAt)
}

// Compile-time interface conformance check.
var _ repositories.InvitationRepository = (*InvitationRepositoryMock)(nil)
