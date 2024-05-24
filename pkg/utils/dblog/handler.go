package dblog

import (
	"context"
	"fmt"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/alexisvisco/amigo/pkg/utils/colors"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	sqldblogger "github.com/simukti/sqldb-logger"
	"strings"
	"time"
)

type DatabaseLogger interface {
	Log(context.Context, sqldblogger.Level, string, map[string]interface{})
	Record(func()) string
	SetRecord(bool)
	FormatRecords() string
	ToggleLogger(bool)
}

type Handler struct {
	record             bool
	log                bool
	queries            []string
	params             [][]any
	syntaxHighlighting bool
}

func NewHandler(syntaxHighlighting bool) *Handler {
	return &Handler{
		syntaxHighlighting: syntaxHighlighting,
	}
}

func (l *Handler) Record(f func()) string {
	l.record = true
	l.queries = nil
	l.params = nil
	f()
	l.record = false

	str := l.FormatRecords()

	return str
}

func (l *Handler) ToggleLogger(b bool) {
	l.log = b
}

func (l *Handler) FormatRecords() string {
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

func (l *Handler) Reset() {
	l.queries = nil
	l.params = nil
}

func (l *Handler) SetRecord(v bool) {
	l.record = v
}

func (l *Handler) Log(_ context.Context, _ sqldblogger.Level, _ string, data map[string]interface{}) {
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

	if l.syntaxHighlighting {
		durLenght := 0
		if mayDuration != nil {
			str := fmt.Sprintf("(%s) ", time.Millisecond*time.Duration(mayDuration.(float64)))
			s.WriteString(colors.Magenta(str))
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
			padding := strings.Repeat(" ", durLenght)
			s.WriteString(colors.Teal(padding + fmt.Sprintf("args: %v", mayArgs)))
		}

	} else {
		if mayDuration != nil {
			s.WriteString(fmt.Sprintf("(%s) ", time.Millisecond*time.Duration(mayDuration.(float64))))
		}

		if mayQuery != nil {
			s.WriteString(mayQuery.(string))
		}

		if mayArgs != nil {
			s.WriteString(fmt.Sprintf("\nargs: %v", mayArgs))
		}
	}

	logger.Info(events.SQLQueryEvent{Query: s.String()})
}
