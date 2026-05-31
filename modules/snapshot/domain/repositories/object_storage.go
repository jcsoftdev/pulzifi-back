package repositories

import (
	"context"
	"io"
	"time"
)

type ObjectStorage interface {
	Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error)
	// Download retrieves an object by its public URL, using the internal storage
	// client to avoid network issues (e.g. localhost URLs unreachable from Docker).
	Download(ctx context.Context, objectURL string) ([]byte, error)
	// Delete removes an object from storage by its object name (path within the
	// bucket/folder). Implementations must be idempotent: deleting a non-existent
	// object should not return an error.
	Delete(ctx context.Context, objectName string) error
	EnsureBucket(ctx context.Context) error
	// ObjectNameFromURL derives the storage object name from a URL previously
	// returned by Upload. Returns "" when the URL cannot be parsed or does not
	// belong to this storage backend; callers must skip Delete for empty results.
	ObjectNameFromURL(storedURL string) string
	// PresignedGetURL returns a time-limited pre-signed URL for reading storedURL.
	// When the bucket is public (SNAPSHOT_BUCKET_PRIVATE=false) implementations
	// MAY return storedURL unchanged as a pass-through. ttl is the requested
	// validity window; a zero ttl means the implementation uses its configured default.
	PresignedGetURL(ctx context.Context, storedURL string, ttl time.Duration) (string, error)
}

// StorageFlagReader is the port that the application layer calls to check whether
// write-time key prefixing (tenant isolation) is enabled for a given tenant schema.
// It is satisfied by an adapter in cmd/wiring/snapshot that bridges schema-name →
// org-ID lookup and featureflags.Reader.IsOn. Default-deny on any error.
type StorageFlagReader interface {
	IsPrivateStorageEnabled(ctx context.Context, schemaName string) (bool, error)
}
