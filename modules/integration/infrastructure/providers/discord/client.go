package discordprovider

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

var (
	AuthorizeURL = "https://discord.com/api/oauth2/authorize"
	TokenURL     = "https://discord.com/api/oauth2/token"
)

const Scopes = "webhook.incoming"

type Client struct {
	clientID, clientSecret string
	http                   *http.Client
}

func New(clientID, clientSecret string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		http:         &http.Client{Timeout: 15 * time.Second},
	}
}

var _ services.ProviderClient = (*Client)(nil)

func (c *Client) ServiceType() string { return "discord" }

func (c *Client) OAuthAuthorizeURL(state, redirect string) (string, error) {
	v := url.Values{}
	v.Set("client_id", c.clientID)
	v.Set("scope", Scopes)
	v.Set("response_type", "code")
	v.Set("redirect_uri", redirect)
	v.Set("state", state)
	return AuthorizeURL + "?" + v.Encode(), nil
}

func (c *Client) HandleCallback(ctx context.Context, code, redirect string) (*entities.OAuthResult, error) {
	body := url.Values{}
	body.Set("client_id", c.clientID)
	body.Set("client_secret", c.clientSecret)
	body.Set("grant_type", "authorization_code")
	body.Set("code", code)
	body.Set("redirect_uri", redirect)
	req, err := http.NewRequestWithContext(ctx, "POST", TokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("discord callback: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("discord callback: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)

	var r struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		Webhook     struct {
			ID        string `json:"id"`
			URL       string `json:"url"`
			ChannelID string `json:"channel_id"`
			Name      string `json:"name"`
			GuildID   string `json:"guild_id"`
		} `json:"webhook"`
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("discord callback: parse: %w", err)
	}
	if r.Error != "" {
		return nil, fmt.Errorf("discord oauth: %s", r.Error)
	}
	if r.Webhook.URL == "" {
		return nil, errors.New("discord oauth: no webhook in response")
	}
	return &entities.OAuthResult{
		AccessToken: r.AccessToken,
		ProviderMeta: map[string]any{
			"webhook_url":  r.Webhook.URL,
			"webhook_id":   r.Webhook.ID,
			"channel_id":   r.Webhook.ChannelID,
			"channel_name": r.Webhook.Name,
			"guild_id":     r.Webhook.GuildID,
		},
	}, nil
}

func (c *Client) RefreshAccessToken(context.Context, string) (*entities.OAuthResult, error) {
	return nil, nil
}

func (c *Client) ListTargets(ctx context.Context, integ *entities.Integration) ([]entities.Target, error) {
	if integ == nil || integ.ProviderMeta == nil {
		return nil, nil
	}
	webhookURL, _ := integ.ProviderMeta["webhook_url"].(string)
	if webhookURL == "" {
		return nil, nil
	}
	name, _ := integ.ProviderMeta["channel_name"].(string)
	if name == "" {
		name = "webhook"
	}
	return []entities.Target{{
		ID:   webhookURL,
		Name: "#" + name,
		Meta: map[string]any{"webhook_url": webhookURL},
	}}, nil
}

func (c *Client) Send(ctx context.Context, integ *entities.Integration, dest *entities.Destination, p *entities.NotificationPayload) (*entities.DeliveryResult, error) {
	webhookURL, _ := dest.Target["webhook_url"].(string)
	if webhookURL == "" && integ != nil && integ.ProviderMeta != nil {
		webhookURL, _ = integ.ProviderMeta["webhook_url"].(string)
	}
	if webhookURL == "" {
		return nil, errors.New("discord: no webhook url")
	}

	body, _ := json.Marshal(map[string]any{
		"embeds": []map[string]any{{
			"title":       p.Title,
			"description": p.Body,
			"url":         p.PageURL,
			"color":       severityColor(p.Severity),
		}},
	})
	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("discord send: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("discord send: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return &entities.DeliveryResult{Code: resp.StatusCode}, nil
	}

	snip := redactWebhookURL(string(raw), webhookURL)
	if len(snip) > 2048 {
		snip = snip[:2048]
	}
	return &entities.DeliveryResult{Code: resp.StatusCode, BodySnip: snip},
		fmt.Errorf("discord send: status=%d", resp.StatusCode)
}

func severityColor(sev string) int {
	switch sev {
	case "critical":
		return 0xE74C3C
	case "warning":
		return 0xF39C12
	default:
		return 0x3498DB
	}
}

// redactWebhookURL removes any occurrence of the webhook URL from a string
// (e.g. an error response body) before logging or returning to upstream callers.
// Discord webhook URLs are credentials and must not leak via logs or BodySnip.
func redactWebhookURL(s, webhookURL string) string {
	if webhookURL == "" {
		return s
	}
	return strings.ReplaceAll(s, webhookURL, "***WEBHOOK_URL_REDACTED***")
}
