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
	cloudName      string
	apiKey         string
	apiSecret      string
	folder         string
	bucketPrivate  bool
	http           *http.Client
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
		cloudName:     cfg.CloudinaryCloudName,
		apiKey:        cfg.CloudinaryAPIKey,
		apiSecret:     cfg.CloudinaryAPISecret,
		folder:        strings.TrimSpace(cfg.CloudinaryFolder),
		bucketPrivate: cfg.SnapshotBucketPrivate,
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

	body, contentTypeHeader, err := c.buildUploadBody(objectName, publicID, signature, params["timestamp"], reader)
	if err != nil {
		return "", err
	}

	uploadURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/%s/upload", c.cloudName, resourceType)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", contentTypeHeader)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	var uploadRes uploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadRes); err != nil {
		return "", err
	}

	return parseUploadResponse(uploadRes, resp.StatusCode)
}

// buildUploadBody assembles the multipart form body for a Cloudinary upload and
// returns the body buffer together with the multipart Content-Type header value.
func (c *Client) buildUploadBody(objectName, publicID, signature, timestamp string, reader io.Reader) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	filePart, err := writer.CreateFormFile("file", path.Base(objectName))
	if err != nil {
		return nil, "", err
	}
	if _, err := io.Copy(filePart, reader); err != nil {
		return nil, "", err
	}

	fields := map[string]string{
		"api_key":   c.apiKey,
		"timestamp": timestamp,
		"signature": signature,
		"public_id": publicID,
	}
	if c.folder != "" {
		fields["folder"] = c.folder
	}
	for name, value := range fields {
		if err := writer.WriteField(name, value); err != nil {
			return nil, "", err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}
	return body, writer.FormDataContentType(), nil
}

// parseUploadResponse maps a decoded Cloudinary response and HTTP status to the
// resulting URL or an error.
func parseUploadResponse(uploadRes uploadResponse, statusCode int) (string, error) {
	if statusCode >= 300 {
		if uploadRes.Error != nil && uploadRes.Error.Message != "" {
			return "", fmt.Errorf("cloudinary upload failed: %s", uploadRes.Error.Message)
		}
		return "", fmt.Errorf("cloudinary upload failed with status %d", statusCode)
	}

	if uploadRes.SecureURL != "" {
		return uploadRes.SecureURL, nil
	}
	if uploadRes.URL != "" {
		return uploadRes.URL, nil
	}

	return "", fmt.Errorf("cloudinary upload did not return a URL")
}

