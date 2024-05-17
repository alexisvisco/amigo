package amigo

import (
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	"io"
	"log/slog"
)

func SetupSlog(showSQL bool, debug bool, json bool, writer io.Writer) {
	logger.ShowSQLEvents = showSQL

	if json {
		slog.SetDefault(slog.New(slog.NewJSONHandler(writer, nil)))
	} else {
		slog.SetDefault(slog.New(logger.NewHandler(writer, nil)))
	}

	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	slog.SetLogLoggerLevel(level)
}
