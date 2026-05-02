package crypto

import (
	"encoding/hex"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key, _ := hex.DecodeString("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
	enc, _ := NewAESGCM(key)
	ct, err := enc.Encrypt([]byte("xoxb-secret-token"))
	if err != nil {
		t.Fatal(err)
	}
	pt, err := enc.Decrypt(ct)
	if err != nil {
		t.Fatal(err)
	}
	if string(pt) != "xoxb-secret-token" {
		t.Fatalf("got %q", pt)
	}
}

func TestRejectsWrongKeyLength(t *testing.T) {
	if _, err := NewAESGCM([]byte("short")); err == nil {
		t.Fatal("expected error")
	}
}

func TestNonceUnique(t *testing.T) {
	key, _ := hex.DecodeString("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
	enc, _ := NewAESGCM(key)
	a, _ := enc.Encrypt([]byte("same"))
	b, _ := enc.Encrypt([]byte("same"))
	if string(a) == string(b) {
		t.Fatal("nonce reuse")
	}
}
