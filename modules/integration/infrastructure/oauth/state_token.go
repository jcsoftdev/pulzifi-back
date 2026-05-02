package oauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// StateClaims holds the data encoded in an OAuth state token.
type StateClaims struct {
	Provider    string    `json:"p"`
	Tenant      string    `json:"t"`
	OrgID       uuid.UUID `json:"o"`
	UserID      uuid.UUID `json:"u"`
	ReturnPath  string    `json:"r"`
	RedirectURI string    `json:"d"`
	Nonce       string    `json:"n"`
	ExpiresAt   int64     `json:"e"`
}

// StateSigner creates and verifies HMAC-SHA256-signed, stateless OAuth state tokens.
// Token format: base64url(json{claims, exp}) + "." + base64url(hmac_sha256(key, header))
type StateSigner struct {
	key []byte
	ttl time.Duration
}

// NewStateSigner returns a StateSigner using the provided HMAC key and TTL.
func NewStateSigner(key []byte, ttl time.Duration) *StateSigner {
	return &StateSigner{key: key, ttl: ttl}
}

// Sign encodes the claims into a signed state token. If Nonce is empty, a UUID is generated.
func (s *StateSigner) Sign(c StateClaims) (string, error) {
	if c.Nonce == "" {
		c.Nonce = uuid.NewString()
	}
	c.ExpiresAt = time.Now().Add(s.ttl).Unix()
	body, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	head := base64.RawURLEncoding.EncodeToString(body)
	mac := hmac.New(sha256.New, s.key)
	mac.Write([]byte(head))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return head + "." + sig, nil
}

// Verify parses and validates the signed state token, returning the embedded claims.
func (s *StateSigner) Verify(token string) (*StateClaims, error) {
	dot := -1
	for i, b := range token {
		if b == '.' {
			dot = i
			break
		}
	}
	if dot <= 0 {
		return nil, errors.New("malformed state")
	}
	head, sig := token[:dot], token[dot+1:]
	if sig == "" {
		return nil, errors.New("malformed state")
	}
	expected := hmac.New(sha256.New, s.key)
	expected.Write([]byte(head))
	expSig := base64.RawURLEncoding.EncodeToString(expected.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expSig)) {
		return nil, errors.New("bad signature")
	}
	body, err := base64.RawURLEncoding.DecodeString(head)
	if err != nil {
		return nil, err
	}
	var c StateClaims
	if err := json.Unmarshal(body, &c); err != nil {
		return nil, err
	}
	if time.Now().Unix() > c.ExpiresAt {
		return nil, errors.New("state expired")
	}
	return &c, nil
}
