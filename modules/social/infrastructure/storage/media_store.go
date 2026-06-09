// Package storage provides the MediaStore implementation for the social module.
// It downloads CDN media URLs (which expire) and re-uploads them to the project's
// configured object storage (MinIO or Cloudinary) via the same provider selection
// as modules/snapshot/infrastructure/storage/provider.go.
package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/services"
	snapshotcloudinary "github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/cloudinary"
	snapshotminio "github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/minio"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
)

// Compile-time interface check.
var _ services.MediaStore = (*MediaStore)(nil)

// objectStorage is the subset of the snapshot ObjectStorage interface used here.
// We only need Upload; the social module does not presign or delete via this path.
type objectStorage interface {
	Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error)
}

// MediaStore implements services.MediaStore over MinIO or Cloudinary.
type MediaStore struct {
	storage    objectStorage
	httpClient *http.Client
}

// NewMediaStore creates a MediaStore backed by the configured object storage provider.
// Provider selection mirrors modules/snapshot/infrastructure/storage/provider.go:
// "" | "minio" | "s3" → MinIO; "cloudinary" → Cloudinary.
func NewMediaStore(cfg *config.Config) (*MediaStore, error) {
	provider := strings.ToLower(strings.TrimSpace(cfg.ObjectStorageProvider))

	var stor objectStorage
	var err error

	switch provider {
	case "", "minio", "s3":
		stor, err = snapshotminio.NewClient(cfg)
		if err != nil {
			return nil, fmt.Errorf("social media store: minio client: %w", err)
		}
	case "cloudinary":
		stor, err = snapshotcloudinary.NewClient(cfg)
		if err != nil {
			return nil, fmt.Errorf("social media store: cloudinary client: %w", err)
		}
	default:
		return nil, fmt.Errorf("social media store: unsupported provider %q", cfg.ObjectStorageProvider)
	}

	return &MediaStore{
		storage: stor,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Store downloads the asset at sourceURL and uploads it to object storage
// under the given key (e.g. "social/{profileID}/posts/{externalID}").
// Returns the durable stored URL.
func (m *MediaStore) Store(ctx context.Context, sourceURL, key string) (string, error) {
	// Download the source asset.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return "", fmt.Errorf("social media store: create download request: %w", err)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("social media store: download %q: %w", sourceURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("social media store: download %q: status %d", sourceURL, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("social media store: read body: %w", err)
	}

	// Determine content type from extension or Content-Type header.
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = guessContentType(sourceURL)
	}

	// Upload to object storage.
	storedURL, err := m.storage.Upload(ctx, key, bytes.NewReader(data), int64(len(data)), contentType)
	if err != nil {
		return "", fmt.Errorf("social media store: upload %q: %w", key, err)
	}

	return storedURL, nil
}

// guessContentType returns a content-type based on the URL file extension.
func guessContentType(rawURL string) string {
	ext := strings.ToLower(path.Ext(rawURL))
	// Strip query string from extension.
	if idx := strings.Index(ext, "?"); idx >= 0 {
		ext = ext[:idx]
	}
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".mp4":
		return "video/mp4"
	default:
		return "application/octet-stream"
	}
}
