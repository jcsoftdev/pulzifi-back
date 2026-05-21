package createreport

import (
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/report/domain/entities"
)

// Request carries the data for creating a new report.
type Request struct {
	PageID     uuid.UUID
	Title      string
	ReportDate time.Time
	Content    entities.Content
	PDFURL     string
	CreatedBy  uuid.UUID
}
