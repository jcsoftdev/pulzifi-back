package providers

import (
	"context"
	"fmt"

	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/resend/resend-go/v2"
	"go.uber.org/zap"
)

// ResendProvider implements EmailProvider using Resend.
type ResendProvider struct {
	client      *resend.Client
	fromAddress string
	fromName    string
}

// NewResendProvider creates a new Resend email provider.
func NewResendProvider(apiKey, fromAddress, fromName string) *ResendProvider {
	return &ResendProvider{
		client:      resend.NewClient(apiKey),
		fromAddress: fromAddress,
		fromName:    fromName,
	}
}

// Send sends an email via Resend.
func (p *ResendProvider) Send(ctx context.Context, to, subject, htmlBody string) error {
	from := fmt.Sprintf("%s <%s>", p.fromName, p.fromAddress)

	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
	}

	_, err := p.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		logger.Error("Failed to send email via Resend",
			zap.String("to", to),
			zap.String("subject", subject),
			zap.Error(err),
		)
		return fmt.Errorf("resend: %w", err)
	}

	logger.Info("Email sent via Resend",
		zap.String("to", to),
		zap.String("subject", subject),
	)
	return nil
}
