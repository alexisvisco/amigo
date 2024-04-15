package dblog

import (
	"context"
	"fmt"
	sqldblogger "github.com/simukti/sqldb-logger"
	"log/slog"
)

type Logger struct {
	l       *slog.Logger
	record  bool
	queries []string
	params  [][]any
}

func NewLogger(l *slog.Logger) *Logger {
	return &Logger{l: l}
}

func (l *Logger) Record(f func()) string {
	l.record = true
	l.queries = nil
	l.params = nil
	f()
	l.record = false

	str := l.String()

	return str
}

func (l *Logger) String() string {
	str := ""

	for i, query := range l.queries {
		str += query
		if l.params[i] != nil {
			str += "\n["
			for j, param := range l.params[i] {
				if j > 0 {
					str += ", "
				}
				str += fmt.Sprintf("%v", param)
			}
			str += "]\n"
		}
		str += "\n"
	}
	return str
}

func (l *Logger) Reset() {
	l.queries = nil
	l.params = nil
}

func (l *Logger) SetRecord(v bool) {
	l.record = v
}

func (l *Logger) Log(ctx context.Context, level sqldblogger.Level, msg string, data map[string]interface{}) {
	if log, ok := data["query"]; ok && l.record {
		l.queries = append(l.queries, log.(string))

		if args, ok := data["args"]; ok {
			l.params = append(l.params, args.([]any))
		} else {
			l.params = append(l.params, nil)
		}
	}

	attrs := make([]slog.Attr, 0, len(data))
	for k, v := range data {
		attrs = append(attrs, slog.Any(k, v))
	}

	var lvl slog.Level
	switch level {
	case sqldblogger.LevelTrace:
		lvl = slog.LevelDebug - 1
		attrs = append(attrs, slog.Any("LOG_LEVEL", level))
	case sqldblogger.LevelDebug:
		lvl = slog.LevelDebug
	case sqldblogger.LevelInfo:
		lvl = slog.LevelInfo
	case sqldblogger.LevelError:
		lvl = slog.LevelError
	default:
		lvl = slog.LevelError
		attrs = append(attrs, slog.Any("LOG_LEVEL", level))
	}
	l.l.LogAttrs(context.Background(), lvl, msg, attrs...)
}
