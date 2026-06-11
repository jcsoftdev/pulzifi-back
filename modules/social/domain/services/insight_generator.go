package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
)

// SocialInsightGenerator builds an AI summary from a detected change and persists
// it (via the change repository's UpdateAISummary). Best-effort: implementations
// log and swallow errors so run_check is never blocked.
type SocialInsightGenerator interface {
	Generate(ctx context.Context, changeID uuid.UUID, handle, platform string, summary entities.ChangeSummary) error
}
