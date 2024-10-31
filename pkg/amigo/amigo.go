package amigo

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/alexisvisco/amigo/pkg/amigoctx"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/templates"
	"github.com/alexisvisco/amigo/pkg/types"
	"github.com/alexisvisco/amigo/pkg/utils"
	"github.com/alexisvisco/amigo/pkg/utils/cmdexec"
	"github.com/gobuffalo/flect"
)

type Amigo struct {
	ctx    *amigoctx.Context
	Driver types.Driver
}

// NewAmigo create a new amigo instance
func NewAmigo(ctx *amigoctx.Context) Amigo {
	return Amigo{
		ctx:    ctx,
		Driver: types.GetDriver(ctx.DSN),
	}
}

// DumpSchema of the database and write it to the writer
func (a Amigo) DumpSchema() (string, error) {
	db, err := schema.ExtractCredentials(a.ctx.GetRealDSN())
	if err != nil {
		return "", err
	}

	ignoreTableName := a.ctx.SchemaVersionTable
	if strings.Contains(ignoreTableName, ".") {
		ignoreTableName = strings.Split(ignoreTableName, ".")[1]
	}

	args := []string{
		a.ctx.PGDumpPath,
		"-d", db.DB,
		"-h", db.Host,
		"-U", db.User,
		"-p", db.Port,
		"-n", a.ctx.Create.DumpSchema,
		"-s",
		"-x",
		"-O",
		"-T", ignoreTableName,
		"--no-comments",
		"--no-owner",
		"--no-privileges",
		"--no-tablespaces",
		"--no-security-labels",
	}

	env := map[string]string{"PGPASSWORD": db.Pass}

	stdout, stderr, err := cmdexec.Exec(a.ctx.ShellPath, []string{"-c", strings.Join(args, " ")}, env)
	if err != nil {
		return "", fmt.Errorf("unable to dump database: %w\n%s", err, stderr)
	}

	// replace all regexp listed below to empty string
	regexpReplace := []string{
		`--
-- Name: .*; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA .*;
`,
	}

	for _, r := range regexpReplace {
		stdout = regexp.MustCompile(r).ReplaceAllString(stdout, "")
	}

	return stdout, nil
}

// GenerateMainFile generate the main.go file in the amigo folder
func (a Amigo) GenerateMainFile(writer io.Writer) error {
	name, err := utils.GetModuleName()
	if err != nil {
		return fmt.Errorf("unable to get module name: %w", err)
	}

	packagePath := path.Join(name, a.ctx.MigrationFolder)

	template, err := templates.GetMainTemplate(templates.MainData{
		PackagePath: packagePath,
		DriverPath:  a.Driver.PackagePath(),
		DriverName:  a.Driver.String(),
	})

	if err != nil {
		return fmt.Errorf("unable to get main template: %w", err)
	}

	_, err = writer.Write([]byte(template))
	if err != nil {
		return fmt.Errorf("unable to write main file: %w", err)
	}

	return nil
}

type GenerateMigrationFileParams struct {
	Name            string
	Up              string
	Down            string
	Change          string
	Type            types.MigrationFileType
	Now             time.Time
	UseSchemaImport bool
	UseFmtImport    bool
	Writer          io.Writer
}

// GenerateMigrationFile generate a migration file in the migrations folder
func (a Amigo) GenerateMigrationFile(params *GenerateMigrationFileParams) error {

	structName := utils.MigrationStructName(params.Now, params.Name)

	orDefault := func(s string) string {
		if s == "" {
			return "// TODO: implement the migration"
		}
		return s
	}

	fileContent, err := templates.GetMigrationTemplate(templates.MigrationData{
		IsSQL:             params.Type == types.MigrationFileTypeSQL,
		Package:           a.ctx.PackagePath,
		StructName:        structName,
		Name:              flect.Underscore(params.Name),
		Type:              params.Type,
		InChange:          orDefault(params.Change),
		InUp:              orDefault(params.Up),
		InDown:            orDefault(params.Down),
		CreatedAt:         params.Now.Format(time.RFC3339),
		PackageDriverName: a.Driver.PackageName(),
		PackageDriverPath: a.Driver.PackageSchemaPath(),
		UseSchemaImport:   params.UseSchemaImport,
		UseFmtImport:      params.UseFmtImport,
	})

	if err != nil {
		return fmt.Errorf("unable to get migration template: %w", err)
	}

	_, err = params.Writer.Write([]byte(fileContent))
	if err != nil {
		return fmt.Errorf("unable to write migration file: %w", err)
	}

	return nil
}

