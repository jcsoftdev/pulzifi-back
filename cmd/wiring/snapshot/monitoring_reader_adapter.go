package snapshotwiring

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	monPersistence "github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/persistence"
	snapEntities "github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/entities"
	snapServices "github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
)

// monitoringReaderAdapter implements snapshot's MonitoringReader port by wrapping
// monitoring infrastructure repositories and direct SQL queries.
type monitoringReaderAdapter struct {
	db *sql.DB
}

// NewMonitoringReader builds a MonitoringReader adapter backed by the shared DB pool.
func NewMonitoringReader(db *sql.DB) snapServices.MonitoringReader {
	return &monitoringReaderAdapter{db: db}
}

func (a *monitoringReaderAdapter) GetConfigByPageID(ctx context.Context, schemaName string, pageID uuid.UUID) (*snapEntities.MonitoringConfig, error) {
	repo := monPersistence.NewMonitoringConfigPostgresRepository(a.db, schemaName)
	mc, err := repo.GetByPageID(ctx, pageID)
	if err != nil || mc == nil {
		return nil, err
	}
	cfg := &snapEntities.MonitoringConfig{
		BlockAdsCookies:        mc.BlockAdsCookies,
		EnabledInsightTypes:    mc.EnabledInsightTypes,
		EnabledAlertConditions: mc.EnabledAlertConditions,
		SelectorType:           mc.SelectorType,
		CSSSelector:            mc.CSSSelector,
		XPathSelector:          mc.XPathSelector,
		IgnoreSelectors:        mc.IgnoreSelectors,
	}
	if mc.SelectorOffsets != nil {
		cfg.SelectorOffsets = &snapEntities.SelectorOffsets{
			Top:    mc.SelectorOffsets.Top,
			Right:  mc.SelectorOffsets.Right,
			Bottom: mc.SelectorOffsets.Bottom,
			Left:   mc.SelectorOffsets.Left,
		}
	}
	return cfg, nil
}

func (a *monitoringReaderAdapter) ListSectionsByPageID(ctx context.Context, schemaName string, pageID uuid.UUID) ([]*snapEntities.MonitoredSection, error) {
	repo := monPersistence.NewMonitoredSectionPostgresRepository(a.db, schemaName)
	sections, err := repo.ListByPageID(ctx, pageID)
	if err != nil {
		return nil, err
	}
	result := make([]*snapEntities.MonitoredSection, len(sections))
	for i, s := range sections {
		sec := &snapEntities.MonitoredSection{
			ID:            s.ID,
			CSSSelector:   s.CSSSelector,
			XPathSelector: s.XPathSelector,
		}
		if s.SelectorOffsets != nil {
			sec.SelectorOffsets = &snapEntities.SelectorOffsets{
				Top:    s.SelectorOffsets.Top,
				Right:  s.SelectorOffsets.Right,
				Bottom: s.SelectorOffsets.Bottom,
				Left:   s.SelectorOffsets.Left,
			}
		}
		result[i] = sec
	}
	return result, nil
}

func (a *monitoringReaderAdapter) GetEmailEnabledByPage(ctx context.Context, schemaName string, pageID uuid.UUID) ([]*snapEntities.NotificationPreference, error) {
	repo := monPersistence.NewNotificationPreferencePostgresRepository(a.db, schemaName)
	prefs, err := repo.GetEmailEnabledByPage(ctx, pageID)
	if err != nil {
		return nil, err
	}
	result := make([]*snapEntities.NotificationPreference, len(prefs))
	for i, p := range prefs {
		result[i] = &snapEntities.NotificationPreference{
			UserID:      p.UserID,
			ChangeTypes: p.ChangeTypes,
		}
	}
	return result, nil
}

func (a *monitoringReaderAdapter) GetUserEmail(ctx context.Context, userID uuid.UUID) (string, error) {
	var email string
	if err := a.db.QueryRowContext(ctx, `SELECT email FROM public.users WHERE id = $1`, userID).Scan(&email); err != nil {
		return "", fmt.Errorf("get user email: %w", err)
	}
	return email, nil
}

func (a *monitoringReaderAdapter) GetWorkspaceIDForPage(ctx context.Context, schemaName string, pageID uuid.UUID) (uuid.UUID, error) {
	if _, err := a.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(schemaName)); err != nil {
		return uuid.Nil, err
	}
	var workspaceID uuid.UUID
	if err := a.db.QueryRowContext(ctx, `SELECT workspace_id FROM pages WHERE id = $1`, pageID).Scan(&workspaceID); err != nil {
		return uuid.Nil, fmt.Errorf("get workspace_id for page %s: %w", pageID, err)
	}
	return workspaceID, nil
}

func (a *monitoringReaderAdapter) GetPageTitle(ctx context.Context, schemaName string, pageID uuid.UUID) (string, error) {
	if _, err := a.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(schemaName)); err != nil {
		return "", err
	}
	var title string
	if err := a.db.QueryRowContext(ctx, `SELECT COALESCE(title,'') FROM pages WHERE id = $1`, pageID).Scan(&title); err != nil {
		return "", fmt.Errorf("get page title for page %s: %w", pageID, err)
	}
	return title, nil
}

func (a *monitoringReaderAdapter) GetOrgIDBySchemaName(ctx context.Context, schemaName string) (uuid.UUID, error) {
	var orgID uuid.UUID
	if err := a.db.QueryRowContext(ctx,
		`SELECT id FROM public.organizations WHERE schema_name = $1 AND deleted_at IS NULL`,
		schemaName,
	).Scan(&orgID); err != nil {
		return uuid.Nil, fmt.Errorf("get org_id for schema %s: %w", schemaName, err)
	}
	return orgID, nil
}

func (a *monitoringReaderAdapter) UpdatePageSnapshotMetadata(ctx context.Context, schemaName string, pageID uuid.UUID, thumbnailURL string, changeDetected bool) error {
	if _, err := a.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(schemaName)); err != nil {
		return err
	}
	q := `UPDATE pages
		SET thumbnail_url = $1,
			last_change_detected_at = CASE WHEN $2 THEN NOW() ELSE last_change_detected_at END
		WHERE id = $3`
	_, err := a.db.ExecContext(ctx, q, thumbnailURL, changeDetected, pageID)
	return err
}

