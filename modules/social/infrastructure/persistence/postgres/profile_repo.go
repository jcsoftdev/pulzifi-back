package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/social/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories"
	valueobjects "github.com/jcsoftdev/pulzifi-back/modules/social/domain/value_objects"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
	"github.com/lib/pq"
)

// Compile-time interface check.
var _ repositories.ProfileRepository = (*ProfilePostgresRepository)(nil)

// ProfilePostgresRepository implements repositories.ProfileRepository using PostgreSQL.
// It is tenant-aware: the tenant schema is set via SET LOCAL search_path in each transaction.
type ProfilePostgresRepository struct {
	db     *sql.DB
	tenant string
}

// NewProfilePostgresRepository creates a new tenant-scoped ProfileRepository.
func NewProfilePostgresRepository(db *sql.DB, tenant string) *ProfilePostgresRepository {
	return &ProfilePostgresRepository{db: db, tenant: tenant}
}

// Create persists a new SocialProfile.
// Returns domainerrors.ErrProfileAlreadyExists on unique constraint violation.
func (r *ProfilePostgresRepository) Create(ctx context.Context, profile *entities.SocialProfile) error {
	q := `
		INSERT INTO social_profiles
			(id, workspace_id, platform, handle, display_name, avatar_url,
			 is_active, check_interval_minutes, next_check_at, last_checked_at,
			 consecutive_failures, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q,
			profile.ID,
			profile.WorkspaceID,
			string(profile.Platform),
			profile.Handle,
			profile.DisplayName,
			profile.AvatarURL,
			profile.IsActive,
			profile.CheckIntervalMinutes,
			profile.NextCheckAt,
			profile.LastCheckedAt,
			profile.ConsecutiveFailures,
			profile.CreatedAt,
			profile.UpdatedAt,
		)
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == "23505" {
				return domainerrors.ErrProfileAlreadyExists
			}
			return fmt.Errorf("profile repo: create: %w", err)
		}
		return nil
	})
}

// GetByID retrieves a profile by its UUID.
// Returns domainerrors.ErrProfileNotFound when no row matches.
func (r *ProfilePostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.SocialProfile, error) {
	q := `
		SELECT id, workspace_id, platform, handle, display_name, avatar_url,
		       is_active, check_interval_minutes, next_check_at, last_checked_at,
		       consecutive_failures, created_at, updated_at
		FROM social_profiles
		WHERE id = $1`

	var profile entities.SocialProfile
	var platform string

	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, q, id).Scan(
			&profile.ID,
			&profile.WorkspaceID,
			&platform,
			&profile.Handle,
			&profile.DisplayName,
			&profile.AvatarURL,
			&profile.IsActive,
			&profile.CheckIntervalMinutes,
			&profile.NextCheckAt,
			&profile.LastCheckedAt,
			&profile.ConsecutiveFailures,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domainerrors.ErrProfileNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("profile repo: get by id: %w", err)
	}
	profile.Platform = valueobjects.Platform(platform)
	return &profile, nil
}

// ListByWorkspace returns all profiles belonging to the given workspace.
func (r *ProfilePostgresRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entities.SocialProfile, error) {
	q := `
		SELECT id, workspace_id, platform, handle, display_name, avatar_url,
		       is_active, check_interval_minutes, next_check_at, last_checked_at,
		       consecutive_failures, created_at, updated_at
		FROM social_profiles
		WHERE workspace_id = $1
		ORDER BY created_at DESC`

	var profiles []*entities.SocialProfile

	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, q, workspaceID)
		if err != nil {
			return fmt.Errorf("profile repo: list by workspace: query: %w", err)
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var p entities.SocialProfile
			var platform string
			if err := rows.Scan(
				&p.ID,
				&p.WorkspaceID,
				&platform,
				&p.Handle,
				&p.DisplayName,
				&p.AvatarURL,
				&p.IsActive,
				&p.CheckIntervalMinutes,
				&p.NextCheckAt,
				&p.LastCheckedAt,
				&p.ConsecutiveFailures,
				&p.CreatedAt,
				&p.UpdatedAt,
			); err != nil {
				return fmt.Errorf("profile repo: list by workspace: scan: %w", err)
			}
			p.Platform = valueobjects.Platform(platform)
			profiles = append(profiles, &p)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return profiles, nil
}

// Update persists changes to an existing profile.
func (r *ProfilePostgresRepository) Update(ctx context.Context, profile *entities.SocialProfile) error {
	q := `
		UPDATE social_profiles
		SET display_name = $1,
		    avatar_url = $2,
		    is_active = $3,
		    check_interval_minutes = $4,
		    next_check_at = $5,
		    last_checked_at = $6,
		    consecutive_failures = $7,
		    updated_at = $8
		WHERE id = $9`

	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, q,
			profile.DisplayName,
			profile.AvatarURL,
			profile.IsActive,
			profile.CheckIntervalMinutes,
			profile.NextCheckAt,
			profile.LastCheckedAt,
			profile.ConsecutiveFailures,
			profile.UpdatedAt,
			profile.ID,
		)
		if err != nil {
			return fmt.Errorf("profile repo: update: %w", err)
		}
		n, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("profile repo: update rows affected: %w", err)
		}
		if n == 0 {
			return domainerrors.ErrProfileNotFound
		}
		return nil
	})
}

// Delete removes a profile. Associated snapshots and changes are removed by ON DELETE CASCADE.
func (r *ProfilePostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM social_profiles WHERE id = $1`

	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, q, id)
		if err != nil {
			return fmt.Errorf("profile repo: delete: %w", err)
		}
		n, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("profile repo: delete rows affected: %w", err)
		}
		if n == 0 {
			return domainerrors.ErrProfileNotFound
		}
		return nil
	})
}

