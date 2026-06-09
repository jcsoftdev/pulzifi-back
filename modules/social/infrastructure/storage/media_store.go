// Package storage provides the MediaStore implementation for the social module.
// It downloads CDN media URLs (which expire) and re-uploads them to the project's
// configured object storage. The underlying storage backend is injected via the
// ObjectUploader interface — the concrete implementation (MinIO or Cloudinary) is
// constructed by cmd/wiring/social and passed in, keeping cross-module imports out
// of the social module (REQ-WIRING-01).
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
)

// Compile-time interface check.
var _ services.MediaStore = (*MediaStore)(nil)

// ObjectUploader is a minimal abstraction over the object storage upload operation.
// Implementors: modules/snapshot/infrastructure/minio.Client and
// modules/snapshot/infrastructure/cloudinary.Client (injected from cmd/wiring/social).
type ObjectUploader interface {
	Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error)
}

// MediaStore implements services.MediaStore over an injected ObjectUploader.
type MediaStore struct {
	uploader   ObjectUploader
	httpClient *http.Client
}

// NewMediaStore creates a MediaStore backed by the given ObjectUploader.
// The concrete uploader is created by cmd/wiring/social to avoid cross-module imports.
func NewMediaStore(uploader ObjectUploader) *MediaStore {
	return &MediaStore{
		uploader: uploader,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
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
	storedURL, err := m.uploader.Upload(ctx, key, bytes.NewReader(data), int64(len(data)), contentType)
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
