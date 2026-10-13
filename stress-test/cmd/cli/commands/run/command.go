package run

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Jordan-J-Brown/casino-go/pkg/casino/casinoclient"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/pkg/ternary"
	"github.com/Jordan-J-Brown/casino-go/stress-test/internal/cli"
	"github.com/Jordan-J-Brown/casino-go/stress-test/internal/games"
	"github.com/Jordan-J-Brown/casino-go/stress-test/internal/stresstest"
)

type Command struct {
	GameID     string     `arg:"" help:"The ID of the game to test"`
	Games      int        `arg:"" help:"Number of games to complete"`
	Workers    int        `optional:"" default:"1" help:"Number of concurrent workers to use"`
	CasinoURL  string     `optional:"" name:"casino_url" help:"override the casino base URL"`
	LogConfig  LogConfig  `embed:"" prefix:"logs."`
	GameConfig GameConfig `required:"" embed:"" prefix:"game."`
}

type LogConfig struct {
	Level string `enum:"debug,info,warn,error" default:"info" help:"Set the minimum level for logs."`
}

func (l *LogConfig) GetLevel() slog.Level {
	switch l.Level {
	case "debug":
		return slog.LevelDebug
	case "error":
		return slog.LevelError
	case "warn":
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}

type GameConfig struct {
	Type games.Type        `required:"" help:"Game type to use (e.g. slot, mines etc)"`
	Args map[string]string `optional:"" help:"Optional: Game specific arguments. Example: --game.args=positions=0,1,2,3"`
}

func (cmd *Command) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	logger := logsx.NewLogger(os.Stdout, logsx.Config{
		Type: logsx.LoggerTypeText,
		HandlerOpts: &slog.HandlerOptions{
			Level: cmd.LogConfig.GetLevel(),
		},
	})

	const bet = int64(1_00)

	controllerFactory, err := cli.ParseArgs(ctx, cmd.GameConfig.Type, cmd.GameConfig.Args)
	if err != nil {
		return err
	}

	runner := stresstest.New(stresstest.Config{
		CasinoBaseURL:         ternary.Eval(cmd.CasinoURL == "", casinoclient.DefaultBaseURL, cmd.CasinoURL),
		Logger:                logger,
		EngineID:              cmd.GameID,
		WorkerCount:           cmd.Workers,
		GameCount:             cmd.Games,
		Bet:                   bet,
		GameControllerFactory: controllerFactory,
	})

	startTime := time.Now()
	go func() {
		err := runner.Run(ctx)
		if err != nil {
			logger.ErrorContext(ctx, "Error from stress test", "error", err)
		}
		stop()
	}()

	<-ctx.Done()

	duration := time.Since(startTime).Seconds()

	logger.DebugContext(ctx, "Metrics", "duration", duration)
	logger.DebugContext(ctx, "Stopping runner...")
	runner.Stop()

	return nil
}
