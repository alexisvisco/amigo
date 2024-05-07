package schema

import (
	"context"
	"database/sql"
	"net/url"
	"strings"
)

// DB is the interface that describes a database connection.
type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// DatabaseCredentials is the struct that holds the database credentials.
type DatabaseCredentials struct {
	Host, Port, User, Pass, DB string
}

// ExtractCredentials extracts the database credentials from the DSN.
func ExtractCredentials(dsn string) (DatabaseCredentials, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return DatabaseCredentials{}, err
	}

	pass, _ := u.User.Password()

	return DatabaseCredentials{
		Host: u.Hostname(),
		Port: u.Port(),
		User: u.User.Username(),
		Pass: pass,
		DB:   strings.TrimLeft(u.Path, "/"),
	}, nil
}
