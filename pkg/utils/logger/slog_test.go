package logger

import (
	"github.com/alexisvisco/amigo/pkg/utils/events"
	"log/slog"
	"os"
	"testing"
)

func TestLog(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	evt := events.FileAddedEvent{FileName: "test.txt"}

	Info(evt)
}
