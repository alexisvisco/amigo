package pg

import (
	"context"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func assertTableExist(t *testing.T, p *Schema, table schema.TableName) {
	var exists bool
	err := p.DB.QueryRowContext(context.Background(), `SELECT EXISTS (
		SELECT 1
		FROM information_schema.tables
		WHERE table_schema = $1
		AND table_name = $2
	);`, table.Schema(), table.Name()).Scan(&exists)

	require.NoError(t, err)
	require.True(t, exists)
}

func assertTableNotExist(t *testing.T, p *Schema, table schema.TableName) {
	var exists bool
	err := p.DB.QueryRowContext(context.Background(), `SELECT EXISTS (
		SELECT 1
		FROM information_schema.tables
		WHERE table_schema = $1
		AND table_name = $2
	);`, table.Schema(), table.Name()).Scan(&exists)

	require.NoError(t, err)
	require.False(t, exists)
}

func TestQuote_Panic(t *testing.T) {
	assert.PanicsWithValue(t, "unsupported value", func() {
		QuoteValue(1)
	})
}

func TestQuote_ID(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{`foo`, `"foo"`},
		{`foo bar baz`, `"foo bar baz"`},
		{`foo"bar`, `"foo""bar"`},
		{"foo\x00bar", `"foo"`},
		{"\x00foo", `""`},
	}

	for _, test := range cases {
		assert.Equal(t, test.want, QuoteIdent(test.input))
	}
}

func TestQuote_Value(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{`foo`, `'foo'`},
		{`foo bar baz`, `'foo bar baz'`},
		{`foo'bar`, `'foo''bar'`},
		{`foo\bar`, ` E'foo\\bar'`},
		{`foo\ba'r`, ` E'foo\\ba''r'`},
		{`foo"bar`, `'foo"bar'`},
		{`foo\x00bar`, ` E'foo\\x00bar'`},
		{`\x00foo`, ` E'\\x00foo'`},
		{`'`, `''''`},
		{`''`, `''''''`},
		{`\`, ` E'\\'`},
		{`'abc'; DROP TABLE users;`, `'''abc''; DROP TABLE users;'`},
		{`\'`, ` E'\\'''`},
		{`E'\''`, ` E'E''\\'''''`},
		{`e'\''`, ` E'e''\\'''''`},
		{`E'\'abc\'; DROP TABLE users;'`, ` E'E''\\''abc\\''; DROP TABLE users;'''`},
		{`e'\'abc\'; DROP TABLE users;'`, ` E'e''\\''abc\\''; DROP TABLE users;'''`},
	}

	for _, test := range cases {
		assert.Equal(t, test.want, QuoteValue(test.input))
	}
}
