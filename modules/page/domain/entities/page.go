package entities

import (
"time"
"github.com/google/uuid"
)

// Page represents a page to monitor
type Page struct {
ID                    uuid.UUID
WorkspaceID           uuid.UUID
Name                  string
URL                   string
ThumbnailURL          string
LastCheckedAt         *time.Time
LastChangeDetectedAt  *time.Time
CheckCount            int
CreatedBy             uuid.UUID
CreatedAt             time.Time
UpdatedAt             time.Time
DeletedAt             *time.Time
}

// NewPage creates a new page
func NewPage(workspaceID uuid.UUID, name, url string, createdBy uuid.UUID) *Page {
return &Page{
ID:          uuid.New(),
WorkspaceID: workspaceID,
Name:        name,
URL:         url,
CheckCount:  0,
CreatedBy:   createdBy,
CreatedAt:   time.Now(),
UpdatedAt:   time.Now(),
}
}
