package logsql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

var (
	driversMu sync.Mutex
	drivers   = make(map[string]string) // wrapped name -> original name
)

// PlaceholderStyle defines how to format query placeholders
type PlaceholderStyle int

const (
	// PlaceholderAuto auto-detects based on driver name
	PlaceholderAuto PlaceholderStyle = iota
	// PlaceholderDollar uses $1, $2, $3 (PostgreSQL)
	PlaceholderDollar
	// PlaceholderQuestion uses ? (MySQL, SQLite)
	PlaceholderQuestion
	// PlaceholderColon uses :1, :2, :3 (Oracle)
	PlaceholderColon
	// PlaceholderNumeric uses arg1, arg2, arg3 (generic)
	PlaceholderNumeric
)

type wrapOpts struct {
	Output      io.Writer
	Placeholder PlaceholderStyle
}

type WrapOptionFunc func(*wrapOpts)

// WrapOptionOutput sets the output writer for SQL logs
func WrapOptionOutput(w io.Writer) WrapOptionFunc {
	return func(opts *wrapOpts) {
		opts.Output = w
	}
}

// WrapOptionPlaceholder sets the placeholder style
func WrapOptionPlaceholder(style PlaceholderStyle) WrapOptionFunc {
	return func(opts *wrapOpts) {
		opts.Placeholder = style
	}
}

func defaultWrapOpts() wrapOpts {
	return wrapOpts{
		Output:      os.Stdout,
		Placeholder: PlaceholderAuto,
	}
}

// WrapDriver wraps a SQL driver to log all queries
// Returns the wrapped driver name to use with sql.Open()
func WrapDriver(driverName string, opts ...WrapOptionFunc) string {
	options := defaultWrapOpts()
	for _, opt := range opts {
		opt(&options)
	}

	if options.Output == nil {
		options.Output = os.Stdout
	}

	// Auto-detect placeholder style
	placeholder := options.Placeholder
	if placeholder == PlaceholderAuto {
		placeholder = detectPlaceholderStyle(driverName)
	}

	wrappedName := "logsql_" + driverName

	driversMu.Lock()
	defer driversMu.Unlock()

	// Check if already registered
	if _, exists := drivers[wrappedName]; exists {
		return wrappedName
	}

	// Register the wrapped driver
	sql.Register(wrappedName, &loggingDriver{
		driverName:  driverName,
		output:      options.Output,
		placeholder: placeholder,
	})

	drivers[wrappedName] = driverName
	return wrappedName
}

// detectPlaceholderStyle auto-detects the placeholder style based on driver name
func detectPlaceholderStyle(driverName string) PlaceholderStyle {
	switch {
	case contains(driverName, "postgres"), contains(driverName, "pg"), contains(driverName, "pgx"):
		return PlaceholderDollar
	case contains(driverName, "mysql"), contains(driverName, "sqlite"):
		return PlaceholderQuestion
	case contains(driverName, "oracle"), contains(driverName, "oci"):
		return PlaceholderColon
	default:
		return PlaceholderNumeric
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)))
}

type loggingDriver struct {
	driverName  string
	output      io.Writer
	placeholder PlaceholderStyle
}

func (d *loggingDriver) Open(name string) (driver.Conn, error) {
	// Open a connection with the original driver
	db, err := sql.Open(d.driverName, name)
	if err != nil {
		return nil, err
	}

	// Get the underlying driver.Conn
	conn, err := db.Driver().Open(name)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &loggingConn{
		Conn:        conn,
		output:      d.output,
		placeholder: d.placeholder,
	}, nil
}

type loggingConn struct {
	driver.Conn
	output      io.Writer
	placeholder PlaceholderStyle
	inTx        bool
}

func (c *loggingConn) logQuery(txPrefix, operation, query string, args interface{}, duration time.Duration, err error) {
	durationMs := colorYellow + fmt.Sprintf("%dms", duration.Milliseconds()) + colorReset
	queryStr := colorCyan + query + colorReset

	fmt.Fprintf(c.output, "SQL DEBUG > [%s%s] [%s] %s", txPrefix, operation, durationMs, queryStr)

	if args != nil {
		formattedArgs := formatArgs(args, c.placeholder)
		if formattedArgs != "" {
			fmt.Fprintf(c.output, " [%s]", formattedArgs)
		}
	}

	if err != nil {
		fmt.Fprintf(c.output, " : %sError %v%s", colorRed, err, colorReset)
	}

	fmt.Fprintf(c.output, "\n")
}

func (c *loggingConn) Prepare(query string) (driver.Stmt, error) {
	stmt, err := c.Conn.Prepare(query)
	if err != nil {
		return nil, err
	}
	return &loggingStmt{Stmt: stmt, query: query, output: c.output, inTx: c.inTx, placeholder: c.placeholder}, nil
}

