package schema

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode"
)

func parentheses(in string) string { return fmt.Sprintf("(%s)", in) }

type replacer map[string]func() string

// Replace replaces the string with the given values {<key>} to the value of the function
func (r replacer) replace(str string) string {
	for k, v := range r {
		str = strings.ReplaceAll(str, "{"+k+"}", v())
	}
	res, err := removeConsecutiveSpaces(str)
	if err != nil {
		res = str
	}

	return res
}

func strfunc(val string) func() string { return func() string { return val } }

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// ForceStopError is an error that stops the migration process even if the `continue_on_error` option is set.
type ForceStopError struct{ error }

// NewForceStopError creates a new ForceStopError.
func NewForceStopError(err error) *ForceStopError {
	return &ForceStopError{err}
}

// Remove all whitespace not between matching unescaped quotes.
// Example: SELECT    * FROM "table" WHERE "column" = 'value  '
// Result: SELECT * FROM "table" WHERE "column" = 'value  '
func removeConsecutiveSpaces(s string) (string, error) {
	rs := make([]rune, 0, len(s))
	for i := 0; i < len(s); i++ {
		r := rune(s[i])
		if r == '\'' || r == '"' {
			prevChar := ' '
			matchedChar := uint8(r)

			// if the text remaining is 'value \' '
			// then the quoteText will be 'value \' '
			// if there is no end quote then it will return an error
			quoteText := string(s[i])

			// jump until the next matching quote character
			for n := i + 1; n < len(s); n++ {
				if s[n] == matchedChar && prevChar != '\\' {
					i = n
					quoteText += string(s[n])
					break
				}
				quoteText += string(s[n])
				prevChar = rune(s[n])
			}

			if quoteText[len(quoteText)-1] != matchedChar || len(quoteText) == 1 {
				err := fmt.Errorf("unmatched unescaped quote: %q", quoteText)
				return "", err
			}

			rs = append(rs, []rune(quoteText)...)
			continue
		}

		if unicode.IsSpace(r) {
			rs = append(rs, r)

			// jump until the next non-space character
			for n := i + 1; n < len(s); n++ {
				if !unicode.IsSpace(rune(s[n])) {
					i = n - 1 // -1 because the loop will increment it
					break
				}
			}

			continue
		}

		if !unicode.IsSpace(r) {
			rs = append(rs, r)
		}

	}

	return strings.Trim(string(rs), " "), nil
}

// ------------------------------------------------------------ tests utils
func osEnvOrDefault(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}

type databaseCredentials struct {
	host, port, user, pass, db string
}

func execCmd(cmd string, args []string, env map[string]string) (string, error) {
	co := exec.Command(cmd, args...)

	for k, v := range env {
		co.Env = append(co.Env, k+"="+v)
	}

	co.Env = append(co.Env, os.Environ()...)

	addToPath := []string{"/opt/homebrew/opt/libpq/bin", "/usr/local/opt/libpq/bin"}
	for i, key := range co.Env {
		if strings.HasPrefix(key, "PATH=") {
			co.Env[i] = key + ":" + strings.Join(addToPath, ":")
			break
		}
	}

	buffer := new(strings.Builder)

	co.Stdout = buffer
	co.Stderr = os.Stderr
	err := co.Run()
	if err != nil {
		return buffer.String(), fmt.Errorf("unable to execute command: %w", err)
	}

	return buffer.String(), nil
}

func Ptr[T any](t T) *T { return &t }

type recorder interface {
	Record(f func()) string
	SetRecord(v bool)
	fmt.Stringer
}
