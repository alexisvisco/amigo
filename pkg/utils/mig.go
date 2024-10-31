package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gobuffalo/flect"
)

func ParseMigrationVersion(f string) (string, error) {
	if TimeRegexp.MatchString(f) {
		return f, nil
	}

	if MigrationFileRegexp.MatchString(f) {
		// get the prefix and remove underscore
		return strings.ReplaceAll(MigrationFileRegexp.FindStringSubmatch(f)[1], "_", ""), nil
	}

	return "", errors.New("invalid version format, should be of form: 20060102150405_migration_name.{go,sql}, 20060102150405")
}

var MigrationFileRegexp = regexp.MustCompile(`(\d{14})_(.*)\.(go|sql)`)
var TimeRegexp = regexp.MustCompile(`\d{14}`)

const FormatTime = "20060102150405"

func MigrationStructName(t time.Time, name string) string {
	return fmt.Sprintf("Migration%s%s", t.UTC().Truncate(time.Second).Format(FormatTime), flect.Pascalize(name))
}

func MigrationFileFormat(t time.Time, name string) string {
	return fmt.Sprintf("%s_%s.go", t.UTC().Truncate(time.Second).Format(FormatTime), flect.Underscore(name))
}
