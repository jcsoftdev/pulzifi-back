package services

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TrialProvisionInput carries the data needed to bootstrap a self-serve org on a trial.
type TrialProvisionInput struct {
	UserID                uuid.UUID
	OrganizationName      string
	OrganizationSubdomain string
	TrialDays             int
}

// TrialProvisionOutput reports the result of provisioning a self-serve trial.
type TrialProvisionOutput struct {
	OrganizationID        uuid.UUID
	OrganizationSubdomain string
	SchemaName            string
	TrialEndsAt           time.Time
}

// TrialProvisioner is auth's port for atomically creating an organization with
// the trial plan and a provisioned tenant schema during self-serve registration.
//
// Implementations live in cmd/wiring/auth/ to keep the auth module decoupled
// from organization, usage, and database packages.
type TrialProvisioner interface {
	Provision(ctx context.Context, in TrialProvisionInput) (*TrialProvisionOutput, error)
}
