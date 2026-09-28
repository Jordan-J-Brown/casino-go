package main

import (
	"log/slog"
	"os"

	"github.com/Jordan-J-Brown/casino-go/engines/mines/internal/config"
	"github.com/Jordan-J-Brown/casino-go/engines/mines/internal/engine"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	logger := logsx.NewLogger(os.Stdout, logsx.Config{
		Type: cfg.Logging.Type,
		HandlerOpts: &slog.HandlerOptions{
			Level: cfg.Logging.Level,
		},
	})

	engine.Start(cfg, logger)
}
