package twilioprovider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// APIBase is the Twilio REST API base URL. Mutable for tests.
// Both Validator and the Client (Task 11) use this.
var APIBase = "https://api.twilio.com"

type Validator struct {
	http *http.Client
}

func NewValidator() *Validator {
	return &Validator{http: &http.Client{Timeout: 10 * time.Second}}
}

// Validate hits GET /Accounts/{SID}.json with HTTP Basic to verify credentials.
// Returns nil if creds are valid and account is active.
// Returns descriptive error otherwise.
func (v *Validator) Validate(ctx context.Context, provider string, creds map[string]string) error {
	if provider != "twilio" {
		return fmt.Errorf("twilio validator: unsupported provider %q", provider)
	}
	sid := creds["account_sid"]
	token := creds["auth_token"]
	if sid == "" {
		return errors.New("twilio: missing account_sid")
	}
	if token == "" {
		return errors.New("twilio: missing auth_token")
	}
	if creds["from_number"] == "" {
		return errors.New("twilio: missing from_number")
	}

	url := fmt.Sprintf("%s/2010-04-01/Accounts/%s.json", APIBase, sid)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("twilio validate: %w", err)
	}
	req.SetBasicAuth(sid, token)

	resp, err := v.http.Do(req)
	if err != nil {
		return fmt.Errorf("twilio validate: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK {
		var r struct {
			Status string `json:"status"`
		}
		_ = json.Unmarshal(body, &r)
		if r.Status != "" && r.Status != "active" {
			return fmt.Errorf("twilio: account status=%s", r.Status)
		}
		return nil
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return errors.New("twilio: invalid credentials")
	}
	return fmt.Errorf("twilio: status=%d body=%s", resp.StatusCode, truncate(string(body), 256))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
