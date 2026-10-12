package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	casinoapp "github.com/Jordan-J-Brown/casino-go/casino/internal/casinoapp"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/config"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := logsx.NewLogger(os.Stdout, logsx.Config{
		Type: cfg.Logging.Type,
		HandlerOpts: &slog.HandlerOptions{
			Level: cfg.Logging.Level,
		},
	})

	app, err := casinoapp.Run(ctx, logger, cfg)
	if err != nil {
		panic(err)
	}
	defer app.Shutdown(ctx)

	<-ctx.Done()

	logger.DebugContext(ctx, "Shutting down...")
}
