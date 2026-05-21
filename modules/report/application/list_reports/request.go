package listreports

import "github.com/google/uuid"

// Request carries optional filter parameters.
type Request struct {
	// PageID filters reports to a specific page when set.
	PageID *uuid.UUID
}
