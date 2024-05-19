// Package logger is almost copy/paste from tint Handler but with some modificatoins
package logger

import (
	"context"
	"encoding"
	"fmt"
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"io"
	"log/slog"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"time"
	"unicode"
)

const errKey = "err"

var (
	defaultLevel      = slog.LevelInfo
	defaultTimeFormat = time.StampMilli
)

var ShowSQLEvents = false

// Options for a slog.Handler that writes tinted logs. A zero Options consists
// entirely of default values.
//
// Options can be used as a drop-in replacement for [slog.HandlerOptions].
type Options struct {
	// Enable source code location (Default: false)
	AddSource bool

	// Minimum level to log (Default: slog.LevelInfo)
	Level slog.Leveler

	// ReplaceAttr is called to rewrite each non-group attribute before it is logged.
	// See https://pkg.go.dev/log/slog#HandlerOptions for details.
	ReplaceAttr func(groups []string, attr slog.Attr) slog.Attr
}

// NewHandler creates a [slog.Handler] that writes tinted logs to Writer w,
// using the default options. If opts is nil, the default options are used.
func NewHandler(w io.Writer, opts *Options) *Handler {
	h := &Handler{
		w:          w,
		level:      defaultLevel,
		timeFormat: defaultTimeFormat,
	}
	if opts == nil {
		return h
	}

	h.addSource = opts.AddSource
	if opts.Level != nil {
		h.level = opts.Level
	}
	h.replaceAttr = opts.ReplaceAttr
	return h
}

// Handler implements a [slog.Handler].
type Handler struct {
	attrsPrefix string
	groupPrefix string
	groups      []string

	mu sync.Mutex
	w  io.Writer

	addSource   bool
	level       slog.Leveler
	replaceAttr func([]string, slog.Attr) slog.Attr
	timeFormat  string
	noColor     bool
}

func (h *Handler) clone() *Handler {
	return &Handler{
		attrsPrefix: h.attrsPrefix,
		groupPrefix: h.groupPrefix,
		groups:      h.groups,
		w:           h.w,
		addSource:   h.addSource,
		level:       h.level,
		replaceAttr: h.replaceAttr,
		timeFormat:  h.timeFormat,
		noColor:     h.noColor,
	}
}

func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *Handler) Handle(_ context.Context, r slog.Record) error {
	// get a buffer from the sync pool
	buf := newBuffer()
	defer buf.Free()

	rep := h.replaceAttr

	// write level
	if rep == nil {
		h.appendLevel(buf, r.Level)
		if r.Level == slog.LevelDebug || r.Level == slog.LevelError || r.Level == slog.LevelWarn {
			buf.WriteByte(' ')
		}
	} else if a := rep(nil /* groups */, slog.Any(slog.LevelKey, r.Level)); a.Key != "" {
		h.appendValue(buf, a.Value, false)
		buf.WriteByte(' ')
	}

	// write source
	if h.addSource {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		if f.File != "" {
			src := &slog.Source{
				Function: f.Function,
				File:     f.File,
				Line:     f.Line,
			}

			if rep == nil {
				h.appendSource(buf, src)
				buf.WriteByte(' ')
			} else if a := rep(nil /* groups */, slog.Any(slog.SourceKey, src)); a.Key != "" {
				h.appendValue(buf, a.Value, false)
				buf.WriteByte(' ')
			}
		}
	}

	// if attrs has an event and

	// write message
	if rep == nil {
		if r.Message != "" {
			buf.WriteString(r.Message)
			buf.WriteByte(' ')
		}
	} else if a := rep(nil /* groups */, slog.String(slog.MessageKey, r.Message)); a.Key != "" {
		h.appendValue(buf, a.Value, false)
		buf.WriteByte(' ')
	}

	// write Handler attributes
	if len(h.attrsPrefix) > 0 {
		buf.WriteString(h.attrsPrefix)
	}

	// write attributes
	r.Attrs(func(attr slog.Attr) bool {
		h.appendAttr(buf, attr, h.groupPrefix, h.groups)
		return true
	})

	if len(*buf) == 0 {
		return nil
	}
	(*buf)[len(*buf)-1] = '\n' // replace last space with newline

	h.mu.Lock()
	defer h.mu.Unlock()

	_, err := h.w.Write(*buf)
	return err
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	h2 := h.clone()

	buf := newBuffer()
	defer buf.Free()

	// write attributes to buffer
	for _, attr := range attrs {
		h.appendAttr(buf, attr, h.groupPrefix, h.groups)
	}
	h2.attrsPrefix = h.attrsPrefix + string(*buf)
	return h2
}

func (h *Handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	h2 := h.clone()
	h2.groupPrefix += name + "."
	h2.groups = append(h2.groups, name)
	return h2
}

func (h *Handler) appendTime(buf *buffer, t time.Time) {
	*buf = t.AppendFormat(*buf, h.timeFormat)
}

func (h *Handler) appendLevel(buf *buffer, level slog.Level) {
	switch {
	case level == slog.LevelDebug:
		buf.WriteString("debug:")
		appendLevelDelta(buf, level-slog.LevelDebug)
	case level == slog.LevelError:
		buf.WriteString("\u001B[31merror:\033[0m")
	case level == slog.LevelWarn:
		buf.WriteString("\u001B[33mwarning:\033[0m")
	}
}

func appendLevelDelta(buf *buffer, delta slog.Level) {
	if delta == 0 {
		return
	} else if delta > 0 {
		buf.WriteByte('+')
	}
	*buf = strconv.AppendInt(*buf, int64(delta), 10)
}

