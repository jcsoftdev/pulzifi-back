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
	"github.com/jcsoftdev/pulzifi-back/shared/database"
)

// Compile-time interface check.
var _ repositories.SnapshotRepository = (*SnapshotPostgresRepository)(nil)

// SnapshotPostgresRepository implements repositories.SnapshotRepository using PostgreSQL.
type SnapshotPostgresRepository struct {
	db     *sql.DB
	tenant string
}

// NewSnapshotPostgresRepository creates a new tenant-scoped SnapshotRepository.
func NewSnapshotPostgresRepository(db *sql.DB, tenant string) *SnapshotPostgresRepository {
	return &SnapshotPostgresRepository{db: db, tenant: tenant}
}

// Save persists a new SocialSnapshot (success or failed).
func (r *SnapshotPostgresRepository) Save(ctx context.Context, snapshot *entities.SocialSnapshot) error {
	var dataJSON []byte
	if snapshot.Data != nil {
		var err error
		dataJSON, err = json.Marshal(snapshot.Data)
		if err != nil {
			return fmt.Errorf("snapshot repo: marshal data: %w", err)
		}
	}

	q := `
		INSERT INTO social_snapshots
			(id, profile_id, captured_at, status, error, followers_count, posts_count, data)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q,
			snapshot.ID,
			snapshot.ProfileID,
			snapshot.CapturedAt,
			string(snapshot.Status),
			snapshot.Error,
			snapshot.FollowersCount,
			snapshot.PostsCount,
			dataJSON,
		)
		if err != nil {
			return fmt.Errorf("snapshot repo: save: %w", err)
		}
		return nil
	})
}

// GetLatestByProfile returns the most recent snapshot for a profile.
// Returns nil, nil when no snapshot exists yet (first-run / baseline case).
func (r *SnapshotPostgresRepository) GetLatestByProfile(ctx context.Context, profileID uuid.UUID) (*entities.SocialSnapshot, error) {
	q := `
		SELECT id, profile_id, captured_at, status, error, followers_count, posts_count, data
		FROM social_snapshots
		WHERE profile_id = $1
		ORDER BY captured_at DESC
		LIMIT 1`

	var snapshot entities.SocialSnapshot
	var status string
	var dataJSON []byte
	var errStr sql.NullString

	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, q, profileID).Scan(
			&snapshot.ID,
			&snapshot.ProfileID,
			&snapshot.CapturedAt,
			&status,
			&errStr,
			&snapshot.FollowersCount,
			&snapshot.PostsCount,
			&dataJSON,
		)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("snapshot repo: get latest by profile: %w", err)
	}

	snapshot.Status = entities.SnapshotStatus(status)
	snapshot.Error = errStr.String

	if len(dataJSON) > 0 {
		var data entities.ProfileData
		if err := json.Unmarshal(dataJSON, &data); err != nil {
			return nil, fmt.Errorf("snapshot repo: unmarshal data: %w", err)
		}
		snapshot.Data = &data
	}

	return &snapshot, nil
}

// GetByID retrieves a snapshot by its UUID. Returns nil, nil when not found.
func (r *SnapshotPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.SocialSnapshot, error) {
	q := `
		SELECT id, profile_id, captured_at, status, error, followers_count, posts_count, data
		FROM social_snapshots
		WHERE id = $1`

	var snapshot entities.SocialSnapshot
	var status string
	var dataJSON []byte
	var errStr sql.NullString

	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, q, id).Scan(
			&snapshot.ID,
			&snapshot.ProfileID,
			&snapshot.CapturedAt,
			&status,
			&errStr,
			&snapshot.FollowersCount,
			&snapshot.PostsCount,
			&dataJSON,
		)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("snapshot repo: get by id: %w", err)
	}

	snapshot.Status = entities.SnapshotStatus(status)
	snapshot.Error = errStr.String

	if len(dataJSON) > 0 {
		var data entities.ProfileData
		if err := json.Unmarshal(dataJSON, &data); err != nil {
			return nil, fmt.Errorf("snapshot repo: unmarshal data: %w", err)
		}
		snapshot.Data = &data
	}

	return &snapshot, nil
}
