package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services/mocks"
)

// stubAuthService is a local fake AuthService that does NOT import infrastructure.
// It implements password hashing/checking with a trivial scheme for tests only.
type stubAuthService struct {
	userRepo *stubUserRepo
	permRepo *stubPermRepo
}

func newStubAuthService(userRepo *stubUserRepo, permRepo *stubPermRepo) *stubAuthService {
	return &stubAuthService{userRepo: userRepo, permRepo: permRepo}
}

func (s *stubAuthService) Authenticate(ctx context.Context, email, password string) (*entities.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, errors.New("invalid credentials")
	}
	if err := s.ValidateCredentials(ctx, user, password); err != nil {
		return nil, err
	}
	switch user.Status {
	case entities.UserStatusPending:
		return nil, errors.New("user not approved")
	case entities.UserStatusRejected:
		return nil, errors.New("user rejected")
	}
	return user, nil
}

func (s *stubAuthService) ValidateCredentials(_ context.Context, user *entities.User, password string) error {
	// Simple stub: stored hash == "hash:" + plain password
	if user.PasswordHash != "hash:"+password {
		return errors.New("invalid credentials")
	}
	return nil
}

func (s *stubAuthService) HashPassword(password string) (string, error) {
	return "hash:" + password, nil
}

func (s *stubAuthService) CheckPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	return s.permRepo.HasPermission(ctx, userID, resource, action)
}

// compile-time check: stubAuthService satisfies the AuthService interface
var _ services.AuthService = (*stubAuthService)(nil)

// compile-time check: MockAuthService (from mocks pkg) satisfies the AuthService interface
var _ services.AuthService = (*mocks.MockAuthService)(nil)

func TestAuthService_Authenticate_Success(t *testing.T) {
	userRepo := newStubUserRepo()
	permRepo := newStubPermRepo()
	svc := newStubAuthService(userRepo, permRepo)

	id := uuid.New()
	_ = userRepo.Create(context.Background(), &entities.User{
		ID:           id,
		Email:        "test@example.com",
		PasswordHash: "hash:secret",
		Status:       entities.UserStatusApproved,
	})

	user, err := svc.Authenticate(context.Background(), "test@example.com", "secret")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID != id {
		t.Errorf("expected user ID %v, got %v", id, user.ID)
	}
}

func TestAuthService_Authenticate_WrongPassword(t *testing.T) {
	userRepo := newStubUserRepo()
	permRepo := newStubPermRepo()
	svc := newStubAuthService(userRepo, permRepo)

	_ = userRepo.Create(context.Background(), &entities.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hash:secret",
		Status:       entities.UserStatusApproved,
	})

	_, err := svc.Authenticate(context.Background(), "test@example.com", "wrong")
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
}

func TestAuthService_Authenticate_PendingUser(t *testing.T) {
	userRepo := newStubUserRepo()
	permRepo := newStubPermRepo()
	svc := newStubAuthService(userRepo, permRepo)

	_ = userRepo.Create(context.Background(), &entities.User{
		ID:           uuid.New(),
		Email:        "pending@example.com",
		PasswordHash: "hash:secret",
		Status:       entities.UserStatusPending,
	})

	_, err := svc.Authenticate(context.Background(), "pending@example.com", "secret")
	if err == nil {
		t.Fatal("expected error for pending user, got nil")
	}
}

func TestAuthService_HashPassword(t *testing.T) {
	svc := newStubAuthService(newStubUserRepo(), newStubPermRepo())

	hash, err := svc.HashPassword("mypassword")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hash == "" {
		t.Error("expected non-empty hash")
	}
}

func TestAuthService_ValidateCredentials(t *testing.T) {
	svc := newStubAuthService(newStubUserRepo(), newStubPermRepo())

	user := &entities.User{PasswordHash: "hash:correct"}

	if err := svc.ValidateCredentials(context.Background(), user, "correct"); err != nil {
		t.Errorf("expected valid credentials, got %v", err)
	}

	if err := svc.ValidateCredentials(context.Background(), user, "wrong"); err == nil {
		t.Error("expected error for wrong credentials, got nil")
	}
}

func TestMockAuthService_SatisfiesInterface(t *testing.T) {
	mock := &mocks.MockAuthService{
		AuthenticateResult: &entities.User{ID: uuid.New(), Email: "mock@example.com"},
	}

	user, err := mock.Authenticate(context.Background(), "mock@example.com", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected non-nil user")
	}
	if mock.AuthenticateCalls != 1 {
		t.Errorf("expected 1 call, got %d", mock.AuthenticateCalls)
	}
}
