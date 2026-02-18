package cloudinary

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jcsoftdev/pulzifi-back/shared/config"
)

type Client struct {
	cloudName string
	apiKey    string
	apiSecret string
	folder    string
	http      *http.Client
}

type uploadResponse struct {
	SecureURL string `json:"secure_url"`
	URL       string `json:"url"`
	Error     *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func NewClient(cfg *config.Config) (*Client, error) {
	if cfg.CloudinaryCloudName == "" || cfg.CloudinaryAPIKey == "" || cfg.CloudinaryAPISecret == "" {
		return nil, fmt.Errorf("cloudinary credentials are required")
	}

	return &Client{
		cloudName: cfg.CloudinaryCloudName,
		apiKey:    cfg.CloudinaryAPIKey,
		apiSecret: cfg.CloudinaryAPISecret,
		folder:    strings.TrimSpace(cfg.CloudinaryFolder),
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (c *Client) EnsureBucket(ctx context.Context) error {
	return nil
}

func (c *Client) Upload(ctx context.Context, objectName string, reader io.Reader, _ int64, contentType string) (string, error) {
	timestamp := time.Now().Unix()
	publicID := strings.TrimSuffix(strings.TrimPrefix(objectName, "/"), path.Ext(objectName))

	params := map[string]string{
		"public_id": publicID,
		"timestamp": strconv.FormatInt(timestamp, 10),
	}
	if c.folder != "" {
		params["folder"] = c.folder
	}

	signature := c.sign(params)
	resourceType := "raw"
	if strings.HasPrefix(contentType, "image/") {
		resourceType = "image"
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	filePart, err := writer.CreateFormFile("file", path.Base(objectName))
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(filePart, reader); err != nil {
		return "", err
	}

	if err := writer.WriteField("api_key", c.apiKey); err != nil {
		return "", err
	}
	if err := writer.WriteField("timestamp", params["timestamp"]); err != nil {
		return "", err
	}
	if err := writer.WriteField("signature", signature); err != nil {
		return "", err
	}
	if err := writer.WriteField("public_id", publicID); err != nil {
		return "", err
	}
	if c.folder != "" {
		if err := writer.WriteField("folder", c.folder); err != nil {
			return "", err
		}
	}
	if err := writer.Close(); err != nil {
		return "", err
	}

	uploadURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/%s/upload", c.cloudName, resourceType)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var uploadRes uploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadRes); err != nil {
		return "", err
	}

	if resp.StatusCode >= 300 {
		if uploadRes.Error != nil && uploadRes.Error.Message != "" {
			return "", fmt.Errorf("cloudinary upload failed: %s", uploadRes.Error.Message)
		}
		return "", fmt.Errorf("cloudinary upload failed with status %d", resp.StatusCode)
	}

	if uploadRes.SecureURL != "" {
		return uploadRes.SecureURL, nil
	}
	if uploadRes.URL != "" {
		return uploadRes.URL, nil
	}

	return "", fmt.Errorf("cloudinary upload did not return a URL")
}

func (c *Client) sign(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, key := range keys {
		value := strings.TrimSpace(params[key])
		if value == "" {
			continue
		}
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
	}

	raw := strings.Join(pairs, "&") + c.apiSecret
	hash := sha1.Sum([]byte(raw))
	return hex.EncodeToString(hash[:])
}
