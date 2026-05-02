package wiring

import (
	"context"

	emailservices "github.com/jcsoftdev/pulzifi-back/modules/email/domain/services"
	emailprovider "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/providers/email"
)

type emailAdapter struct{ provider emailservices.EmailProvider }

func NewEmailAdapter(p emailservices.EmailProvider) emailprovider.EmailSender {
	return &emailAdapter{provider: p}
}

func (a *emailAdapter) Send(ctx context.Context, to []string, subject, htmlBody string) error {
	for _, addr := range to {
		if err := a.provider.Send(ctx, addr, subject, htmlBody); err != nil {
			return err
		}
	}
	return nil
}
