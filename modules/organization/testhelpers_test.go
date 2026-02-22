package main_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// getEnvOrDefault returns the environment variable value or a default.
func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// getRESTBaseURL returns the REST base URL from env or default.
func getRESTBaseURL() string {
	return getEnvOrDefault("TEST_REST_BASE_URL", "http://localhost:8082")
}

// getGRPCAddress returns the gRPC address from env or default.
func getGRPCAddress() string {
	return getEnvOrDefault("TEST_GRPC_ADDRESS", "localhost:9082")
}

// generateTestJWT creates a real JWT signed with the test secret.
func generateTestJWT(userID string) string {
	secret := getEnvOrDefault("JWT_SECRET", "secret")

	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(fmt.Sprintf("failed to sign test JWT: %v", err))
	}
	return signed
}

// skipUnlessE2E skips the test unless RUN_E2E_TESTS=true.
func skipUnlessE2E(t *testing.T) {
	t.Helper()
	if os.Getenv("RUN_E2E_TESTS") != "true" {
		t.Skip("Skipping e2e test (set RUN_E2E_TESTS=true to run)")
	}
}
