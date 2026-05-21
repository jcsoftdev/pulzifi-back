// Package jobs hosts background workers that run inside cmd/worker.
//
// trial_expirer is a small daily cron: every interval it scans for
// organization_plans rows with code='trial', trial_ends_at in the past, and
// converted_at NULL. For each such row it flips every member's user.status
// to 'trial_expired' (idempotent: skips users already flipped) and emits one
// trial-expired email per member.
package jobs

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	authentities "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	emailservices "github.com/jcsoftdev/pulzifi-back/modules/email/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/email/infrastructure/templates"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// TrialExpirer flips expired trials to status='trial_expired' and sends the
// trial-expired email. It is safe to run multiple times per day — the update
// only matches users that are NOT already trial_expired.
type TrialExpirer struct {
	db          *sql.DB
	emailProv   emailservices.EmailProvider
	frontendURL string
	interval    time.Duration
}

// NewTrialExpirer constructs the job.
func NewTrialExpirer(db *sql.DB, emailProv emailservices.EmailProvider, frontendURL string, interval time.Duration) *TrialExpirer {
	if interval <= 0 {
		interval = 1 * time.Hour
	}
	return &TrialExpirer{
		db:          db,
		emailProv:   emailProv,
		frontendURL: frontendURL,
		interval:    interval,
	}
}

// Run loops on the configured interval until ctx is cancelled. It runs one
// tick immediately so an operator can see whether the wiring works.
func (e *TrialExpirer) Run(ctx context.Context) error {
	logger.Info("Starting trial expirer", zap.Duration("interval", e.interval))
	if err := e.runOnce(ctx); err != nil {
		logger.Warn("trial expirer first tick errored", zap.Error(err))
	}
	t := time.NewTicker(e.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			logger.Info("Trial expirer stopping")
			return ctx.Err()
		case <-t.C:
			if err := e.runOnce(ctx); err != nil {
				logger.Warn("trial expirer tick errored", zap.Error(err))
			}
		}
	}
}

// expiredMember is one user member of an expired trial org.
type expiredMember struct {
	UserID     uuid.UUID
	OrgID      uuid.UUID
	OrgSubdomain string
	Email      string
	FirstName  string
	UserStatus string
}

// runOnce performs one expiry sweep. Exported as a method for tests.
func (e *TrialExpirer) runOnce(ctx context.Context) error {
	members, err := e.findExpiredMembers(ctx)
	if err != nil {
		return err
	}
	for _, m := range members {
		if err := e.flipUser(ctx, m.UserID); err != nil {
			logger.Warn("trial expirer: failed to flip user",
				zap.String("user_id", m.UserID.String()),
				zap.Error(err),
			)
			continue
		}
		e.sendExpiredEmail(ctx, m)
	}
	if len(members) > 0 {
		logger.Info("trial expirer processed expired members", zap.Int("count", len(members)))
	}
	return nil
}

func (e *TrialExpirer) findExpiredMembers(ctx context.Context) ([]expiredMember, error) {
	rows, err := e.db.QueryContext(ctx, `
		SELECT u.id, o.id, o.subdomain, u.email, u.first_name, u.status
		  FROM public.organization_plans op
		  JOIN public.plans p ON p.id = op.plan_id AND p.code = 'trial'
		  JOIN public.organizations o ON o.id = op.organization_id
		  JOIN public.organization_members om ON om.organization_id = o.id
		  JOIN public.users u ON u.id = om.user_id
		 WHERE op.status = 'active'
		   AND op.deleted_at IS NULL
		   AND op.trial_ends_at < NOW()
		   AND op.converted_at IS NULL
		   AND u.status <> $1
		   AND u.deleted_at IS NULL
	`, authentities.UserStatusTrialExpired)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []expiredMember
	for rows.Next() {
		var m expiredMember
		if err := rows.Scan(&m.UserID, &m.OrgID, &m.OrgSubdomain, &m.Email, &m.FirstName, &m.UserStatus); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (e *TrialExpirer) flipUser(ctx context.Context, userID uuid.UUID) error {
	_, err := e.db.ExecContext(ctx, `
		UPDATE public.users
		   SET status = $2, updated_at = NOW()
		 WHERE id = $1 AND status <> $2
	`, userID, authentities.UserStatusTrialExpired)
	return err
}

func (e *TrialExpirer) sendExpiredEmail(ctx context.Context, m expiredMember) {
	if e.emailProv == nil {
		return
	}
	upgradeURL := e.frontendURL
	if upgradeURL == "" {
		upgradeURL = "https://" + m.OrgSubdomain + ".pulzifi.com/billing"
	}
	subject, html := templates.TrialExpired(m.FirstName, upgradeURL+"/billing")
	if err := e.emailProv.Send(ctx, m.Email, subject, html); err != nil {
		logger.Warn("trial expirer: failed to send trial_expired email",
			zap.String("to", m.Email),
			zap.Error(err),
		)
	}
}
