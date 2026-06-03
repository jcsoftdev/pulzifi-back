package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	"golang.org/x/crypto/bcrypt"
)

type BcryptAuthService struct {
	userRepo repositories.UserRepository
	permRepo repositories.PermissionRepository
}

func NewBcryptAuthService(
	userRepo repositories.UserRepository,
	permRepo repositories.PermissionRepository,
) *BcryptAuthService {
	return &BcryptAuthService{
		userRepo: userRepo,
		permRepo: permRepo,
	}
}

func (s *BcryptAuthService) Authenticate(ctx context.Context, email, password string) (*entities.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	if err := s.ValidateCredentials(ctx, user, password); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *BcryptAuthService) ValidateCredentials(ctx context.Context, user *entities.User, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
}

func (s *BcryptAuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *BcryptAuthService) CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	return s.permRepo.HasPermission(ctx, userID, resource, action)
}
