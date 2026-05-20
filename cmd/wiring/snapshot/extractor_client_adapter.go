package snapshotwiring

import (
	"context"

	snapinfraextractor "github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/extractor"
	snapServices "github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/services"
)

// extractorClientAdapter implements snapshot's ExtractorClient port by
// wrapping snapshot/infrastructure/extractor.HTTPClient.
type extractorClientAdapter struct {
	inner *snapinfraextractor.HTTPClient
}

// NewExtractorClient builds an ExtractorClient adapter backed by the given HTTPClient.
func NewExtractorClient(inner *snapinfraextractor.HTTPClient) snapServices.ExtractorClient {
	return &extractorClientAdapter{inner: inner}
}

func (a *extractorClientAdapter) Extract(ctx context.Context, url string, opts snapServices.ExtractOptions) (*snapServices.ExtractorResult, error) {
	infraOpts := snapinfraextractor.ExtractOptions{
		BlockAdsCookies: opts.BlockAdsCookies,
		Selector:        opts.Selector,
		SelectorXPath:   opts.SelectorXPath,
		IgnoreSelectors: opts.IgnoreSelectors,
	}
	if opts.SelectorOffsets != nil {
		infraOpts.SelectorOffsets = &snapinfraextractor.SelectorOffsets{
			Top:    opts.SelectorOffsets.Top,
			Right:  opts.SelectorOffsets.Right,
			Bottom: opts.SelectorOffsets.Bottom,
			Left:   opts.SelectorOffsets.Left,
		}
	}
	for _, s := range opts.Sections {
		opt := snapinfraextractor.SectionExtractOption{
			ID:            s.ID,
			Selector:      s.Selector,
			SelectorXPath: s.SelectorXPath,
		}
		if s.Offsets != nil {
			opt.Offsets = &snapinfraextractor.SelectorOffsets{
				Top:    s.Offsets.Top,
				Right:  s.Offsets.Right,
				Bottom: s.Offsets.Bottom,
				Left:   s.Offsets.Left,
			}
		}
		infraOpts.Sections = append(infraOpts.Sections, opt)
	}

	res, err := a.inner.Extract(ctx, url, infraOpts)
	if err != nil {
		return nil, err
	}

	result := &snapServices.ExtractorResult{
		Title:            res.Title,
		HTML:             res.HTML,
		Text:             res.Text,
		ScreenshotBase64: res.ScreenshotBase64,
		SelectorMatched:  res.SelectorMatched,
	}
	for _, s := range res.Sections {
		result.Sections = append(result.Sections, snapServices.SectionExtractResult{
			ID:               s.ID,
			ScreenshotBase64: s.ScreenshotBase64,
			HTML:             s.HTML,
			Text:             s.Text,
			SelectorMatched:  s.SelectorMatched,
		})
	}
	return result, nil
}
