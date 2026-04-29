package repositories

import (
	"context"
	"io"
)

type ObjectStorage interface {
	Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error)
	// Download retrieves an object by its public URL, using the internal storage
	// client to avoid network issues (e.g. localhost URLs unreachable from Docker).
	Download(ctx context.Context, objectURL string) ([]byte, error)
	EnsureBucket(ctx context.Context) error
}
