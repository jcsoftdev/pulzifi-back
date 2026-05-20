package oauth

import (
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
)

// StateSignerAdapter wraps *StateSigner to satisfy the domain services.StateSignerPort interface.
type StateSignerAdapter struct {
	signer *StateSigner
}

// NewStateSignerAdapter wraps a *StateSigner as a domain port.
func NewStateSignerAdapter(s *StateSigner) services.StateSignerPort {
	return &StateSignerAdapter{signer: s}
}

func (a *StateSignerAdapter) Sign(c services.StateClaims) (string, error) {
	return a.signer.Sign(StateClaims{
		Provider:    c.Provider,
		Tenant:      c.Tenant,
		OrgID:       c.OrgID,
		UserID:      c.UserID,
		ReturnPath:  c.ReturnPath,
		RedirectURI: c.RedirectURI,
		ReturnHost:  c.ReturnHost,
	})
}

func (a *StateSignerAdapter) Verify(token string) (*services.StateClaims, error) {
	c, err := a.signer.Verify(token)
	if err != nil {
		return nil, err
	}
	return &services.StateClaims{
		Provider:    c.Provider,
		Tenant:      c.Tenant,
		OrgID:       c.OrgID,
		UserID:      c.UserID,
		ReturnPath:  c.ReturnPath,
		RedirectURI: c.RedirectURI,
		ReturnHost:  c.ReturnHost,
	}, nil
}
