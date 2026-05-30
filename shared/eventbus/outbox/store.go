package outbox

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Row is one row from public.outbox_events as returned by the relay.
type Row struct {
	ID             uuid.UUID
	Topic          string
	MsgKey         string
	Payload        []byte
	Tenant         string
	Status         string
	Attempts       int
	LastError      string
	NextAttemptAt  time.Time
	CreatedAt      time.Time
	PublishedAt    *time.Time
}

// Store is the data-access layer for public.outbox_events. It is a separate
// interface so the relay can be tested with a fake store without a real DB.
type Store interface {
	// Insert writes one event row inside an existing transaction.
	Insert(ctx context.Context, tx *sql.Tx, row Row) error

	// InsertStandalone writes one event row in its own short transaction.
	InsertStandalone(ctx context.Context, row Row) error

	// ResetStale moves rows stuck in status='publishing' beyond the stale
	// threshold back to status='pending' without incrementing attempts. This
	// self-heals crashed relay instances.
	ResetStale(ctx context.Context, threshold time.Duration) error

	// Claim atomically moves up to batchSize pending-eligible rows to
	// status='publishing' and returns them. Uses SKIP LOCKED so multiple relay
	// instances can coexist safely (future-proofing).
	Claim(ctx context.Context, batchSize int) ([]Row, error)

	// MarkPublished transitions a row to status='published'.
	MarkPublished(ctx context.Context, id uuid.UUID) error

	// MarkFailed increments attempts, records the error, calculates the next
	// retry time, and either sets status='pending' (retry) or status='dead'
	// (max attempts reached). nextAttemptAt and dead are pre-computed by the
	// relay using backoff logic so the store stays pure SQL.
	MarkFailed(ctx context.Context, id uuid.UUID, errMsg string, nextAttemptAt time.Time, dead bool) error
}

// postgresStore is the real PostgreSQL implementation of Store.
type postgresStore struct {
	db *sql.DB
}

// NewPostgresStore constructs a Store backed by the given *sql.DB.
func NewPostgresStore(db *sql.DB) Store {
	return &postgresStore{db: db}
}

func (s *postgresStore) Insert(ctx context.Context, tx *sql.Tx, row Row) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO public.outbox_events
		    (id, topic, msg_key, payload, tenant, status, attempts, last_error, next_attempt_at, created_at)
		VALUES ($1, $2, $3, $4, $5, 'pending', 0, '', now(), now())`,
		row.ID, row.Topic, row.MsgKey, row.Payload, row.Tenant,
	)
	return err
}

func (s *postgresStore) InsertStandalone(ctx context.Context, row Row) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO public.outbox_events
		    (id, topic, msg_key, payload, tenant, status, attempts, last_error, next_attempt_at, created_at)
		VALUES ($1, $2, $3, $4, $5, 'pending', 0, '', now(), now())`,
		row.ID, row.Topic, row.MsgKey, row.Payload, row.Tenant,
	)
	return err
}

func (s *postgresStore) ResetStale(ctx context.Context, threshold time.Duration) error {
	cutoff := time.Now().Add(-threshold)
	_, err := s.db.ExecContext(ctx, `
		UPDATE public.outbox_events
		SET    status = 'pending'
		WHERE  status = 'publishing'
		  AND  created_at < $1`,
		cutoff,
	)
	return err
}

func (s *postgresStore) Claim(ctx context.Context, batchSize int) ([]Row, error) {
	rows, err := s.db.QueryContext(ctx, `
		UPDATE public.outbox_events
		SET    status = 'publishing'
		WHERE  id IN (
		    SELECT id
		    FROM   public.outbox_events
		    WHERE  status = 'pending'
		      AND  next_attempt_at <= now()
		    ORDER  BY next_attempt_at
		    LIMIT  $1
		    FOR UPDATE SKIP LOCKED
		)
		RETURNING id, topic, msg_key, payload, tenant, status, attempts,
		          last_error, next_attempt_at, created_at, published_at`,
		batchSize,
	)
	if err != nil {
		return nil, fmt.Errorf("outbox: claim: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []Row
	for rows.Next() {
		var r Row
		if err := rows.Scan(
			&r.ID, &r.Topic, &r.MsgKey, &r.Payload, &r.Tenant, &r.Status,
			&r.Attempts, &r.LastError, &r.NextAttemptAt, &r.CreatedAt, &r.PublishedAt,
		); err != nil {
			return nil, fmt.Errorf("outbox: scan row: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *postgresStore) MarkPublished(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE public.outbox_events
		SET    status = 'published', published_at = now()
		WHERE  id = $1`,
		id,
	)
	return err
}

func (s *postgresStore) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string, nextAttemptAt time.Time, dead bool) error {
	newStatus := "pending"
	if dead {
		newStatus = "dead"
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE public.outbox_events
		SET    attempts        = attempts + 1,
		       last_error      = $2,
		       next_attempt_at = $3,
		       status          = $4
		WHERE  id = $1`,
		id, errMsg, nextAttemptAt, newStatus,
	)
	return err
}
