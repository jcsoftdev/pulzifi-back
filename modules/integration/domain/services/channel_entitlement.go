package services

import (
	"context"

	"github.com/google/uuid"
)

// ChannelEntitlement is the port that the dispatch_event handler uses to
// determine whether an organisation may use non-email notification channels.
//
// Rules:
//   - IsPaid returns (true, nil)  → paid org; all channels are allowed.
//   - IsPaid returns (false, nil) → free org; only the "email" channel is allowed.
//   - IsPaid returns (_, err)     → entitlement check failed; the handler
//     fails open to email-only (conservative safe-default) and logs the error.
type ChannelEntitlement interface {
	IsPaid(ctx context.Context, orgID uuid.UUID) (bool, error)
}
