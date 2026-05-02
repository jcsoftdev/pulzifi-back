package teamsprovider

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
	domerrors "github.com/jcsoftdev/pulzifi-back/modules/integration/domain/errors"
	services "github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
)

var (
	AuthorizeURL        = "https://login.microsoftonline.com/common/oauth2/v2.0/authorize"
	TokenURL            = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
	GraphBase           = "https://graph.microsoft.com/v1.0"
	AdminConsentURLBase = "https://login.microsoftonline.com/common/adminconsent"
)

const Scopes = "https://graph.microsoft.com/ChannelMessage.Send https://graph.microsoft.com/Team.ReadBasic.All https://graph.microsoft.com/Channel.ReadBasic.All offline_access"

type Client struct {
	clientID, clientSecret string
	http                   *http.Client
}

func New(clientID, clientSecret string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		http:         &http.Client{Timeout: 30 * time.Second},
	}
}

var _ services.ProviderClient = (*Client)(nil)

func (c *Client) ServiceType() string { return "teams" }

func (c *Client) OAuthAuthorizeURL(state, redirect string) (string, error) {
	v := url.Values{}
	v.Set("client_id", c.clientID)
	v.Set("scope", Scopes)
	v.Set("response_type", "code")
	v.Set("redirect_uri", redirect)
	v.Set("state", state)
	v.Set("response_mode", "query")
	v.Set("prompt", "select_account")
	return AuthorizeURL + "?" + v.Encode(), nil
}

func (c *Client) adminConsentURL(redirect string) string {
	v := url.Values{}
	v.Set("client_id", c.clientID)
	v.Set("redirect_uri", redirect)
	return AdminConsentURLBase + "?" + v.Encode()
}

// BuildAdminConsentURL is exported for the HTTP layer's consent-required
// handler (T11). Pure URL synthesis; no HTTP call.
func (c *Client) BuildAdminConsentURL(redirect string) *domerrors.ErrConsentRequired {
	return &domerrors.ErrConsentRequired{AdminConsentURL: c.adminConsentURL(redirect)}
}

func (c *Client) HandleCallback(ctx context.Context, code, redirect string) (*entities.OAuthResult, error) {
	body := url.Values{}
	body.Set("client_id", c.clientID)
	body.Set("client_secret", c.clientSecret)
	body.Set("grant_type", "authorization_code")
	body.Set("code", code)
	body.Set("redirect_uri", redirect)
	body.Set("scope", Scopes)
	return c.tokenExchange(ctx, body, true)
}

func (c *Client) RefreshAccessToken(ctx context.Context, refreshToken string) (*entities.OAuthResult, error) {
	body := url.Values{}
	body.Set("client_id", c.clientID)
	body.Set("client_secret", c.clientSecret)
	body.Set("grant_type", "refresh_token")
	body.Set("refresh_token", refreshToken)
	body.Set("scope", Scopes)
	return c.tokenExchange(ctx, body, true)
}

func (c *Client) tokenExchange(ctx context.Context, body url.Values, requireTokens bool) (*entities.OAuthResult, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", TokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("teams token: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("teams token: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("teams token: status=%d body=%s", resp.StatusCode, string(raw))
	}

	var r struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
		TokenType    string `json:"token_type"`
		Error        string `json:"error"`
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("teams token: parse: %w", err)
	}
	if r.Error != "" {
		return nil, fmt.Errorf("teams token: %s", r.Error)
	}
	if requireTokens {
		if r.AccessToken == "" {
			return nil, errors.New("teams token: empty access_token")
		}
		if r.RefreshToken == "" {
			return nil, errors.New("teams token: empty refresh_token")
		}
	}

	expiresAt := time.Now().UTC().Add(time.Duration(r.ExpiresIn) * time.Second)
	return &entities.OAuthResult{
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		ExpiresAt:    &expiresAt,
		ProviderMeta: map[string]any{},
	}, nil
}

type idName struct {
	ID   string
	Name string
}

