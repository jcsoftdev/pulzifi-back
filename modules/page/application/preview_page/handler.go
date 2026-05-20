package previewpage

import (
	"context"

	pageservices "github.com/jcsoftdev/pulzifi-back/modules/page/domain/services"
)

type PreviewPageHandler struct {
	previewer pageservices.PagePreviewer
}

func NewPreviewPageHandler(previewer pageservices.PagePreviewer) *PreviewPageHandler {
	return &PreviewPageHandler{previewer: previewer}
}

func (h *PreviewPageHandler) Handle(ctx context.Context, req *PreviewPageRequest) (*PreviewPageResponse, error) {
	result, err := h.previewer.Preview(ctx, req.URL, req.BlockAdsCookies)
	if err != nil {
		return nil, err
	}

	elements := make([]PreviewElement, len(result.Elements))
	for i, el := range result.Elements {
		elements[i] = PreviewElement{
			Selector:     el.Selector,
			XPath:        el.XPath,
			Tag:          el.Tag,
			Rect:         ElementRect{X: el.Rect.X, Y: el.Rect.Y, W: el.Rect.W, H: el.Rect.H},
			TextPreview:  el.TextPreview,
			SemanticRole: el.SemanticRole,
		}
	}

	return &PreviewPageResponse{
		ScreenshotBase64: result.ScreenshotBase64,
		Viewport:         PreviewViewport{Width: result.Viewport.Width, Height: result.Viewport.Height},
		PageHeight:       result.PageHeight,
		Elements:         elements,
	}, nil
}
