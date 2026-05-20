package services

import "context"

// ExtractOptions holds options for the page extraction request.
type ExtractOptions struct {
	BlockAdsCookies bool
	Selector        string
	SelectorXPath   string
	SelectorOffsets *SelectorOffsets
	Sections        []SectionExtractOption
	IgnoreSelectors []string
}

// SelectorOffsets matches the extractor's offset type.
type SelectorOffsets struct {
	Top    int
	Right  int
	Bottom int
	Left   int
}

// SectionExtractOption describes a monitored section for multi-section extraction.
type SectionExtractOption struct {
	ID            string
	Selector      string
	SelectorXPath string
	Offsets       *SelectorOffsets
}

// ExtractorResult holds the result of a page extraction.
type ExtractorResult struct {
	Title            string
	HTML             string
	Text             string
	ScreenshotBase64 string
	SelectorMatched  bool
	Sections         []SectionExtractResult
}

// SectionExtractResult holds the extraction result for a single monitored section.
type SectionExtractResult struct {
	ID               string
	ScreenshotBase64 string
	HTML             string
	Text             string
	SelectorMatched  bool
}

// ExtractorClient is the port for calling the scraper/extractor service.
// The concrete adapter wraps snapshot/infrastructure/extractor.HTTPClient.
type ExtractorClient interface {
	Extract(ctx context.Context, url string, opts ExtractOptions) (*ExtractorResult, error)
}
