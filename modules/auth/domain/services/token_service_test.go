package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services/mocks"
)

// stubTokenService is a local fake TokenService that does NOT import infrastructure.
// It generates trivially inspectable tokens for unit tests only.
type stubTokenService struct {
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	roleRepo      *stubRoleRepo
}

func newStubTokenService(roleRepo *stubRoleRepo) *stubTokenService {
	return &stubTokenService{
		accessExpiry:  15 * time.Minute,
		refreshExpiry: 7 * 24 * time.Hour,
		roleRepo:      roleRepo,
	}
}

func (s *stubTokenService) GenerateAccessToken(_ context.Context, userID uuid.UUID, email string) (string, error) {
	if userID == uuid.Nil {
		return "", errors.New("invalid user ID")
	}
	// Trivial: "access:<userID>:<email>"
	return "access:" + userID.String() + ":" + email, nil
}

func (s *stubTokenService) GenerateRefreshToken(_ context.Context, userID uuid.UUID) (string, error) {
	if userID == uuid.Nil {
		return "", errors.New("invalid user ID")
	}
	return "refresh:" + userID.String(), nil
}

func (s *stubTokenService) ValidateToken(_ context.Context, token string) (*services.TokenClaims, error) {
	if token == "" {
		return nil, errors.New("empty token")
	}
	// Return a fixed claims set for the stub
	return &services.TokenClaims{
		UserID: uuid.New(),
		Email:  "stub@example.com",
		Roles:  []string{"USER"},
	}, nil
}

func (s *stubTokenService) GetTokenExpiration() time.Duration {
	return s.accessExpiry
}

func (s *stubTokenService) GetRefreshTokenExpiration() time.Duration {
	return s.refreshExpiry
}

// compile-time check: stubTokenService satisfies the TokenService interface
var _ services.TokenService = (*stubTokenService)(nil)

// compile-time check: MockTokenService (from mocks pkg) satisfies the TokenService interface
var _ services.TokenService = (*mocks.MockTokenService)(nil)

func TestTokenService_GenerateAccessToken(t *testing.T) {
	svc := newStubTokenService(newStubRoleRepo())

	id := uuid.New()
	token, err := svc.GenerateAccessToken(context.Background(), id, "user@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty access token")
	}
}

func TestTokenService_GenerateRefreshToken(t *testing.T) {
	svc := newStubTokenService(newStubRoleRepo())

	id := uuid.New()
	token, err := svc.GenerateRefreshToken(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty refresh token")
	}
}

func TestTokenService_GenerateAccessToken_NilID(t *testing.T) {
	svc := newStubTokenService(newStubRoleRepo())

	_, err := svc.GenerateAccessToken(context.Background(), uuid.Nil, "user@example.com")
	if err == nil {
		t.Fatal("expected error for nil user ID, got nil")
	}
}

func TestTokenService_ValidateToken(t *testing.T) {
	svc := newStubTokenService(newStubRoleRepo())

	claims, err := svc.ValidateToken(context.Background(), "some-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims == nil {
		t.Fatal("expected non-nil claims")
	}
}

func TestTokenService_ValidateToken_Empty(t *testing.T) {
	svc := newStubTokenService(newStubRoleRepo())

	_, err := svc.ValidateToken(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty token, got nil")
	}
}

func TestTokenService_Expirations(t *testing.T) {
	svc := newStubTokenService(newStubRoleRepo())

	if svc.GetTokenExpiration() != 15*time.Minute {
		t.Errorf("expected 15m, got %v", svc.GetTokenExpiration())
	}
	if svc.GetRefreshTokenExpiration() != 7*24*time.Hour {
		t.Errorf("expected 7d, got %v", svc.GetRefreshTokenExpiration())
	}
}

func TestMockTokenService_SatisfiesInterface(t *testing.T) {
	id := uuid.New()
	mock := &mocks.MockTokenService{
		GenerateAccessTokenResult: "mock-access-token",
	}

	token, err := mock.GenerateAccessToken(context.Background(), id, "user@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "mock-access-token" {
		t.Errorf("expected 'mock-access-token', got %q", token)
	}
	if mock.GenerateAccessTokenCalls != 1 {
		t.Errorf("expected 1 call, got %d", mock.GenerateAccessTokenCalls)
	}
}