func (c *loggingConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if connCtx, ok := c.Conn.(driver.ConnPrepareContext); ok {
		stmt, err := connCtx.PrepareContext(ctx, query)
		if err != nil {
			return nil, err
		}
		return &loggingStmt{Stmt: stmt, query: query, output: c.output, inTx: c.inTx, placeholder: c.placeholder}, nil
	}
	return c.Prepare(query)
}

func (c *loggingConn) Begin() (driver.Tx, error) {
	start := time.Now()
	tx, err := c.Conn.Begin()
	duration := time.Since(start)
	c.logQuery("TX ", "QUERY", "BEGIN", nil, duration, err)

	if err != nil {
		return nil, err
	}
	c.inTx = true
	return &loggingTx{Tx: tx, output: c.output, placeholder: c.placeholder, conn: c}, nil
}

func (c *loggingConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if connCtx, ok := c.Conn.(driver.ConnBeginTx); ok {
		start := time.Now()
		tx, err := connCtx.BeginTx(ctx, opts)
		duration := time.Since(start)
		c.logQuery("TX ", "QUERY", "BEGIN", nil, duration, err)

		if err != nil {
			return nil, err
		}
		c.inTx = true
		return &loggingTx{Tx: tx, output: c.output, placeholder: c.placeholder, conn: c}, nil
	}
	return c.Begin()
}

type loggingStmt struct {
	driver.Stmt
	query       string
	output      io.Writer
	inTx        bool
	placeholder PlaceholderStyle
}

func (s *loggingStmt) logQuery(operation string, args interface{}, duration time.Duration, err error) {
	durationMs := colorYellow + fmt.Sprintf("%dms", duration.Milliseconds()) + colorReset
	queryStr := colorCyan + s.query + colorReset

	txPrefix := ""
	if s.inTx {
		txPrefix = "TX "
	}

	fmt.Fprintf(s.output, "SQL DEBUG > [%s%s] [%s] %s", txPrefix, operation, durationMs, queryStr)

	if args != nil {
		formattedArgs := formatArgs(args, s.placeholder)
		if formattedArgs != "" {
			fmt.Fprintf(s.output, " [%s]", formattedArgs)
		}
	}

	if err != nil {
		fmt.Fprintf(s.output, " : %sError %v%s", colorRed, err, colorReset)
	}

	fmt.Fprintf(s.output, "\n")
}

func (s *loggingStmt) Exec(args []driver.Value) (driver.Result, error) {
	start := time.Now()
	result, err := s.Stmt.Exec(args)
	duration := time.Since(start)
	s.logQuery("EXEC", args, duration, err)
	return result, err
}

func (s *loggingStmt) Query(args []driver.Value) (driver.Rows, error) {
	start := time.Now()
	rows, err := s.Stmt.Query(args)
	duration := time.Since(start)
	s.logQuery("QUERY", args, duration, err)
	return rows, err
}

func (s *loggingStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	if stmtCtx, ok := s.Stmt.(driver.StmtExecContext); ok {
		start := time.Now()
		result, err := stmtCtx.ExecContext(ctx, args)
		duration := time.Since(start)
		s.logQuery("EXEC", namedValuesToValues(args), duration, err)
		return result, err
	}
	return s.Exec(namedValuesToDriverValues(args))
}

func (s *loggingStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	if stmtCtx, ok := s.Stmt.(driver.StmtQueryContext); ok {
		start := time.Now()
		rows, err := stmtCtx.QueryContext(ctx, args)
		duration := time.Since(start)
		s.logQuery("QUERY", namedValuesToValues(args), duration, err)
		return rows, err
	}
	return s.Query(namedValuesToDriverValues(args))
}

type loggingTx struct {
	driver.Tx
	output      io.Writer
	placeholder PlaceholderStyle
	conn        *loggingConn
}

func (t *loggingTx) logQuery(operation, query string, duration time.Duration, err error) {
	durationMs := colorYellow + fmt.Sprintf("%dms", duration.Milliseconds()) + colorReset
	queryStr := colorCyan + query + colorReset

	fmt.Fprintf(t.output, "SQL DEBUG > [TX %s] [%s] %s", operation, durationMs, queryStr)

	if err != nil {
		fmt.Fprintf(t.output, " : %sError %v%s", colorRed, err, colorReset)
	}

	fmt.Fprintf(t.output, "\n")
}

func (t *loggingTx) Commit() error {
	start := time.Now()
	err := t.Tx.Commit()
	duration := time.Since(start)
	t.logQuery("QUERY", "COMMIT", duration, err)
	t.conn.inTx = false
	return err
}

func (t *loggingTx) Rollback() error {
	start := time.Now()
	err := t.Tx.Rollback()
	duration := time.Since(start)
	t.logQuery("QUERY", "ROLLBACK", duration, err)
	t.conn.inTx = false
	return err
}

