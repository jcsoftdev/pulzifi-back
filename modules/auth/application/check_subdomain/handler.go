package checksubdomain

import (
	"context"
	"strings"

	adminrepos "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories"
	orgrepos "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/repositories"
	orgservices "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// Handler checks whether a subdomain is available for registration
type Handler struct {
	regReqRepo adminrepos.RegistrationRequestRepository
	orgRepo    orgrepos.OrganizationRepository
	orgService *orgservices.OrganizationService
}

// NewHandler creates a new handler instance
func NewHandler(
	regReqRepo adminrepos.RegistrationRequestRepository,
	orgRepo orgrepos.OrganizationRepository,
	orgService *orgservices.OrganizationService,
) *Handler {
	return &Handler{regReqRepo: regReqRepo, orgRepo: orgRepo, orgService: orgService}
}

// Response is the result of a subdomain availability check
type Response struct {
	Available bool   `json:"available"`
	Message   string `json:"message,omitempty"`
}

// Handle executes the subdomain availability check
func (h *Handler) Handle(ctx context.Context, subdomain string) (*Response, error) {
	subdomain = strings.TrimSpace(strings.ToLower(subdomain))

	if err := h.orgService.ValidateSubdomain(subdomain); err != nil {
		return &Response{Available: false, Message: err.Error()}, nil
	}

	count, err := h.orgRepo.CountBySubdomain(ctx, subdomain)
	if err != nil {
		logger.Error("Failed to check subdomain uniqueness", zap.Error(err))
		return nil, err
	}
	if count > 0 {
		return &Response{Available: false, Message: "subdomain is already in use"}, nil
	}

	pendingExists, err := h.regReqRepo.ExistsPendingBySubdomain(ctx, subdomain)
	if err != nil {
		logger.Error("Failed to check pending subdomain", zap.Error(err))
		return nil, err
	}
	if pendingExists {
		return &Response{Available: false, Message: "subdomain is already pending registration approval"}, nil
	}

	return &Response{Available: true}, nil
}
