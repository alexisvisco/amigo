package amigo

import (
	"reflect"
	"testing"
)

func Test_parseSQLMigration(t *testing.T) {
	config := Configuration{
		SQLFileUpAnnotation:   "-- +migrate Up",
		SQLFileDownAnnotation: "-- +migrate Down",
	}

	tests := []struct {
		name    string
		content string
		want    SQLMigration
	}{
		{
			name: "basic up and down",
			content: `-- +migrate Up
CREATE TABLE users (id INT);
INSERT INTO users VALUES (1);
-- +migrate Down
DROP TABLE users;`,
			want: SQLMigration{
				up:     "CREATE TABLE users (id INT);\nINSERT INTO users VALUES (1);",
				down:   "DROP TABLE users;",
				txUp:   true,
				txDown: true,
			},
		},
		{
			name: "tx=false on up",
			content: `-- +migrate Up tx=false
CREATE INDEX CONCURRENTLY idx ON users(id);
-- +migrate Down
DROP INDEX idx;`,
			want: SQLMigration{
				up:     "CREATE INDEX CONCURRENTLY idx ON users(id);",
				down:   "DROP INDEX idx;",
				txUp:   false,
				txDown: true,
			},
		},
		{
			name: "tx=false on both",
			content: `-- +migrate Up tx=false
CREATE INDEX CONCURRENTLY idx ON users(id);
-- +migrate Down tx=false
DROP INDEX CONCURRENTLY idx;`,
			want: SQLMigration{
				up:     "CREATE INDEX CONCURRENTLY idx ON users(id);",
				down:   "DROP INDEX CONCURRENTLY idx;",
				txUp:   false,
				txDown: false,
			},
		},
		{
			name: "content before up is ignored",
			content: `-- some comment
-- another comment

-- +migrate Up
CREATE TABLE users (id INT);
-- +migrate Down
DROP TABLE users;`,
			want: SQLMigration{
				up:     "CREATE TABLE users (id INT);",
				down:   "DROP TABLE users;",
				txUp:   true,
				txDown: true,
			},
		},
		{
			name:    "empty file",
			content: "",
			want: SQLMigration{
				up:     "",
				down:   "",
				txUp:   true,
				txDown: true,
			},
		},
		{
			name: "only up section",
			content: `-- +migrate Up
CREATE TABLE users (id INT);`,
			want: SQLMigration{
				up:     "CREATE TABLE users (id INT);",
				down:   "",
				txUp:   true,
				txDown: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSQLFile([]byte(tt.content), config)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.up != tt.want.up {
				t.Errorf("up:\ngot:  %q\nwant: %q", got.up, tt.want.up)
			}
			if got.down != tt.want.down {
				t.Errorf("down:\ngot:  %q\nwant: %q", got.down, tt.want.down)
			}
			if got.txUp != tt.want.txUp {
				t.Errorf("txUp: got %v, want %v", got.txUp, tt.want.txUp)
			}
			if got.txDown != tt.want.txDown {
				t.Errorf("txDown: got %v, want %v", got.txDown, tt.want.txDown)
			}
		})
	}
}

func Test_parseFileName(t *testing.T) {
	tests := []struct {
		name     string
		filepath string
		wantName string
		wantDate int64
		wantErr  bool
	}{
		{
			name:     "valid migration file",
			filepath: "20240101120000_create_users_table.sql",
			wantName: "create_users_table",
			wantDate: 20240101120000,
			wantErr:  false,
		},
		{
			name:     "multiple underscores in name",
			filepath: "20240101120000_create_users_table_with_indexes.sql",
			wantName: "create_users_table_with_indexes",
			wantDate: 20240101120000,
			wantErr:  false,
		},
		{
			name:     "no underscore",
			filepath: "20240101120000.sql",
			wantErr:  true,
		},
		{
			name:    "empty string",
			wantErr: true,
		},
		{
			name:     "invalid date format",
			filepath: "notadate_create_users.sql",
			wantErr:  true,
		},
		{
			name:     "underscore but empty name",
			filepath: "20240101120000_",
			wantErr:  true,
		},
		{
			name:     "underscore at start",
			filepath: "_create_users.sql",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotDate, err := parseFileName(tt.filepath)

			if (err != nil) != tt.wantErr {
				t.Fatalf("error: got %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if gotName != tt.wantName {
				t.Errorf("name: got %q, want %q", gotName, tt.wantName)
			}
			if gotDate != tt.wantDate {
				t.Errorf("date: got %d, want %d", gotDate, tt.wantDate)
			}
		})
	}
}

