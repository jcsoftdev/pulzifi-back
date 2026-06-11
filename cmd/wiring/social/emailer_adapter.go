package socialwiring

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	emailservices "github.com/jcsoftdev/pulzifi-back/modules/email/domain/services"
	socialservices "github.com/jcsoftdev/pulzifi-back/modules/social/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// emailerAdapter implements social services.SocialNotificationEmailer by resolving
// the workspace's active members and sending each a transactional email through
// the email module's EmailProvider. Recipients = active workspace members with
// email_notifications_enabled = true.
type emailerAdapter struct {
	db          *sql.DB
	provider    emailservices.EmailProvider
	tenant      string
	frontendURL string
}

var _ socialservices.SocialNotificationEmailer = (*emailerAdapter)(nil)

// NewEmailer constructs a tenant-scoped SocialNotificationEmailer.
func NewEmailer(db *sql.DB, provider emailservices.EmailProvider, tenant, frontendURL string) socialservices.SocialNotificationEmailer {
	return &emailerAdapter{db: db, provider: provider, tenant: tenant, frontendURL: frontendURL}
}

// SendChangeEmail resolves recipients and emails each one. Best-effort per recipient.
func (a *emailerAdapter) SendChangeEmail(ctx context.Context, workspaceID uuid.UUID, handle, platform, summary, dashboardURL string) error {
	if a.provider == nil {
		return nil
	}
	if dashboardURL == "" {
		dashboardURL = fmt.Sprintf("%s/workspaces", a.frontendURL)
	}

	emails, err := a.recipientEmails(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("social emailer: resolve recipients: %w", err)
	}
	if len(emails) == 0 {
		return nil
	}

	subject := fmt.Sprintf("@%s — social changes detected", handle)
	body := buildChangeEmailHTML(handle, platform, summary, dashboardURL)

	for _, to := range emails {
		if err := a.provider.Send(ctx, to, subject, body); err != nil {
			logger.Warn("social emailer: send failed",
				zap.String("to", to), zap.String("handle", handle), zap.Error(err))
		}
	}
	return nil
}

// recipientEmails returns active workspace members' emails (notifications enabled).
func (a *emailerAdapter) recipientEmails(ctx context.Context, workspaceID uuid.UUID) ([]string, error) {
	q := `
		SELECT u.email
		FROM workspace_members wm
		JOIN public.users u ON u.id = wm.user_id
		WHERE wm.workspace_id = $1
		  AND wm.removed_at IS NULL
		  AND u.deleted_at IS NULL
		  AND COALESCE(u.email_notifications_enabled, TRUE) = TRUE`

	var emails []string
	err := database.WithTenant(ctx, a.db, a.tenant, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, q, workspaceID)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var email string
			if err := rows.Scan(&email); err != nil {
				return err
			}
			if strings.TrimSpace(email) != "" {
				emails = append(emails, email)
			}
		}
		return rows.Err()
	})
	return emails, err
}

// buildChangeEmailHTML renders a minimal HTML body for the change notification.
func buildChangeEmailHTML(handle, platform, summary, dashboardURL string) string {
	return fmt.Sprintf(`<div style="font-family:sans-serif;max-width:480px">
<h2>Changes detected on @%s</h2>
<p>Platform: %s</p>
<p>%s</p>
<p><a href="%s">View in Pulzifi</a></p>
</div>`, handle, platform, summary, dashboardURL)
}
