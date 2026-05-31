package application

import (
	"context"
	"testing"
)

// stubStorageFlagReader implements repositories.StorageFlagReader for tests.
type stubStorageFlagReader struct {
	enabled bool
	err     error
}

func (s *stubStorageFlagReader) IsPrivateStorageEnabled(_ context.Context, _ string) (bool, error) {
	return s.enabled, s.err
}

func TestSnapshotWorker_objectKey_prefixEnabled(t *testing.T) {
	w := &SnapshotWorker{
		storageFlagReader: &stubStorageFlagReader{enabled: true},
	}

	got := w.objectKey(context.Background(), "tenant1", "page-uuid/1234.png")
	want := "tenant1/page-uuid/1234.png"
	if got != want {
		t.Errorf("objectKey (prefix on) = %q, want %q", got, want)
	}
}

func TestSnapshotWorker_objectKey_prefixDisabled(t *testing.T) {
	w := &SnapshotWorker{
		storageFlagReader: &stubStorageFlagReader{enabled: false},
	}

	got := w.objectKey(context.Background(), "tenant1", "page-uuid/1234.png")
	want := "page-uuid/1234.png"
	if got != want {
		t.Errorf("objectKey (prefix off) = %q, want %q", got, want)
	}
}

func TestSnapshotWorker_objectKey_nilFlagReader_defaultDeny(t *testing.T) {
	// When storageFlagReader is nil, objectKey must NOT prefix (default deny / safe public mode).
	w := &SnapshotWorker{storageFlagReader: nil}

	got := w.objectKey(context.Background(), "tenant1", "page-uuid/1234.png")
	want := "page-uuid/1234.png"
	if got != want {
		t.Errorf("objectKey (nil reader) = %q, want %q (expected no prefix)", got, want)
	}
}

func TestSnapshotWorker_objectKey_flagError_defaultDeny(t *testing.T) {
	// When the flag reader errors, objectKey must NOT prefix (safe public fallback).
	w := &SnapshotWorker{
		storageFlagReader: &stubStorageFlagReader{err: context.DeadlineExceeded},
	}

	got := w.objectKey(context.Background(), "tenant1", "page-uuid/1234.png")
	want := "page-uuid/1234.png"
	if got != want {
		t.Errorf("objectKey (error) = %q, want %q (expected no prefix)", got, want)
	}
}

func TestSnapshotWorker_objectKey_emptySchema(t *testing.T) {
	// Empty schemaName with prefix enabled should prefix with empty string — no double slash.
	w := &SnapshotWorker{
		storageFlagReader: &stubStorageFlagReader{enabled: true},
	}
	got := w.objectKey(context.Background(), "", "page-uuid/1234.png")
	want := "page-uuid/1234.png" // empty schema → treat as no prefix (same as default deny)
	if got != want {
		t.Errorf("objectKey (empty schema) = %q, want %q", got, want)
	}
}
