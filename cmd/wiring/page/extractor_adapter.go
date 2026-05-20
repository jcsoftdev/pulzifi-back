// Package pagewiring provides cross-module adapters for the page module.
// It bridges the snapshot extractor client without creating direct
// module-to-module imports in the page module.
package pagewiring

import (
	"context"
	"io"

	pageservices "github.com/jcsoftdev/pulzifi-back/modules/page/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/extractor"
)

// ExtractorPreviewerAdapter adapts extractor.HTTPClient to page's PagePreviewer interface.
type ExtractorPreviewerAdapter struct {
	client *extractor.HTTPClient
}

// NewExtractorPreviewerAdapter wraps an extractor HTTPClient as PagePreviewer.
func NewExtractorPreviewerAdapter(client *extractor.HTTPClient) pageservices.PagePreviewer {
	return &ExtractorPreviewerAdapter{client: client}
}

func (a *ExtractorPreviewerAdapter) Preview(ctx context.Context, url string, blockAdsCookies bool) (*pageservices.PreviewResult, error) {
	result, err := a.client.Preview(ctx, url, blockAdsCookies)
	if err != nil {
		return nil, err
	}

	elements := make([]pageservices.PreviewElement, len(result.Elements))
	for i, el := range result.Elements {
		elements[i] = pageservices.PreviewElement{
			Selector:     el.Selector,
			XPath:        el.XPath,
			Tag:          el.Tag,
			Rect:         pageservices.ElementRect{X: el.Rect.X, Y: el.Rect.Y, W: el.Rect.W, H: el.Rect.H},
			TextPreview:  el.TextPreview,
			SemanticRole: el.SemanticRole,
		}
	}

	return &pageservices.PreviewResult{
		ScreenshotBase64: result.ScreenshotBase64,
		Viewport:         pageservices.PreviewViewport{Width: result.Viewport.Width, Height: result.Viewport.Height},
		PageHeight:       result.PageHeight,
		Elements:         elements,
	}, nil
}

// ExtractorPreviewStreamerAdapter adapts extractor.HTTPClient to page's PagePreviewStreamer interface.
type ExtractorPreviewStreamerAdapter struct {
	client *extractor.HTTPClient
}

// NewExtractorPreviewStreamerAdapter wraps an extractor HTTPClient as PagePreviewStreamer.
func NewExtractorPreviewStreamerAdapter(client *extractor.HTTPClient) pageservices.PagePreviewStreamer {
	return &ExtractorPreviewStreamerAdapter{client: client}
}

func (a *ExtractorPreviewStreamerAdapter) PreviewStream(ctx context.Context, url string, blockAdsCookies bool) (io.ReadCloser, error) {
	resp, err := a.client.PreviewStream(ctx, url, blockAdsCookies)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
