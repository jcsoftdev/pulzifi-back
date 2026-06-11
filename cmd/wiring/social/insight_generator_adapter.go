package socialwiring

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	socialservices "github.com/jcsoftdev/pulzifi-back/modules/social/domain/services"
	socialpostgres "github.com/jcsoftdev/pulzifi-back/modules/social/infrastructure/persistence/postgres"
	"github.com/jcsoftdev/pulzifi-back/shared/ai"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// completer is the minimal OpenRouter surface this adapter needs.
type completer interface {
	Complete(ctx context.Context, messages []ai.Message) (string, error)
}

// insightGeneratorAdapter implements social services.SocialInsightGenerator using
// an OpenRouter completion. It persists the generated summary via the change repo.
type insightGeneratorAdapter struct {
	client completer
	db     *sql.DB
	tenant string
}

var _ socialservices.SocialInsightGenerator = (*insightGeneratorAdapter)(nil)

// NewInsightGenerator constructs a tenant-scoped SocialInsightGenerator.
func NewInsightGenerator(client completer, db *sql.DB, tenant string) socialservices.SocialInsightGenerator {
	return &insightGeneratorAdapter{client: client, db: db, tenant: tenant}
}

// Generate builds a concise AI summary and stores it on the change row.
func (a *insightGeneratorAdapter) Generate(ctx context.Context, changeID uuid.UUID, handle, platform string, summary entities.ChangeSummary) error {
	if a.client == nil {
		return nil
	}
	prompt := buildInsightPrompt(handle, platform, summary)
	out, err := a.client.Complete(ctx, []ai.Message{
		{Role: "system", Content: "You are a concise social-media monitoring analyst. Summarize the change in 1-2 sentences."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return fmt.Errorf("social insight: complete: %w", err)
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return nil
	}
	changeRepo := socialpostgres.NewChangePostgresRepository(a.db, a.tenant)
	if err := changeRepo.UpdateAISummary(ctx, changeID, out); err != nil {
		logger.Warn("social insight: persist summary failed",
			zap.String("change_id", changeID.String()), zap.Error(err))
		return err
	}
	return nil
}

// buildInsightPrompt renders the change into a prompt for the LLM.
func buildInsightPrompt(handle, platform string, s entities.ChangeSummary) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Profile @%s on %s changed.\n", handle, platform)
	if s.FollowersDelta != 0 {
		fmt.Fprintf(&b, "- Followers delta: %+d\n", s.FollowersDelta)
	}
	if len(s.NewPosts) > 0 {
		fmt.Fprintf(&b, "- New posts: %d\n", len(s.NewPosts))
	}
	if len(s.RemovedPosts) > 0 {
		fmt.Fprintf(&b, "- Removed posts: %d\n", len(s.RemovedPosts))
	}
	if s.Bio != nil {
		fmt.Fprintf(&b, "- Bio: %q -> %q\n", s.Bio.From, s.Bio.To)
	}
	if s.DisplayName != nil {
		fmt.Fprintf(&b, "- Display name: %q -> %q\n", s.DisplayName.From, s.DisplayName.To)
	}
	for _, c := range s.EditedCaptions {
		fmt.Fprintf(&b, "- Caption edited on %s: %q -> %q\n", c.ExternalID, c.From, c.To)
	}
	return b.String()
}
