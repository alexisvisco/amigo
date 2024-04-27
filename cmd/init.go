package cmd

import (
	"fmt"
	"github.com/alexisvisco/mig/internal/cli"
	"github.com/alexisvisco/mig/pkg/cmdexec"
	"github.com/alexisvisco/mig/pkg/schema"
	"github.com/spf13/cobra"
	"os"
	"path"
	"regexp"
	"strings"
)

var (
	initDumpFlag   bool
	initDumpShell  string
	initPGDumpPath string
	initDumpSchema string
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "will initialize migrations folder",
	RunE: func(cmd *cobra.Command, args []string) error {
		printer := cli.NewPrinter()

		err := os.MkdirAll(folderFlag, 0755)
		if err != nil {
			return fmt.Errorf("unable to create migration folder: %w", err)
		}

		printer.AddEvent(cli.FolderAddedEvent{FolderName: folderFlag})

		var migrations []string
		if initDumpFlag {
			structName, err := initDumpDatabase(printer)
			if err != nil {
				return err
			}

			migrations = append(migrations, structName)
		}

		err = cli.GenerateMigrationsFile(folderFlag, packageFlag, path.Join(folderFlag, "migrations.go"))
		if err != nil {
			return err
		}

		// todo: migration migration table into the database

		printer.
			AddEvent(cli.FileAddedEvent{FileName: path.Join(folderFlag, "migrations.go")}).
			Measure().
			Print(jsonFlag)

		return nil
	},
}

func initDumpDatabase(printer *cli.Printer) (structName string, err error) {
	if rootDSNFlag == "" {
		return "", fmt.Errorf("dsn is required, example: postgres://user:password@localhost:5432/dbname?sslmode=disable")
	}

	db, err := schema.ExtractCredentials(rootDSNFlag)
	if err != nil {
		return "", err
	}

	args := []string{
		initPGDumpPath,
		"-d", db.DB,
		"-h", db.Host,
		"-U", db.User,
		"-p", db.Port,
		"-n", initDumpSchema,
		"-s",
		"--no-comments",
		"--no-owner",
		"--no-privileges",
		"--no-tablespaces",
		"--no-security-labels",
	}

	env := map[string]string{"PGPASSWORD": db.Pass}

	stdout, stderr, err := cmdexec.Exec(initDumpShell, []string{"-c", strings.Join(args, " ")}, env)
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

	inUp := fmt.Sprintf("t.Exec(`%s`)", stdout)
	fileCreated, structName, err := cli.CreateMigrationFile(cli.CreateMigrationFileOptions{
		Name:    "init",
		Folder:  folderFlag,
		Driver:  "postgres",
		Package: packageFlag,
		MigType: "classic",
		InUp:    inUp,
		InDown:  "",
	})
	if err != nil {
		return "", err
	}

	printer.AddEvent(cli.FileAddedEvent{FileName: fileCreated})

	return structName, nil
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&initDumpFlag, "dump", false, "dump the migration")
	initCmd.Flags().StringVar(&initDumpShell, "shell", "/bin/bash", "the shell to use for the dump")
	initCmd.Flags().StringVar(&initPGDumpPath, "pg-dump", "pg_dump", "the path to the pg_dump command")
	initCmd.Flags().StringVar(&initDumpSchema, "schema", "public", "the schema to dump")
}
