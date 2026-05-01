package acceptinvitation

import (
	"context"
	"strings"

	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories"
	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
)

// Handler accepts a pending invitation, hashing the user-supplied password
// and delegating to the repository's transactional AcceptInvitation. The
// repository is responsible for invitation status, user creation, org row,
// membership, and tenant schema provisioning via ProvisionFunc.
type Handler struct {
	repo          repositories.InvitationRepository
	authService   authservices.AuthService
	provisionFunc func(schemaName string) error
}

// New wires the dependencies. provisionFunc is the tenant-schema migration
// callback (typically shared/database.ProvisionTenantSchema) which the
// repository invokes inside its transaction.
func New(repo repositories.InvitationRepository, authService authservices.AuthService, provisionFunc func(string) error) *Handler {
	return &Handler{repo: repo, authService: authService, provisionFunc: provisionFunc}
}

// Handle hashes the password, normalizes the subdomain, and forwards to the
// repository. Domain errors (ErrInvitationNotFound, ErrSubdomainTaken, etc.)
// propagate untouched so the HTTP layer can map to status codes.
func (h *Handler) Handle(ctx context.Context, token string, req Request) (*Response, error) {
	hashedPwd, err := h.authService.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	out, err := h.repo.AcceptInvitation(ctx, repositories.AcceptInvitationInput{
		Token:         token,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		PasswordHash:  hashedPwd,
		OrgName:       req.OrganizationName,
		OrgSubdomain:  strings.ToLower(req.OrganizationSubdomain),
		ProvisionFunc: h.provisionFunc,
	})
	if err != nil {
		return nil, err
	}

	return &Response{
		UserID:       out.UserID.String(),
		OrgID:        out.OrgID.String(),
		OrgSubdomain: out.OrgSubdomain,
	}, nil
}
