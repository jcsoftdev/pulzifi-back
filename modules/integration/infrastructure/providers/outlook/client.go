package outlookprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	services "github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
)

var _ services.ProviderClient = (*Client)(nil)

// Package-level URL vars so tests can swap to httptest.Server URLs.
var (
	// /organizations endpoint = work/school accounts only; avoids PKCE requirement for personal accounts.
	AuthorizeURL = "https://login.microsoftonline.com/organizations/oauth2/v2.0/authorize"
	TokenURL     = "https://login.microsoftonline.com/organizations/oauth2/v2.0/token"
	UserinfoURL  = "https://graph.microsoft.com/v1.0/me"
	SendURL      = "https://graph.microsoft.com/v1.0/me/sendMail"
)

const Scopes = "https://graph.microsoft.com/Mail.Send offline_access"

// Client implements services.ProviderClient for Outlook / Microsoft 365.
type Client struct {
	clientID     string
	clientSecret string
	redirectBase string
	http         *http.Client
}

// New returns a Client with sane HTTP timeout defaults.
func New(clientID, clientSecret, redirectBase string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectBase: redirectBase,
		http:         &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) ServiceType() string { return "outlook" }

// OAuthAuthorizeURL builds the Microsoft OAuth authorization URL.
func (c *Client) OAuthAuthorizeURL(state, redirectURI string) (string, error) {
	v := url.Values{}
	v.Set("client_id", c.clientID)
	v.Set("redirect_uri", redirectURI)
	v.Set("response_type", "code")
	v.Set("scope", Scopes)
	v.Set("state", state)
	return AuthorizeURL + "?" + v.Encode(), nil
}

// HandleCallback exchanges the code for tokens and fetches the authenticated email via /me.
// Returns an error if the token exchange fails or /me cannot determine the email —
// the integration is not stored in either failure case.
func (c *Client) HandleCallback(ctx context.Context, code, redirectURI string) (*entities.OAuthResult, error) {
	body := url.Values{}
	body.Set("client_id", c.clientID)
	body.Set("client_secret", c.clientSecret)
	body.Set("code", code)
	body.Set("redirect_uri", redirectURI)
	body.Set("grant_type", "authorization_code")
	body.Set("scope", Scopes)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, TokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("outlook: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("outlook: token request: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("outlook: token endpoint: status %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}

	var tok struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}
	if err := json.Unmarshal(raw, &tok); err != nil {
		return nil, fmt.Errorf("outlook: parse token response: %w", err)
	}
	if tok.Error != "" {
		return nil, fmt.Errorf("outlook oauth: %s: %s", tok.Error, tok.ErrorDesc)
	}
	if tok.AccessToken == "" {
		return nil, errors.New("outlook: empty access token in response")
	}

	email, err := c.fetchUserEmail(ctx, tok.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("outlook: fetch userinfo: %w", err)
	}

	var expiresAt *time.Time
	if tok.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)
		expiresAt = &t
	}

	return &entities.OAuthResult{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		ExpiresAt:    expiresAt,
		ProviderMeta: map[string]any{"email": email},
	}, nil
}

