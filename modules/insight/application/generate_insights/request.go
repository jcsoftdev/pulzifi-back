package generateinsights

import "github.com/google/uuid"

// Request holds the input data for the GenerateInsights use case.
type Request struct {
	PageID              uuid.UUID
	CheckID             uuid.UUID
	PageURL             string
	PrevText            string
	NewText             string
	SchemaName          string
	EnabledInsightTypes []string // subset to generate; defaults to ["marketing","market_analysis"]
}
