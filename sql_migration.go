package amigo

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"embed"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type SQLMigration struct {
	up   string
	down string
	name string
	date int64

	txUp   bool
	txDown bool

	splitStatements bool
}

func (s SQLMigration) Up(ctx context.Context, db *sql.DB) error {
	if s.txUp {
		return Tx(ctx, db, func(tx *sql.Tx) error {
			return s.execSQL(ctx, tx, s.up)
		})
	}

	return s.execSQLDB(ctx, db, s.up)
}

func (s SQLMigration) Down(ctx context.Context, db *sql.DB) error {
	if s.txDown {
		return Tx(ctx, db, func(tx *sql.Tx) error {
			return s.execSQL(ctx, tx, s.down)
		})
	}

	return s.execSQLDB(ctx, db, s.down)
}

// execSQL executes SQL within a transaction, either as a single exec or split by statements
func (s SQLMigration) execSQL(ctx context.Context, tx *sql.Tx, query string) error {
	if !s.splitStatements {
		_, err := tx.ExecContext(ctx, query)
		return err
	}

	statements := splitSQLStatementsWithAnnotations(query)
	for _, stmt := range statements {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

// execSQLDB executes SQL on a DB, either as a single exec or split by statements
func (s SQLMigration) execSQLDB(ctx context.Context, db *sql.DB, query string) error {
	if !s.splitStatements {
		_, err := db.ExecContext(ctx, query)
		return err
	}

	statements := splitSQLStatementsWithAnnotations(query)
	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s SQLMigration) Name() string {
	return s.name
}

func (s SQLMigration) Date() int64 {
	return s.date
}

// SQLFileToMigration converts a sql file from an embedded filesystem to a Migration struct
func SQLFileToMigration(fs embed.FS, filepath string, config Configuration) Migration {
	file, err := fs.ReadFile(filepath)
	if err != nil {
		panic(fmt.Sprintf("failed to read migration file %s: %v", filepath, err))
	}

	name, date, err := parseFileName(filepath)
	if err != nil {
		panic(fmt.Sprintf("failed to parse migration file name %s: %v", filepath, err))
	}
	migration, err := parseSQLFile(file, config)
	if err != nil {
		panic(fmt.Sprintf("failed to parse migration file %s: %v", filepath, err))
	}
	migration.name = name
	migration.date = date
	migration.splitStatements = config.SplitStatements

	return migration
}

// parseFileName parses the migration file name to extract the name and date
//
//	ex: "20240101120000_create_users_table.sql" -> gives "20240101120000", "create_users_table"
func parseFileName(filePath string) (name string, date int64, err error) {
	// Extract just the filename from the path
	filename := filepath.Base(filePath)

	n := strings.SplitN(filename, "_", 2)
	if len(n) != 2 {
		return "", 0, fmt.Errorf("invalid migration file name: %s", filePath)
	}

	toDate, err := parseVersionToDate(n[0])
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse migration file name: %w", err)
	}

	// Remove .sql extension
	name = strings.TrimSuffix(n[1], ".sql")

	if name == "" {
		return "", 0, fmt.Errorf("invalid migration file name: empty name")
	}

	return name, toDate, nil
}

func parseVersionToDate(version any) (int64, error) {
	vStr, ok := version.(string)
	if !ok {
		return 0, fmt.Errorf("version is not a string: %v", version)
	}

	var date int64
	_, err := fmt.Sscanf(vStr, "%d", &date)
	if err != nil {
		return 0, fmt.Errorf("failed to parse version to date: %w", err)
	}

	return date, nil
}

// parseSQLFile parses the content of a SQL file into an SQLMigration struct
// It looks for the up and down annotations to split the file into up and down migrations
// example of file:
// -- migrate:up tx=true
// CREATE TABLE users (id INT PRIMARY KEY, name TEXT);
// -- migrate:down tx=false
// DROP TABLE users;
// In this example, the up migration will be run in a transaction, while the down migration will not
func parseSQLFile(fileContent []byte, config Configuration) (SQLMigration, error) {
	file := SQLMigration{
		up:     "",
		down:   "",
		txUp:   true,
		txDown: true,
	}

	var upLines, downLines [][]byte
	var current *[][]byte // nil = before up, &upLines = in up, &downLines = in down

	txRegexp := regexp.MustCompile(`tx=(true|false)`)
	scanner := bufio.NewScanner(bytes.NewReader(fileContent))
	for scanner.Scan() {
		line := scanner.Bytes()

		if bytes.HasPrefix(line, []byte(config.SQLFileUpAnnotation)) {
			parseTxAnnotation(scanner.Text(), &file.txUp, txRegexp)
			current = &upLines
			continue
		}
		if bytes.HasPrefix(line, []byte(config.SQLFileDownAnnotation)) {
			parseTxAnnotation(scanner.Text(), &file.txDown, txRegexp)
			current = &downLines
			continue
		}

		if current != nil {
			*current = append(*current, bytes.Clone(line)) // Clone car scanner réutilise le buffer
		}
	}

	if err := scanner.Err(); err != nil {
		return file, fmt.Errorf("failed to scan file: %w", err)
	}

	file.up = string(bytes.Join(upLines, []byte("\n")))
	file.down = string(bytes.Join(downLines, []byte("\n")))

	return file, nil
}

// parseTxAnnotation parses the tx annotation from the given text and sets the value of b accordingly
// A typical refexp is regexp.MustCompile(`tx=(true|false)`)
func parseTxAnnotation(line string, b *bool, annotation *regexp.Regexp) {
	matches := annotation.FindStringSubmatch(line)
	if len(matches) == 2 {
		if matches[1] == "false" {
			*b = false
		} else if matches[1] == "true" {
			*b = true
		}
	}
}

const (
	statementBeginAnnotation = "-- amigo:statement:begin"
	statementEndAnnotation   = "-- amigo:statement:end"
)

// splitSQLStatementsWithAnnotations splits a SQL string into individual statements,
// respecting -- amigo:statement:begin/end annotations that protect complex statements
// (like PostgreSQL functions with dollar-quoted strings) from being split incorrectly.
func splitSQLStatementsWithAnnotations(sql string) []string {
	var statements []string
	var current strings.Builder
	var annotatedBlock strings.Builder
	inAnnotatedBlock := false

	lines := strings.Split(sql, "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == statementBeginAnnotation {
			inAnnotatedBlock = true
			continue
		}

		if trimmedLine == statementEndAnnotation {
			inAnnotatedBlock = false
			// Flush the annotated block as a single statement
			stmt := strings.TrimSpace(annotatedBlock.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			annotatedBlock.Reset()
			continue
		}

		if inAnnotatedBlock {
			if annotatedBlock.Len() > 0 {
				annotatedBlock.WriteByte('\n')
			}
			annotatedBlock.WriteString(line)
		} else {
			// Split by semicolons (respecting quotes) for non-annotated content
			stmts := splitLineByStatements(line, &current)
			statements = append(statements, stmts...)
		}
	}

	// Add any remaining content
	stmt := strings.TrimSpace(current.String())
	if stmt != "" {
		statements = append(statements, stmt)
	}

	return statements
}

// splitLineByStatements splits a line by semicolons, respecting quotes,
// and accumulates partial statements across lines.
func splitLineByStatements(line string, current *strings.Builder) []string {
	var statements []string
	inSingleQuote := false
	inDoubleQuote := false

	for i := 0; i < len(line); i++ {
		ch := line[i]

		// Handle escape sequences
		if i > 0 && line[i-1] == '\\' {
			current.WriteByte(ch)
			continue
		}

		// Toggle quote states
		if ch == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
			current.WriteByte(ch)
			continue
		}

		if ch == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
			current.WriteByte(ch)
			continue
		}

		// Split on semicolon if not in quotes
		if ch == ';' && !inSingleQuote && !inDoubleQuote {
			stmt := strings.TrimSpace(current.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
			continue
		}

		current.WriteByte(ch)
	}

	// Add newline if there's content (to preserve multiline statements)
	if current.Len() > 0 {
		current.WriteByte('\n')
	}

	return statements
}
