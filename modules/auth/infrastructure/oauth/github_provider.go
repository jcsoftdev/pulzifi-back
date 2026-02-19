package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// GitHubProvider handles GitHub OAuth authentication.
type GitHubProvider struct {
	config *oauth2.Config
}

// GitHubUser represents the user profile returned by GitHub.
type GitHubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// GitHubEmail represents an email from the GitHub emails API.
type GitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

// NewGitHubProvider creates a new GitHub OAuth provider.
func NewGitHubProvider(clientID, clientSecret, redirectURL string) *GitHubProvider {
	return &GitHubProvider{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"user:email", "read:user"},
			Endpoint:     github.Endpoint,
		},
	}
}

// AuthCodeURL returns the URL to redirect the user to for authorization.
func (p *GitHubProvider) AuthCodeURL(state string) string {
	return p.config.AuthCodeURL(state)
}

// Exchange exchanges an authorization code for a token and fetches the user profile.
func (p *GitHubProvider) Exchange(ctx context.Context, code string) (*OAuthUser, *oauth2.Token, error) {
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("github: failed to exchange code: %w", err)
	}

	client := p.config.Client(ctx, token)

	// Fetch user profile
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, nil, fmt.Errorf("github: failed to fetch user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("github: user API returned %d: %s", resp.StatusCode, body)
	}

	var gu GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&gu); err != nil {
		return nil, nil, fmt.Errorf("github: failed to decode user info: %w", err)
	}

	email := gu.Email
	if email == "" {
		// Fetch primary email from emails API
		email, err = p.fetchPrimaryEmail(ctx, client)
		if err != nil {
			return nil, nil, fmt.Errorf("github: failed to fetch email: %w", err)
		}
	}

	firstName := gu.Name
	lastName := ""
	// Split name if possible
	if gu.Name == "" {
		firstName = gu.Login
	}

	return &OAuthUser{
		ProviderID: fmt.Sprintf("%d", gu.ID),
		Email:      email,
		FirstName:  firstName,
		LastName:   lastName,
		AvatarURL:  gu.AvatarURL,
	}, token, nil
}

func (p *GitHubProvider) fetchPrimaryEmail(ctx context.Context, client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("emails API returned %d: %s", resp.StatusCode, body)
	}

	var emails []GitHubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	for _, e := range emails {
		if e.Verified {
			return e.Email, nil
		}
	}

	return "", fmt.Errorf("no verified email found")
}
