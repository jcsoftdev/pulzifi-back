package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
	"github.com/lib/pq"
)

// Compile-time interface check.
var _ repositories.DestinationRepository = (*DestinationPostgresRepository)(nil)

// DestinationPostgresRepository implements DestinationRepository using PostgreSQL (tenant schema).
type DestinationPostgresRepository struct {
	db     *sql.DB
	tenant string
}

// NewDestinationPostgresRepository creates a new tenant-scoped DestinationRepository.
func NewDestinationPostgresRepository(db *sql.DB, tenant string) repositories.DestinationRepository {
	return &DestinationPostgresRepository{db: db, tenant: tenant}
}

// Create inserts a new destination. created_at / updated_at are server-side NOW().
func (r *DestinationPostgresRepository) Create(ctx context.Context, d *entities.Destination) error {
	targetJSON, err := json.Marshal(d.Target)
	if err != nil {
		return fmt.Errorf("destination repo: marshal target: %w", err)
	}

	q := `INSERT INTO integration_destinations
		(id, integration_id, service_type, scope_type, scope_id, target, events, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())`

	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q,
			d.ID,
			nullUUID(d.IntegrationID),
			d.ServiceType,
			string(d.ScopeType),
			d.ScopeID,
			targetJSON,
			pq.Array(d.Events),
			d.Enabled,
		)
		if err != nil {
			return fmt.Errorf("destination repo: create: %w", err)
		}
		return nil
	})
}

// Update modifies an existing destination. updated_at is server-side NOW().
func (r *DestinationPostgresRepository) Update(ctx context.Context, d *entities.Destination) error {
	targetJSON, err := json.Marshal(d.Target)
	if err != nil {
		return fmt.Errorf("destination repo: marshal target: %w", err)
	}

	q := `UPDATE integration_destinations
		SET integration_id = $1,
		    service_type   = $2,
		    scope_type     = $3,
		    scope_id       = $4,
		    target         = $5,
		    events         = $6,
		    enabled        = $7,
		    updated_at     = NOW()
		WHERE id = $8`

	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q,
			nullUUID(d.IntegrationID),
			d.ServiceType,
			string(d.ScopeType),
			d.ScopeID,
			targetJSON,
			pq.Array(d.Events),
			d.Enabled,
			d.ID,
		)
		if err != nil {
			return fmt.Errorf("destination repo: update: %w", err)
		}
		return nil
	})
}

