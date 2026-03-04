package previewpage

type PreviewPageResponse struct {
	ScreenshotBase64 string           `json:"screenshot_base64"`
	Viewport         PreviewViewport  `json:"viewport"`
	PageHeight       int              `json:"page_height"`
	Elements         []PreviewElement `json:"elements"`
}

type PreviewViewport struct {
	Width  int `json:"width"`
	Height int `json:"height"`
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
