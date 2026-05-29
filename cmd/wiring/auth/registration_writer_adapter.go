package authwiring

import (
	"context"

	adminentities "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/entities"
	adminrepos "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories"
	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
)

// registrationWriterAdapter implements authservices.RegistrationRequestWriter
// by wrapping admin's RegistrationRequestRepository. Lives in cmd/wiring/auth so
// the auth module itself never imports from the admin module.
type registrationWriterAdapter struct {
	repo adminrepos.RegistrationRequestRepository
}

// NewRegistrationWriterAdapter creates a RegistrationRequestWriter backed by
// admin's RegistrationRequestRepository.
func NewRegistrationWriterAdapter(repo adminrepos.RegistrationRequestRepository) authservices.RegistrationRequestWriter {
	return &registrationWriterAdapter{repo: repo}
}

func (a *registrationWriterAdapter) Create(ctx context.Context, req *authservices.PendingRegistration) error {
	adminReq := &adminentities.RegistrationRequest{
		ID:                    req.ID,
		UserID:                req.UserID,
		OrganizationName:      req.OrganizationName,
		OrganizationSubdomain: req.OrganizationSubdomain,
		Status:                adminentities.RegistrationStatusPending,
		CreatedAt:             req.CreatedAt,
		UpdatedAt:             req.UpdatedAt,
	}
	return a.repo.Create(ctx, adminReq)
}

func (a *registrationWriterAdapter) ExistsPendingBySubdomain(ctx context.Context, subdomain string) (bool, error) {
	return a.repo.ExistsPendingBySubdomain(ctx, subdomain)
}