func (c *Client) ListTargets(ctx context.Context, integ *entities.Integration) ([]entities.Target, error) {
	if integ == nil || integ.AccessToken == "" {
		return nil, errors.New("teams list_targets: no access token")
	}

	teams, err := c.listJoinedTeams(ctx, integ.AccessToken)
	if err != nil {
		return nil, err
	}

	var all []entities.Target
	for _, team := range teams {
		channels, err := c.listChannels(ctx, integ.AccessToken, team.ID)
		if err != nil {
			return nil, err
		}
		for _, ch := range channels {
			all = append(all, entities.Target{
				ID:   team.ID + "|" + ch.ID,
				Name: fmt.Sprintf("%s / #%s", team.Name, ch.Name),
				Meta: map[string]any{
					"team_id":      team.ID,
					"channel_id":   ch.ID,
					"team_name":    team.Name,
					"channel_name": ch.Name,
				},
			})
		}
	}
	return all, nil
}

func (c *Client) listJoinedTeams(ctx context.Context, token string) ([]idName, error) {
	var all []idName
	nextLink := GraphBase + "/me/joinedTeams"
	for i := 0; i < 20 && nextLink != ""; i++ {
		out, next, err := c.fetchIDName(ctx, token, nextLink)
		if err != nil {
			return nil, fmt.Errorf("teams list_joined_teams: %w", err)
		}
		all = append(all, out...)
		nextLink = next
	}
	return all, nil
}

func (c *Client) listChannels(ctx context.Context, token, teamID string) ([]idName, error) {
	var all []idName
	nextLink := GraphBase + "/teams/" + teamID + "/channels"
	for i := 0; i < 50 && nextLink != ""; i++ {
		out, next, err := c.fetchIDName(ctx, token, nextLink)
		if err != nil {
			return nil, fmt.Errorf("teams list_channels: %w", err)
		}
		all = append(all, out...)
		nextLink = next
	}
	return all, nil
}

func (c *Client) fetchIDName(ctx context.Context, token, u string) ([]idName, string, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("status=%d body=%s", resp.StatusCode, string(raw))
	}
	var r struct {
		Value []struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"value"`
		NextLink string `json:"@odata.nextLink"`
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, "", err
	}
	out := make([]idName, 0, len(r.Value))
	for _, v := range r.Value {
		out = append(out, idName{ID: v.ID, Name: v.DisplayName})
	}
	return out, r.NextLink, nil
}

func (c *Client) Send(ctx context.Context, integ *entities.Integration, dest *entities.Destination, p *entities.NotificationPayload) (*entities.DeliveryResult, error) {
	teamID, channelID := splitTarget(dest.Target)
	if teamID == "" || channelID == "" {
		return nil, errors.New("teams: missing team_id/channel_id (target)")
	}

	body, _ := json.Marshal(map[string]any{
		"subject": p.Title,
		"body": map[string]any{
			"contentType": "html",
			"content":     buildHTMLBody(p),
		},
	})
	u := fmt.Sprintf("%s/teams/%s/channels/%s/messages", GraphBase, teamID, channelID)
	req, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("teams send: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+integ.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("teams send: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return &entities.DeliveryResult{Code: resp.StatusCode}, nil
	}

	snip := string(raw)
	if len(snip) > 2048 {
		snip = snip[:2048]
	}
	if resp.StatusCode == 401 {
		return &entities.DeliveryResult{Code: 401, BodySnip: snip},
			fmt.Errorf("teams send auth: status=401")
	}
	return &entities.DeliveryResult{Code: resp.StatusCode, BodySnip: snip},
		fmt.Errorf("teams send: status=%d", resp.StatusCode)
}

// splitTarget accepts either composite "channel_target": "team_id|channel_id"
// or split form ("team_id" + "channel_id" keys).
func splitTarget(target map[string]any) (teamID, channelID string) {
	if v, _ := target["team_id"].(string); v != "" {
		teamID = v
	}
	if v, _ := target["channel_id"].(string); v != "" {
		channelID = v
	}
	if teamID != "" && channelID != "" {
		return
	}
	if v, ok := target["channel_target"].(string); ok {
		parts := strings.SplitN(v, "|", 2)
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	}
	return
}
