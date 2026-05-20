package services

import (
	"context"
	"io"
)

// PreviewResult holds the result of a page preview (screenshot + elements).
type PreviewResult struct {
	ScreenshotBase64 string
	Viewport         PreviewViewport
	PageHeight       int
	Elements         []PreviewElement
}

// PreviewViewport holds the viewport dimensions.
type PreviewViewport struct {
	Width  int
	Height int
}

// PreviewElement holds a single DOM element from the preview.
type PreviewElement struct {
	Selector     string
	XPath        string
	Tag          string
	Rect         ElementRect
	TextPreview  string
	SemanticRole string
}

// ElementRect holds the position and size of a DOM element.
type ElementRect struct {
	X int
	Y int
	W int
	H int
}

// PagePreviewer requests a page preview (screenshot + elements) from the extractor.
type PagePreviewer interface {
	Preview(ctx context.Context, url string, blockAdsCookies bool) (*PreviewResult, error)
}

// PagePreviewStreamer requests a streaming (SSE) page preview from the extractor.
// It returns an io.ReadCloser that the HTTP layer can pipe directly to the client.
type PagePreviewStreamer interface {
	PreviewStream(ctx context.Context, url string, blockAdsCookies bool) (io.ReadCloser, error)
}
