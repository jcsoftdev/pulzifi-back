package persistence

import (
	"context"
	"database/sql"
)

// SQLTxBeginner wraps *sql.DB to implement repositories.TxBeginner.
type SQLTxBeginner struct {
	db *sql.DB
}

// NewSQLTxBeginner creates a new SQLTxBeginner.
func NewSQLTxBeginner(db *sql.DB) *SQLTxBeginner {
	return &SQLTxBeginner{db: db}
}

// Begin starts a new transaction. Returns *sql.Tx as interface{}.
func (b *SQLTxBeginner) Begin(ctx context.Context) (interface{}, error) {
	return b.db.BeginTx(ctx, nil)
}

// Commit commits a *sql.Tx received from Begin.
func (b *SQLTxBeginner) Commit(tx interface{}) error {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok || sqlTx == nil {
		return nil
	}
	return sqlTx.Commit()
}

// Rollback rolls back a *sql.Tx. Safe to call after Commit.
func (b *SQLTxBeginner) Rollback(tx interface{}) {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok || sqlTx == nil {
		return
	}
	sqlTx.Rollback() //nolint:errcheck
}
