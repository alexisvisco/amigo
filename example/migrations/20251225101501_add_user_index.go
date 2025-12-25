package migrations

import (
	"context"
	"database/sql"

	"github.com/alexisvisco/amigo"
)

type Migration20251225101501AddUserIndex struct{}

func (m Migration20251225101501AddUserIndex) Name() string {
	return "add_user_index"
}

func (m Migration20251225101501AddUserIndex) Date() int64 {
	return 20251225101501
}

func (m Migration20251225101501AddUserIndex) Up(ctx context.Context, db *sql.DB) error {
	return amigo.Tx(ctx, db, func(tx *sql.Tx) error {
		return amigo.NewChainExecTx(ctx, tx).
			Exec(`CREATE INDEX idx_users_created_at ON users(created_at)`).
			Err()
	})
}

func (m Migration20251225101501AddUserIndex) Down(ctx context.Context, db *sql.DB) error {
	return amigo.Tx(ctx, db, func(tx *sql.Tx) error {
		return amigo.NewChainExecTx(ctx, tx).
			Exec(`DROP INDEX idx_users_created_at`).
			Err()
	})
}
