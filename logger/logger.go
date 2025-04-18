package logger

import (
	"context"
	"log/slog"
	"os"
)

type ctxLogger struct{}

func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxLogger{}, l)
}

func getLoggerFromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxLogger{}).(*slog.Logger); ok {
		return l
	}
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func Debug(ctx context.Context, msg string, args ...any) {
	getLoggerFromContext(ctx).Debug(msg, args...)
}

func Info(ctx context.Context, msg string, args ...any) {
	getLoggerFromContext(ctx).Info(msg, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	getLoggerFromContext(ctx).Warn(msg, args...)
}

func Error(ctx context.Context, msg string, args ...any) {
	getLoggerFromContext(ctx).Error(msg, args...)
}