// fetchUserEmail resolves the mailbox address via /me.
// Prefers the mail field; falls back to userPrincipalName; errors if both are absent.
func (c *Client) fetchUserEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, UserinfoURL, nil)
	if err != nil {
		return "", fmt.Errorf("build me request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("me request: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("me: status %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}

	var me struct {
		Mail              string `json:"mail"`
		UserPrincipalName string `json:"userPrincipalName"`
	}
	if err := json.Unmarshal(raw, &me); err != nil {
		return "", fmt.Errorf("parse me response: %w", err)
	}
	if me.Mail != "" {
		return me.Mail, nil
	}
	if me.UserPrincipalName != "" {
		return me.UserPrincipalName, nil
	}
	return "", errors.New("outlook: could not determine email from /me response (mail and userPrincipalName both empty)")
}

// RefreshAccessToken exchanges the refresh token for a new access token.
// Microsoft issues a new refresh token on each refresh (sliding 90-day window).
// The caller must check result.RefreshToken and persist it if non-empty.
func (c *Client) RefreshAccessToken(ctx context.Context, refreshToken string) (*entities.OAuthResult, error) {
	body := url.Values{}
	body.Set("client_id", c.clientID)
	body.Set("client_secret", c.clientSecret)
	body.Set("refresh_token", refreshToken)
	body.Set("grant_type", "refresh_token")
	body.Set("scope", Scopes)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, TokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("outlook: build refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("outlook: refresh request: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("outlook: token endpoint: status %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}

	var tok struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}
	if err := json.Unmarshal(raw, &tok); err != nil {
		return nil, fmt.Errorf("outlook: parse refresh response: %w", err)
	}
	if tok.Error != "" {
		return nil, fmt.Errorf("outlook refresh: %s: %s", tok.Error, tok.ErrorDesc)
	}
	if tok.AccessToken == "" {
		return nil, errors.New("outlook: empty access token in refresh response")
	}

	var expiresAt *time.Time
	if tok.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)
		expiresAt = &t
	}

	return &entities.OAuthResult{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// ListTargets returns an empty slice — Outlook recipients come from Destination.Target["emails"].
func (c *Client) ListTargets(_ context.Context, _ *entities.Integration) ([]entities.Target, error) {
	return nil, nil
}

// Send delivers a notification email via the Microsoft Graph sendMail API.
// One API call per recipient. Returns Code 401 if the access token is rejected.
func (c *Client) Send(ctx context.Context, integ *entities.Integration, dest *entities.Destination, p *entities.NotificationPayload) (*entities.DeliveryResult, error) {
	if integ == nil {
		return nil, errors.New("outlook: integration required")
	}

	fromEmail, _ := integ.ProviderMeta["email"].(string)
	if fromEmail == "" {
		return nil, errors.New("outlook: missing sender email in provider_meta")
	}

	raw, _ := dest.Target["emails"].([]any)
	var emails []string
	for _, v := range raw {
		if s, ok := v.(string); ok && strings.Contains(s, "@") {
			emails = append(emails, s)
		}
	}
	if len(emails) == 0 {
		return nil, errors.New("outlook: no recipients in destination target")
	}

	htmlBody := fmt.Sprintf(`<h2>%s</h2><p>%s</p><p><a href="%s">View page</a></p>`,
		p.Title, p.Body, p.PageURL)

	for _, to := range emails {
		if err := c.sendOne(ctx, integ.AccessToken, fromEmail, to, p.Title, htmlBody); err != nil {
			var authErr *errUnauthorized
			if errors.As(err, &authErr) {
				return &entities.DeliveryResult{Code: http.StatusUnauthorized, BodySnip: authErr.Error()},
					fmt.Errorf("outlook send: %w", err)
			}
			return nil, fmt.Errorf("outlook send to %s: %w", to, err)
		}
	}

	return &entities.DeliveryResult{Code: http.StatusOK}, nil
}

// errUnauthorized is returned by sendOne when the Graph API responds with 401.
type errUnauthorized struct{ body string }

func (e *errUnauthorized) Error() string { return "401: " + e.body }

// sendOne sends one message via the Graph sendMail endpoint.
func (c *Client) sendOne(ctx context.Context, accessToken, from, to, subject, htmlBody string) error {
	msg := map[string]any{
		"message": map[string]any{
			"subject": subject,
			"body": map[string]any{
				"contentType": "HTML",
				"content":     htmlBody,
			},
			"toRecipients": []map[string]any{
				{"emailAddress": map[string]any{"address": to}},
			},
			"from": map[string]any{
				"emailAddress": map[string]any{"address": from},
			},
		},
		"saveToSentItems": false,
	}

	payload, _ := json.Marshal(msg)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, SendURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized {
		return &errUnauthorized{body: truncate(string(raw), 200)}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("status %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
