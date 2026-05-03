package crypto

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Minimal in-process SQL driver for tests
// ---------------------------------------------------------------------------

func init() {
	sql.Register("fakedb", &fakeDriver{})
}

// fakeDriver is a simple in-memory driver whose behaviour is set per-connection
// via a *fakeConnState shared by reference.
type fakeDriver struct{}

var registry = struct {
	mu     sync.Mutex
	states map[string]*fakeConnState
}{states: make(map[string]*fakeConnState)}

func (d *fakeDriver) Open(name string) (driver.Conn, error) {
	registry.mu.Lock()
	state, ok := registry.states[name]
	registry.mu.Unlock()
	if !ok {
		return nil, errors.New("fakedb: unknown dsn " + name)
	}
	return &fakeConn{state: state}, nil
}

func registerFakeState(dsn string, state *fakeConnState) {
	registry.mu.Lock()
	registry.states[dsn] = state
	registry.mu.Unlock()
}

// fakeConnState holds the sequence of scripted query/exec results.
type fakeConnState struct {
	mu      sync.Mutex
	results []fakeResult
}

type fakeResult struct {
	// For QueryRow / Query: rows to return.
	rows [][]driver.Value
	cols []string
	// For Exec: rows affected.
	rowsAffected int64
	// error to return
	err error
}

func (s *fakeConnState) push(r fakeResult) {
	s.mu.Lock()
	s.results = append(s.results, r)
	s.mu.Unlock()
}

func (s *fakeConnState) pop() (fakeResult, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.results) == 0 {
		return fakeResult{}, false
	}
	r := s.results[0]
	s.results = s.results[1:]
	return r, true
}

// fakeConn implements driver.Conn, driver.QueryerContext, driver.ExecerContext.
type fakeConn struct{ state *fakeConnState }

func (c *fakeConn) Prepare(query string) (driver.Stmt, error) { return &fakeStmt{conn: c}, nil }
func (c *fakeConn) Close() error                              { return nil }
func (c *fakeConn) Begin() (driver.Tx, error)                 { return &fakeTx{}, nil }

type fakeTx struct{}

func (t *fakeTx) Commit() error   { return nil }
func (t *fakeTx) Rollback() error { return nil }

type fakeStmt struct{ conn *fakeConn }

func (s *fakeStmt) Close() error                                    { return nil }
func (s *fakeStmt) NumInput() int                                   { return -1 }
func (s *fakeStmt) Exec(args []driver.Value) (driver.Result, error) { return s.conn.execResult() }
func (s *fakeStmt) Query(args []driver.Value) (driver.Rows, error)  { return s.conn.queryRows() }

func (c *fakeConn) execResult() (driver.Result, error) {
	r, ok := c.state.pop()
	if !ok {
		return nil, errors.New("fakedb: no scripted exec result")
	}
	if r.err != nil {
		return nil, r.err
	}
	return driver.RowsAffected(r.rowsAffected), nil
}

func (c *fakeConn) queryRows() (driver.Rows, error) {
	r, ok := c.state.pop()
	if !ok {
		return nil, errors.New("fakedb: no scripted query result")
	}
	if r.err != nil {
		return nil, r.err
	}
	return &fakeRows{cols: r.cols, data: r.rows}, nil
}

type fakeRows struct {
	cols    []string
	data    [][]driver.Value
	current int
}

