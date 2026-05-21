package getreport

import "github.com/google/uuid"

// Request carries the report ID to fetch.
type Request struct {
	ReportID uuid.UUID
}
