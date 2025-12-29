package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
)

type AuthService interface {
	Authenticate(ctx context.Context, email, password string) (*entities.User, error)
	ValidateCredentials(ctx context.Context, user *entities.User, password string) error
	HashPassword(password string) (string, error)
	CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error)
}
