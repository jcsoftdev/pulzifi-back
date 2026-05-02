package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/crypto"
)

type IntegrationPostgresRepository struct {
	db  *sql.DB
	enc *crypto.AESGCM
}

func NewIntegrationPostgresRepository(db *sql.DB, enc *crypto.AESGCM) repositories.IntegrationRepository {
	return &IntegrationPostgresRepository{db: db, enc: enc}
}

func (r *IntegrationPostgresRepository) Create(ctx context.Context, i *entities.Integration) error {
	accessEnc, err := r.enc.Encrypt([]byte(i.AccessToken))
	if err != nil {
		return fmt.Errorf("integration repo: encrypt access token: %w", err)
	}

	var refreshEnc []byte
	if i.RefreshToken != "" {
		refreshEnc, err = r.enc.Encrypt([]byte(i.RefreshToken))
		if err != nil {
			return fmt.Errorf("integration repo: encrypt refresh token: %w", err)
		}
	}

	metaJSON, err := json.Marshal(i.ProviderMeta)
	if err != nil {
		return fmt.Errorf("integration repo: marshal provider_meta: %w", err)
	}

	q := `INSERT INTO integrations
		(id, org_id, service_type, status, access_token_encrypted, refresh_token_encrypted,
		 token_expires_at, provider_meta, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())`

	_, err = r.db.ExecContext(ctx, q,
		i.ID, i.OrgID, i.ServiceType, string(i.Status),
		accessEnc, nullBytes(refreshEnc),
		nullTime(i.TokenExpiresAt),
		metaJSON, i.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("integration repo: create: %w", err)
	}
	return nil
}

func (r *IntegrationPostgresRepository) Update(ctx context.Context, i *entities.Integration) error {
	accessEnc, err := r.enc.Encrypt([]byte(i.AccessToken))
	if err != nil {
		return fmt.Errorf("integration repo: encrypt access token: %w", err)
	}

	var refreshEnc []byte
	if i.RefreshToken != "" {
		refreshEnc, err = r.enc.Encrypt([]byte(i.RefreshToken))
		if err != nil {
			return fmt.Errorf("integration repo: encrypt refresh token: %w", err)
		}
	}

	metaJSON, err := json.Marshal(i.ProviderMeta)
	if err != nil {
		return fmt.Errorf("integration repo: marshal provider_meta: %w", err)
	}

	q := `UPDATE integrations
		SET status = $1,
		    access_token_encrypted = $2,
		    refresh_token_encrypted = $3,
		    token_expires_at = $4,
		    provider_meta = $5,
		    updated_at = NOW()
		WHERE id = $6 AND deleted_at IS NULL`

	_, err = r.db.ExecContext(ctx, q,
		string(i.Status),
		accessEnc, nullBytes(refreshEnc),
		nullTime(i.TokenExpiresAt),
		metaJSON,
		i.ID,
	)
	if err != nil {
		return fmt.Errorf("integration repo: update: %w", err)
	}
	return nil
}

func (r *IntegrationPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Integration, error) {
	q := `SELECT id, org_id, service_type, status, access_token_encrypted, refresh_token_encrypted,
		         token_expires_at, provider_meta, created_by, created_at, updated_at, deleted_at
		  FROM integrations
		  WHERE id = $1 AND deleted_at IS NULL
		  LIMIT 1`

	row := r.db.QueryRowContext(ctx, q, id)
	i, err := r.scanRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("integration repo: get by id: %w", err)
	}
	return i, nil
}

func (r *IntegrationPostgresRepository) GetByOrgAndService(ctx context.Context, orgID uuid.UUID, serviceType string) (*entities.Integration, error) {
	q := `SELECT id, org_id, service_type, status, access_token_encrypted, refresh_token_encrypted,
		         token_expires_at, provider_meta, created_by, created_at, updated_at, deleted_at
		  FROM integrations
		  WHERE org_id = $1 AND service_type = $2 AND deleted_at IS NULL
		  LIMIT 1`

	row := r.db.QueryRowContext(ctx, q, orgID, serviceType)
	i, err := r.scanRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("integration repo: get by org and service: %w", err)
	}
	return i, nil
}

func (r *IntegrationPostgresRepository) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*entities.Integration, error) {
	q := `SELECT id, org_id, service_type, status, access_token_encrypted, refresh_token_encrypted,
		         token_expires_at, provider_meta, created_by, created_at, updated_at, deleted_at
		  FROM integrations
		  WHERE org_id = $1 AND deleted_at IS NULL
		  ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, q, orgID)
	if err != nil {
		return nil, fmt.Errorf("integration repo: list by org: %w", err)
	}
	defer rows.Close()

	var result []*entities.Integration
	for rows.Next() {
		i, err := r.scanRow(rows)
		if err != nil {
			return nil, fmt.Errorf("integration repo: list by org scan: %w", err)
		}
		result = append(result, i)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("integration repo: list by org rows: %w", err)
	}
	return result, nil
}

func (r *IntegrationPostgresRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	q := `UPDATE integrations
		SET deleted_at = NOW(), status = 'disconnected', updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("integration repo: soft delete: %w", err)
	}
	return nil
}

// scanner is a common interface satisfied by both *sql.Row and *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func (r *IntegrationPostgresRepository) scanRow(s scanner) (*entities.Integration, error) {
	var i entities.Integration
	var statusStr string
	var accessEnc, refreshEnc []byte
	var tokenExpiresAt sql.NullTime
	var deletedAt sql.NullTime
	var metaJSON []byte

	err := s.Scan(
		&i.ID, &i.OrgID, &i.ServiceType, &statusStr,
		&accessEnc, &refreshEnc,
		&tokenExpiresAt, &metaJSON,
		&i.CreatedBy, &i.CreatedAt, &i.UpdatedAt, &deletedAt,
	)
	if err != nil {
		return nil, err
	}

	i.Status = entities.IntegrationStatus(statusStr)

	if len(accessEnc) > 0 {
		plain, err := r.enc.Decrypt(accessEnc)
		if err != nil {
			return nil, fmt.Errorf("decrypt access token: %w", err)
		}
		i.AccessToken = string(plain)
	}

	if len(refreshEnc) > 0 {
		plain, err := r.enc.Decrypt(refreshEnc)
		if err != nil {
			return nil, fmt.Errorf("decrypt refresh token: %w", err)
		}
		i.RefreshToken = string(plain)
	}

	if tokenExpiresAt.Valid {
		t := tokenExpiresAt.Time
		i.TokenExpiresAt = &t
	}

	if deletedAt.Valid {
		t := deletedAt.Time
		i.DeletedAt = &t
	}

	i.ProviderMeta = map[string]any{}
	if len(metaJSON) > 0 {
		if err := json.Unmarshal(metaJSON, &i.ProviderMeta); err != nil {
			i.ProviderMeta = map[string]any{}
		}
	}

	return &i, nil
}

// nullBytes returns nil if b is empty, otherwise returns b (for BYTEA NULL handling).
func nullBytes(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return b
}

// nullTime returns nil if t is nil, otherwise returns *t (for TIMESTAMPTZ NULL handling).
func nullTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return *t
}
