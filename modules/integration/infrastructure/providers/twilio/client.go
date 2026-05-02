package twilioprovider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	services "github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
)

type tier string

const (
	TierFree       tier = "free"
	TierPaid       tier = "paid"
	TierEnterprise tier = "enterprise"
)

// OrgPlanLookup is the interface Twilio uses to resolve tier.
// Implementation lives in cmd/wiring/integration/ (composition root).
type OrgPlanLookup interface {
	PlanCode(ctx context.Context, orgID uuid.UUID) (string, error)
}

// QuotaTracker enforces per-org monthly send limits. Implementation lives in
// cmd/wiring/integration/ and wraps shared/integrationusage.Tracker with a
// plan-driven AllowedFor closure.
type QuotaTracker interface {
	CheckAndIncrement(ctx context.Context, orgID uuid.UUID, serviceType string) error
}

// ErrQuotaExceeded is the user-facing sentinel surfaced when the org's
// monthly Twilio quota is exhausted. The wiring adapter wraps
// integrationusage.ErrQuotaExceeded; this provider-local sentinel lets
// callers switch on it without importing the shared package.
var ErrQuotaExceeded = errors.New("twilio: monthly SMS quota exceeded")

type Config struct {
	PaidPlans          []string // plan codes that grant access to platform Twilio
	PlatformAccountSID string
	PlatformAuthToken  string
	PlatformFromNumber string
}

type Client struct {
	cfg    Config
	plans  OrgPlanLookup
	quotas QuotaTracker // nil-safe (skipped when nil — for tests / pre-T12)
	http   *http.Client
}

func New(cfg Config, plans OrgPlanLookup, quotas QuotaTracker) *Client {
	return &Client{
		cfg:    cfg,
		plans:  plans,
		quotas: quotas,
		http:   &http.Client{Timeout: 15 * time.Second},
	}
}

var _ services.ProviderClient = (*Client)(nil)

func (c *Client) ServiceType() string { return "twilio" }

// OAuthAuthorizeURL — no-op. Twilio uses HTTP Basic.
func (c *Client) OAuthAuthorizeURL(string, string) (string, error) {
	return "", errors.New("twilio: no oauth")
}

func (c *Client) HandleCallback(context.Context, string, string) (*entities.OAuthResult, error) {
	return nil, errors.New("twilio: no oauth")
}

func (c *Client) RefreshAccessToken(context.Context, string) (*entities.OAuthResult, error) {
	return nil, nil
}

func (c *Client) ListTargets(context.Context, *entities.Integration) ([]entities.Target, error) {
	return nil, nil
}

// credsFromIntegration encapsulates the BYO field-mapping convention (spec D4):
// AccessToken=AuthToken, RefreshToken=AccountSID, ProviderMeta.from_number.
// NEVER read integ.AccessToken/RefreshToken directly elsewhere in this package —
// always go through this helper to avoid swap bugs.
func credsFromIntegration(integ *entities.Integration) (sid, token, from string) {
	if integ == nil {
		return "", "", ""
	}
	sid = integ.RefreshToken
	token = integ.AccessToken
	if integ.ProviderMeta != nil {
		from, _ = integ.ProviderMeta["from_number"].(string)
	}
	return
}

// tierFor resolves which credential set to use for this org+integ combo.
// Enterprise = BYO row exists. Paid = no BYO + plan in PaidPlans. Free = neither.
// Tier lookup requires org-scoped destinations (workspace/page scope deferred to Phase 3).
func (c *Client) tierFor(ctx context.Context, dest *entities.Destination, integ *entities.Integration) (tier, error) {
	if integ != nil {
		return TierEnterprise, nil
	}
	if dest.ScopeType != entities.ScopeOrg {
		return "", errors.New("twilio: tier lookup requires org-scoped destination")
	}
	code, err := c.plans.PlanCode(ctx, dest.ScopeID)
	if err != nil {
		return "", err
	}
	for _, p := range c.cfg.PaidPlans {
		if p == code {
			return TierPaid, nil
		}
	}
	return TierFree, nil
}

func (c *Client) Send(ctx context.Context, integ *entities.Integration, dest *entities.Destination, p *entities.NotificationPayload) (*entities.DeliveryResult, error) {
	t, err := c.tierFor(ctx, dest, integ)
	if err != nil {
		return nil, fmt.Errorf("twilio: tier lookup: %w", err)
	}

	var sid, token, from string
	switch t {
	case TierFree:
		return nil, errors.New("twilio: SMS not available on free plan")
	case TierPaid:
		if c.quotas != nil {
			if err := c.quotas.CheckAndIncrement(ctx, dest.ScopeID, "twilio"); err != nil {
				// Map shared package's ErrQuotaExceeded (string-match avoids importing it here)
				if strings.Contains(err.Error(), "quota exceeded") {
					return nil, ErrQuotaExceeded
				}
				return nil, fmt.Errorf("twilio quota: %w", err)
			}
		}
		sid, token, from = c.cfg.PlatformAccountSID, c.cfg.PlatformAuthToken, c.cfg.PlatformFromNumber
		if sid == "" || token == "" || from == "" {
			return nil, errors.New("twilio: platform credentials not configured")
		}
	case TierEnterprise:
		sid, token, from = credsFromIntegration(integ)
		if sid == "" || token == "" || from == "" {
			return nil, errors.New("twilio: enterprise integration missing required fields")
		}
	}

	rawNums, _ := dest.Target["phone_numbers"].([]any)
	if len(rawNums) == 0 {
		return nil, errors.New("twilio: no phone numbers")
	}
	body := truncate(p.Title+": "+p.Body, 1500)

	var last *entities.DeliveryResult
	for _, rn := range rawNums {
		n, _ := rn.(string)
		if n == "" {
			continue
		}
		r, err := c.sendOne(ctx, sid, token, from, n, body)
		if err != nil {
			return r, err
		}
		last = r
	}
	return last, nil
}

func (c *Client) sendOne(ctx context.Context, sid, token, from, to, body string) (*entities.DeliveryResult, error) {
	form := url.Values{}
	form.Set("From", from)
	form.Set("To", to)
	form.Set("Body", body)

	apiURL := fmt.Sprintf("%s/2010-04-01/Accounts/%s/Messages.json", APIBase, sid)
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("twilio send: %w", err)
	}
	req.SetBasicAuth(sid, token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("twilio send: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return &entities.DeliveryResult{Code: resp.StatusCode}, nil
	}

	var errResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(raw, &errResp)

	snip := truncate(string(raw), 2048)
	if isAuthError(errResp.Code) {
		return &entities.DeliveryResult{Code: http.StatusUnauthorized, BodySnip: snip},
			fmt.Errorf("twilio auth: code=%d %s", errResp.Code, errResp.Message)
	}
	return &entities.DeliveryResult{Code: resp.StatusCode, BodySnip: snip},
		fmt.Errorf("twilio: code=%d %s", errResp.Code, errResp.Message)
}

// isAuthError returns true for Twilio error codes that indicate credential issues.
// 20003 = Authenticate (invalid creds). 20429 = Too Many Requests (rate-limited; treat as auth-tier dead since retry will keep failing).
func isAuthError(code int) bool {
	return code == 20003 || code == 20429
}
