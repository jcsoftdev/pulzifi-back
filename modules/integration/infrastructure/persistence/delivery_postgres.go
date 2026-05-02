package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
)

// Compile-time interface check.
var _ repositories.DeliveryRepository = (*DeliveryPostgresRepository)(nil)

// DeliveryPostgresRepository implements DeliveryRepository using PostgreSQL (tenant schema).
type DeliveryPostgresRepository struct {
	db     *sql.DB
	tenant string
}

// NewDeliveryPostgresRepository creates a new tenant-scoped DeliveryRepository.
func NewDeliveryPostgresRepository(db *sql.DB, tenant string) repositories.DeliveryRepository {
	return &DeliveryPostgresRepository{db: db, tenant: tenant}
}

// Create inserts a new delivery with status='pending' and next_attempt_at=NOW().
// The entity's Status/Attempts/NextAttemptAt/CreatedAt are ignored; server defaults apply.
func (r *DeliveryPostgresRepository) Create(ctx context.Context, d *entities.Delivery) error {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return fmt.Errorf("delivery repo: set search path: %w", err)
	}

	payloadJSON, err := json.Marshal(d.EventPayload)
	if err != nil {
		return fmt.Errorf("delivery repo: marshal event payload: %w", err)
	}

	q := `INSERT INTO integration_deliveries
		(id, destination_id, event_type, event_payload, status, attempts, next_attempt_at, created_at)
		VALUES ($1, $2, $3, $4, 'pending', 0, NOW(), NOW())`

	_, err = r.db.ExecContext(ctx, q,
		d.ID,
		d.DestinationID,
		d.EventType,
		payloadJSON,
	)
	if err != nil {
		return fmt.Errorf("delivery repo: create: %w", err)
	}
	return nil
}

// ClaimPending atomically claims up to `limit` pending deliveries whose
// next_attempt_at <= now, setting their status to 'in_flight'.
// Uses FOR UPDATE SKIP LOCKED for parallel-worker safety.
func (r *DeliveryPostgresRepository) ClaimPending(ctx context.Context, limit int, now time.Time) ([]*entities.Delivery, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, fmt.Errorf("delivery repo: set search path: %w", err)
	}

	q := `UPDATE integration_deliveries
	   SET status = 'in_flight',
	       last_attempt_at = $2,
	       attempts = attempts + 1
	 WHERE id IN (
	     SELECT id FROM integration_deliveries
	      WHERE status = 'pending' AND next_attempt_at <= $2
	      ORDER BY next_attempt_at
	      LIMIT $1
	      FOR UPDATE SKIP LOCKED
	 )
	RETURNING id, destination_id, event_type, event_payload, status, attempts,
	          last_attempt_at, next_attempt_at, response_code, response_body,
	          error_message, delivered_at, created_at`

	rows, err := r.db.QueryContext(ctx, q, limit, now)
	if err != nil {
		return nil, fmt.Errorf("delivery repo: claim pending: %w", err)
	}
	defer rows.Close()

	var result []*entities.Delivery
	for rows.Next() {
		d, err := scanDelivery(rows)
		if err != nil {
			return nil, fmt.Errorf("delivery repo: claim pending scan: %w", err)
		}
		result = append(result, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("delivery repo: claim pending rows: %w", err)
	}
	return result, nil
}

// MarkDelivered transitions a delivery to 'delivered' and records the response.
func (r *DeliveryPostgresRepository) MarkDelivered(ctx context.Context, id uuid.UUID, code int, bodySnip string) error {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return fmt.Errorf("delivery repo: set search path: %w", err)
	}

	q := `UPDATE integration_deliveries
	   SET status = 'delivered', response_code = $2, response_body = $3,
	       delivered_at = NOW(), error_message = NULL
	 WHERE id = $1`

	_, err := r.db.ExecContext(ctx, q, id, code, bodySnip)
	if err != nil {
		return fmt.Errorf("delivery repo: mark delivered: %w", err)
	}
	return nil
}

