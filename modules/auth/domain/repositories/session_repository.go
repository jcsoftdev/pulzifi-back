package repositories

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
)

type SessionRepository interface {
	Create(ctx context.Context, session *entities.Session) error
	FindByID(ctx context.Context, id string) (*entities.Session, error)
	DeleteByID(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) error
}