func (h *Handler) appendSource(buf *buffer, src *slog.Source) {
	dir, file := filepath.Split(src.File)

	buf.WriteString(filepath.Join(filepath.Base(dir), file))
	buf.WriteByte(':')
	buf.WriteString(strconv.Itoa(src.Line))
}

func (h *Handler) appendAttr(buf *buffer, attr slog.Attr, groupsPrefix string, groups []string) {
	if attr.Key != "event" {
		return
	}

	attr.Value = attr.Value.Resolve()
	if rep := h.replaceAttr; rep != nil && attr.Value.Kind() != slog.KindGroup {
		attr = rep(groups, attr)
		attr.Value = attr.Value.Resolve()
	}

	if attr.Equal(slog.Attr{}) {
		return
	}

	if attr.Value.Kind() == slog.KindGroup {
		if attr.Key != "" && attr.Key != "event" {
			groupsPrefix += attr.Key + "."
			groups = append(groups, attr.Key)
		}
		for _, groupAttr := range attr.Value.Group() {
			h.appendAttr(buf, groupAttr, groupsPrefix, groups)
		}
	} else if err, ok := attr.Value.Any().(tintError); ok {
		// append tintError
		h.appendTintError(buf, err, groupsPrefix)
		buf.WriteByte(' ')
	} else {
		if attr.Key != "event" {
			h.appendKey(buf, attr.Key, groupsPrefix)
		}
		h.appendValue(buf, attr.Value, true)
		buf.WriteByte(' ')
	}
}

func (h *Handler) appendKey(buf *buffer, key, groups string) {
	appendString(buf, groups+key, true)
	buf.WriteByte('=')
}

func (h *Handler) appendValue(buf *buffer, v slog.Value, quote bool) {
	switch v.Kind() {
	case slog.KindString:
		appendString(buf, v.String(), quote)
	case slog.KindInt64:
		*buf = strconv.AppendInt(*buf, v.Int64(), 10)
	case slog.KindUint64:
		*buf = strconv.AppendUint(*buf, v.Uint64(), 10)
	case slog.KindFloat64:
		*buf = strconv.AppendFloat(*buf, v.Float64(), 'g', -1, 64)
	case slog.KindBool:
		*buf = strconv.AppendBool(*buf, v.Bool())
	case slog.KindDuration:
		appendString(buf, v.Duration().String(), quote)
	case slog.KindTime:
		appendString(buf, v.Time().String(), quote)
	case slog.KindAny:
		switch cv := v.Any().(type) {
		case slog.Level:
			h.appendLevel(buf, cv)
		case encoding.TextMarshaler:
			data, err := cv.MarshalText()
			if err != nil {
				break
			}
			appendString(buf, string(data), quote)
		case fmt.Stringer:
			appendString(buf, cv.String(), quote)
		case *slog.Source:
			h.appendSource(buf, cv)
		default:
			appendString(buf, fmt.Sprintf("%+v", v.Any()), quote)
		}
	}
}

func (h *Handler) appendTintError(buf *buffer, err error, groupsPrefix string) {
	appendString(buf, groupsPrefix+errKey, true)
	buf.WriteByte('=')
	appendString(buf, err.Error(), true)
}

func appendString(buf *buffer, s string, _ bool) {
	buf.WriteString(s)
}

func needsQuoting(s string) bool {
	if len(s) == 0 {
		return true
	}
	for _, r := range s {
		if unicode.IsSpace(r) || r == '"' || r == '=' || !unicode.IsPrint(r) {
			return true
		}
	}
	return false
}

type tintError struct{ error }

// Err returns a tinted (colorized) [slog.Attr] that will be written in red color
// by the [tint.Handler]. When used with any other [slog.Handler], it behaves as
//
//	slog.Any("err", err)
func Err(err error) slog.Attr {
	if err != nil {
		err = tintError{err}
	}
	return slog.Any(errKey, err)
}

type buffer []byte

var bufPool = sync.Pool{
	New: func() any {
		b := make(buffer, 0, 1024)
		return &b
	},
}

func newBuffer() *buffer {
	return bufPool.Get().(*buffer)
}

func (b *buffer) Free() {
	// To reduce peak allocation, return only smaller buffers to the pool.
	const maxBufferSize = 16 << 10
	if cap(*b) <= maxBufferSize {
		*b = (*b)[:0]
		bufPool.Put(b)
	}
}
func (b *buffer) Write(bytes []byte) (int, error) {
	*b = append(*b, bytes...)
	return len(bytes), nil
}

func (b *buffer) WriteByte(char byte) error {
	*b = append(*b, char)
	return nil
}

func (b *buffer) WriteString(str string) (int, error) {
	*b = append(*b, str...)
	return len(str), nil
}

func (b *buffer) WriteStringIf(ok bool, str string) (int, error) {
	if !ok {
		return 0, nil
	}
	return b.WriteString(str)
}

func event(event any) *slog.Logger {
	name := reflect.TypeOf(event).Name()
	if en, ok := event.(events.EventName); ok {
		name = en.EventName()
	}

	return slog.With(slog.Any("event", event), slog.String("event_name", name))
}

func Info(evt any) {
	if !canLogEvent(evt) {
		return
	}
	event(evt).Info("")
}

func Error(evt any) {
	if !canLogEvent(evt) {
		return
	}
	event(evt).Error("")
}

func Debug(evt any) {
	if !canLogEvent(evt) {
		return
	}
	event(evt).Debug("")
}

func Warn(evt any) {
	if !canLogEvent(evt) {
		return
	}

	event(evt).Warn("")
}

func isSQLQueryEvent(event any) bool {
	_, ok := event.(events.SQLQueryEvent)
	return ok
}

func canLogEvent(event any) bool {
	if isSQLQueryEvent(event) && !ShowSQLEvents {
		return false
	}

	return true
}
