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
	httpClient      *http.Client
	streamingClient *http.Client
}

func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
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

// Extract sends an extraction request to the scraper service.
// Uses the streaming client with a 5-minute context deadline so that large
// pages (which may take minutes to scroll/screenshot) are not killed prematurely.
func (c *HTTPClient) Extract(ctx context.Context, url string, opts ExtractOptions) (*ExtractorResult, error) {
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

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	extractCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(extractCtx, "POST", c.baseURL+"/extract", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.streamingClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("extractor service returned status: %d", resp.StatusCode)
	}

	var result ExtractorResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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

	resp, err := c.streamingClient.Do(req)
	if err != nil {
		cancel()
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
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
