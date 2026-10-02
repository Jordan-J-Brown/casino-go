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
	With(args ...any) Logger
}

type LoggerType string

const (
	LoggerTypeJSON LoggerType = "json"
	LoggerTypeText LoggerType = "text"
)

type Config struct {
	Type        LoggerType
	HandlerOpts *slog.HandlerOptions
}

type logger struct {
	*slog.Logger
}

func NewLogger(w io.Writer, cfg Config) Logger {
	var handler slog.Handler
	if cfg.Type == LoggerTypeText {
		handler = NewTextHandler(w, cfg.HandlerOpts)
	} else {
		//TODO: add custom JSON handler
		handler = slog.NewJSONHandler(w, cfg.HandlerOpts)
	}

	return &logger{
		Logger: slog.New(handler),
	}
}

func (l *logger) With(args ...any) Logger {
	return &logger{l.Logger.With(args...)}
}

type NopLogger struct{}

func (n NopLogger) InfoContext(ctx context.Context, msg string, args ...any)  {}
func (n NopLogger) WarnContext(ctx context.Context, msg string, args ...any)  {}
func (n NopLogger) ErrorContext(ctx context.Context, msg string, args ...any) {}
func (n NopLogger) DebugContext(ctx context.Context, msg string, args ...any) {}
func (n NopLogger) With(args ...any) Logger {
	return n
}
