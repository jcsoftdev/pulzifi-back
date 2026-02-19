package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
)

type IntegrationPostgresRepository struct {
	db     *sql.DB
	tenant string
}

func NewIntegrationPostgresRepository(db *sql.DB, tenant string) *IntegrationPostgresRepository {
	return &IntegrationPostgresRepository{db: db, tenant: tenant}
}

func (r *IntegrationPostgresRepository) Create(ctx context.Context, integration *entities.Integration) error {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return err
	}

	configJSON, err := json.Marshal(integration.Config)
	if err != nil {
		return err
	}

	q := `INSERT INTO integrations (id, service_type, config, enabled, created_by, created_at, updated_at)
	      VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = r.db.ExecContext(ctx, q,
		integration.ID, integration.ServiceType, configJSON,
		integration.Enabled, integration.CreatedBy,
		integration.CreatedAt, integration.UpdatedAt,
	)
	return err
}

func (r *IntegrationPostgresRepository) List(ctx context.Context) ([]*entities.Integration, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, err
	}

	q := `SELECT id, service_type, config, enabled, created_by, created_at, updated_at
	      FROM integrations WHERE deleted_at IS NULL ORDER BY created_at ASC`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var integrations []*entities.Integration
	for rows.Next() {
		var i entities.Integration
		var configJSON []byte
		if err := rows.Scan(&i.ID, &i.ServiceType, &configJSON, &i.Enabled, &i.CreatedBy, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(configJSON, &i.Config); err != nil {
			i.Config = map[string]interface{}{}
		}
		integrations = append(integrations, &i)
	}
	return integrations, rows.Err()
}

func (r *IntegrationPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Integration, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, err
	}

	var i entities.Integration
	var configJSON []byte
	q := `SELECT id, service_type, config, enabled, created_by, created_at, updated_at
	      FROM integrations WHERE id = $1 AND deleted_at IS NULL LIMIT 1`
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&i.ID, &i.ServiceType, &configJSON, &i.Enabled, &i.CreatedBy, &i.CreatedAt, &i.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(configJSON, &i.Config); err != nil {
		i.Config = map[string]interface{}{}
	}
	return &i, nil
}

func (r *IntegrationPostgresRepository) ListByServiceType(ctx context.Context, serviceType string) ([]*entities.Integration, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, err
	}

	q := `SELECT id, service_type, config, enabled, created_by, created_at, updated_at
	      FROM integrations WHERE service_type = $1 AND deleted_at IS NULL ORDER BY created_at ASC`
	rows, err := r.db.QueryContext(ctx, q, serviceType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var integrations []*entities.Integration
	for rows.Next() {
		var i entities.Integration
		var configJSON []byte
		if err := rows.Scan(&i.ID, &i.ServiceType, &configJSON, &i.Enabled, &i.CreatedBy, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(configJSON, &i.Config); err != nil {
			i.Config = map[string]interface{}{}
		}
		integrations = append(integrations, &i)
	}
	return integrations, rows.Err()
}

func (r *IntegrationPostgresRepository) GetByServiceType(ctx context.Context, serviceType string) (*entities.Integration, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, err
	}

	var i entities.Integration
	var configJSON []byte
	q := `SELECT id, service_type, config, enabled, created_by, created_at, updated_at
	      FROM integrations WHERE service_type = $1 AND deleted_at IS NULL LIMIT 1`
	err := r.db.QueryRowContext(ctx, q, serviceType).Scan(
		&i.ID, &i.ServiceType, &configJSON, &i.Enabled, &i.CreatedBy, &i.CreatedAt, &i.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(configJSON, &i.Config); err != nil {
		i.Config = map[string]interface{}{}
	}
	return &i, nil
}

func (r *IntegrationPostgresRepository) Update(ctx context.Context, integration *entities.Integration) error {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return err
	}

	configJSON, err := json.Marshal(integration.Config)
	if err != nil {
		return err
	}

	integration.UpdatedAt = time.Now()
	q := `UPDATE integrations SET config = $1, enabled = $2, updated_at = $3 WHERE id = $4`
	_, err = r.db.ExecContext(ctx, q, configJSON, integration.Enabled, integration.UpdatedAt, integration.ID)
	return err
}

func (r *IntegrationPostgresRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return err
	}

	now := time.Now()
	q := `UPDATE integrations SET deleted_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, q, now, id)
	return err
}
