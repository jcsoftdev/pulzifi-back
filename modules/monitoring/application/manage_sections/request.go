package managesections

// SectionOffsetsDTO represents pixel offsets for an element bounding box.
type SectionOffsetsDTO struct {
	Top    int `json:"top"`
	Right  int `json:"right"`
	Bottom int `json:"bottom"`
	Left   int `json:"left"`
}

// SectionRectDTO holds pixel coordinates of a section in the original viewport space.
type SectionRectDTO struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

// SectionDTO represents a single monitored section in requests.
type SectionDTO struct {
	ID              string             `json:"id,omitempty"`
	Name            string             `json:"name"`
	CSSSelector     string             `json:"css_selector"`
	XPathSelector   string             `json:"xpath_selector,omitempty"`
	SelectorOffsets *SectionOffsetsDTO `json:"selector_offsets,omitempty"`
	Rect            *SectionRectDTO    `json:"rect,omitempty"`
	ViewportWidth   int                `json:"viewport_width,omitempty"`
	SortOrder       int                `json:"sort_order"`
}

// SaveSectionsRequest replaces all sections for a page.
type SaveSectionsRequest struct {
	Sections []SectionDTO `json:"sections"`
}
