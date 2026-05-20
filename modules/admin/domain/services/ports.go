package services

import (
	"context"

	"github.com/google/uuid"
)

// PendingUser is admin's own domain type for a user awaiting approval.
// It contains only the fields that the admin module needs to read.
type PendingUser struct {
	ID        uuid.UUID
	Email     string
	FirstName string
	LastName  string
}

// PendingUserReader retrieves user information needed by the approval workflow.
// Implemented by cmd/wiring/admin — wraps the auth module's UserRepository.
type PendingUserReader interface {
	// GetByID returns a PendingUser by their ID, or (nil, nil) if not found.
	GetByID(ctx context.Context, id uuid.UUID) (*PendingUser, error)
}

// ApprovalInput carries the data needed to provision an organization on approval.
type ApprovalInput struct {
	UserID                uuid.UUID
	OrganizationName      string
	OrganizationSubdomain string
	ReviewerID            uuid.UUID
	RequestID             uuid.UUID
}

// ApprovalProvisioner atomically approves a user registration:
//   - updates user status to approved
//   - updates the registration request status
//   - creates or reuses the organization
//   - assigns the owner/member role
//   - assigns the ADMIN role
//   - assigns the default plan
//   - provisions the tenant schema (outside the transaction)
//
// Implemented by cmd/wiring/admin — encapsulates the transaction and
// cross-module coordination so the application layer stays free of DB concerns.
type ApprovalProvisioner interface {
	Provision(ctx context.Context, input ApprovalInput) error
}

// RejectionInput carries the data needed to reject a user registration.
type RejectionInput struct {
	UserID     uuid.UUID
	ReviewerID uuid.UUID
	RequestID  uuid.UUID
}

// RejectionProvisioner atomically rejects a user registration:
//   - updates user status to rejected
//   - updates the registration request status
//
// Implemented by cmd/wiring/admin.
type RejectionProvisioner interface {
	Reject(ctx context.Context, input RejectionInput) error
}

// RegistrationNotifier sends email notifications for approval/rejection outcomes.
// Implemented by cmd/wiring/admin — wraps email module's provider + templates.
type RegistrationNotifier interface {
	// SendApproval sends an approval confirmation email to the user.
	SendApproval(ctx context.Context, toEmail, firstName, orgSubdomain, loginURL string) error

	// SendRejection sends a rejection notification email to the user.
	SendRejection(ctx context.Context, toEmail, firstName string) error
}