// GetByID fetches a destination by primary key. Returns nil, nil when not found.
func (r *DestinationPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Destination, error) {
	q := `SELECT id, integration_id, service_type, scope_type, scope_id,
		         target, events, enabled, created_at, updated_at
		  FROM integration_destinations
		  WHERE id = $1
		  LIMIT 1`

	var result *entities.Destination
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		row := tx.QueryRowContext(ctx, q, id)
		d, err := r.scanRow(row)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("destination repo: get by id: %w", err)
		}
		result = d
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Delete hard-deletes a destination by primary key.
func (r *DestinationPostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM integration_destinations WHERE id = $1`
	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q, id)
		if err != nil {
			return fmt.Errorf("destination repo: delete: %w", err)
		}
		return nil
	})
}

// ListByScope returns all destinations (enabled and disabled) for a given scope.
// For workspace and page scopes the workspace_name / page_name fields are populated
// via LEFT JOINs so callers can display human-readable scope labels without an
// extra round-trip.
func (r *DestinationPostgresRepository) ListByScope(ctx context.Context, scope entities.ScopeType, scopeID uuid.UUID) ([]*entities.Destination, error) {
	q := `SELECT d.id, d.integration_id, d.service_type, d.scope_type, d.scope_id,
		         d.target, d.events, d.enabled, d.created_at, d.updated_at,
		         COALESCE(w.name, '') AS workspace_name,
		         COALESCE(p.name, '') AS page_name
		  FROM integration_destinations d
		  LEFT JOIN workspaces w ON d.scope_type = 'workspace' AND w.id = d.scope_id
		  LEFT JOIN pages      p ON d.scope_type = 'page'      AND p.id = d.scope_id
		  WHERE d.scope_type = $1 AND d.scope_id = $2
		  ORDER BY d.created_at ASC`

	var result []*entities.Destination
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, q, string(scope), scopeID)
		if err != nil {
			return fmt.Errorf("destination repo: list by scope: %w", err)
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			d, err := r.scanRowWithNames(rows)
			if err != nil {
				return fmt.Errorf("destination repo: list by scope scan: %w", err)
			}
			result = append(result, d)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("destination repo: list by scope rows: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ResolveForEvent returns the winning destinations for a given event using
// scope-override semantics: page > workspace > org, per service_type.
// Only enabled destinations are considered.
// workspaceID and pageID may be nil when the event has no workspace/page context.
func (r *DestinationPostgresRepository) ResolveForEvent(
	ctx context.Context,
	eventType string,
	orgID uuid.UUID,
	workspaceID, pageID *uuid.UUID,
) ([]*entities.Destination, error) {
	// Build the scope filter and args dynamically based on which IDs are provided.
	// Always include org; optionally include workspace and page.
	scopeClauses := []string{`(scope_type = 'org' AND scope_id = $2)`}
	args := []any{eventType, orgID}
	argIdx := 3

	if workspaceID != nil {
		scopeClauses = append(scopeClauses, fmt.Sprintf(`(scope_type = 'workspace' AND scope_id = $%d)`, argIdx))
		args = append(args, *workspaceID)
		argIdx++
	}
	if pageID != nil {
		scopeClauses = append(scopeClauses, fmt.Sprintf(`(scope_type = 'page' AND scope_id = $%d)`, argIdx))
		args = append(args, *pageID)
	}

	scopeFilter := strings.Join(scopeClauses, " OR ")

	q := fmt.Sprintf(`
WITH candidates AS (
  SELECT *
  FROM integration_destinations
  WHERE enabled
    AND $1 = ANY(events)
    AND (%s)
),
ranked AS (
  SELECT *,
         CASE scope_type
           WHEN 'page'      THEN 3
           WHEN 'workspace' THEN 2
           ELSE                  1
         END AS specificity
  FROM candidates
)
SELECT r1.id, r1.integration_id, r1.service_type, r1.scope_type, r1.scope_id,
       r1.target, r1.events, r1.enabled, r1.created_at, r1.updated_at
FROM ranked r1
WHERE NOT EXISTS (
  SELECT 1 FROM ranked r2
  WHERE r2.service_type = r1.service_type
    AND r2.specificity > r1.specificity
)`, scopeFilter)

	var result []*entities.Destination
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, q, args...)
		if err != nil {
			return fmt.Errorf("destination repo: resolve for event: %w", err)
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			d, err := r.scanRow(rows)
			if err != nil {
				return fmt.Errorf("destination repo: resolve for event scan: %w", err)
			}
			result = append(result, d)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("destination repo: resolve for event rows: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// DisableByIntegrationID sets enabled = FALSE for all destinations linked to an integration.
func (r *DestinationPostgresRepository) DisableByIntegrationID(ctx context.Context, integrationID uuid.UUID) error {
	q := `UPDATE integration_destinations
		SET enabled = FALSE, updated_at = NOW()
		WHERE integration_id = $1`

	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q, integrationID)
		if err != nil {
			return fmt.Errorf("destination repo: disable by integration id: %w", err)
		}
		return nil
	})
}

// --- internal helpers ---

// scanner is already declared in integration_postgres.go; both files share the same package.
// We reuse it here via the scanRow method.

func (r *DestinationPostgresRepository) scanRow(s scanner) (*entities.Destination, error) {
	var d entities.Destination
	var integrationID uuid.NullUUID
	var scopeTypeStr string
	var targetJSON []byte

	err := s.Scan(
		&d.ID,
		&integrationID,
		&d.ServiceType,
		&scopeTypeStr,
		&d.ScopeID,
		&targetJSON,
		pq.Array(&d.Events),
		&d.Enabled,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if integrationID.Valid {
		id := integrationID.UUID
		d.IntegrationID = &id
	}

	d.ScopeType = entities.ScopeType(scopeTypeStr)

	d.Target = map[string]any{}
	if len(targetJSON) > 0 {
		if err := json.Unmarshal(targetJSON, &d.Target); err != nil {
			d.Target = map[string]any{}
		}
	}

	return &d, nil
}

// scanRowWithNames is like scanRow but also reads the two extra columns
// (workspace_name, page_name) produced by the ListByScope JOIN query.
func (r *DestinationPostgresRepository) scanRowWithNames(s scanner) (*entities.Destination, error) {
	var d entities.Destination
	var integrationID uuid.NullUUID
	var scopeTypeStr string
	var targetJSON []byte

	err := s.Scan(
		&d.ID,
		&integrationID,
		&d.ServiceType,
		&scopeTypeStr,
		&d.ScopeID,
		&targetJSON,
		pq.Array(&d.Events),
		&d.Enabled,
		&d.CreatedAt,
		&d.UpdatedAt,
		&d.WorkspaceName,
		&d.PageName,
	)
	if err != nil {
		return nil, err
	}

	if integrationID.Valid {
		id := integrationID.UUID
		d.IntegrationID = &id
	}

	d.ScopeType = entities.ScopeType(scopeTypeStr)

	d.Target = map[string]any{}
	if len(targetJSON) > 0 {
		if err := json.Unmarshal(targetJSON, &d.Target); err != nil {
			d.Target = map[string]any{}
		}
	}

	return &d, nil
}

// nullUUID converts a *uuid.UUID to a value suitable for a nullable UUID column.
func nullUUID(id *uuid.UUID) any {
	if id == nil {
		return nil
	}
	return *id
}
