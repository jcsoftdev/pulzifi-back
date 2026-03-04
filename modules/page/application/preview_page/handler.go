package previewpage

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/extractor"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type PreviewPageHandler struct {
	extractorClient *extractor.HTTPClient
}

func NewPreviewPageHandler(extractorClient *extractor.HTTPClient) *PreviewPageHandler {
	return &PreviewPageHandler{extractorClient: extractorClient}
}

func (h *PreviewPageHandler) Handle(ctx context.Context, req *PreviewPageRequest) (*PreviewPageResponse, error) {
	result, err := h.extractorClient.Preview(ctx, req.URL, req.BlockAdsCookies)
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

func (h *PreviewPageHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	var req PreviewPageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	resp, err := h.Handle(r.Context(), &req)
	if err != nil {
		logger.Error("Failed to preview page", zap.Error(err))
		http.Error(w, "Failed to preview page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
