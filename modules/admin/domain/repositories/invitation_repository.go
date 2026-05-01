package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/entities"
)

type CreateInvitationInput struct {
	Email              string
	Token              string
	InvitedBy          uuid.UUID
	ExpiresAt          time.Time
	DailyCapPerInviter int
	DailyCapGlobal     int
}

type AcceptInvitationInput struct {
	Token         string
	FirstName     string
	LastName      string
	PasswordHash  string
	OrgName       string
	OrgSubdomain  string
	ProvisionFunc func(schemaName string) error
}

type AcceptInvitationOutput struct {
	UserID        uuid.UUID
	OrgID         uuid.UUID
	OrgSubdomain  string
	OrgSchemaName string
}

type ListInvitationsFilter struct {
	Status string
	Limit  int
	Offset int
}

type InvitationRepository interface {
	Create(ctx context.Context, in CreateInvitationInput) (*entities.Invitation, error)
	GetByToken(ctx context.Context, token string) (*entities.Invitation, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Invitation, error)
	List(ctx context.Context, f ListInvitationsFilter) ([]*entities.Invitation, int, error)
	AcceptInvitation(ctx context.Context, in AcceptInvitationInput) (*AcceptInvitationOutput, error)
	Revoke(ctx context.Context, id uuid.UUID, revokedBy uuid.UUID) error
	UpdateEmailDeliveryStatus(ctx context.Context, id uuid.UUID, sentAt *time.Time, sendErr *string) error
	RotateForResend(ctx context.Context, id uuid.UUID, newToken string, newExpiresAt time.Time) error
}
