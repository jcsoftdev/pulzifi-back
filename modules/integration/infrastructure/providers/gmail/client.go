package gmailprovider

import (
	"bytes"
	"context"
	"encoding/base64"
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
	AuthorizeURL = "https://accounts.google.com/o/oauth2/v2/auth"
	TokenURL     = "https://oauth2.googleapis.com/token"
	UserinfoURL  = "https://www.googleapis.com/oauth2/v2/userinfo"
	SendURL      = "https://gmail.googleapis.com/gmail/v1/users/me/messages/send"
)

const Scopes = "https://www.googleapis.com/auth/gmail.send"

// Client implements services.ProviderClient for Gmail.
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

func (c *Client) ServiceType() string { return "gmail" }

// OAuthAuthorizeURL builds the Google OAuth URL requesting offline gmail.send access.
func (c *Client) OAuthAuthorizeURL(state, redirectURI string) (string, error) {
	v := url.Values{}
	v.Set("client_id", c.clientID)
	v.Set("redirect_uri", redirectURI)
	v.Set("response_type", "code")
	v.Set("scope", Scopes)
	v.Set("access_type", "offline")
	v.Set("prompt", "consent")
	v.Set("state", state)
	return AuthorizeURL + "?" + v.Encode(), nil
}

// HandleCallback exchanges the code for tokens and fetches the authenticated email.
// Returns an error if the token exchange fails or userinfo cannot be retrieved —
// the integration is not stored in either failure case.
func (c *Client) HandleCallback(ctx context.Context, code, redirectURI string) (*entities.OAuthResult, error) {
	body := url.Values{}
	body.Set("client_id", c.clientID)
	body.Set("client_secret", c.clientSecret)
	body.Set("code", code)
	body.Set("redirect_uri", redirectURI)
	body.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, TokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("gmail: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gmail: token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)

	var tok struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Error        string `json:"error"`
	}
	if err := json.Unmarshal(raw, &tok); err != nil {
		return nil, fmt.Errorf("gmail: parse token response: %w", err)
	}
	if tok.Error != "" {
		return nil, fmt.Errorf("gmail oauth: %s", tok.Error)
	}
	if tok.AccessToken == "" {
		return nil, errors.New("gmail: empty access token in response")
	}

	email, err := c.fetchUserEmail(ctx, tok.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("gmail: fetch userinfo: %w", err)
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

// fetchUserEmail calls the Google userinfo endpoint to get the sender address.
func (c *Client) fetchUserEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, UserinfoURL, nil)
	if err != nil {
		return "", fmt.Errorf("build userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("userinfo request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("userinfo: status %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}

	var u struct {
		Email string `json:"email"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(raw, &u); err != nil {
		return "", fmt.Errorf("parse userinfo: %w", err)
	}
	if u.Error != "" {
		return "", fmt.Errorf("userinfo: %s", u.Error)
	}
	if u.Email == "" {
		return "", errors.New("userinfo: empty email field")
	}
	return u.Email, nil
}

// RefreshAccessToken exchanges the refresh token for a new access token.
// Called by delivery_processor.go before Send when the token is near expiry.
func (c *Client) RefreshAccessToken(ctx context.Context, refreshToken string) (*entities.OAuthResult, error) {
	body := url.Values{}
	body.Set("client_id", c.clientID)
	body.Set("client_secret", c.clientSecret)
	body.Set("refresh_token", refreshToken)
	body.Set("grant_type", "refresh_token")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, TokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("gmail: build refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gmail: refresh request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)

	var tok struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Error       string `json:"error"`
	}
	if err := json.Unmarshal(raw, &tok); err != nil {
		return nil, fmt.Errorf("gmail: parse refresh response: %w", err)
	}
	if tok.Error != "" {
		return nil, fmt.Errorf("gmail refresh: %s", tok.Error)
	}
	if tok.AccessToken == "" {
		return nil, errors.New("gmail: empty access token in refresh response")
	}

	var expiresAt *time.Time
	if tok.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)
		expiresAt = &t
	}

	return &entities.OAuthResult{
		AccessToken: tok.AccessToken,
		ExpiresAt:   expiresAt,
	}, nil
}

// ListTargets returns an empty slice — Gmail recipients come from Destination.Target["emails"].
func (c *Client) ListTargets(_ context.Context, _ *entities.Integration) ([]entities.Target, error) {
	return nil, nil
}

// Send delivers a notification email via the Gmail API.
// One API call per recipient. Returns Code 401 if the access token is rejected.
func (c *Client) Send(ctx context.Context, integ *entities.Integration, dest *entities.Destination, p *entities.NotificationPayload) (*entities.DeliveryResult, error) {
	if integ == nil {
		return nil, errors.New("gmail: integration required")
	}

	fromEmail, _ := integ.ProviderMeta["email"].(string)
	if fromEmail == "" {
		return nil, errors.New("gmail: missing sender email in provider_meta")
	}

	raw, _ := dest.Target["emails"].([]any)
	var emails []string
	for _, v := range raw {
		if s, ok := v.(string); ok && strings.Contains(s, "@") {
			emails = append(emails, s)
		}
	}
	if len(emails) == 0 {
		return nil, errors.New("gmail: no recipients in destination target")
	}

	htmlBody := fmt.Sprintf(`<h2>%s</h2><p>%s</p><p><a href="%s">View page</a></p>`,
		p.Title, p.Body, p.PageURL)

	for _, to := range emails {
		if err := c.sendOne(ctx, integ.AccessToken, fromEmail, to, p.Title, htmlBody); err != nil {
			var authErr *errUnauthorized
			if errors.As(err, &authErr) {
				return &entities.DeliveryResult{Code: http.StatusUnauthorized, BodySnip: authErr.Error()},
					fmt.Errorf("gmail send: %w", err)
			}
			return nil, fmt.Errorf("gmail send to %s: %w", to, err)
		}
	}

	return &entities.DeliveryResult{Code: http.StatusOK}, nil
}

// errUnauthorized is returned by sendOne when the API responds with 401.
type errUnauthorized struct{ body string }

func (e *errUnauthorized) Error() string { return "401: " + e.body }

// sendOne sends one RFC 2822 message via the Gmail messages.send endpoint.
func (c *Client) sendOne(ctx context.Context, accessToken, from, to, subject, htmlBody string) error {
	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		from, to, subject, htmlBody,
	)
	encoded := base64.RawURLEncoding.EncodeToString([]byte(msg))
	payload, _ := json.Marshal(map[string]string{"raw": encoded})

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
	defer func() { _ = resp.Body.Close() }()
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
