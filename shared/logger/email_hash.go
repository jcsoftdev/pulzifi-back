package logger

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

var emailHashKey = []byte("dev-only-hash-key")

// InitEmailHashKey sets the HMAC key used by HashEmail. Empty key keeps the
// dev default. Call once at startup from cmd/server/main.go.
func InitEmailHashKey(key string) {
	if key != "" {
		emailHashKey = []byte(key)
	}
}

// HashEmail returns a 16-char HMAC-SHA256 hex digest of the lowercased email,
// safe to log without revealing the address. Reversible only with the key.
func HashEmail(email string) string {
	mac := hmac.New(sha256.New, emailHashKey)
	mac.Write([]byte(strings.ToLower(email)))
	return hex.EncodeToString(mac.Sum(nil))[:16]
}
