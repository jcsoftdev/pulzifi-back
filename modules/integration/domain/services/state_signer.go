package services

import "github.com/google/uuid"

// StateClaims is the domain-level representation of OAuth state claims.
// It mirrors the infrastructure StateClaims without importing infrastructure.
type StateClaims struct {
	Provider    string
	Tenant      string
	OrgID       uuid.UUID
	UserID      uuid.UUID
	ReturnPath  string
	RedirectURI string
	ReturnHost  string
}

// StateSignerPort abstracts signing and verifying OAuth state tokens.
// The concrete implementation lives in infrastructure/oauth.
type StateSignerPort interface {
	Sign(claims StateClaims) (string, error)
	Verify(token string) (*StateClaims, error)
}