// MarkFailed transitions a delivery back to 'pending' for retry with a new next_attempt_at.
func (r *DeliveryPostgresRepository) MarkFailed(ctx context.Context, id uuid.UUID, code *int, body, errMsg string, nextAttempt time.Time) error {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return fmt.Errorf("delivery repo: set search path: %w", err)
	}

	var nullCode sql.NullInt32
	if code != nil {
		nullCode = sql.NullInt32{Int32: int32(*code), Valid: true}
	}

	q := `UPDATE integration_deliveries
	   SET status = 'pending', response_code = $2, response_body = $3,
	       error_message = $4, next_attempt_at = $5
	 WHERE id = $1`

	_, err := r.db.ExecContext(ctx, q, id, nullCode, body, errMsg, nextAttempt)
	if err != nil {
		return fmt.Errorf("delivery repo: mark failed: %w", err)
	}
	return nil
}

// MarkDead transitions a delivery to 'dead' — no further retries.
func (r *DeliveryPostgresRepository) MarkDead(ctx context.Context, id uuid.UUID, errMsg string) error {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return fmt.Errorf("delivery repo: set search path: %w", err)
	}

	q := `UPDATE integration_deliveries
	   SET status = 'dead', error_message = $2
	 WHERE id = $1`

	_, err := r.db.ExecContext(ctx, q, id, errMsg)
	if err != nil {
		return fmt.Errorf("delivery repo: mark dead: %w", err)
	}
	return nil
}

// ListByDestination returns deliveries for a given destination, newest first.
func (r *DeliveryPostgresRepository) ListByDestination(ctx context.Context, destinationID uuid.UUID, limit, offset int) ([]*entities.Delivery, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, fmt.Errorf("delivery repo: set search path: %w", err)
	}

	q := `SELECT id, destination_id, event_type, event_payload, status, attempts,
	             last_attempt_at, next_attempt_at, response_code, response_body,
	             error_message, delivered_at, created_at
	      FROM integration_deliveries
	      WHERE destination_id = $1
	      ORDER BY created_at DESC
	      LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, q, destinationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("delivery repo: list by destination: %w", err)
	}
	defer rows.Close()

	var result []*entities.Delivery
	for rows.Next() {
		d, err := scanDelivery(rows)
		if err != nil {
			return nil, fmt.Errorf("delivery repo: list by destination scan: %w", err)
		}
		result = append(result, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("delivery repo: list by destination rows: %w", err)
	}
	return result, nil
}

// --- internal helpers ---

// scanDelivery reads one row from a scanner into a Delivery entity.
func scanDelivery(s scanner) (*entities.Delivery, error) {
	var d entities.Delivery
	var payloadJSON []byte
	var lastAttemptAt sql.NullTime
	var nextAttemptAt time.Time
	var responseCode sql.NullInt32
	var responseBody sql.NullString
	var errorMessage sql.NullString
	var deliveredAt sql.NullTime

	err := s.Scan(
		&d.ID,
		&d.DestinationID,
		&d.EventType,
		&payloadJSON,
		&d.Status,
		&d.Attempts,
		&lastAttemptAt,
		&nextAttemptAt,
		&responseCode,
		&responseBody,
		&errorMessage,
		&deliveredAt,
		&d.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	d.EventPayload = map[string]any{}
	if len(payloadJSON) > 0 {
		if err := json.Unmarshal(payloadJSON, &d.EventPayload); err != nil {
			d.EventPayload = map[string]any{}
		}
	}

	if lastAttemptAt.Valid {
		t := lastAttemptAt.Time
		d.LastAttemptAt = &t
	}
	d.NextAttemptAt = &nextAttemptAt

	if responseCode.Valid {
		v := int(responseCode.Int32)
		d.ResponseCode = &v
	}
	if responseBody.Valid {
		v := responseBody.String
		d.ResponseBody = &v
	}
	if errorMessage.Valid {
		v := errorMessage.String
		d.ErrorMessage = &v
	}
	if deliveredAt.Valid {
		t := deliveredAt.Time
		d.DeliveredAt = &t
	}

	return &d, nil
}
