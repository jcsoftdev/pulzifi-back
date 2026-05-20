package entities

import (
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
