package previewpage

type PreviewPageRequest struct {
	URL             string `json:"url"`
	BlockAdsCookies bool   `json:"block_ads_cookies"`
}
