package logger

import (
	"log/slog"
	"os"
)

func parseLevel(s string) slog.Level {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(s)); err != nil {
		// fallback
		return slog.LevelInfo
	}
	return lvl
}

func InitLogger(level string, enabled bool, path string, withAttrs ...any) error {
	if !enabled {
		slog.SetDefault(slog.New(DiscardHandler{}).With(withAttrs...))
		return nil
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	opts := &slog.HandlerOptions{Level: parseLevel(level), AddSource: true}
	slog.SetDefault(
		slog.New(slog.NewTextHandler(file, opts)).With(withAttrs...),
	)

	return nil
}
