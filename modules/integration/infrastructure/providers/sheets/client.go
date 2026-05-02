package sheetsprovider

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
	AuthorizeBase = "https://accounts.google.com/o/oauth2/v2/auth"
	TokenURL      = "https://oauth2.googleapis.com/token"
	DriveListURL  = "https://www.googleapis.com/drive/v3/files"
	SheetsBase    = "https://sheets.googleapis.com/v4"
)

const Scopes = "https://www.googleapis.com/auth/spreadsheets https://www.googleapis.com/auth/drive.readonly"

// Client implements services.ProviderClient for Google Sheets.
type Client struct {
	clientID, clientSecret string
	http                   *http.Client
}

// New returns a Client with sane HTTP timeout defaults.
func New(clientID, clientSecret string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		http:         &http.Client{Timeout: 30 * time.Second},
	}
}

var _ services.ProviderClient = (*Client)(nil)

func (c *Client) ServiceType() string { return "sheets" }

// OAuthAuthorizeURL builds the Google OAuth2 authorization URL with offline access.
func (c *Client) OAuthAuthorizeURL(state, redirect string) (string, error) {
	v := url.Values{}
	v.Set("client_id", c.clientID)
	v.Set("scope", Scopes)
	v.Set("response_type", "code")
	v.Set("redirect_uri", redirect)
	v.Set("state", state)
	v.Set("access_type", "offline")
	v.Set("prompt", "consent")
	v.Set("include_granted_scopes", "true")
	return AuthorizeBase + "?" + v.Encode(), nil
}

// HandleCallback exchanges the OAuth code for tokens; requires refresh_token in response.
func (c *Client) HandleCallback(ctx context.Context, code, redirect string) (*entities.OAuthResult, error) {
	body := url.Values{}
	body.Set("client_id", c.clientID)
	body.Set("client_secret", c.clientSecret)
	body.Set("grant_type", "authorization_code")
	body.Set("code", code)
	body.Set("redirect_uri", redirect)
	return c.tokenExchange(ctx, body, true /* require refresh_token */)
}

// RefreshAccessToken uses a refresh token to obtain a new access token.
func (c *Client) RefreshAccessToken(ctx context.Context, refreshToken string) (*entities.OAuthResult, error) {
	body := url.Values{}
	body.Set("client_id", c.clientID)
	body.Set("client_secret", c.clientSecret)
	body.Set("grant_type", "refresh_token")
	body.Set("refresh_token", refreshToken)
	return c.tokenExchange(ctx, body, false /* refresh response may omit refresh_token */)
}

func (c *Client) tokenExchange(ctx context.Context, body url.Values, requireRefresh bool) (*entities.OAuthResult, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", TokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("sheets token: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sheets token: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("sheets token: status=%d body=%s", resp.StatusCode, string(raw))
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
		return nil, fmt.Errorf("sheets token: parse: %w", err)
	}
	if r.Error != "" {
		return nil, fmt.Errorf("sheets token: %s", r.Error)
	}
	if r.AccessToken == "" {
		return nil, errors.New("sheets token: empty access_token")
	}
	if requireRefresh && r.RefreshToken == "" {
		return nil, errors.New("sheets token: no refresh_token in response")
	}

	expiresAt := time.Now().UTC().Add(time.Duration(r.ExpiresIn) * time.Second)
	return &entities.OAuthResult{
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken, // empty preserved on refresh path
		ExpiresAt:    &expiresAt,
		ProviderMeta: map[string]any{},
	}, nil
}

// ListTargets fetches all Google Sheets spreadsheets via Drive API (paginated).
func (c *Client) ListTargets(ctx context.Context, integ *entities.Integration) ([]entities.Target, error) {
	if integ == nil || integ.AccessToken == "" {
		return nil, errors.New("sheets list_targets: no access token")
	}
	var (
		all       []entities.Target
		pageToken string
	)
	for i := 0; i < 50; i++ { // cap at 50 pages × 200 = 10K
		u := fmt.Sprintf("%s?q=%s&fields=files(id,name),nextPageToken&pageSize=200",
			DriveListURL,
			url.QueryEscape("mimeType='application/vnd.google-apps.spreadsheet'"),
		)
		if pageToken != "" {
			u += "&pageToken=" + url.QueryEscape(pageToken)
		}

		req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
		req.Header.Set("Authorization", "Bearer "+integ.AccessToken)

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("sheets list_targets: %w", err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			return nil, fmt.Errorf("sheets list_targets: auth failure: status=%d", resp.StatusCode)
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("sheets list_targets: status=%d body=%s", resp.StatusCode, string(body))
		}

		var r struct {
			Files         []struct{ ID, Name string } `json:"files"`
			NextPageToken string                       `json:"nextPageToken"`
		}
		if err := json.Unmarshal(body, &r); err != nil {
			return nil, fmt.Errorf("sheets list_targets: parse: %w", err)
		}
		for _, f := range r.Files {
			all = append(all, entities.Target{ID: f.ID, Name: f.Name})
		}
		if r.NextPageToken == "" {
			break
		}
		pageToken = r.NextPageToken
	}
	return all, nil
}

// Send appends a notification row to a Google Sheets spreadsheet via the Sheets API v4.
func (c *Client) Send(ctx context.Context, integ *entities.Integration, dest *entities.Destination, p *entities.NotificationPayload) (*entities.DeliveryResult, error) {
	spreadsheetID, _ := dest.Target["spreadsheet_id"].(string)
	tabName, _ := dest.Target["tab_name"].(string)
	if spreadsheetID == "" {
		return nil, errors.New("sheets: missing spreadsheet_id")
	}
	if tabName == "" {
		return nil, errors.New("sheets: missing tab_name")
	}

	rangeRef := tabName + "!A:E"
	u := fmt.Sprintf("%s/spreadsheets/%s/values/%s:append?valueInputOption=USER_ENTERED&insertDataOption=INSERT_ROWS",
		SheetsBase, spreadsheetID, url.PathEscape(rangeRef))

	row := []any{
		time.Now().UTC().Format(time.RFC3339),
		p.EventType,
		p.PageURL,
		p.Title,
		p.Severity,
	}
	body, _ := json.Marshal(map[string]any{"values": [][]any{row}})

	req, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("sheets send: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+integ.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sheets send: %w", err)
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

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return &entities.DeliveryResult{Code: http.StatusUnauthorized, BodySnip: snip},
			fmt.Errorf("sheets send auth: status=%d", resp.StatusCode)
	}
	return &entities.DeliveryResult{Code: resp.StatusCode, BodySnip: snip},
		fmt.Errorf("sheets send: status=%d", resp.StatusCode)
}
