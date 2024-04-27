package cli

import (
	"fmt"
	"time"
)

type FileAddedEvent struct{ FileName string }

func (p FileAddedEvent) String() string { return fmt.Sprintf("+ file: %s", p.FileName) }

type FileModifiedEvent struct{ FileName string }

func (p FileModifiedEvent) String() string { return fmt.Sprintf("~ file: %s", p.FileName) }

type FolderAddedEvent struct{ FolderName string }

func (p FolderAddedEvent) String() string { return fmt.Sprintf("+ folder: %s", p.FolderName) }

type InfoEvent struct{ Message string }

func (p InfoEvent) String() string { return fmt.Sprintf("  %s", p.Message) }

type MeasurementEvent struct{ TimeElapsed time.Duration }

func (m MeasurementEvent) String() string { return fmt.Sprintf("  done in %s", m.TimeElapsed) }
