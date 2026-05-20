package startoauth

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
)

type Request struct {
	Provider    string
	Tenant      string
	ReturnPath  string
	RedirectURI string
	ReturnHost  string
	OrgID       uuid.UUID
	UserID      uuid.UUID
}

type Response struct {
	AuthorizeURL string
}

type Handler struct {
	registry services.ProviderRegistry
	signer   services.StateSignerPort
}

func NewHandler(reg services.ProviderRegistry, signer services.StateSignerPort) *Handler {
	return &Handler{registry: reg, signer: signer}
}

func (h *Handler) Handle(ctx context.Context, req Request) (*Response, error) {
	client, ok := h.registry.Get(req.Provider)
	if !ok {
		return nil, errors.New("unknown provider")
	}
	state, err := h.signer.Sign(services.StateClaims{
		Provider:    req.Provider,
		Tenant:      req.Tenant,
		OrgID:       req.OrgID,
		UserID:      req.UserID,
		ReturnPath:  req.ReturnPath,
		RedirectURI: req.RedirectURI,
		ReturnHost:  req.ReturnHost,
	})
	if err != nil {
		return nil, err
	}
	url, err := client.OAuthAuthorizeURL(state, req.RedirectURI)
	if err != nil {
		return nil, err
	}
	return &Response{AuthorizeURL: url}, nil
}
