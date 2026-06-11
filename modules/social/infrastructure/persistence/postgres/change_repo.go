package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories"
	valueobjects "github.com/jcsoftdev/pulzifi-back/modules/social/domain/value_objects"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
	"github.com/lib/pq"
)

// Compile-time interface check.
var _ repositories.ChangeRepository = (*ChangePostgresRepository)(nil)

// ChangePostgresRepository implements repositories.ChangeRepository using PostgreSQL.
type ChangePostgresRepository struct {
	db     *sql.DB
	tenant string
}

// NewChangePostgresRepository creates a new tenant-scoped ChangeRepository.
func NewChangePostgresRepository(db *sql.DB, tenant string) *ChangePostgresRepository {
	return &ChangePostgresRepository{db: db, tenant: tenant}
}

// Save persists a new SocialChange.
func (r *ChangePostgresRepository) Save(ctx context.Context, change *entities.SocialChange) error {
	// Serialize change_types as TEXT[]
	changeTypeStrs := make([]string, len(change.ChangeTypes))
	for i, ct := range change.ChangeTypes {
		changeTypeStrs[i] = string(ct)
	}

	summaryJSON, err := json.Marshal(change.Summary)
	if err != nil {
		return fmt.Errorf("change repo: marshal summary: %w", err)
	}

	q := `
		INSERT INTO social_changes
			(id, profile_id, from_snapshot_id, to_snapshot_id, change_types, summary, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q,
			change.ID,
			change.ProfileID,
			change.FromSnapshotID,
			change.ToSnapshotID,
			pq.Array(changeTypeStrs),
			summaryJSON,
			change.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("change repo: save: %w", err)
		}
		return nil
	})
}

// List returns changes matching the filter in descending created_at order.
// When filter.ProfileID is set, results are restricted to that profile.
// When filter.WorkspaceID is set (and ProfileID is nil), results span all profiles in that workspace.
func (r *ChangePostgresRepository) List(ctx context.Context, filter repositories.ChangeFilter) ([]*entities.SocialChange, error) {
	var (
		q    string
		args []interface{}
	)

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset

	switch {
	case filter.ProfileID != nil:
		q = `
			SELECT id, profile_id, from_snapshot_id, to_snapshot_id, change_types, summary, created_at, ai_summary
			FROM social_changes
			WHERE profile_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`
		args = []interface{}{*filter.ProfileID, limit, offset}
	case filter.WorkspaceID != nil:
		// Join via social_profiles to scope to workspace
		q = `
			SELECT sc.id, sc.profile_id, sc.from_snapshot_id, sc.to_snapshot_id,
			       sc.change_types, sc.summary, sc.created_at, sc.ai_summary
			FROM social_changes sc
			JOIN social_profiles sp ON sc.profile_id = sp.id
			WHERE sp.workspace_id = $1
			ORDER BY sc.created_at DESC
			LIMIT $2 OFFSET $3`
		args = []interface{}{*filter.WorkspaceID, limit, offset}
	default:
		q = `
			SELECT id, profile_id, from_snapshot_id, to_snapshot_id, change_types, summary, created_at, ai_summary
			FROM social_changes
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2`
		args = []interface{}{limit, offset}
	}

	var changes []*entities.SocialChange

	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, q, args...)
		if err != nil {
			return fmt.Errorf("change repo: list: query: %w", err)
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var c entities.SocialChange
			var changeTypeStrs []string
			var summaryJSON []byte
			var aiSummary sql.NullString

			if err := rows.Scan(
				&c.ID,
				&c.ProfileID,
				&c.FromSnapshotID,
				&c.ToSnapshotID,
				pq.Array(&changeTypeStrs),
				&summaryJSON,
				&c.CreatedAt,
				&aiSummary,
			); err != nil {
				return fmt.Errorf("change repo: list: scan: %w", err)
			}

			c.ChangeTypes = make([]valueobjects.ChangeType, len(changeTypeStrs))
			for i, s := range changeTypeStrs {
				c.ChangeTypes[i] = valueobjects.ChangeType(s)
			}

			if err := json.Unmarshal(summaryJSON, &c.Summary); err != nil {
				return fmt.Errorf("change repo: list: unmarshal summary: %w", err)
			}

			if aiSummary.Valid {
				v := aiSummary.String
				c.AISummary = &v
			}

			changes = append(changes, &c)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return changes, nil
}

// GetByID retrieves a change by its UUID. Returns nil, nil when not found.
func (r *ChangePostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.SocialChange, error) {
	q := `
		SELECT id, profile_id, from_snapshot_id, to_snapshot_id, change_types, summary, created_at, ai_summary
		FROM social_changes
		WHERE id = $1`

	var c entities.SocialChange
	var changeTypeStrs []string
	var summaryJSON []byte
	var aiSummary sql.NullString

	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, q, id).Scan(
			&c.ID,
			&c.ProfileID,
			&c.FromSnapshotID,
			&c.ToSnapshotID,
			pq.Array(&changeTypeStrs),
			&summaryJSON,
			&c.CreatedAt,
			&aiSummary,
		)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("change repo: get by id: %w", err)
	}

	c.ChangeTypes = make([]valueobjects.ChangeType, len(changeTypeStrs))
	for i, s := range changeTypeStrs {
		c.ChangeTypes[i] = valueobjects.ChangeType(s)
	}

	if err := json.Unmarshal(summaryJSON, &c.Summary); err != nil {
		return nil, fmt.Errorf("change repo: get by id: unmarshal summary: %w", err)
	}

	if aiSummary.Valid {
		v := aiSummary.String
		c.AISummary = &v
	}

	return &c, nil
}

// UpdateAISummary sets the ai_summary column for a change.
func (r *ChangePostgresRepository) UpdateAISummary(ctx context.Context, changeID uuid.UUID, summary string) error {
	q := `UPDATE social_changes SET ai_summary = $1 WHERE id = $2`
	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q, summary, changeID)
		if err != nil {
			return fmt.Errorf("change repo: update ai_summary: %w", err)
		}
		return nil
	})
}