func Test_splitSQLStatementsWithAnnotations(t *testing.T) {
	tests := []struct {
		name string
		sql  string
		want []string
	}{
		{
			name: "simple statements",
			sql:  "CREATE TABLE foo (id INT);\nINSERT INTO foo VALUES (1);",
			want: []string{"CREATE TABLE foo (id INT)", "INSERT INTO foo VALUES (1)"},
		},
		{
			name: "single statement no semicolon",
			sql:  "CREATE TABLE foo (id INT)",
			want: []string{"CREATE TABLE foo (id INT)"},
		},
		{
			name: "annotated block preserves internal semicolons",
			sql: `CREATE TABLE foo (id INT);
-- amigo:statement:begin
CREATE FUNCTION test() RETURNS void AS $$
BEGIN
  SELECT 1;
  SELECT 2;
END;
$$ LANGUAGE plpgsql;
-- amigo:statement:end
CREATE TABLE bar (id INT);`,
			want: []string{
				"CREATE TABLE foo (id INT)",
				`CREATE FUNCTION test() RETURNS void AS $$
BEGIN
  SELECT 1;
  SELECT 2;
END;
$$ LANGUAGE plpgsql;`,
				"CREATE TABLE bar (id INT)",
			},
		},
		{
			name: "multiple annotated blocks",
			sql: `CREATE TABLE foo (id INT);
-- amigo:statement:begin
CREATE FUNCTION func1() AS $$
  SELECT 1;
$$;
-- amigo:statement:end
CREATE TABLE bar (id INT);
-- amigo:statement:begin
CREATE FUNCTION func2() AS $$
  SELECT 2;
$$;
-- amigo:statement:end
CREATE TABLE baz (id INT);`,
			want: []string{
				"CREATE TABLE foo (id INT)",
				`CREATE FUNCTION func1() AS $$
  SELECT 1;
$$;`,
				"CREATE TABLE bar (id INT)",
				`CREATE FUNCTION func2() AS $$
  SELECT 2;
$$;`,
				"CREATE TABLE baz (id INT)",
			},
		},
		{
			name: "semicolon in single quotes",
			sql:  `INSERT INTO foo VALUES ('hello; world');`,
			want: []string{`INSERT INTO foo VALUES ('hello; world')`},
		},
		{
			name: "semicolon in double quotes",
			sql:  `INSERT INTO foo VALUES ("hello; world");`,
			want: []string{`INSERT INTO foo VALUES ("hello; world")`},
		},
		{
			name: "empty input",
			sql:  "",
			want: nil,
		},
		{
			name: "whitespace only",
			sql:  "   \n   ",
			want: nil,
		},
		{
			name: "multiline statement",
			sql: `CREATE TABLE foo (
  id INT,
  name TEXT
);`,
			want: []string{`CREATE TABLE foo (
  id INT,
  name TEXT
)`},
		},
		{
			name: "only annotated block",
			sql: `-- amigo:statement:begin
CREATE FUNCTION test() AS $$
  SELECT 1;
$$;
-- amigo:statement:end`,
			want: []string{`CREATE FUNCTION test() AS $$
  SELECT 1;
$$;`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitSQLStatementsWithAnnotations(tt.sql)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitSQLStatementsWithAnnotations():\ngot:  %#v\nwant: %#v", got, tt.want)
			}
		})
	}
}
