package amigo

import (
	"io"
	"log/slog"

	"github.com/alexisvisco/amigo/pkg/utils/logger"
)

func (a Amigo) SetupSlog(writer io.Writer, mayLogger *slog.Logger) {
	logger.ShowSQLEvents = a.Config.ShowSQL
	if writer == nil && mayLogger == nil {
		logger.Logger = slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: slog.LevelError}))
		return
	}

	if mayLogger != nil {
		logger.Logger = mayLogger
		return
	}

	level := slog.LevelInfo
	if a.Config.Debug {
		level = slog.LevelDebug
	}

	if a.Config.JSON {
		logger.Logger = slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: level}))
	} else {
		logger.Logger = slog.New(logger.NewHandler(writer, &logger.Options{Level: level}))
	}
}
