package schema

import "regexp"

func isTableDoesNotExists(err error) bool {
	if err == nil {
		return false
	}

	re := []*regexp.Regexp{
		regexp.MustCompile(`Error 1146 \(42S02\): Table '.*' doesn't exist`),
		regexp.MustCompile(`ERROR: relation ".*" does not exist \(SQLSTATE 42P01\)`),
		regexp.MustCompile(`no such table: .*`),
		regexp.MustCompile(`.*does not exist \(SQLSTATE=42P01\).*`),
	}

	for _, r := range re {
		if r.MatchString(err.Error()) {
			return true
		}
	}

	return false
}
