package amigo

import (
	"fmt"
	"time"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

// cliOutput provides helper methods for formatted CLI output
type cliOutput struct{}

// newCLIOutput creates a new cliOutput helper
func newCLIOutput() *cliOutput {
	return &cliOutput{}
}

// path formats a file path in cyan
func (o *cliOutput) path(path string) string {
	return colorCyan + path + colorReset
}

// duration formats a duration in yellow (in milliseconds)
func (o *cliOutput) duration(d time.Duration) string {
	ms := d.Milliseconds()
	return colorYellow + fmt.Sprintf("%dms", ms) + colorReset
}

// error formats an error message in red
func (o *cliOutput) error(msg string) string {
	return colorRed + msg + colorReset
}

// timestamp formats a timestamp in green
func (o *cliOutput) timestamp(t time.Time) string {
	return colorGreen + t.Format("2006-01-02 15:04:05") + colorReset
}

// timestampNow formats the current time in green
func (o *cliOutput) timestampNow() string {
	return o.timestamp(time.Now())
}

// date formats a date/timestamp integer (YYYYMMDDHHMMSS) in green
func (o *cliOutput) date(date int64) string {
	return colorGreen + fmt.Sprintf("%d", date) + colorReset
}