// Delete removes an asset from Cloudinary by its public_id (object name without extension).
// A 404 response is treated as success (idempotent).
func (c *Client) Delete(ctx context.Context, objectName string) error {
	// Derive resource type from extension: images use "image", everything else "raw".
	resourceType := "raw"
	ext := strings.ToLower(path.Ext(objectName))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp":
		resourceType = "image"
	}

	// publicID is the objectName without extension, with optional folder prefix.
	publicID := strings.TrimSuffix(objectName, path.Ext(objectName))
	if c.folder != "" {
		publicID = c.folder + "/" + publicID
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	params := map[string]string{
		"public_id":     publicID,
		"timestamp":     timestamp,
		"resource_type": resourceType,
	}
	signature := c.sign(params)

	form := fmt.Sprintf("public_id=%s&timestamp=%s&api_key=%s&signature=%s",
		publicID, timestamp, c.apiKey, signature)

	destroyURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/%s/destroy",
		c.cloudName, resourceType)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, destroyURL,
		strings.NewReader(form))
	if err != nil {
		return fmt.Errorf("cloudinary delete: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("cloudinary delete: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil // already deleted
	}

	var result struct {
		Result string `json:"result"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("cloudinary delete: decode response: %w", err)
	}
	if result.Error != nil && result.Error.Message != "" {
		return fmt.Errorf("cloudinary delete failed: %s", result.Error.Message)
	}
	// "not found" result is also idempotent success
	if result.Result != "ok" && result.Result != "not found" {
		return fmt.Errorf("cloudinary delete: unexpected result %q", result.Result)
	}
	return nil
}

// Download retrieves an object by fetching its public URL via HTTP.
// Cloudinary URLs are always publicly accessible CDN endpoints.
func (c *Client) Download(ctx context.Context, objectURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, objectURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return data, nil
}

// ObjectNameFromURL implements repositories.ObjectStorage.
// Cloudinary URLs have the shape:
//
//	https://res.cloudinary.com/<cloud>/<resource_type>/upload[/v<ts>]/<public_id>.<ext>
//
// The original objectName is the public_id with its extension re-added, minus any
// folder prefix that the client prepends during Upload. Returns "" when the URL
// cannot be parsed; callers must skip Delete for empty results.
func (c *Client) ObjectNameFromURL(storedURL string) string {
	// Find the "/upload/" marker in the URL.
	const uploadMarker = "/upload/"
	idx := strings.Index(storedURL, uploadMarker)
	if idx < 0 {
		return ""
	}
	rest := storedURL[idx+len(uploadMarker):]

	// Strip optional version segment (e.g. "v1234567890/").
	if len(rest) > 1 && rest[0] == 'v' {
		if slashIdx := strings.Index(rest, "/"); slashIdx > 1 {
			// Only strip if what follows the 'v' and before the slash looks like digits.
			candidate := rest[1:slashIdx]
			allDigits := true
			for _, ch := range candidate {
				if ch < '0' || ch > '9' {
					allDigits = false
					break
				}
			}
			if allDigits {
				rest = rest[slashIdx+1:]
			}
		}
	}

	if rest == "" {
		return ""
	}

	// Strip the folder prefix that the client prepends during Upload.
	if c.folder != "" {
		prefix := c.folder + "/"
		rest = strings.TrimPrefix(rest, prefix)
	}

	return rest
}

// PresignedGetURL returns a time-limited signed delivery URL for private Cloudinary
// assets (type=authenticated). For public assets (no "/authenticated/" segment in
// the URL) it returns storedURL unchanged. An empty storedURL is returned as-is.
//
// The signed URL follows Cloudinary's SHA-1 signed URL pattern:
// https://cloudinary.com/documentation/advanced_url_delivery_options#generating_delivery_url_signatures
func (c *Client) PresignedGetURL(_ context.Context, storedURL string, ttl time.Duration) (string, error) {
	if storedURL == "" {
		return "", nil
	}

	// Public assets: pass-through.
	if !strings.Contains(storedURL, "/authenticated/") {
		return storedURL, nil
	}

	// Build expiry timestamp.
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	expiry := time.Now().Add(ttl).Unix()
	timestamp := strconv.FormatInt(expiry, 10)

	// Extract the public_id from the URL (everything after "/authenticated/[v<ts>/]").
	const authMarker = "/authenticated/"
	idx := strings.Index(storedURL, authMarker)
	rest := storedURL[idx+len(authMarker):]

	// Strip optional version segment.
	if len(rest) > 1 && rest[0] == 'v' {
		if slashIdx := strings.Index(rest, "/"); slashIdx > 1 {
			candidate := rest[1:slashIdx]
			allDigits := true
			for _, ch := range candidate {
				if ch < '0' || ch > '9' {
					allDigits = false
					break
				}
			}
			if allDigits {
				rest = rest[slashIdx+1:]
			}
		}
	}

	// public_id is the path without extension.
	publicID := strings.TrimSuffix(rest, path.Ext(rest))

	// Determine resource type from the URL path.
	resourceType := "image"
	if strings.Contains(storedURL, "/raw/") {
		resourceType = "raw"
	}

	// Sign: public_id=<id>&timestamp=<ts><apiSecret>
	params := map[string]string{
		"public_id": publicID,
		"timestamp": timestamp,
	}
	signature := c.sign(params)

	// Construct the signed delivery URL.
	signedURL := fmt.Sprintf(
		"https://res.cloudinary.com/%s/%s/authenticated/%s?timestamp=%s&signature=%s&api_key=%s",
		c.cloudName, resourceType, rest, timestamp, signature, c.apiKey,
	)

	return signedURL, nil
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
