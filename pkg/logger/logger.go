package logger

import (
	"context"
	"log/slog"
	"os"
)

type contextKey string

const loggerContextKey contextKey = "request_logger"

func New(env string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	if env == "production" {
		return slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}

func WithContext(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, log)
}

func FromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}
	log, ok := ctx.Value(loggerContextKey).(*slog.Logger)
	if !ok || log == nil {
		return slog.Default()
	}
	return log
}
