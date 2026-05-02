package services

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
)

type ProviderClient interface {
	ServiceType() string
	OAuthAuthorizeURL(state, redirectURI string) (string, error)
	HandleCallback(ctx context.Context, code, redirectURI string) (*entities.OAuthResult, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*entities.OAuthResult, error)
	ListTargets(ctx context.Context, integ *entities.Integration) ([]entities.Target, error)
	Send(ctx context.Context, integ *entities.Integration, dest *entities.Destination, payload *entities.NotificationPayload) (*entities.DeliveryResult, error)
}

type ProviderRegistry interface {
	Get(serviceType string) (ProviderClient, bool)
}
