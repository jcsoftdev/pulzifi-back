package handleoauthcallback

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
	intoauth "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/oauth"
)

// Request carries the OAuth callback parameters received from the provider redirect.
type Request struct {
	Provider string
	Code     string
	State    string
}

// Response contains routing information extracted from the verified state token.
type Response struct {
	Tenant     string
	ReturnPath string
	Provider   string
	ReturnHost string
}

// Handler verifies the OAuth state token, exchanges the code for tokens, and
// upserts an Integration row (replace existing if an active integration for the
// same org + service already exists).
type Handler struct {
	registry services.ProviderRegistry
	signer   *intoauth.StateSigner
	repo     repositories.IntegrationRepository
}

// NewHandler constructs a Handler with its dependencies.
func NewHandler(r services.ProviderRegistry, signer *intoauth.StateSigner, repo repositories.IntegrationRepository) *Handler {
	return &Handler{registry: r, signer: signer, repo: repo}
}

// Handle executes the OAuth callback use case.
func (h *Handler) Handle(ctx context.Context, req Request) (*Response, error) {
	st, err := h.signer.Verify(req.State)
	if err != nil {
		return nil, errors.New("invalid or expired state: " + err.Error())
	}
	if st.Provider != req.Provider {
		return nil, errors.New("provider mismatch")
	}

	client, ok := h.registry.Get(req.Provider)
	if !ok {
		return nil, errors.New("unknown provider")
	}

	tok, err := client.HandleCallback(ctx, req.Code, st.RedirectURI)
	if err != nil {
		return nil, err
	}

	// Upsert: if existing active integration for org+service, replace credentials.
	existing, _ := h.repo.GetByOrgAndService(ctx, st.OrgID, req.Provider)
	integ := &entities.Integration{
		ID:             uuid.New(),
		OrgID:          st.OrgID,
		ServiceType:    req.Provider,
		Status:         entities.IntegrationActive,
		AccessToken:    tok.AccessToken,
		RefreshToken:   tok.RefreshToken,
		TokenExpiresAt: tok.ExpiresAt,
		ProviderMeta:   tok.ProviderMeta,
		CreatedBy:      st.UserID,
	}
	if existing != nil {
		integ.ID = existing.ID
		if err := h.repo.Update(ctx, integ); err != nil {
			return nil, err
		}
	} else {
		if err := h.repo.Create(ctx, integ); err != nil {
			return nil, err
		}
	}

	return &Response{
		Tenant:     st.Tenant,
		ReturnPath: st.ReturnPath,
		Provider:   req.Provider,
		ReturnHost: st.ReturnHost,
	}, nil
}
