package schema

import "database/sql"

// Schema is the interface that need to be implemented to support migrations.
type Schema interface {
	AddVersion(version string)
	RemoveVersion(version string)
	FindAppliedVersions() []string

	Exec(query string, args ...interface{})
	Query(query string, args []any, rowHandler func(row *sql.Rows) error)
}
