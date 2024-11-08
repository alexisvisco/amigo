package amigo

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/utils/cmdexec"
)

func (a Amigo) DumpSchema(writer io.Writer, ignoreSchemaVersionTable bool) error {
	db, err := schema.ExtractCredentials(a.ctx.GetRealDSN())
	if err != nil {
		return err
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
		"-n", a.ctx.SchemaDBDumpSchema,
		"-s",
		"-x",
		"-O",
		"--no-comments",
		"--no-owner",
		"--no-privileges",
		"--no-tablespaces",
		"--no-security-labels",
	}

	if !ignoreSchemaVersionTable {
		args = append(args, "-T="+ignoreTableName)
	}

	env := map[string]string{"PGPASSWORD": db.Pass}

	stdout, stderr, err := cmdexec.Exec(a.ctx.ShellPath, []string{"-c", strings.Join(args, " ")}, env)
	if err != nil {
		return fmt.Errorf("unable to dump database: %w\n%s", err, stderr)
	}

	// Generate extension statements
	extensionsToAdd := autoDetectExtensions(stdout)
	var extensionStatements strings.Builder
	extensionStatements.WriteString("\n-- Create extensions if they don't exist\n")
	for _, ext := range extensionsToAdd {
		extensionStatements.WriteString(fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS \"%s\";\n", ext))
	}
	extensionStatements.WriteString("\n")

	dateGenerated := fmt.Sprintf("-- Generated at: %s\n", time.Now().Format(time.RFC3339))

	schemaPattern := fmt.Sprintf(
		`(?s)(.*?)(?:--\s*\n--\s*Name:\s*%s;\s*Type:\s*SCHEMA.*?\n--\s*\n\s*CREATE\s+SCHEMA\s+%s;\s*\n)(.*)`,
		regexp.QuoteMeta(a.ctx.SchemaDBDumpSchema),
		regexp.QuoteMeta(a.ctx.SchemaDBDumpSchema),
	)

	dumpParts := regexp.MustCompile(schemaPattern).FindStringSubmatch(stdout)
	if len(dumpParts) != 3 {
		return fmt.Errorf("failed to parse schema dump: unexpected format")
	}

	setSchemaPath := fmt.Sprintf("SET search_path TO %s;\n", a.ctx.SchemaDBDumpSchema)

	// Combine all parts with the proper ordering
	result := dateGenerated +
		dumpParts[1] + // Content before schema
		setSchemaPath +
		extensionStatements.String() +
		dumpParts[2] // Content after schema

	writer.Write([]byte(result))

	return nil
}

func autoDetectExtensions(stdout string) []string {
	extensions := make(map[string]bool)

	// Common patterns that indicate extension usage
	patterns := map[string][]string{
		"uuid-ossp": {
			`uuid_generate_v`,
			`gen_random_uuid`,
		},
		"hstore": {
			`hstore_to_json`,
			`hstore_to_array`,
		},
		"postgis": {
			`geometry_columns`,
			`spatial_ref_sys`,
			`ST_`,
		},
		"pg_trgm": {
			`similarity`,
			`show_trgm`,
		},
		"pgcrypto": {
			`crypt(`,
			`gen_salt`,
		},
		"ltree": {
			`ltree_gist`,
			`lquery`,
		},
		"citext": {
			`citext_ops`,
		},
		"tablefunc": {
			`crosstab`,
			`normal_rand`,
		},
	}

	// Check each extension's patterns
	for ext, searchPatterns := range patterns {
		for _, pattern := range searchPatterns {
			if strings.Contains(stdout, pattern) {
				extensions[ext] = true
				break
			}
		}
	}

	// Convert map keys to sorted slice
	result := make([]string, 0, len(extensions))
	for ext := range extensions {
		result = append(result, ext)
	}
	sort.Strings(result)

	return result
}
