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

func InitLogger(level string, withAttrs ...any) {
	handler := slog.Handler(
		//slog.NewJSONHandler(
		//	os.Stdout,
		//	&slog.HandlerOptions{Level: parseLevel(level), AddSource: true},
		//),
		slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: parseLevel(level), AddSource: true},
		),
	)
	handler = NewHandlerMiddleware(handler)
	slog.SetDefault(slog.New(handler).With(withAttrs...))
}
