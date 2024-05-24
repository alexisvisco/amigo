package amigo

import (
	"github.com/alexisvisco/amigo/pkg/utils/logger"
	"io"
	"log/slog"
)

func (a Amigo) SetupSlog(writer io.Writer) {
	logger.ShowSQLEvents = a.ctx.ShowSQL

	if a.ctx.JSON {
		slog.SetDefault(slog.New(slog.NewJSONHandler(writer, nil)))
	} else {
		slog.SetDefault(slog.New(logger.NewHandler(writer, nil)))
	}

	level := slog.LevelInfo
	if a.ctx.Debug {
		level = slog.LevelDebug
	}

	slog.SetLogLoggerLevel(level)
}
