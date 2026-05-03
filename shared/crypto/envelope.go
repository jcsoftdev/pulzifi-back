package crypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"
)

var ErrUnknownMEKVersion = errors.New("crypto: unknown MEK version in ciphertext")

// Envelope encrypts/decrypts opaque bytes with per-org DEKs that are
// themselves wrapped (encrypted) with a versioned Master Encryption Key.
//
// Ciphertext layout: [1 byte mek_version][12 bytes nonce][gcm ciphertext+tag]
//
// The version byte makes Decrypt able to pick the right MEK to unwrap the DEK
// regardless of when the row was written. RotateMEK changes the wrapped_dek
// in storage; existing ciphertext continues to decrypt because the version
// byte references the old MEK (which must remain in the meks map until the
// rotation is verified).
type Envelope struct {
	db        *sql.DB
	meks      map[int][]byte
	activeMEK int

	mu    sync.Mutex
	cache map[uuid.UUID][]byte // unwrapped DEK cache; cleared on RotateMEK
}

// NewEnvelope validates inputs and returns a new Envelope.
// Each MEK must be exactly 32 bytes. activeMEK must exist in meks.
func NewEnvelope(db *sql.DB, meks map[int][]byte, activeMEK int) (*Envelope, error) {
	if db == nil {
		return nil, errors.New("crypto: db is nil")
	}
	if len(meks) == 0 {
		return nil, errors.New("crypto: meks map is empty")
	}
	for v, k := range meks {
		if len(k) != 32 {
			return nil, fmt.Errorf("crypto: MEK v%d must be 32 bytes, got %d", v, len(k))
		}
	}
	if _, ok := meks[activeMEK]; !ok {
		return nil, fmt.Errorf("crypto: activeMEK v%d not in meks map", activeMEK)
	}
	return &Envelope{
		db:        db,
		meks:      meks,
		activeMEK: activeMEK,
		cache:     make(map[uuid.UUID][]byte),
	}, nil
}

// Encrypt encrypts plaintext for the given org. Lazily creates a DEK if
// the org doesn't have one yet (concurrent-safe via INSERT ON CONFLICT).
func (e *Envelope) Encrypt(ctx context.Context, orgID uuid.UUID, plaintext []byte) ([]byte, error) {
	dek, err := e.getOrCreateDEK(ctx, orgID)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, fmt.Errorf("crypto: aes cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: gcm: %w", err)
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("crypto: nonce: %w", err)
	}
	sealed := aead.Seal(nil, nonce, plaintext, nil)

	out := make([]byte, 0, 1+aead.NonceSize()+len(sealed))
	out = append(out, byte(e.activeMEK))
	out = append(out, nonce...)
	out = append(out, sealed...)
	return out, nil
}

// Decrypt parses the version byte, unwraps the DEK with the matching MEK,
// and decrypts the ciphertext with the DEK.
func (e *Envelope) Decrypt(ctx context.Context, orgID uuid.UUID, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < 1+12+16 {
		return nil, errors.New("crypto: ciphertext too short")
	}
	version := int(ciphertext[0])
	_, ok := e.meks[version]
	if !ok {
		return nil, fmt.Errorf("%w: v%d", ErrUnknownMEKVersion, version)
	}
	nonce := ciphertext[1:13]
	sealed := ciphertext[13:]

	dek, err := e.unwrapDEK(ctx, orgID, version)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, fmt.Errorf("crypto: aes cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: gcm: %w", err)
	}
	return aead.Open(nil, nonce, sealed, nil)
}

// RotateMEK re-wraps every DEK with mek_version=fromVersion using toVersion's MEK,
// updating the row in place. Idempotent — safe to re-run if interrupted.
// Caches are cleared so the next access reads fresh wrapped_dek bytes.
func (e *Envelope) RotateMEK(ctx context.Context, fromVersion, toVersion int) error {
	fromMEK, ok := e.meks[fromVersion]
	if !ok {
		return fmt.Errorf("crypto: MEK v%d not in map", fromVersion)
	}
	toMEK, ok := e.meks[toVersion]
	if !ok {
		return fmt.Errorf("crypto: MEK v%d not in map", toVersion)
	}

	rows, err := e.db.QueryContext(ctx,
		`SELECT org_id, wrapped_dek FROM org_data_keys WHERE mek_version = $1`, fromVersion)
	if err != nil {
		return fmt.Errorf("crypto: rotate: query: %w", err)
	}
	defer rows.Close()

	type item struct {
		orgID      uuid.UUID
		wrappedDEK []byte
	}
	var batch []item
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.orgID, &it.wrappedDEK); err != nil {
			return fmt.Errorf("crypto: rotate: scan: %w", err)
		}
		batch = append(batch, it)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, it := range batch {
		plain, err := unwrap(fromMEK, it.orgID, it.wrappedDEK)
		if err != nil {
			return fmt.Errorf("crypto: rotate: unwrap org %s: %w", it.orgID, err)
		}
		rewrap, err := wrap(toMEK, it.orgID, plain)
		if err != nil {
			return fmt.Errorf("crypto: rotate: wrap org %s: %w", it.orgID, err)
		}
		_, err = e.db.ExecContext(ctx, `
            UPDATE org_data_keys
               SET wrapped_dek = $2, mek_version = $3, updated_at = NOW()
             WHERE org_id = $1
        `, it.orgID, rewrap, toVersion)
		if err != nil {
			return fmt.Errorf("crypto: rotate: update org %s: %w", it.orgID, err)
		}
	}

	e.mu.Lock()
	e.cache = make(map[uuid.UUID][]byte)
	e.mu.Unlock()
	return nil
}

