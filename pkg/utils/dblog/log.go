package dblog

import (
	"context"
	"fmt"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/alexisvisco/mig/pkg/utils/tracker"
	"github.com/charmbracelet/lipgloss"
	sqldblogger "github.com/simukti/sqldb-logger"
	"log/slog"
	"strings"
)

var (
	magenta = lipgloss.Color("#EC96EB")
	blue    = lipgloss.Color("#11DAF9")
)

type Logger struct {
	l       *slog.Logger
	record  bool
	queries []string
	params  [][]any
	tracker tracker.Tracker
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

	//attrs := make([]slog.Attr, 0, len(data))
	//for k, v := range data {
	//	attrs = append(attrs, slog.Any(k, v))
	//}

	//var lvl slog.Level
	//switch level {
	//case sqldblogger.LevelTrace:
	//	lvl = slog.LevelDebug - 1
	//	attrs = append(attrs, slog.Any("LOG_LEVEL", level))
	//case sqldblogger.LevelDebug:
	//	lvl = slog.LevelDebug
	//case sqldblogger.LevelInfo:
	//	lvl = slog.LevelInfo
	//case sqldblogger.LevelError:
	//	lvl = slog.LevelError
	//default:
	//	lvl = slog.LevelError
	//	attrs = append(attrs, slog.Any("LOG_LEVEL", level))
	//}

	mayDuration := data["duration"]
	mayQuery := data["query"]
	mayArgs := data["args"]

	s := &strings.Builder{}

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

	fmt.Println(s.String())
}
