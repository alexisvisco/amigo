package testutils

import "fmt"

// Recorder is an interface to record a function.
type Recorder interface {
	Record(f func()) string
	SetRecord(v bool)
	fmt.Stringer
}
