package amigo

import (
	"fmt"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/utils/cmdexec"
	"regexp"
	"strings"
)

type DumpSchemaOptions struct {
	DSN                string
	MigrationTableName string
	PGDumpPath         string
	Schema             string
	Shell              string
}

func DumpSchema(opts *DumpSchemaOptions) (schemaDump string, err error) {
	db, err := schema.ExtractCredentials(opts.DSN)
	if err != nil {
		return "", err
	}

	ignoreTableName := opts.MigrationTableName
	if strings.Contains(opts.MigrationTableName, ".") {
		ignoreTableName = strings.Split(opts.MigrationTableName, ".")[1]
	}

	args := []string{
		opts.PGDumpPath,
		"-d", db.DB,
		"-h", db.Host,
		"-U", db.User,
		"-p", db.Port,
		"-n", opts.Schema,
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

	stdout, stderr, err := cmdexec.Exec(opts.Shell, []string{"-c", strings.Join(args, " ")}, env)
	if err != nil {
		return "", fmt.Errorf("unable to dump database: %w\n%s", err, stderr)
	}

	// replace all regexp listed below to empty string
	regexpReplace := []string{
		`--
-- Name: .*; MigrationDirection: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA .*;

`,
	}

	for _, r := range regexpReplace {
		stdout = regexp.MustCompile(r).ReplaceAllString(stdout, "")
	}

	return stdout, nil
}