// GenerateMigrationsFiles generate the migrations file in the migrations folder
// It's used to keep track of all migrations
func (a Amigo) GenerateMigrationsFiles(writer io.Writer) error {
	migrationFiles, keys, err := a.GetMigrationFiles(true)
	if err != nil {
		return err
	}

	var migrations []string
	var mustImportSchemaPackage *string
	for _, k := range keys {
		if migrationFiles[k].IsSQL {
			// schema.NewSQLMigration[*pg.Schema](sqlMigrationsFS, "20240602081806_drop_index.sql", "2024-06-02T10:18:06+02:00", "---- down:"),
			line := fmt.Sprintf("schema.NewSQLMigration[%s](sqlMigrationsFS, \"%s\", \"%s\", \"%s\")",
				a.Driver.StructName(),
				migrationFiles[k].FulName,
				k.Format(time.RFC3339),
				a.ctx.Create.SQLSeparator,
			)

			migrations = append(migrations, line)

			if mustImportSchemaPackage == nil {
				v := a.Driver.PackageSchemaPath()
				mustImportSchemaPackage = &v
			}
		} else {
			migrations = append(migrations, fmt.Sprintf("&%s{}", utils.MigrationStructName(k, migrationFiles[k].Name)))

		}
	}

	content, err := templates.GetMigrationsTemplate(templates.MigrationsData{
		Package:             a.ctx.PackagePath,
		Migrations:          migrations,
		ImportSchemaPackage: mustImportSchemaPackage,
	})

	if err != nil {
		return fmt.Errorf("unable to get migrations template: %w", err)
	}

	_, err = writer.Write([]byte(content))
	if err != nil {
		return fmt.Errorf("unable to write migrations file: %w", err)
	}

	return nil
}

// GetStatus return the state of the database
func (a Amigo) GetStatus(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT id FROM " + a.ctx.SchemaVersionTable + " ORDER BY id desc")
	if err != nil {
		return nil, fmt.Errorf("unable to get state: %w", err)
	}

	var state []string
	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("unable to scan state: %w", err)
		}
		state = append(state, id)
	}

	return state, nil
}

type MigrationFile struct {
	Name    string
	FulName string
	IsSQL   bool
}

func (a Amigo) GetMigrationFiles(ascending bool) (map[time.Time]MigrationFile, []time.Time, error) {
	migrationFiles := make(map[time.Time]MigrationFile)

	// get the list of structs by the file name
	err := filepath.Walk(a.ctx.MigrationFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			if utils.MigrationFileRegexp.MatchString(info.Name()) {
				matches := utils.MigrationFileRegexp.FindStringSubmatch(info.Name())
				fileTime := matches[1]
				migrationName := matches[2]
				ext := matches[3]

				t, _ := time.Parse(utils.FormatTime, fileTime)
				migrationFiles[t] = MigrationFile{Name: migrationName, IsSQL: ext == "sql", FulName: info.Name()}
			}
		}

		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("unable to walk through the migration folder: %w", err)
	}

	// sort the files
	var keys []time.Time
	for k := range migrationFiles {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		if ascending {
			return keys[i].Unix() < keys[j].Unix()
		} else {
			return keys[i].Unix() > keys[j].Unix()
		}
	})

	return migrationFiles, keys, nil
}

func (a Amigo) SkipMigrationFile(db *sql.DB) error {
	_, err := db.Exec("INSERT INTO "+a.ctx.SchemaVersionTable+" (id) VALUES ($1)", a.ctx.Create.Version)
	if err != nil {
		return fmt.Errorf("unable to skip migration file: %w", err)
	}

	return nil
}
