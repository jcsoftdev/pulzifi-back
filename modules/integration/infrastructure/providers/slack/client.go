package slackprovider

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

// Package-level vars so tests can swap to httptest.Server URLs.
var (
	AuthorizeURL = "https://slack.com/oauth/v2/authorize"
	TokenURL     = "https://slack.com/api/oauth.v2.access"
	PostMessage  = "https://slack.com/api/chat.postMessage"
	ListChannels = "https://slack.com/api/conversations.list"
)

const Scopes = "chat:write,channels:read,groups:read"

// maxPaginationPages caps the number of pages fetched by ListTargets.
const maxPaginationPages = 50

// Client implements services.ProviderClient for Slack.
type Client struct {
	clientID     string
	clientSecret string
	http         *http.Client
}

// New returns a Client with sane HTTP timeout defaults.
func New(clientID, clientSecret string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		http:         &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) ServiceType() string { return "slack" }

// OAuthAuthorizeURL builds the Slack OAuth v2 authorization URL.
func (c *Client) OAuthAuthorizeURL(state, redirectURI string) (string, error) {
	v := url.Values{}
	v.Set("client_id", c.clientID)
	v.Set("scope", Scopes)
	v.Set("redirect_uri", redirectURI)
	v.Set("state", state)
	return AuthorizeURL + "?" + v.Encode(), nil
}

// HandleCallback exchanges the OAuth code for an access token via oauth.v2.access.
func (c *Client) HandleCallback(ctx context.Context, code, redirectURI string) (*entities.OAuthResult, error) {
	body := url.Values{}
	body.Set("client_id", c.clientID)
	body.Set("client_secret", c.clientSecret)
	body.Set("code", code)
	body.Set("redirect_uri", redirectURI)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, TokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("slack: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("slack: token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("slack: read token response: %w", err)
	}

	var r struct {
		OK          bool   `json:"ok"`
		Error       string `json:"error"`
		AccessToken string `json:"access_token"`
		Team        struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"team"`
		BotUserID string `json:"bot_user_id"`
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("slack: parse token response: %w", err)
	}
	if !r.OK {
		return nil, fmt.Errorf("slack oauth: %s", r.Error)
	}

	return &entities.OAuthResult{
		AccessToken: r.AccessToken,
		ProviderMeta: map[string]any{
			"team_id":     r.Team.ID,
			"team_name":   r.Team.Name,
			"bot_user_id": r.BotUserID,
		},
	}, nil
}

// RefreshAccessToken is a no-op for Slack (tokens do not expire).
func (c *Client) RefreshAccessToken(_ context.Context, _ string) (*entities.OAuthResult, error) {
	return nil, nil
}

// ListTargets fetches all public and private channels via conversations.list (paginated).
func (c *Client) ListTargets(ctx context.Context, integ *entities.Integration) ([]entities.Target, error) {
	var targets []entities.Target
	cursor := ""

	for page := 0; page < maxPaginationPages; page++ {
		q := url.Values{}
		q.Set("types", "public_channel,private_channel")
		q.Set("exclude_archived", "true")
		q.Set("limit", "200")
		if cursor != "" {
			q.Set("cursor", cursor)
		}

		reqURL := ListChannels + "?" + q.Encode()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			return nil, fmt.Errorf("slack: build channels request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+integ.AccessToken)

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("slack: channels request: %w", err)
		}
		raw, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("slack: read channels response: %w", err)
		}

		var r struct {
			OK       bool `json:"ok"`
			Error    string `json:"error"`
			Channels []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"channels"`
			ResponseMetadata struct {
				NextCursor string `json:"next_cursor"`
			} `json:"response_metadata"`
		}
		if err := json.Unmarshal(raw, &r); err != nil {
			return nil, fmt.Errorf("slack: parse channels response: %w", err)
		}
		if !r.OK {
			return nil, fmt.Errorf("slack channels: %s", r.Error)
		}

		for _, ch := range r.Channels {
			targets = append(targets, entities.Target{
				ID:   ch.ID,
				Name: "#" + ch.Name,
			})
		}

		cursor = r.ResponseMetadata.NextCursor
		if cursor == "" {
			break
		}
	}

	return targets, nil
}

// Send posts a message to a Slack channel via chat.postMessage.
func (c *Client) Send(ctx context.Context, integ *entities.Integration, dest *entities.Destination, p *entities.NotificationPayload) (*entities.DeliveryResult, error) {
	channelID, _ := dest.Target["channel_id"].(string)
	if channelID == "" {
		return nil, errors.New("slack: missing channel_id")
	}

	msgBody, err := json.Marshal(map[string]any{
		"channel": channelID,
		"blocks":  buildBlocks(p),
	})
	if err != nil {
		return nil, fmt.Errorf("slack: marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, PostMessage, bytes.NewReader(msgBody))
	if err != nil {
		return nil, fmt.Errorf("slack: build post request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+integ.AccessToken)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("slack: post message: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("slack: read post response: %w", err)
	}

	var r struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("slack: parse post response: %w", err)
	}

	if !r.OK {
		snip := string(raw)
		if len(snip) > 2048 {
			snip = snip[:2048]
		}
		if r.Error == "invalid_auth" || r.Error == "token_revoked" {
			return &entities.DeliveryResult{Code: http.StatusUnauthorized, BodySnip: snip},
				fmt.Errorf("slack auth invalid: %s", r.Error)
		}
		return &entities.DeliveryResult{Code: resp.StatusCode, BodySnip: snip},
			fmt.Errorf("slack: %s", r.Error)
	}

	return &entities.DeliveryResult{Code: resp.StatusCode}, nil
}
