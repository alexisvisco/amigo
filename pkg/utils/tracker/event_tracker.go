package tracker

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"time"
)

type Event struct {
	Type  string
	Event fmt.Stringer
}

type EventName interface {
	EventName() string
}

type Tracker interface {
	AddEvent(event fmt.Stringer) Tracker
	Measure() Tracker
	io.Writer
}

type EventLogger struct {
	Events []Event

	writer io.Writer
	json   bool

	timeTracker time.Time
}

func NewLogger(json bool, writer io.Writer) *EventLogger {
	return &EventLogger{
		timeTracker: time.Now(),
		writer:      writer,
		json:        json,
	}
}

func (p *EventLogger) AddEvent(event fmt.Stringer) Tracker {
	name := reflect.TypeOf(event).Name()
	if en, ok := event.(EventName); ok {
		name = en.EventName()
	}

	p.Events = append(p.Events, Event{
		Type:  name,
		Event: event,
	})

	if p.writer != nil {
		str := event.String()
		if p.json {
			indent, _ := json.Marshal(p.Events)
			str = string(indent) + "\n"
		}

		// if last char is not a newline, add one
		if len(str) > 0 && str[len(str)-1] != '\n' {
			str += "\n"
		}

		_, _ = p.Write([]byte(str))
	}

	return p
}

func (p *EventLogger) Write(p0 []byte) (n int, err error) {
	return p.writer.Write(p0)
}
func (p *EventLogger) Measure() Tracker {
	p.AddEvent(MeasurementEvent{TimeElapsed: time.Since(p.timeTracker)})
	p.timeTracker = time.Now()
	return p
}