// Prepare creates a prepared statement within the transaction
func (t *loggingTx) Prepare(query string) (driver.Stmt, error) {
	stmt, err := t.Tx.(interface {
		Prepare(string) (driver.Stmt, error)
	}).Prepare(query)
	if err != nil {
		return nil, err
	}
	return &loggingStmt{Stmt: stmt, query: query, output: t.output, inTx: true, placeholder: t.placeholder}, nil
}

// PrepareContext creates a prepared statement within the transaction
func (t *loggingTx) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if stmtCtx, ok := t.Tx.(driver.ConnPrepareContext); ok {
		stmt, err := stmtCtx.PrepareContext(ctx, query)
		if err != nil {
			return nil, err
		}
		return &loggingStmt{Stmt: stmt, query: query, output: t.output, inTx: true, placeholder: t.placeholder}, nil
	}
	return t.Prepare(query)
}

// ExecContext intercepts exec calls within a transaction
func (t *loggingTx) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if execer, ok := t.Tx.(driver.ExecerContext); ok {
		start := time.Now()
		result, err := execer.ExecContext(ctx, query, args)
		duration := time.Since(start)

		durationMs := colorYellow + fmt.Sprintf("%dms", duration.Milliseconds()) + colorReset
		queryStr := colorCyan + query + colorReset

		fmt.Fprintf(t.output, "SQL DEBUG > [TX EXEC] [%s] %s", durationMs, queryStr)

		if len(args) > 0 {
			formattedArgs := formatArgs(namedValuesToValues(args), t.placeholder)
			if formattedArgs != "" {
				fmt.Fprintf(t.output, " [%s]", formattedArgs)
			}
		}

		if err != nil {
			fmt.Fprintf(t.output, " : %sError %v%s", colorRed, err, colorReset)
		}

		fmt.Fprintf(t.output, "\n")
		return result, err
	}
	// Fallback to prepare + exec
	stmt, err := t.conn.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// Force inTx to true
	if ls, ok := stmt.(*loggingStmt); ok {
		ls.inTx = true
	}

	return stmt.(*loggingStmt).ExecContext(ctx, args)
}

// QueryContext intercepts query calls within a transaction
func (t *loggingTx) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if queryer, ok := t.Tx.(driver.QueryerContext); ok {
		start := time.Now()
		rows, err := queryer.QueryContext(ctx, query, args)
		duration := time.Since(start)

		durationMs := colorYellow + fmt.Sprintf("%dms", duration.Milliseconds()) + colorReset
		queryStr := colorCyan + query + colorReset

		fmt.Fprintf(t.output, "SQL DEBUG > [TX QUERY] [%s] %s", durationMs, queryStr)

		if len(args) > 0 {
			formattedArgs := formatArgs(namedValuesToValues(args), t.placeholder)
			if formattedArgs != "" {
				fmt.Fprintf(t.output, " [%s]", formattedArgs)
			}
		}

		if err != nil {
			fmt.Fprintf(t.output, " : %sError %v%s", colorRed, err, colorReset)
		}

		fmt.Fprintf(t.output, "\n")
		return rows, err
	}
	// Fallback to prepare + query
	stmt, err := t.conn.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// Force inTx to true
	if ls, ok := stmt.(*loggingStmt); ok {
		ls.inTx = true
	}

	return stmt.(*loggingStmt).QueryContext(ctx, args)
}

// Helper functions to convert between driver value types
func namedValuesToValues(args []driver.NamedValue) []interface{} {
	values := make([]interface{}, len(args))
	for i, arg := range args {
		values[i] = arg.Value
	}
	return values
}

func namedValuesToDriverValues(args []driver.NamedValue) []driver.Value {
	values := make([]driver.Value, len(args))
	for i, arg := range args {
		values[i] = arg.Value
	}
	return values
}

// formatArgs formats arguments for display
func formatArgs(args interface{}, placeholder PlaceholderStyle) string {
	var values []interface{}

	switch v := args.(type) {
	case []driver.Value:
		if len(v) == 0 {
			return ""
		}
		values = make([]interface{}, len(v))
		for i, val := range v {
			values[i] = val
		}
	case []interface{}:
		if len(v) == 0 {
			return ""
		}
		values = v
	default:
		return fmt.Sprintf("%v", args)
	}

	result := ""
	for i, arg := range values {
		if i > 0 {
			result += ", "
		}

		switch placeholder {
		case PlaceholderDollar:
			result += fmt.Sprintf("$%d=%v", i+1, arg)
		case PlaceholderQuestion:
			result += fmt.Sprintf("?=%v", arg)
		case PlaceholderColon:
			result += fmt.Sprintf(":%d=%v", i+1, arg)
		case PlaceholderNumeric:
			result += fmt.Sprintf("arg%d=%v", i+1, arg)
		default:
			result += fmt.Sprintf("arg%d=%v", i+1, arg)
		}
	}

	return result
}
