package dblog

import (
	"context"
	"fmt"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	"github.com/charmbracelet/lipgloss"
	sqldblogger "github.com/simukti/sqldb-logger"
	"strings"
)

var (
	magenta = lipgloss.Color("#EC96EB")
	blue    = lipgloss.Color("#11DAF9")
)

type DatabaseLogger interface {
	Log(context.Context, sqldblogger.Level, string, map[string]interface{})
	Record(func()) string
	SetRecord(bool)
	FormatRecords() string
	ToggleLogger(bool)
}

type Logger struct {
	record  bool
	log     bool
	queries []string
	params  [][]any
}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) Record(f func()) string {
	l.record = true
	l.queries = nil
	l.params = nil
	f()
	l.record = false

	str := l.FormatRecords()

	return str
}

func (l *Logger) ToggleLogger(b bool) {
	l.log = b
}

func (l *Logger) FormatRecords() string {
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

func (l *Logger) Log(_ context.Context, _ sqldblogger.Level, _ string, data map[string]interface{}) {
	if !l.log {
		return
	}

	if log, ok := data["query"]; ok && l.record {
		l.queries = append(l.queries, log.(string))

		if args, ok := data["args"]; ok {
			l.params = append(l.params, args.([]any))
		} else {
			l.params = append(l.params, nil)
		}
	}

	mayDuration := data["duration"]
	mayQuery := data["query"]
	mayArgs := data["args"]

	s := &strings.Builder{}

	if mayQuery == nil || mayQuery.(string) == "" {
		return
	}

	durLenght := 0
	if mayDuration != nil {
		str := fmt.Sprintf("(%v) ", mayDuration)
		render := lipgloss.NewStyle().Foreground(magenta).SetString(str)
		s.WriteString(render.String())
		durLenght = len(str)
	}

	if mayQuery != nil {
		if err := quick.Highlight(s, strings.ReplaceAll(mayQuery.(string), "\n", " "), "sql", "terminal256",
			"native"); err != nil {
			return
		}
	}

	if mayArgs != nil {
		s.WriteString("\n")
		s.WriteString(lipgloss.NewStyle().PaddingLeft(durLenght).Foreground(blue).Render(fmt.Sprintf("args: %v",
			mayArgs)))
	}

	logger.Info(events.SQLQueryEvent{Query: s.String()})
}
