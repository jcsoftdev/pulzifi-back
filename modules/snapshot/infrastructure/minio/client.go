package minio

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

type Client struct {
	minioClient   *minio.Client
	presignClient *minio.Client // bound to the externally reachable host for presigning
	bucketName    string
	publicURL     string
	bucketPrivate bool
}

func NewClient(cfg *config.Config) (*Client, error) {
	client, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: cfg.MinIOUseSSL,
	})
	if err != nil {
		return nil, err
	}

	// Build a second client bound to the public-reachable host for presigning.
	// When no separate public URL is configured, reuse the internal client.
	var presignClient *minio.Client
	if cfg.MinIOPublicURL != "" {
		pubEndpoint := cfg.MinIOPublicURL
		// Strip scheme for the minio client constructor.
		pubEndpoint = strings.TrimPrefix(pubEndpoint, "https://")
		pubEndpoint = strings.TrimPrefix(pubEndpoint, "http://")
		// Strip trailing slash/path — only host:port matters.
		if idx := strings.Index(pubEndpoint, "/"); idx >= 0 {
			pubEndpoint = pubEndpoint[:idx]
		}
		useSSL := strings.HasPrefix(cfg.MinIOPublicURL, "https://")
		presignClient, err = minio.New(pubEndpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
			Secure: useSSL,
		})
		if err != nil {
			// Non-fatal — fall back to the internal client for presigning.
			logger.Warn("Failed to build presign client; falling back to internal client", zap.Error(err))
			presignClient = client
		}
	} else {
		presignClient = client
	}

	return &Client{
		minioClient:   client,
		presignClient: presignClient,
		bucketName:    cfg.MinIOBucket,
		publicURL:     cfg.MinIOPublicURL,
		bucketPrivate: cfg.SnapshotBucketPrivate,
	}, nil
}

func (c *Client) EnsureBucket(ctx context.Context) error {
	exists, err := c.minioClient.BucketExists(ctx, c.bucketName)
	if err != nil {
		return err
	}
	if !exists {
		err = c.minioClient.MakeBucket(ctx, c.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
	}
	// Only set a public-read policy when the bucket is NOT private.
	if c.bucketPrivate {
		return nil
	}
	policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`, c.bucketName)
	return c.minioClient.SetBucketPolicy(ctx, c.bucketName, policy)
}

func (c *Client) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := c.minioClient.PutObject(ctx, c.bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	// Construct public URL
	// If public URL doesn't end with slash, add it
	baseURL := c.publicURL
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	return fmt.Sprintf("%s%s/%s", baseURL, c.bucketName, objectName), nil
}

// Delete removes an object from the MinIO bucket by its object name.
// A missing object is treated as already deleted (idempotent).
func (c *Client) Delete(ctx context.Context, objectName string) error {
	err := c.minioClient.RemoveObject(ctx, c.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		// Treat "not found" responses as success (idempotent).
		resp := minio.ToErrorResponse(err)
		if resp.StatusCode == 404 || resp.Code == "NoSuchKey" {
			return nil
		}
		return fmt.Errorf("minio RemoveObject failed: %w", err)
	}
	return nil
}

// Download retrieves an object via the internal MinIO client, bypassing the public URL.
// This avoids issues where the public URL (e.g. http://localhost:4566) is unreachable
// from within Docker containers.
func (c *Client) Download(ctx context.Context, objectURL string) ([]byte, error) {
	objectName := c.extractObjectName(objectURL)
	if objectName == "" {
		return nil, fmt.Errorf("could not extract object name from URL: %s", objectURL)
	}

	obj, err := c.minioClient.GetObject(ctx, c.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("minio GetObject failed: %w", err)
	}
	defer func() { _ = obj.Close() }()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to read object body: %w", err)
	}

	logger.Debug("Downloaded object via internal MinIO client",
		zap.String("object", objectName), zap.Int("bytes", len(data)))

	return data, nil
}

// PresignedGetURL returns a time-limited pre-signed GET URL for the given storedURL.
// When the bucket is public (bucketPrivate=false) it returns storedURL unchanged
// (pass-through). When bucketPrivate=true it mints a presigned URL using the
// externally reachable host so the browser can access the object.
func (c *Client) PresignedGetURL(ctx context.Context, storedURL string, ttl time.Duration) (string, error) {
	if storedURL == "" {
		return "", nil
	}

	// Pass-through when the bucket is public — no presigning needed.
	if !c.bucketPrivate {
		return storedURL, nil
	}

	objectName := c.extractObjectName(storedURL)
	if objectName == "" {
		// Cannot parse the key; return the stored URL and let the caller handle it.
		return storedURL, nil
	}

	if ttl <= 0 {
		ttl = 15 * time.Minute
	}

	presignedURL, err := c.presignClient.PresignedGetObject(ctx, c.bucketName, objectName, ttl, url.Values{})
	if err != nil {
		return "", fmt.Errorf("minio PresignedGetObject: %w", err)
	}

	return presignedURL.String(), nil
}

// extractObjectName parses the object name from a public URL.
// E.g. "http://localhost:4566/snapshots/page-id/123.png" → "page-id/123.png"
func (c *Client) extractObjectName(objectURL string) string {
	// Look for the bucket name in the URL path and return everything after it.
	marker := "/" + c.bucketName + "/"
	idx := strings.Index(objectURL, marker)
	if idx >= 0 {
		key := objectURL[idx+len(marker):]
		if key == "" {
			return ""
		}
		// Strip query string if present (e.g. from already-presigned URLs).
		if qIdx := strings.Index(key, "?"); qIdx >= 0 {
			key = key[:qIdx]
		}
		return key
	}
	return ""
}

// ObjectNameFromURL implements repositories.ObjectStorage.
// It delegates to the existing extractObjectName which keys off the bucket marker.
// Returns "" when the URL does not contain the expected bucket path; callers must
// skip Delete for empty results.
func (c *Client) ObjectNameFromURL(storedURL string) string {
	return c.extractObjectName(storedURL)
}