// CountActiveByWorkspace returns the number of active profiles in a workspace.
func (r *ProfilePostgresRepository) CountActiveByWorkspace(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	q := `SELECT COUNT(*) FROM social_profiles WHERE workspace_id = $1 AND is_active = TRUE`

	var count int
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, q, workspaceID).Scan(&count)
	})
	if err != nil {
		return 0, fmt.Errorf("profile repo: count active: %w", err)
	}
	return count, nil
}

// claimWindow is how far into the future next_check_at is pushed during the
// atomic claim step. This prevents a second worker from re-selecting the same
// row before run_check sets the real next_check_at on completion (REQ-SCHED-02).
const claimWindow = 10 * time.Minute

// ListDue atomically claims and returns active profiles whose next_check_at is
// at or before now. The claim is implemented via a single UPDATE … WHERE id IN
// (SELECT … FOR UPDATE SKIP LOCKED) RETURNING statement so that the row lock,
// the next_check_at bump, and the row read happen in one implicit transaction —
// no separate BEGIN/COMMIT is needed. This ensures cross-instance safety even
// with multiple concurrent worker processes (REQ-SCHED-02).
func (r *ProfilePostgresRepository) ListDue(ctx context.Context, now time.Time, limit int) ([]*entities.SocialProfile, error) {
	qTenant := pq.QuoteIdentifier(r.tenant)

	// next_check_at is advanced by claimWindow so that concurrent workers using
	// SKIP LOCKED will skip these rows until run_check sets the real value.
	provisional := now.Add(claimWindow)

	q := fmt.Sprintf(`
		UPDATE %s.social_profiles
		SET    next_check_at = $3,
		       updated_at    = $1
		WHERE  id IN (
		    SELECT id
		    FROM   %s.social_profiles
		    WHERE  is_active      = TRUE
		      AND  next_check_at IS NOT NULL
		      AND  next_check_at <= $1
		    ORDER BY next_check_at ASC
		    LIMIT  $2
		    FOR UPDATE SKIP LOCKED
		)
		RETURNING id, workspace_id, platform, handle, display_name, avatar_url,
		          is_active, check_interval_minutes, next_check_at, last_checked_at,
		          consecutive_failures, created_at, updated_at`,
		qTenant, qTenant)

	rows, err := r.db.QueryContext(ctx, q, now, limit, provisional)
	if err != nil {
		return nil, fmt.Errorf("profile repo: list due: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var profiles []*entities.SocialProfile
	for rows.Next() {
		var p entities.SocialProfile
		var platform string
		if err := rows.Scan(
			&p.ID,
			&p.WorkspaceID,
			&platform,
			&p.Handle,
			&p.DisplayName,
			&p.AvatarURL,
			&p.IsActive,
			&p.CheckIntervalMinutes,
			&p.NextCheckAt,
			&p.LastCheckedAt,
			&p.ConsecutiveFailures,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("profile repo: list due: scan: %w", err)
		}
		p.Platform = valueobjects.Platform(platform)
		profiles = append(profiles, &p)
	}
	return profiles, rows.Err()
}
