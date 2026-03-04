package extractor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ExtractorResult struct {
	Title            string `json:"title"`
	HTML             string `json:"html"`
	Text             string `json:"text"`
	ScreenshotBase64 string `json:"screenshot_base64"`
	SelectorMatched  bool   `json:"selector_matched"`
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
	baseURL    string
	httpClient *http.Client
}

func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

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

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/extract", bytes.NewBuffer(body))
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
