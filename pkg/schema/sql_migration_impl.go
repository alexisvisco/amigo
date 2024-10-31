package schema

import (
	"bytes"
	"embed"
	"fmt"
	"time"
)

type SQLMigration[T Schema] struct {
	fs        embed.FS
	name      string
	time      string // RFC3339
	delimiter string
}

func NewSQLMigration[T Schema](fs embed.FS, name string, time string, delimiter string) DetailedMigration[T] {
	return &SQLMigration[T]{fs: fs, name: name, time: time, delimiter: delimiter}
}

func (s SQLMigration[T]) Up(x T) {
	up, _, err := s.parseContent()

	if err != nil {
		panic(err)
	}

	x.Exec(up)
}

func (s SQLMigration[T]) Down(x T) {
	_, down, err := s.parseContent()

	if err != nil {
		panic(err)
	}

	x.Exec(down)
}

func (s SQLMigration[T]) Name() string {
	return s.name
}

func (s SQLMigration[T]) Date() time.Time {
	t, _ := time.Parse(time.RFC3339, s.time)
	return t
}

func (s SQLMigration[T]) parseContent() (string, string, error) {
	file, err := s.fs.ReadFile(s.name)
	if err != nil {
		return "", "", fmt.Errorf("unable to read file %s: %w", s.name, err)
	}

	split := bytes.Split(file, []byte(s.delimiter))
	if len(split) != 2 {
		return "", "", fmt.Errorf("invalid content, expected 2 parts, got %d", len(split))
	}

	return string(split[0]), string(split[1]), nil
}
