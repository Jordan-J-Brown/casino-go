package main

import (
	"log/slog"
	"os"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/app"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/config"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
)

func main() {
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

	err = app.Run(app.AppConfig{
		ServiceName:    cfg.ServiceName,
		ServiceVersion: cfg.ServiceVersion,
		NatsCfg:        &cfg.Nats,
		TracerCfg:      &cfg.Tracer,
		Logger:         logger,
		Config:         cfg,
	})
	if err != nil {
		panic(err)
	}
}
