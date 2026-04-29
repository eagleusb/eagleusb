package util

import (
	"log/slog"
	"os"
)

func InitLogger(verbose bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})

	slog.SetDefault(slog.New(handler))
}

func Logger(feature string) *slog.Logger {
	return slog.Default().With("feature", feature)
}
