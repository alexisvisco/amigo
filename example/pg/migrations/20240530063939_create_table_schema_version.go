package migrations

import (
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/schema/pg"
	"time"
)

type Migration20240530063939CreateTableSchemaVersion struct{}

func (m Migration20240530063939CreateTableSchemaVersion) Up(s *pg.Schema) {
	s.CreateTable("public.mig_schema_versions", func(s *pg.PostgresTableDef) {
		s.String("id")
	}, schema.TableOptions{IfNotExists: true})
}

func (m Migration20240530063939CreateTableSchemaVersion) Down(s *pg.Schema) {
	// nothing to do to keep the schema version table
}

func (m Migration20240530063939CreateTableSchemaVersion) Name() string {
	return "create_table_schema_version"
}

func (m Migration20240530063939CreateTableSchemaVersion) Date() time.Time {
	t, _ := time.Parse(time.RFC3339, "2024-05-30T08:39:39+02:00")
	return t
}
