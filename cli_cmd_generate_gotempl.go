package amigo

const goTemplate = `package {{.PackageName}}

import (
	"context"
	"database/sql"
{{if .Transactional}}
	"github.com/alexisvisco/amigo"
{{end}})

type Migration{{.Timestamp}}{{.StructName}} struct{}

func (m Migration{{.Timestamp}}{{.StructName}}) Name() string {
	return "{{.Name}}"
}

func (m Migration{{.Timestamp}}{{.StructName}}) Date() int64 {
	return {{.Timestamp}}
}

func (m Migration{{.Timestamp}}{{.StructName}}) Up(ctx context.Context, db *sql.DB) error {
{{if .Transactional}}	return amigo.Tx(ctx, db, func(tx *sql.Tx) error {
		// TODO: implement up migration
		return nil
	})
{{else}}	// TODO: implement up migration
	return nil
{{end}}}

func (m Migration{{.Timestamp}}{{.StructName}}) Down(ctx context.Context, db *sql.DB) error {
{{if .Transactional}}	return amigo.Tx(ctx, db, func(tx *sql.Tx) error {
		// TODO: implement down migration
		return nil
	})
{{else}}	// TODO: implement down migration
	return nil
{{end}}}
`
