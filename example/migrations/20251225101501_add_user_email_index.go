package migrations

import (
	"context"
	"database/sql"

	"github.com/alexisvisco/amigo"
)

type Migration20251225101501AddUserEmailIndex struct{}

func (m Migration20251225101501AddUserEmailIndex) Name() string {
	return "add_user_email_index"
}

func (m Migration20251225101501AddUserEmailIndex) Date() int64 {
	return 20251225101501
}

func (m Migration20251225101501AddUserEmailIndex) Up(ctx context.Context, db *sql.DB) error {
	return amigo.Tx(ctx, db, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `CREATE INDEX idx_users_created_at ON users(created_at)`)
		return err
	})
}

func (m Migration20251225101501AddUserEmailIndex) Down(ctx context.Context, db *sql.DB) error {
	return amigo.Tx(ctx, db, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `DROP INDEX idx_users_created_at`)
		return err
	})
}
