package config

import (
	"log/slog"

	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/spf13/viper"
)

const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

type Logging struct {
	Type  logsx.LoggerType
	Level slog.Level
}

func (l *Logging) load() error {
	loggerType := viper.Get("logging.type")
	switch loggerType {
	case "text":
		l.Type = logsx.LoggerTypeText
	default:
		l.Type = logsx.LoggerTypeJSON
	}

	level := viper.GetString("logging.level")
	switch level {
	case LogLevelDebug:
		l.Level = slog.LevelDebug
	case LogLevelInfo:
		l.Level = slog.LevelInfo
	case LogLevelWarn:
		l.Level = slog.LevelWarn
	default:
		l.Level = slog.LevelError
	}

	return nil
}
