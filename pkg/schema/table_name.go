package schema

import (
	"fmt"
	"strings"
)

type TableName string

// Schema returns the schema part of the table name.
func (t TableName) Schema() string {
	index := strings.IndexByte(string(t), '.')
	if index != -1 {
		return string(t[:index])
	}
	return "public"
}

// HasSchema returns if the table name has a schema.
func (t TableName) HasSchema() bool {
	return strings.Contains(string(t), ".")
}

// Name returns the name part of the table name.
func (t TableName) Name() string {
	index := strings.IndexByte(string(t), '.')
	if index != -1 {
		return string(t[index+1:])
	}
	return string(t)
}

// String returns the string representation of the table name.
func (t TableName) String() string {
	return string(t)
}

// Table returns a new TableName with the schema and table name.
// But you can also use regular string, this function is when you have dynamic schema names. (like in the tests)
func Table(name string, schema ...string) TableName {
	if len(schema) == 0 {
		return TableName(name)
	}
	return TableName(fmt.Sprintf("%s.%s", schema[0], name))
}
