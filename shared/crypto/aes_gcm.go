package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

type AESGCM struct{ aead cipher.AEAD }

func NewAESGCM(key []byte) (*AESGCM, error) {
	if len(key) != 32 {
		return nil, errors.New("crypto: key must be 32 bytes (AES-256)")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &AESGCM{aead: aead}, nil
}

func (a *AESGCM) Encrypt(plain []byte) ([]byte, error) {
	nonce := make([]byte, a.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return a.aead.Seal(nonce, nonce, plain, nil), nil
}

func (a *AESGCM) Decrypt(blob []byte) ([]byte, error) {
	if len(blob) < a.aead.NonceSize() {
		return nil, errors.New("crypto: ciphertext too short")
	}
	n := a.aead.NonceSize()
	return a.aead.Open(nil, blob[:n], blob[n:], nil)
}
