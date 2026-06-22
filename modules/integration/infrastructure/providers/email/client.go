package emailprovider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	services "github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
)

var _ services.ProviderClient = (*Client)(nil)

// EmailSender is a narrow interface for sending emails.
// The wiring layer (Task 25) injects an adapter that wraps modules/email's Resend sender.
// No direct import of modules/email here — keeps cross-module dependency rule clean.
type EmailSender interface {
	Send(ctx context.Context, to []string, subject, htmlBody string) error
}

// Client implements services.ProviderClient for email delivery.
// OAuth methods are no-ops because email needs no OAuth flow.
type Client struct{ sender EmailSender }

// New returns a Client backed by the given EmailSender.
func New(sender EmailSender) *Client { return &Client{sender: sender} }

func (c *Client) ServiceType() string { return "email" }

func (c *Client) OAuthAuthorizeURL(string, string) (string, error) {
	return "", errors.New("email: no oauth")
}

func (c *Client) HandleCallback(context.Context, string, string) (*entities.OAuthResult, error) {
	return nil, errors.New("email: no oauth")
}

func (c *Client) RefreshAccessToken(context.Context, string) (*entities.OAuthResult, error) {
	return nil, nil
}

func (c *Client) ListTargets(context.Context, *entities.Integration) ([]entities.Target, error) {
	return nil, nil
}

func (c *Client) Send(ctx context.Context, _ *entities.Integration, dest *entities.Destination, p *entities.NotificationPayload) (*entities.DeliveryResult, error) {
	raw, _ := dest.Target["emails"].([]any)
	// Collect recipients, trimming whitespace and removing case-insensitive
	// duplicates so an address that appears both as a workspace member (seeded
	// default) and as a manually-added recipient is emailed only once.
	seen := make(map[string]struct{})
	var emails []string
	for _, v := range raw {
		s, ok := v.(string)
		if !ok {
			continue
		}
		s = strings.TrimSpace(s)
		if !strings.Contains(s, "@") {
			continue
		}
		key := strings.ToLower(s)
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		emails = append(emails, s)
	}
	if len(emails) == 0 {
		return nil, errors.New("email: no recipients")
	}
	body := fmt.Sprintf(`<h2>%s</h2><p>%s</p><p><a href="%s">View page</a></p>`, p.Title, p.Body, p.PageURL)
	if err := c.sender.Send(ctx, emails, p.Title, body); err != nil {
		return nil, err
	}
	return &entities.DeliveryResult{Code: 200}, nil
}
