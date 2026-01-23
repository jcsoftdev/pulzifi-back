package minio

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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
