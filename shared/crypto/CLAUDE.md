# Crypto Package (`shared/crypto/`)

AES-256-GCM authenticated encryption for secrets stored in the database.

## Files

- `aes_gcm.go` — AES-GCM encrypt/decrypt implementation
- `aes_gcm_test.go` — Unit tests

## Exported API

### Types
- `AESGCM` — Cipher wrapper around `cipher.AEAD`

### Functions
- `NewAESGCM(key []byte) (*AESGCM, error)` — Creates cipher with a 32-byte key (AES-256). Returns error if key is not exactly 32 bytes.

### Methods (`*AESGCM`)
- `Encrypt(plain []byte) ([]byte, error)` — Encrypts with random nonce prepended to ciphertext
- `Decrypt(blob []byte) ([]byte, error)` — Decrypts; nonce extracted from blob prefix

## Usage

Used by the integration module persistence layer to encrypt OAuth access tokens and webhook secrets stored in the `integrations` table. Key comes from `INTEGRATION_TOKEN_KEY` (32-byte hex) in config.

## Watch Out

- Key MUST be exactly 32 bytes. Fatal on startup in production if key is missing or wrong length.
- The `INTEGRATION_TOKEN_KEY` defaults to an all-zeros dev key; **never use the default in production**.
