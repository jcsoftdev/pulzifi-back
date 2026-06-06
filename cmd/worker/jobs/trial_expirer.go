// Package jobs hosts background workers that run inside cmd/worker.
//
// trial_expirer is a small daily cron: every interval it scans for
// organization_plans rows with code='trial', trial_ends_at in the past, and
// converted_at NULL. For each such org it DOWNGRADES the org to the canonical
// 'free' plan (deactivates the trial row, inserts an active free row) and emits
// one trial-expired upsell email per member. Downgrading to free — instead of
// flipping users to a locked 'trial_expired' status — keeps the org working at
// the free tier (limited pages/insights, email-only alerts) rather than walling
// it off. The trial_guard middleware keys on plan='trial', so once the org is on
// free it stops returning 402 on writes automatically.
package jobs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	emailservices "github.com/jcsoftdev/pulzifi-back/modules/email/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/email/infrastructure/templates"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// TrialExpirer downgrades expired trials to the free plan and sends the
// trial-expired email. It is safe to run multiple times per day — once an org
// is moved off the trial plan it no longer matches the expiry sweep.
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
	UserID       uuid.UUID
	OrgID        uuid.UUID
	OrgSubdomain string
	Email        string
	FirstName    string
}

// runOnce performs one expiry sweep: it groups expired-trial members by org,
// downgrades each org to free exactly once, then emails that org's members.
// If a downgrade fails the org's emails are skipped so the next tick retries.
func (e *TrialExpirer) runOnce(ctx context.Context) error {
	members, err := e.findExpiredMembers(ctx)
	if err != nil {
		return err
	}

	byOrg := make(map[uuid.UUID][]expiredMember)
	var orgOrder []uuid.UUID
	for _, m := range members {
		if _, seen := byOrg[m.OrgID]; !seen {
			orgOrder = append(orgOrder, m.OrgID)
		}
		byOrg[m.OrgID] = append(byOrg[m.OrgID], m)
	}

	downgraded := 0
	for _, orgID := range orgOrder {
		if err := e.downgradeOrgToFree(ctx, orgID); err != nil {
			logger.Warn("trial expirer: failed to downgrade org to free",
				zap.String("org_id", orgID.String()),
				zap.Error(err),
			)
			continue
		}
		downgraded++
		for _, m := range byOrg[orgID] {
			e.sendExpiredEmail(ctx, m)
		}
	}

	if downgraded > 0 {
		logger.Info("trial expirer downgraded expired orgs to free",
			zap.Int("orgs", downgraded),
		)
	}
	return nil
}

func (e *TrialExpirer) findExpiredMembers(ctx context.Context) ([]expiredMember, error) {
	rows, err := e.db.QueryContext(ctx, `
		SELECT u.id, o.id, o.subdomain, u.email, u.first_name
		  FROM public.organization_plans op
		  JOIN public.plans p ON p.id = op.plan_id AND p.code = 'trial'
		  JOIN public.organizations o ON o.id = op.organization_id
		  JOIN public.organization_members om ON om.organization_id = o.id
		  JOIN public.users u ON u.id = om.user_id
		 WHERE op.status = 'active'
		   AND op.deleted_at IS NULL
		   AND op.trial_ends_at < NOW()
		   AND op.converted_at IS NULL
		   AND u.deleted_at IS NULL
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []expiredMember
	for rows.Next() {
		var m expiredMember
		if err := rows.Scan(&m.UserID, &m.OrgID, &m.OrgSubdomain, &m.Email, &m.FirstName); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// downgradeOrgToFree deactivates the org's active (trial) plan row and inserts a
// fresh active row pointing at the canonical 'free' plan, in one transaction.
// Mirrors the deactivate-then-insert pattern used by the billing PlanAssigner.
func (e *TrialExpirer) downgradeOrgToFree(ctx context.Context, orgID uuid.UUID) error {
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var freePlanID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		SELECT id FROM public.plans WHERE code = 'free' AND is_active = TRUE LIMIT 1
	`).Scan(&freePlanID)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("free plan not seeded (migration 000027 pending?)")
	}
	if err != nil {
		return fmt.Errorf("resolve free plan: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `
		UPDATE public.organization_plans
		   SET status = 'inactive', ended_at = NOW(), updated_at = NOW()
		 WHERE organization_id = $1
		   AND status = 'active'
		   AND deleted_at IS NULL
	`, orgID); err != nil {
		return fmt.Errorf("deactivate trial plan: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO public.organization_plans
		       (organization_id, plan_id, status, started_at, created_at, updated_at)
		VALUES ($1, $2, 'active', NOW(), NOW(), NOW())
	`, orgID, freePlanID); err != nil {
		return fmt.Errorf("insert free plan: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	committed = true
	return nil
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
