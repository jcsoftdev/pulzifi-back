package repositories

import "context"

// TxBeginner abstracts the ability to begin a database transaction.
// This allows application handlers to start transactions without importing database/sql.
type TxBeginner interface {
	// Begin starts a new transaction and returns an opaque handle.
	// The returned handle must be passed to OrganizationPlanRepository.WithTx.
	Begin(ctx context.Context) (interface{}, error)

	// Commit commits the transaction returned by Begin.
	Commit(tx interface{}) error

	// Rollback rolls back the transaction. Safe to call after Commit.
	Rollback(tx interface{})
}
