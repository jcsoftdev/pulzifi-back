package entities

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Alert struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	PageID      uuid.UUID
	CheckID     uuid.UUID
	Type        string
	Title       string
	Description string
	Metadata    Metadata
	ReadAt      *time.Time
	CreatedAt   time.Time
}

// Metadata is flexible JSON data
type Metadata map[string]interface{}

func (m Metadata) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *Metadata) Scan(value interface{}) error {
	if value == nil {
		*m = make(Metadata)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, &m)
}

func NewAlert(workspaceID, pageID, checkID uuid.UUID, alertType, title, description string) *Alert {
	return &Alert{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		PageID:      pageID,
		CheckID:     checkID,
		Type:        alertType,
		Title:       title,
		Description: description,
		CreatedAt:   time.Now(),
	}
}
