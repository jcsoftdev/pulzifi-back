package minio

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

type Client struct {
	minioClient *minio.Client
	bucketName  string
	publicURL   string
}

func NewClient(cfg *config.Config) (*Client, error) {
	client, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: cfg.MinIOUseSSL,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		minioClient: client,
		bucketName:  cfg.MinIOBucket,
		publicURL:   cfg.MinIOPublicURL,
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
	// Set policy to public read
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

// extractObjectName parses the object name from a public URL.
// E.g. "http://localhost:4566/snapshots/page-id/123.png" → "page-id/123.png"
func (c *Client) extractObjectName(objectURL string) string {
	// Look for the bucket name in the URL path and return everything after it.
	marker := "/" + c.bucketName + "/"
	idx := strings.Index(objectURL, marker)
	if idx >= 0 {
		return objectURL[idx+len(marker):]
	}
	return ""
}
