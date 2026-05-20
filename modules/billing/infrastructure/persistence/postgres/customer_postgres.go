package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/repositories"
)

// Compile-time interface check.
var _ repositories.CustomerRepository = (*CustomerPostgresRepository)(nil)

// CustomerPostgresRepository implements CustomerRepository using the public.organizations table.
// The stripe_customer_id column is part of public.organizations (added in migration 000018).
type CustomerPostgresRepository struct {
	db *sql.DB
}

// NewCustomerPostgresRepository returns a new CustomerPostgresRepository.
func NewCustomerPostgresRepository(db *sql.DB) *CustomerPostgresRepository {
	return &CustomerPostgresRepository{db: db}
}

// FindByOrgID returns the customer mapping for the given org, or nil if no Stripe customer exists yet.
func (r *CustomerPostgresRepository) FindByOrgID(ctx context.Context, orgID uuid.UUID) (*entities.Customer, error) {
	const query = `
		SELECT id, stripe_customer_id
		FROM public.organizations
		WHERE id = $1 AND stripe_customer_id IS NOT NULL
		LIMIT 1
	`
	var (
		id         uuid.UUID
		customerID string
	)
	err := r.db.QueryRowContext(ctx, query, orgID).Scan(&id, &customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &entities.Customer{OrgID: id, StripeCustomerID: customerID}, nil
}

// FindByStripeCustomerID looks up by Stripe customer identifier.
func (r *CustomerPostgresRepository) FindByStripeCustomerID(ctx context.Context, customerID string) (*entities.Customer, error) {
	const query = `
		SELECT id, stripe_customer_id
		FROM public.organizations
		WHERE stripe_customer_id = $1
		LIMIT 1
	`
	var (
		id     uuid.UUID
		custID string
	)
	err := r.db.QueryRowContext(ctx, query, customerID).Scan(&id, &custID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &entities.Customer{OrgID: id, StripeCustomerID: custID}, nil
}

// Save persists the Stripe customer ID on the organization row.
// This is an UPDATE rather than an INSERT because the organization must already exist.
func (r *CustomerPostgresRepository) Save(ctx context.Context, customer *entities.Customer) error {
	const query = `
		UPDATE public.organizations
		SET stripe_customer_id = $2
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, customer.OrgID, customer.StripeCustomerID)
	return err
}
