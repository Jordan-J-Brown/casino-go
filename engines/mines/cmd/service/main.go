package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/Jordan-J-Brown/casino-go/engines/mines/internal/config"
	"github.com/Jordan-J-Brown/casino-go/engines/mines/internal/engine"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

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

	engine, err := engine.Start(ctx, cfg, logger)
	if err != nil {
		panic(err)
	}
	defer engine.Shutdown()

	logger.InfoContext(ctx, "Engine running...")

	<-ctx.Done()

	logger.InfoContext(ctx, "Shutting down...")
}