func (r *fakeRows) Columns() []string { return r.cols }
func (r *fakeRows) Close() error      { return nil }
func (r *fakeRows) Next(dest []driver.Value) error {
	if r.current >= len(r.data) {
		return io.EOF
	}
	copy(dest, r.data[r.current])
	r.current++
	return nil
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

var dsnSeq int
var dsnMu sync.Mutex

func newFakeDB(t *testing.T) (*sql.DB, *fakeConnState) {
	t.Helper()
	dsnMu.Lock()
	dsnSeq++
	dsn := fmt.Sprintf("test%d", dsnSeq)
	dsnMu.Unlock()

	state := &fakeConnState{}
	registerFakeState(dsn, state)
	db, err := sql.Open("fakedb", dsn)
	if err != nil {
		t.Fatalf("open fakedb: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, state
}

func testMEK(b byte) []byte {
	key := make([]byte, 32)
	for i := range key {
		key[i] = b
	}
	return key
}

func newTestEnvelope(t *testing.T, db *sql.DB, meks map[int][]byte, active int) *Envelope {
	t.Helper()
	e, err := NewEnvelope(db, meks, active)
	if err != nil {
		t.Fatalf("NewEnvelope: %v", err)
	}
	return e
}

// wrappedDEKForTest produces a valid wrapped DEK using the same wrap() function
// so mock rows contain bytes the Envelope can actually unwrap.
func wrappedDEKForTest(t *testing.T, mek []byte, orgID uuid.UUID) []byte {
	t.Helper()
	dek := make([]byte, 32)
	for i := range dek {
		dek[i] = 0xDE
	}
	wrapped, err := wrap(mek, orgID, dek)
	if err != nil {
		t.Fatalf("wrappedDEKForTest: %v", err)
	}
	return wrapped
}

// ---------------------------------------------------------------------------
// NewEnvelope validation
// ---------------------------------------------------------------------------

func TestNewEnvelope_NilDB(t *testing.T) {
	_, err := NewEnvelope(nil, map[int][]byte{1: testMEK(0xAA)}, 1)
	if err == nil {
		t.Fatal("expected error for nil db")
	}
}

func TestNewEnvelope_EmptyMEKs(t *testing.T) {
	db, _ := newFakeDB(t)
	_, err := NewEnvelope(db, map[int][]byte{}, 1)
	if err == nil {
		t.Fatal("expected error for empty meks")
	}
}

func TestNewEnvelope_ShortMEK(t *testing.T) {
	db, _ := newFakeDB(t)
	_, err := NewEnvelope(db, map[int][]byte{1: []byte("tooshort")}, 1)
	if err == nil {
		t.Fatal("expected error for short mek")
	}
}

func TestNewEnvelope_ActiveMEKNotInMap(t *testing.T) {
	db, _ := newFakeDB(t)
	_, err := NewEnvelope(db, map[int][]byte{1: testMEK(0xAA)}, 99)
	if err == nil {
		t.Fatal("expected error when activeMEK not in meks")
	}
}

func TestNewEnvelope_Valid(t *testing.T) {
	db, _ := newFakeDB(t)
	_, err := NewEnvelope(db, map[int][]byte{1: testMEK(0xAA)}, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Encrypt / Decrypt round-trip
// ---------------------------------------------------------------------------

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	db, state := newFakeDB(t)
	orgID := uuid.New()
	mek1 := testMEK(0x11)
	plaintext := []byte("super secret token")

	// Script the INSERT ON CONFLICT...RETURNING query.
	state.push(fakeResult{
		cols: []string{"wrapped_dek", "mek_version"},
		rows: [][]driver.Value{{wrappedDEKForTest(t, mek1, orgID), int64(1)}},
	})

	e := newTestEnvelope(t, db, map[int][]byte{1: mek1}, 1)
	ctx := context.Background()

	ciphertext, err := e.Encrypt(ctx, orgID, plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if len(ciphertext) < 1+12+16 {
		t.Fatalf("ciphertext too short: %d bytes", len(ciphertext))
	}
	if ciphertext[0] != 1 {
		t.Errorf("version byte = %d, want 1", ciphertext[0])
	}

	// Decrypt uses the in-memory cache — no additional DB call.
	got, err := e.Decrypt(ctx, orgID, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if string(got) != string(plaintext) {
		t.Errorf("got %q, want %q", got, plaintext)
	}
}

func TestEncryptDecrypt_DifferentCiphertexts(t *testing.T) {
	db, state := newFakeDB(t)
	orgID := uuid.New()
	mek1 := testMEK(0x22)
	plaintext := []byte("hello")

	// Only one DB call even though we call Encrypt twice (cache).
	state.push(fakeResult{
		cols: []string{"wrapped_dek", "mek_version"},
		rows: [][]driver.Value{{wrappedDEKForTest(t, mek1, orgID), int64(1)}},
	})

	e := newTestEnvelope(t, db, map[int][]byte{1: mek1}, 1)
	ctx := context.Background()

	c1, err := e.Encrypt(ctx, orgID, plaintext)
	if err != nil {
		t.Fatalf("first Encrypt: %v", err)
	}
	c2, err := e.Encrypt(ctx, orgID, plaintext)
	if err != nil {
		t.Fatalf("second Encrypt: %v", err)
	}
	if string(c1) == string(c2) {
		t.Error("two Encrypt calls returned identical ciphertext (nonce not random)")
	}
}

// ---------------------------------------------------------------------------
// Decrypt error paths
// ---------------------------------------------------------------------------

func TestDecrypt_TooShort(t *testing.T) {
	db, _ := newFakeDB(t)
	e := newTestEnvelope(t, db, map[int][]byte{1: testMEK(0x33)}, 1)
	_, err := e.Decrypt(context.Background(), uuid.New(), []byte{0x01, 0x02})
	if err == nil {
		t.Fatal("expected error for too-short ciphertext")
	}
}

func TestDecrypt_UnknownMEKVersion(t *testing.T) {
	db, _ := newFakeDB(t)
	e := newTestEnvelope(t, db, map[int][]byte{1: testMEK(0x44)}, 1)

	// Build a syntactically valid-length ciphertext with version byte = 9 (not in map).
	buf := make([]byte, 1+12+16)
	buf[0] = 9
	_, err := e.Decrypt(context.Background(), uuid.New(), buf)
	if !errors.Is(err, ErrUnknownMEKVersion) {
		t.Errorf("expected ErrUnknownMEKVersion, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// DEK cache hit — only one DB call for multiple Encrypt calls
// ---------------------------------------------------------------------------

func TestEncrypt_DEKCacheHit(t *testing.T) {
	db, state := newFakeDB(t)
	orgID := uuid.New()
	mek1 := testMEK(0x99)

	// Script exactly ONE result; a second DB call would return an error.
	state.push(fakeResult{
		cols: []string{"wrapped_dek", "mek_version"},
		rows: [][]driver.Value{{wrappedDEKForTest(t, mek1, orgID), int64(1)}},
	})

	e := newTestEnvelope(t, db, map[int][]byte{1: mek1}, 1)
	ctx := context.Background()

	if _, err := e.Encrypt(ctx, orgID, []byte("first")); err != nil {
		t.Fatalf("first Encrypt: %v", err)
	}
	// Second Encrypt must NOT hit the DB (cache hit).
	if _, err := e.Encrypt(ctx, orgID, []byte("second")); err != nil {
		t.Fatalf("second Encrypt: %v", err)
	}
}

// ---------------------------------------------------------------------------
// RotateMEK
// ---------------------------------------------------------------------------

func TestRotateMEK_ReWrapsAndClears(t *testing.T) {
	db, state := newFakeDB(t)
	orgID := uuid.New()
	mek1 := testMEK(0x55)
	mek2 := testMEK(0x66)

	originalWrapped := wrappedDEKForTest(t, mek1, orgID)

	// Script SELECT query for rows with mek_version=1.
	// Pass org_id as string; uuid.UUID implements sql.Scanner and accepts strings.
	state.push(fakeResult{
		cols: []string{"org_id", "wrapped_dek"},
		rows: [][]driver.Value{{orgID.String(), originalWrapped}},
	})
	// Script UPDATE exec.
	state.push(fakeResult{rowsAffected: 1})

	e := newTestEnvelope(t, db, map[int][]byte{1: mek1, 2: mek2}, 1)

	// Pre-populate cache so we can confirm it was cleared.
	e.mu.Lock()
	e.cache[orgID] = []byte("cached_dek")
	e.mu.Unlock()

	if err := e.RotateMEK(context.Background(), 1, 2); err != nil {
		t.Fatalf("RotateMEK: %v", err)
	}

	e.mu.Lock()
	_, cacheHit := e.cache[orgID]
	e.mu.Unlock()
	if cacheHit {
		t.Error("cache should have been cleared after RotateMEK")
	}
}

func TestRotateMEK_UnknownFromVersion(t *testing.T) {
	db, _ := newFakeDB(t)
	e := newTestEnvelope(t, db, map[int][]byte{1: testMEK(0x77)}, 1)
	err := e.RotateMEK(context.Background(), 99, 1)
	if err == nil {
		t.Fatal("expected error for unknown fromVersion")
	}
}

func TestRotateMEK_UnknownToVersion(t *testing.T) {
	db, _ := newFakeDB(t)
	e := newTestEnvelope(t, db, map[int][]byte{1: testMEK(0x88)}, 1)
	err := e.RotateMEK(context.Background(), 1, 99)
	if err == nil {
		t.Fatal("expected error for unknown toVersion")
	}
}
