package utils

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
)

func GetModuleName() (string, error) {
	goModBytes, err := os.ReadFile("go.mod")
	if err != nil {
		return "", fmt.Errorf("failed to read go.mod file: %w", err)
	}

	modName := modulePath(goModBytes)

	return modName, nil
}

var (
	slashSlash = []byte("//")
	moduleStr  = []byte("module")
)

// modulePath returns the module path from the gomod file text.
// If it cannot find a module path, it returns an empty string.
// It is tolerant of unrelated problems in the go.mod file.
func modulePath(mod []byte) string {
	for len(mod) > 0 {
		line := mod
		mod = nil
		if i := bytes.IndexByte(line, '\n'); i >= 0 {
			line, mod = line[:i], line[i+1:]
		}
		if i := bytes.Index(line, slashSlash); i >= 0 {
			line = line[:i]
		}
		line = bytes.TrimSpace(line)
		if !bytes.HasPrefix(line, moduleStr) {
			continue
		}
		line = line[len(moduleStr):]
		n := len(line)
		line = bytes.TrimSpace(line)
		if len(line) == n || len(line) == 0 {
			continue
		}

		if line[0] == '"' || line[0] == '`' {
			p, err := strconv.Unquote(string(line))
			if err != nil {
				return "" // malformed quoted string or multiline module path
			}
			return p
		}

		return string(line)
	}
	return "" // missing module path
}
