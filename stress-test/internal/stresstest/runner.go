package stresstest

import (
	"context"
	"fmt"

	"github.com/Jordan-J-Brown/casino-go/pkg/casino/casinoclient"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/pkg/ternary"
	"github.com/Jordan-J-Brown/casino-go/stress-test/internal/games"
	"github.com/Jordan-J-Brown/casino-go/stress-test/internal/worker"
)

type Runner struct {
	cfg            Config
	workers        []*worker.W
	workerMessages chan worker.OutboundMsg
	errors         chan error
	finished       chan struct{}
}

type Config struct {
	CasinoBaseURL         string
	EngineID              string
	Logger                logsx.Logger
	WorkerCount           int
	GameCount             int
	Bet                   int64
	GameControllerFactory func() games.Controller
}

func New(cfg Config) *Runner {
	return &Runner{
		cfg:            cfg,
		workerMessages: make(chan worker.OutboundMsg, cfg.WorkerCount),
		errors:         make(chan error, cfg.WorkerCount),
		finished:       make(chan struct{}, 1),
	}
}

func (runner *Runner) Run(ctx context.Context) error {
	runner.initWorkers(ctx, &runner.cfg)
	go runner.processWorkerMessages(ctx)

	for _, worker := range runner.workers {
		go worker.Run(ctx)
	}

	select {
	case <-ctx.Done():
	case err := <-runner.errors:
		return err
	case <-runner.finished:
	}

	return nil
}

func (runner *Runner) Stop() {
	for _, worker := range runner.workers {
		worker.Stop()
	}
}

type workerConfig struct {
	gameCount int
	bet       int64
	balance   int64
}

func (runner *Runner) initWorkers(ctx context.Context, cfg *Config) {
	gameCount := cfg.GameCount / cfg.WorkerCount
	remaining := cfg.GameCount % cfg.WorkerCount
	configs := make([]workerConfig, cfg.WorkerCount)
	for i := range configs {
		gameCount := ternary.Eval(i == cfg.WorkerCount-1 && remaining > 0, remaining, gameCount)
		configs[i] = workerConfig{
			gameCount: gameCount,
			bet:       cfg.Bet,
			balance:   cfg.Bet * int64(gameCount),
		}
	}

	runner.workers = make([]*worker.W, len(configs))
	for i := range configs {
		runner.workers[i] = worker.New(ctx, worker.Config{
			Client:         casinoclient.New(runner.cfg.CasinoBaseURL),
			Logger:         runner.cfg.Logger,
			EngineID:       runner.cfg.EngineID,
			GameCount:      configs[i].gameCount,
			Bet:            configs[i].bet,
			Balance:        configs[i].balance,
			GameController: runner.cfg.GameControllerFactory(),
			Outbound:       runner.workerMessages,
		})
	}
}

func (runner *Runner) processWorkerMessages(ctx context.Context) {
	finished := 0
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-runner.workerMessages:
			switch msg.Type {
			case worker.MessageTypeError:
				runner.errors <- msg.Value.(error)
			case worker.MessageTypeFinished:
				runner.cfg.Logger.DebugContext(ctx, "worker finished")
				finished++
				if finished == runner.cfg.WorkerCount {
					runner.finished <- struct{}{}
					return
				}
			default:
				runner.errors <- fmt.Errorf("unknown message type: %T", msg.Type)
			}
		}
	}
}
