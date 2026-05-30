package persistence

import (
	"context"
	"database/sql"

	"github.com/jcsoftdev/pulzifi-back/modules/dashboard/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/dashboard/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
)

type DashboardPostgresRepository struct {
	db     *sql.DB
	tenant string
}

func NewDashboardPostgresRepository(db *sql.DB, tenant string) repositories.DashboardRepository {
	return &DashboardPostgresRepository{db: db, tenant: tenant}
}

func (r *DashboardPostgresRepository) GetStats(ctx context.Context) (*entities.DashboardStats, error) {
	stats := &entities.DashboardStats{
		ChangesPerWorkspace: []entities.WorkspaceChanges{},
		RecentAlerts:        []entities.RecentAlert{},
		RecentInsights:      []entities.RecentInsight{},
	}

	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM workspaces WHERE deleted_at IS NULL`,
		).Scan(&stats.WorkspacesCount); err != nil {
			return err
		}

		if err := tx.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM pages WHERE deleted_at IS NULL`,
		).Scan(&stats.PagesCount); err != nil {
			return err
		}

		if err := tx.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM checks WHERE DATE(checked_at) = CURRENT_DATE`,
		).Scan(&stats.TodayChecksCount); err != nil {
			return err
		}

		changesRows, err := tx.QueryContext(ctx, `
			SELECT w.name, COUNT(c.id) FILTER (WHERE c.change_detected = TRUE) AS changes_count
			FROM workspaces w
			LEFT JOIN pages p ON p.workspace_id = w.id AND p.deleted_at IS NULL
			LEFT JOIN checks c ON c.page_id = p.id
			WHERE w.deleted_at IS NULL
			GROUP BY w.id, w.name
			ORDER BY changes_count DESC
		`)
		if err != nil {
			return err
		}
		defer func() { _ = changesRows.Close() }()

		for changesRows.Next() {
			var wc entities.WorkspaceChanges
			if err := changesRows.Scan(&wc.WorkspaceName, &wc.DetectedChanges); err != nil {
				return err
			}
			stats.ChangesPerWorkspace = append(stats.ChangesPerWorkspace, wc)
		}

		alertRows, err := tx.QueryContext(ctx, `
			SELECT c.checked_at, w.name, COALESCE(c.change_type, ''), p.url
			FROM checks c
			JOIN pages p ON c.page_id = p.id AND p.deleted_at IS NULL
			JOIN workspaces w ON p.workspace_id = w.id AND w.deleted_at IS NULL
			WHERE c.change_detected = TRUE
			ORDER BY c.checked_at DESC
			LIMIT 10
		`)
		if err != nil {
			return err
		}
		defer func() { _ = alertRows.Close() }()

		for alertRows.Next() {
			var a entities.RecentAlert
			if err := alertRows.Scan(&a.CheckedAt, &a.WorkspaceName, &a.ChangeType, &a.PageURL); err != nil {
				return err
			}
			stats.RecentAlerts = append(stats.RecentAlerts, a)
		}

		insightRows, err := tx.QueryContext(ctx, `
			SELECT i.created_at, w.name, p.url, COALESCE(i.title, ''), COALESCE(i.content, '')
			FROM insights i
			JOIN pages p ON i.page_id = p.id AND p.deleted_at IS NULL
			JOIN workspaces w ON p.workspace_id = w.id AND w.deleted_at IS NULL
			WHERE i.deleted_at IS NULL
			ORDER BY i.created_at DESC
			LIMIT 5
		`)
		if err != nil {
			return err
		}
		defer func() { _ = insightRows.Close() }()

		for insightRows.Next() {
			var ins entities.RecentInsight
			if err := insightRows.Scan(&ins.CreatedAt, &ins.WorkspaceName, &ins.PageURL, &ins.Title, &ins.Content); err != nil {
				return err
			}
			stats.RecentInsights = append(stats.RecentInsights, ins)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return stats, nil
}
