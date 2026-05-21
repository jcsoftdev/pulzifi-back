// Package services_test validates the TokenService interface contract using the
// concrete JWTService implementation from infrastructure/services.
package services_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	infraservices "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/services"
)

// ── Stub repos ────────────────────────────────────────────────────────────────

// Compile-time interface checks.
var _ repositories.RoleRepository = (*stubRoleRepo)(nil)
var _ repositories.PermissionRepository = (*stubPermRepo)(nil)

type stubRoleRepo struct{}

func (r *stubRoleRepo) GetByID(_ context.Context, _ uuid.UUID) (*entities.Role, error) {
	return nil, nil
}
func (r *stubRoleRepo) GetByName(_ context.Context, _ string) (*entities.Role, error) {
	return nil, nil
}
func (r *stubRoleRepo) GetUserRoles(_ context.Context, _ uuid.UUID) ([]*entities.Role, error) {
	return nil, nil
}
func (r *stubRoleRepo) GetRolePermissions(_ context.Context, _ uuid.UUID) ([]*entities.Permission, error) {
	return nil, nil
}
func (r *stubRoleRepo) AssignRoleToUser(_ context.Context, _, _ uuid.UUID) error { return nil }

type stubPermRepo struct{}

func (p *stubPermRepo) GetByID(_ context.Context, _ uuid.UUID) (*entities.Permission, error) {
	return nil, nil
}
func (p *stubPermRepo) GetByName(_ context.Context, _ string) (*entities.Permission, error) {
	return nil, nil
}
func (p *stubPermRepo) GetUserPermissions(_ context.Context, _ uuid.UUID) ([]*entities.Permission, error) {
	return nil, nil
}
func (p *stubPermRepo) HasPermission(_ context.Context, _ uuid.UUID, _, _ string) (bool, error) {
	return false, nil
}

// buildJWTService creates a JWTService with in-memory stub repos and the given secret/expiry.
func buildJWTService(t *testing.T, secret string, accessExpiry time.Duration) *infraservices.JWTService {
	t.Helper()
	return infraservices.NewJWTService(secret, accessExpiry, 7*24*time.Hour, &stubRoleRepo{}, &stubPermRepo{})
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestJWTService_GenerateAndValidateAccessToken_RoundTrip(t *testing.T) {
	svc := buildJWTService(t, "test-secret-key-long-enough", 15*time.Minute)
	userID := uuid.New()
	email := "roundtrip@example.com"

	token, err := svc.GenerateAccessToken(context.Background(), userID, email)
	if err != nil {
		t.Fatalf("GenerateAccessToken: unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token string")
	}

	claims, err := svc.ValidateToken(context.Background(), token)
	if err != nil {
		t.Fatalf("ValidateToken: unexpected error: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("user ID: want %v, got %v", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("email: want %q, got %q", email, claims.Email)
	}
}

func TestJWTService_ExpiredToken_Rejected(t *testing.T) {
	// Negative expiry means the token is already expired at creation time.
	svc := buildJWTService(t, "test-secret-key-long-enough", -1*time.Second)
	userID := uuid.New()

	token, err := svc.GenerateAccessToken(context.Background(), userID, "expired@example.com")
	if err != nil {
		t.Fatalf("GenerateAccessToken: unexpected error: %v", err)
	}

	_, err = svc.ValidateToken(context.Background(), token)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestJWTService_TamperedSignature_Rejected(t *testing.T) {
	svc := buildJWTService(t, "test-secret-key-long-enough", 15*time.Minute)

	// Construct a JWT signed with a DIFFERENT key — this guarantees a valid-format
	// token with an invalid HMAC, regardless of base64 encoding edge cases.
	svcDifferentKey := buildJWTService(t, "completely-different-secret-key", 15*time.Minute)
	userID := uuid.New()

	// Sign the token with the original key.
	_, err := svc.GenerateAccessToken(context.Background(), userID, "tamper@example.com")
	if err != nil {
		t.Fatalf("GenerateAccessToken (original): unexpected error: %v", err)
	}

	// Generate a token with the wrong key — this will have an incorrect signature
	// from the perspective of svc's key.
	wrongKeyToken, err := svcDifferentKey.GenerateAccessToken(context.Background(), userID, "tamper@example.com")
	if err != nil {
		t.Fatalf("GenerateAccessToken (wrong key): unexpected error: %v", err)
	}

	// Validate with the ORIGINAL service — should fail because the signature was made with a different key.
	_, err = svc.ValidateToken(context.Background(), wrongKeyToken)
	if err == nil {
		t.Fatal("expected error when validating a token signed with a different key, got nil")
	}

	// Also test with a literally mangled token: replace the entire signature segment.
	goodToken, _ := svc.GenerateAccessToken(context.Background(), userID, "mangled@example.com")
	parts := strings.Split(goodToken, ".")
	mangledToken := parts[0] + "." + parts[1] + ".AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	_, err = svc.ValidateToken(context.Background(), mangledToken)
	if err == nil {
		t.Fatal("expected error for mangled signature segment, got nil")
	}
}

func TestJWTService_GetTokenExpiration_ReturnsConfigured(t *testing.T) {
	expiry := 30 * time.Minute
	svc := buildJWTService(t, "test-secret-key-long-enough", expiry)
	if got := svc.GetTokenExpiration(); got != expiry {
		t.Errorf("GetTokenExpiration: want %v, got %v", expiry, got)
	}
}
