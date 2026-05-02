package services

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
)

type ProviderClient interface {
	ServiceType() string
	OAuthAuthorizeURL(state, redirectURI string) (string, error)
	HandleCallback(ctx context.Context, code, redirectURI string) (*entities.OAuthResult, error)

	// RefreshAccessToken exchanges a refresh_token for a new access_token.
	//
	// CONTRACT:
	//   - On success: return (*OAuthResult, nil) with non-nil AccessToken + ExpiresAt.
	//     RefreshToken may be empty (provider didn't rotate) or new (provider rotated).
	//   - On failure (revoked, expired, network error): return (nil, error).
	//   - DO NOT return (nil, nil). For providers whose tokens never expire (Slack
	//     bot tokens) or have no OAuth (Email, Twilio), simply leave
	//     Integration.TokenExpiresAt nil so the worker never invokes the refresh
	//     path. This contract is enforced defensively by the worker.
	RefreshAccessToken(ctx context.Context, refreshToken string) (*entities.OAuthResult, error)

	ListTargets(ctx context.Context, integ *entities.Integration) ([]entities.Target, error)
	Send(ctx context.Context, integ *entities.Integration, dest *entities.Destination, payload *entities.NotificationPayload) (*entities.DeliveryResult, error)
}

type ProviderRegistry interface {
	Get(serviceType string) (ProviderClient, bool)
}
