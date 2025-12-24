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
