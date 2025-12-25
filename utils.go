package amigo

import (
	"context"
	"database/sql"
	"fmt"
)

func Tx(ctx context.Context, db *sql.DB, f func(*sql.Tx) error) (err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // re-panic after rollback
		}
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	err = f(tx)
	if err != nil {
		return fmt.Errorf("transaction function failed: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ChainExec provides a fluent interface for executing multiple SQL statements
// with automatic error accumulation. Errors are accumulated and returned at the end.
//
// Example usage:
//
//	func (m Migration) Up(ctx context.Context, db *sql.DB) error {
//	    return amigo.ChainExec(ctx, db).
//	        Exec(`CREATE TABLE users (id INT)`).
//	        Exec(`CREATE INDEX idx_users ON users(id)`).
//	        Err()
//	}
type ChainExec struct {
	db  *sql.DB
	ctx context.Context
	err error
}

// NewChainExec creates a new ChainExec for the given context and database connection
func NewChainExec(ctx context.Context, db *sql.DB) *ChainExec {
	return &ChainExec{
		db:  db,
		ctx: ctx,
	}
}

// Exec executes a SQL statement. If a previous error occurred, this is a no-op.
func (c *ChainExec) Exec(query string, args ...any) *ChainExec {
	if c.err != nil {
		return c
	}
	_, c.err = c.db.ExecContext(c.ctx, query, args...)
	return c
}

// Err returns the first error that occurred during the chain, or nil if no errors occurred
func (c *ChainExec) Err() error {
	return c.err
}

// ChainExecTx provides a fluent interface for executing multiple SQL statements
// within a transaction with automatic error accumulation.
//
// Example usage:
//
//	func (m Migration) Up(ctx context.Context, db *sql.DB) error {
//	    return amigo.Tx(ctx, db, func(tx *sql.Tx) error {
//	        return amigo.ChainExecTx(ctx, tx).
//	            Exec(`CREATE TABLE users (id INT)`).
//	            Exec(`CREATE INDEX idx_users ON users(id)`).
//	            Err()
//	    })
//	}
type ChainExecTx struct {
	tx  *sql.Tx
	ctx context.Context
	err error
}

// NewChainExecTx creates a new ChainExecTx for the given context and transaction
func NewChainExecTx(ctx context.Context, tx *sql.Tx) *ChainExecTx {
	return &ChainExecTx{
		tx:  tx,
		ctx: ctx,
	}
}

// Exec executes a SQL statement within the transaction. If a previous error occurred, this is a no-op.
func (c *ChainExecTx) Exec(query string, args ...any) *ChainExecTx {
	if c.err != nil {
		return c
	}
	_, c.err = c.tx.ExecContext(c.ctx, query, args...)
	return c
}

// Err returns the first error that occurred during the chain, or nil if no errors occurred
func (c *ChainExecTx) Err() error {
	return c.err
}