// getOrCreateDEK returns the unwrapped DEK for the org. Creates a new one
// (random 32 bytes) wrapped with the active MEK if no row exists.
func (e *Envelope) getOrCreateDEK(ctx context.Context, orgID uuid.UUID) ([]byte, error) {
	e.mu.Lock()
	if dek, ok := e.cache[orgID]; ok {
		e.mu.Unlock()
		return dek, nil
	}
	e.mu.Unlock()

	// Try INSERT first; if conflict, SELECT existing.
	activeMEK := e.meks[e.activeMEK]
	fresh := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, fresh); err != nil {
		return nil, fmt.Errorf("crypto: random dek: %w", err)
	}
	wrappedFresh, err := wrap(activeMEK, orgID, fresh)
	if err != nil {
		return nil, fmt.Errorf("crypto: wrap fresh: %w", err)
	}

	var (
		wrappedFromDB []byte
		version       int
	)
	err = e.db.QueryRowContext(ctx, `
        INSERT INTO org_data_keys (org_id, wrapped_dek, mek_version, created_at, updated_at)
        VALUES ($1, $2, $3, NOW(), NOW())
        ON CONFLICT (org_id) DO UPDATE SET org_id = org_data_keys.org_id
        RETURNING wrapped_dek, mek_version
    `, orgID, wrappedFresh, e.activeMEK).Scan(&wrappedFromDB, &version)
	if err != nil {
		return nil, fmt.Errorf("crypto: upsert dek: %w", err)
	}

	var dek []byte
	if version == e.activeMEK {
		// Either we just inserted, or a concurrent insert won and used the same
		// active version. Either way, wrappedFromDB unwraps with activeMEK.
		dek, err = unwrap(activeMEK, orgID, wrappedFromDB)
	} else {
		// Existing row from earlier MEK version.
		mek, ok := e.meks[version]
		if !ok {
			return nil, fmt.Errorf("crypto: stored MEK v%d not in map", version)
		}
		dek, err = unwrap(mek, orgID, wrappedFromDB)
	}
	if err != nil {
		return nil, err
	}

	e.mu.Lock()
	e.cache[orgID] = dek
	e.mu.Unlock()
	return dek, nil
}

func (e *Envelope) unwrapDEK(ctx context.Context, orgID uuid.UUID, _ int) ([]byte, error) {
	e.mu.Lock()
	if dek, ok := e.cache[orgID]; ok {
		e.mu.Unlock()
		return dek, nil
	}
	e.mu.Unlock()

	var wrapped []byte
	var rowVersion int
	err := e.db.QueryRowContext(ctx,
		`SELECT wrapped_dek, mek_version FROM org_data_keys WHERE org_id = $1`, orgID,
	).Scan(&wrapped, &rowVersion)
	if err != nil {
		return nil, fmt.Errorf("crypto: load dek: %w", err)
	}

	// Use the row's mek_version (which may differ from the ciphertext byte —
	// e.g. after RotateMEK, ciphertext byte references old version but row was
	// re-wrapped). Pick the right MEK from our map.
	actualMEK, ok := e.meks[rowVersion]
	if !ok {
		return nil, fmt.Errorf("%w: row v%d", ErrUnknownMEKVersion, rowVersion)
	}

	dek, err := unwrap(actualMEK, orgID, wrapped)
	if err != nil {
		return nil, err
	}

	e.mu.Lock()
	e.cache[orgID] = dek
	e.mu.Unlock()
	return dek, nil
}

// wrap encrypts plain with mek using a deterministic nonce derived from orgID.
// Deterministic nonce keeps wrapped_dek stable across re-wraps for the same
// MEK version (helps idempotency).
func wrap(mek []byte, orgID uuid.UUID, plain []byte) ([]byte, error) {
	block, err := aes.NewCipher(mek)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := deriveNonce(orgID, mek)
	return aead.Seal(nil, nonce, plain, nil), nil
}

func unwrap(mek []byte, orgID uuid.UUID, wrapped []byte) ([]byte, error) {
	block, err := aes.NewCipher(mek)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := deriveNonce(orgID, mek)
	return aead.Open(nil, nonce, wrapped, nil)
}

// deriveNonce produces a deterministic 12-byte nonce from orgID + a piece of
// the MEK. Deterministic-per-(org, mek_version) is acceptable here because we
// only use this for wrapping the DEK — never reused for token plaintext.
func deriveNonce(orgID uuid.UUID, mek []byte) []byte {
	h := sha256.New()
	h.Write(orgID[:])
	h.Write(mek[:8]) // tag the nonce with a slice of MEK so different MEKs => different nonces
	sum := h.Sum(nil)
	return sum[:12]
}

// Unused but kept for future tooling: timestamp helper for migration paths.
var _ = time.Now
