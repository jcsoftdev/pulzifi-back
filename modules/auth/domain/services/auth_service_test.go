// Package services_test tests the AuthService interface contract using the
// concrete BcryptAuthService implementation from infrastructure/services.
package services_test

import (
	"testing"

	infraservices "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/services"
)

func buildBcryptAuthService() *infraservices.BcryptAuthService {
	return infraservices.NewBcryptAuthService(&stubUserRepo{}, &stubPermRepo{})
}

func TestBcryptAuthService_HashAndCheck_RoundTrip(t *testing.T) {
	svc := buildBcryptAuthService()
	plaintext := "supersecretpassword"

	hash, err := svc.HashPassword(plaintext)
	if err != nil {
		t.Fatalf("HashPassword: unexpected error: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	if hash == plaintext {
		t.Error("hash must not equal plaintext")
	}

	// ValidateCredentials accepts a user entity; simulate by building minimal user
	// with the hash we just computed.
	user := &stubUserWithHash{hash: hash}
	if err := svc.ValidateCredentials(nil, user.asEntity(), plaintext); err != nil {
		t.Errorf("CheckPassword with correct password: expected nil error, got %v", err)
	}
}

func TestBcryptAuthService_WrongPassword_ReturnsFalse(t *testing.T) {
	svc := buildBcryptAuthService()

	hash, err := svc.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("HashPassword: unexpected error: %v", err)
	}

	user := &stubUserWithHash{hash: hash}
	err = svc.ValidateCredentials(nil, user.asEntity(), "wrong-password")
	if err == nil {
		t.Error("expected error for wrong password, got nil")
	}
}

func TestBcryptAuthService_HashPassword_DifferentHashesSameInput(t *testing.T) {
	svc := buildBcryptAuthService()
	pw := "mypassword"

	hash1, _ := svc.HashPassword(pw)
	hash2, _ := svc.HashPassword(pw)

	// bcrypt always produces different salts, so two hashes of the same input differ.
	if hash1 == hash2 {
		t.Error("expected two hashes of the same password to differ (bcrypt salting)")
	}
}
