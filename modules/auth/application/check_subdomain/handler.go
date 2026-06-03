package checksubdomain

import (
	"context"
	"strings"

	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// Handler checks whether a subdomain is available for registration
type Handler struct {
	orgDirectory authservices.OrganizationDirectory
}

// NewHandler creates a new handler instance
func NewHandler(
	orgDirectory authservices.OrganizationDirectory,
) *Handler {
	return &Handler{orgDirectory: orgDirectory}
}

// Response is the result of a subdomain availability check
type Response struct {
	Available bool   `json:"available"`
	Message   string `json:"message,omitempty"`
}

// Handle executes the subdomain availability check
func (h *Handler) Handle(ctx context.Context, subdomain string) (*Response, error) {
	subdomain = strings.TrimSpace(strings.ToLower(subdomain))

	if err := h.orgDirectory.ValidateSubdomain(subdomain); err != nil {
		return &Response{Available: false, Message: err.Error()}, nil
	}

	count, err := h.orgDirectory.CountBySubdomain(ctx, subdomain)
	if err != nil {
		logger.Error("Failed to check subdomain uniqueness", zap.Error(err))
		return nil, err
	}
	if count > 0 {
		return &Response{Available: false, Message: "subdomain is already in use"}, nil
	}

	return &Response{Available: true}, nil
}
