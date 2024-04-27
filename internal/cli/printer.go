package cli

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

type PrinterEvent struct {
	Type  string
	Event fmt.Stringer
}

type Printer struct {
	Events []PrinterEvent

	timeTracker time.Time
}

func NewPrinter() *Printer {
	return &Printer{
		timeTracker: time.Now(),
	}
}

func (p *Printer) AddEvent(event fmt.Stringer) *Printer {
	// get the name of the struct
	reflectType := reflect.TypeOf(event)

	p.Events = append(p.Events, PrinterEvent{
		Type:  reflectType.Name(),
		Event: event,
	})

	return p
}

func (p *Printer) Measure() *Printer {
	p.AddEvent(MeasurementEvent{TimeElapsed: time.Since(p.timeTracker)})
	p.timeTracker = time.Now()
	return p
}

func (p *Printer) Print(json bool) {
	if json {
		p.JSON()
		return
	}
	str := ""
	for _, event := range p.Events {
		str += event.Event.String() + "\n"
	}

	fmt.Print(str)
}

func (p *Printer) JSON() {
	indent, _ := json.MarshalIndent(p, "", "  ")
	fmt.Println(string(indent))
}
