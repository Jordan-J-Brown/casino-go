package logsx

import (
	"context"
	"io"
	"log/slog"
)

type Logger interface {
	InfoContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
	DebugContext(ctx context.Context, msg string, args ...any)
}

// TODO: create more in depth logger that adds correlation IDs and custom record handlers.
func NewLogger(w io.Writer) Logger {
	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

type NopLogger struct{}

func (n NopLogger) InfoContext(ctx context.Context, msg string, args ...any)  {}
func (n NopLogger) WarnContext(ctx context.Context, msg string, args ...any)  {}
func (n NopLogger) ErrorContext(ctx context.Context, msg string, args ...any) {}
func (n NopLogger) DebugContext(ctx context.Context, msg string, args ...any) {}
