package providers

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// NoopProvider is a no-op email provider for development/testing.
type NoopProvider struct{}

// NewNoopProvider creates a new no-op email provider.
func NewNoopProvider() *NoopProvider {
	return &NoopProvider{}
}

// Send logs the email instead of sending it.
func (p *NoopProvider) Send(ctx context.Context, to, subject, htmlBody string) error {
	logger.Info("NoopProvider: email would be sent",
		zap.String("to", to),
		zap.String("subject", subject),
	)
	return nil
}
