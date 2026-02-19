package entities

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Report struct {
	ID         uuid.UUID  `json:"id"`
	PageID     uuid.UUID  `json:"page_id"`
	Title      string     `json:"title"`
	ReportDate time.Time  `json:"report_date"`
	Content    Content    `json:"content"`
	PDFURL     string     `json:"pdf_url,omitempty"`
	CreatedBy  uuid.UUID  `json:"created_by"`
	CreatedAt  time.Time  `json:"created_at"`
	DeletedAt  *time.Time `json:"-"`
}

type Content map[string]interface{}

func (c Content) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

func (c *Content) Scan(value interface{}) error {
	if value == nil {
		*c = make(Content)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}
