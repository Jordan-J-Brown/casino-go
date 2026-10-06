package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/Jordan-J-Brown/casino-go/pkg/casinoclient"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/stress-test/internal/casinosession"
	"github.com/Jordan-J-Brown/casino-go/stress-test/internal/games/mineshooks"
)

func main() {
	ctx := context.Background()

	client := casinoclient.DefaultClient

	logger := logsx.NewLogger(os.Stdout, logsx.Config{
		Type: logsx.LoggerTypeText,
		HandlerOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	})

	gameHooks := &mineshooks.Hooks{
		Mines:         3,
		IndexesToPick: []int{0, 1, 2, 3, 4, 5},
		Logger:        logger,
	}

	casinoSession, err := casinosession.New(ctx, "mines", 100_00, client, logger, gameHooks)
	if err != nil {
		panic("failed to create session: " + err.Error())
	}

	for range 1000 {
		if err := casinoSession.RunGame(ctx, 1_00); err != nil {
			panic("error from session: " + err.Error())
		}
	}
}
