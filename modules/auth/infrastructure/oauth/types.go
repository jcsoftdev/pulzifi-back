package oauth

import (
	"context"

	"golang.org/x/oauth2"
)

// OAuthUser represents a normalized user profile from any OAuth provider.
type OAuthUser struct {
	ProviderID string
	Email      string
	FirstName  string
	LastName   string
	AvatarURL  string
}

// Provider defines the interface for OAuth providers.
type Provider interface {
	AuthCodeURL(state string) string
	Exchange(ctx context.Context, code string) (*OAuthUser, *oauth2.Token, error)
}
