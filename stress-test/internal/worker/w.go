package worker

import (
	"context"

	"github.com/Jordan-J-Brown/casino-go/pkg/casino/casinoclient"
	"github.com/Jordan-J-Brown/casino-go/pkg/chans"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/stress-test/internal/casinosession"
	"github.com/Jordan-J-Brown/casino-go/stress-test/internal/games"
)

type W struct {
	cfg      Config
	outbound chan<- OutboundMsg
	stop     chan struct{}
}

type Config struct {
	Client         *casinoclient.Client
	Logger         logsx.Logger
	EngineID       string
	GameCount      int
	Bet            int64
	Balance        int64
	GameController games.Controller
	Outbound       chan<- OutboundMsg
}

func New(ctx context.Context, cfg Config) *W {
	if cfg.Outbound == nil {
		panic("outbound channel is nil")
	}

	return &W{
		cfg:      cfg,
		stop:     make(chan struct{}, 1),
		outbound: cfg.Outbound,
	}
}

func (worker *W) Run(ctx context.Context) {
	worker.cfg.Logger.DebugContext(ctx, "worker started", "game_count", worker.cfg.GameCount)

	session, err := casinosession.New(ctx, casinosession.Config{
		EngineID:       worker.cfg.EngineID,
		Balance:        worker.cfg.Balance,
		Client:         worker.cfg.Client,
		Logger:         worker.cfg.Logger,
		GameController: worker.cfg.GameController,
	})
	if err != nil {
		worker.publishError(err)
		return
	}

	for range worker.cfg.GameCount {
		if _, stop := chans.TryRecv(worker.stop); stop || ctx.Err() != nil {
			break
		}

		err := session.RunGame(ctx, worker.cfg.Bet)
		if err != nil {
			worker.publishError(err)
			return
		}
	}

	worker.publishFinish()
}

func (worker *W) Stop() {
	worker.stop <- struct{}{}
}

func (worker *W) publishFinish() {
	worker.outbound <- OutboundMsg{
		Type: MessageTypeFinished,
	}
}

func (worker *W) publishError(err error) {
	worker.outbound <- OutboundMsg{
		Type:  MessageTypeError,
		Value: err,
	}
}
