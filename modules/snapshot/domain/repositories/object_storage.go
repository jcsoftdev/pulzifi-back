package repositories

import (
	"context"
	"io"
)

type ObjectStorage interface {
	Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error)
	EnsureBucket(ctx context.Context) error
}
