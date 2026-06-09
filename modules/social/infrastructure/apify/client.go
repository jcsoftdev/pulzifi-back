// Package apify provides the Apify actor client for the social module.
// It implements the SocialFetcher port using Apify's run-sync-get-dataset-items API.
package apify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/social/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/services"
	valueobjects "github.com/jcsoftdev/pulzifi-back/modules/social/domain/value_objects"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

const (
	// apifyRunSyncURL is the Apify endpoint for synchronous actor runs.
	apifyRunSyncURL = "https://api.apify.com/v2/acts/%s/run-sync-get-dataset-items"

	// defaultTimeout is the maximum time to wait for an Apify actor run.
	// Instagram scrapes take 30–60s; 90s gives comfortable headroom.
	defaultTimeout = 90 * time.Second
)

// Compile-time interface check.
var _ services.SocialFetcher = (*Client)(nil)

// Client is the Apify actor client implementing services.SocialFetcher.
type Client struct {
	token              string
	actorInstagram     string
	actorTikTok        string
	actorFacebook      string
	httpClient         *http.Client
}

// NewClient creates a new Apify client.
//
//   - token            — Apify API token (APIFY_TOKEN)
//   - actorInstagram   — actor ID for Instagram (APIFY_ACTOR_INSTAGRAM)
//   - actorTikTok      — actor ID for TikTok (APIFY_ACTOR_TIKTOK, Phase 2)
//   - actorFacebook    — actor ID for Facebook (APIFY_ACTOR_FACEBOOK, Phase 2)
func NewClient(token, actorInstagram, actorTikTok, actorFacebook string) *Client {
	return &Client{
		token:          token,
		actorInstagram: actorInstagram,
		actorTikTok:    actorTikTok,
		actorFacebook:  actorFacebook,
		httpClient: &http.Client{
			// Timeout is also enforced via context.WithTimeout in FetchProfile;
			// the transport-level timeout here is a safety net.
			Timeout: defaultTimeout + 10*time.Second,
		},
	}
}

// FetchProfile retrieves the current state of a social profile via the appropriate
// Apify actor for the given platform.
//
// Returns domainerrors.ErrPlatformNotSupported for unsupported platforms.
// Returns domainerrors.ErrFetchFailed on provider errors, timeouts, or empty results.
// One retry is performed on 5xx responses before returning ErrFetchFailed.
func (c *Client) FetchProfile(
	ctx context.Context,
	platform valueobjects.Platform,
	handle string,
	postsLimit int,
) (*entities.ProfileData, error) {
	actorID, mapper, err := c.resolveAdapter(platform)
	if err != nil {
		return nil, err
	}

	body, bodyErr := c.buildRequestBody(platform, handle, postsLimit)
	if bodyErr != nil {
		return nil, fmt.Errorf("%w: build request: %v", domainerrors.ErrFetchFailed, bodyErr)
	}

	// Apply a timeout to cap the Apify run duration (design §6: 60–90s).
	runCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	items, fetchErr := c.runWithRetry(runCtx, actorID, postsLimit, body)
	if fetchErr != nil {
		return nil, fetchErr
	}

	data, mapErr := mapper(items)
	if mapErr != nil {
		return nil, fmt.Errorf("%w: map response: %v", domainerrors.ErrFetchFailed, mapErr)
	}

	return data, nil
}

// resolveAdapter returns the actor ID and mapper function for the given platform.
func (c *Client) resolveAdapter(platform valueobjects.Platform) (string, mapperFunc, error) {
	switch platform {
	case valueobjects.PlatformInstagram:
		if c.actorInstagram == "" {
			return "", nil, fmt.Errorf("%w: Instagram actor not configured", domainerrors.ErrPlatformNotSupported)
		}
		return c.actorInstagram, mapInstagramItems, nil
	case valueobjects.PlatformTikTok:
		if c.actorTikTok == "" {
			return "", nil, domainerrors.ErrPlatformNotSupported
		}
		return c.actorTikTok, mapInstagramItems, nil // placeholder until mapper exists
	case valueobjects.PlatformFacebook:
		if c.actorFacebook == "" {
			return "", nil, domainerrors.ErrPlatformNotSupported
		}
		return c.actorFacebook, mapInstagramItems, nil // placeholder until mapper exists
	default:
		return "", nil, domainerrors.ErrPlatformNotSupported
	}
}

// buildRequestBody constructs the JSON payload for the Apify actor run.
func (c *Client) buildRequestBody(platform valueobjects.Platform, handle string, postsLimit int) ([]byte, error) {
	switch platform {
	case valueobjects.PlatformInstagram:
		payload := map[string]interface{}{
			"usernames":      []string{handle},
			"resultsLimit":   postsLimit,
			"addParentData":  false,
		}
		return json.Marshal(payload)
	default:
		// Generic fallback for future platforms.
		payload := map[string]interface{}{
			"usernames":    []string{handle},
			"resultsLimit": postsLimit,
		}
		return json.Marshal(payload)
	}
}

// runWithRetry calls the Apify run-sync endpoint. On 5xx it retries once.
func (c *Client) runWithRetry(ctx context.Context, actorID string, postsLimit int, body []byte) ([]json.RawMessage, error) {
	items, statusCode, err := c.doRun(ctx, actorID, postsLimit, body)
	if err == nil {
		return items, nil
	}

	// Retry once on 5xx (transient provider errors).
	if statusCode >= 500 {
		logger.Warn("Apify actor run returned 5xx; retrying once",
			zap.String("actor", actorID),
			zap.Int("status", statusCode),
		)
		items, _, retryErr := c.doRun(ctx, actorID, postsLimit, body)
		if retryErr != nil {
			return nil, retryErr
		}
		return items, nil
	}

	return nil, err
}

// doRun performs a single POST to Apify run-sync-get-dataset-items.
// Returns the raw dataset items, the HTTP status code, and any error.
func (c *Client) doRun(ctx context.Context, actorID string, postsLimit int, body []byte) ([]json.RawMessage, int, error) {
	endpoint := fmt.Sprintf(apifyRunSyncURL, url.PathEscape(actorID))

	// Build query params: token + resultsLimit cap
	params := url.Values{}
	params.Set("token", c.token)
	params.Set("resultsLimit", fmt.Sprintf("%d", postsLimit))
	fullURL := endpoint + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewReader(body))
	if err != nil {
		return nil, 0, fmt.Errorf("%w: create request: %v", domainerrors.ErrFetchFailed, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: request: %v", domainerrors.ErrFetchFailed, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 500 {
		return nil, resp.StatusCode, fmt.Errorf("%w: apify returned status %d", domainerrors.ErrFetchFailed, resp.StatusCode)
	}

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, resp.StatusCode, fmt.Errorf("%w: apify returned status %d: %s",
			domainerrors.ErrFetchFailed, resp.StatusCode, string(bodyBytes))
	}

	var items []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("%w: decode response: %v", domainerrors.ErrFetchFailed, err)
	}

	if len(items) == 0 {
		return nil, resp.StatusCode, fmt.Errorf("%w: actor returned empty dataset (profile may not exist or is private)",
			domainerrors.ErrFetchFailed)
	}

	return items, resp.StatusCode, nil
}

// mapperFunc is the signature for actor-specific response mappers.
type mapperFunc func(items []json.RawMessage) (*entities.ProfileData, error)
