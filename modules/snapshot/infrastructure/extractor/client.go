package extractor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type SectionExtractOption struct {
	ID            string           `json:"id"`
	Selector      string           `json:"selector,omitempty"`
	SelectorXPath string           `json:"selectorXpath,omitempty"`
	Offsets       *SelectorOffsets `json:"selectorOffsets,omitempty"`
}

type SectionExtractResult struct {
	ID               string `json:"id"`
	ScreenshotBase64 string `json:"screenshot_base64"`
	HTML             string `json:"html"`
	Text             string `json:"text"`
	SelectorMatched  bool   `json:"selector_matched"`
}

type ExtractorResult struct {
	Title            string                 `json:"title"`
	HTML             string                 `json:"html"`
	Text             string                 `json:"text"`
	ScreenshotBase64 string                 `json:"screenshot_base64"`
	SelectorMatched  bool                   `json:"selector_matched"`
	Sections         []SectionExtractResult `json:"sections,omitempty"`
}

type SelectorOffsets struct {
	Top    int `json:"top"`
	Right  int `json:"right"`
	Bottom int `json:"bottom"`
	Left   int `json:"left"`
}

type ExtractOptions struct {
	BlockAdsCookies bool
	Selector        string
	SelectorXPath   string
	SelectorOffsets *SelectorOffsets
	Sections        []SectionExtractOption
	IgnoreSelectors []string
}

type PreviewElement struct {
	Selector     string      `json:"selector"`
	XPath        string      `json:"xpath"`
	Tag          string      `json:"tag"`
	Rect         ElementRect `json:"rect"`
	TextPreview  string      `json:"text_preview"`
	SemanticRole string      `json:"semantic_role"`
}

type ElementRect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type PreviewResult struct {
	ScreenshotBase64 string           `json:"screenshot_base64"`
	Viewport         PreviewViewport  `json:"viewport"`
	PageHeight       int              `json:"page_height"`
	Elements         []PreviewElement `json:"elements"`
}

type PreviewViewport struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type HTTPClient struct {
	baseURL         string
	apiKey          string
	httpClient      *http.Client
	streamingClient *http.Client
}

func NewHTTPClient(baseURL string) *HTTPClient {
	return NewHTTPClientWithKey(baseURL, "")
}

// NewHTTPClientWithKey creates an HTTPClient that sets X-API-Key on every
// request. Use this in production where EXTRACTOR_API_KEY is configured.
func NewHTTPClientWithKey(baseURL, apiKey string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		// streamingClient has no global timeout so SSE streams for large
		// pages aren't killed prematurely. Callers should use context
		// cancellation instead.
		streamingClient: &http.Client{
			Timeout: 0,
		},
	}
}

// setAuthHeader attaches the X-API-Key header when an API key is configured.
func (c *HTTPClient) setAuthHeader(req *http.Request) {
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}
}

// Extract sends an extraction request to the scraper service.
// Uses the streaming client with a 5-minute context deadline so that large
// pages (which may take minutes to scroll/screenshot) are not killed prematurely.
func (c *HTTPClient) Extract(ctx context.Context, url string, opts ExtractOptions) (*ExtractorResult, error) {
	body, err := buildExtractPayload(url, opts)
	if err != nil {
		return nil, err
	}

	extractCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			select {
			case <-extractCtx.Done():
				return nil, extractCtx.Err()
			case <-time.After(time.Duration(attempt) * 2 * time.Second):
			}
		}

		req, err := c.newExtractRequest(extractCtx, body)
		if err != nil {
			return nil, err
		}

		result, retry, err := c.doExtractAttempt(req)
		if err != nil {
			if retry {
				lastErr = err
				continue
			}
			return nil, err
		}
		return result, nil
	}

	return nil, fmt.Errorf("extractor failed after 3 attempts: %w", lastErr)
}

// buildExtractPayload marshals the extract request body from the options.
func buildExtractPayload(url string, opts ExtractOptions) ([]byte, error) {
	payload := map[string]interface{}{
		"url":               url,
		"block_ads_cookies": opts.BlockAdsCookies,
	}
	if opts.Selector != "" {
		payload["selector"] = opts.Selector
	}
	if opts.SelectorXPath != "" {
		payload["selector_xpath"] = opts.SelectorXPath
	}
	if opts.SelectorOffsets != nil {
		payload["selector_offsets"] = opts.SelectorOffsets
	}
	if len(opts.Sections) > 0 {
		payload["sections"] = opts.Sections
	}
	if len(opts.IgnoreSelectors) > 0 {
		payload["ignore_selectors"] = opts.IgnoreSelectors
	}
	return json.Marshal(payload)
}

// newExtractRequest builds a POST /extract request with the standard headers.
func (c *HTTPClient) newExtractRequest(ctx context.Context, body []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/extract", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(req)
	return req, nil
}

// doExtractAttempt performs a single extraction request. It returns the decoded
// result on success. The retry flag is true when the caller should retry the
// request (transport error or 5xx response).
func (c *HTTPClient) doExtractAttempt(req *http.Request) (result *ExtractorResult, retry bool, err error) {
	resp, err := c.streamingClient.Do(req)
	if err != nil {
		return nil, true, err
	}

	if resp.StatusCode >= 500 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		_ = resp.Body.Close()
		return nil, true, fmt.Errorf("extractor service returned status: %d, body: %s", resp.StatusCode, string(respBody))
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		_ = resp.Body.Close()
		return nil, false, fmt.Errorf("extractor service returned status: %d, body: %s", resp.StatusCode, string(respBody))
	}

	defer func() { _ = resp.Body.Close() }()
	var decoded ExtractorResult
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, false, err
	}
	return &decoded, false, nil
}

func (c *HTTPClient) Preview(ctx context.Context, url string, blockAdsCookies bool) (*PreviewResult, error) {
	payload := map[string]interface{}{
		"url":               url,
		"block_ads_cookies": blockAdsCookies,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/preview", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("extractor preview returned status: %d", resp.StatusCode)
	}

	var result PreviewResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// PreviewStream sends a preview request to the extractor and returns the raw
// SSE stream response. The caller is responsible for closing resp.Body.
// The extractor streams "progress" and "result"/"error" SSE events.
// Uses a dedicated streaming client without a global timeout so that large
// pages (which may take minutes to scroll/screenshot) are not killed prematurely.
// A 5-minute context deadline is applied as an upper bound.
func (c *HTTPClient) PreviewStream(ctx context.Context, url string, blockAdsCookies bool) (*http.Response, error) {
	payload := map[string]interface{}{
		"url":               url,
		"block_ads_cookies": blockAdsCookies,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	streamCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)

	req, err := http.NewRequestWithContext(streamCtx, "POST", c.baseURL+"/preview", bytes.NewBuffer(body))
	if err != nil {
		cancel()
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(req)

	resp, err := c.streamingClient.Do(req)
	if err != nil {
		cancel()
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		cancel()
		return nil, fmt.Errorf("extractor preview returned status: %d", resp.StatusCode)
	}

	// Wrap the body so that the context cancel is called when the body is closed.
	resp.Body = &cancelOnClose{ReadCloser: resp.Body, cancel: cancel}
	return resp, nil
}

// cancelOnClose wraps an io.ReadCloser and calls a cancel function on Close.
type cancelOnClose struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (c *cancelOnClose) Close() error {
	err := c.ReadCloser.Close()
	c.cancel()
	return err
}
