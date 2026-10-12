package casinosession

import (
	"context"

	"github.com/Jordan-J-Brown/casino-go/casino/pkg/requests"
	"github.com/Jordan-J-Brown/casino-go/casino/pkg/responses"
	"github.com/Jordan-J-Brown/casino-go/pkg/casinoclient"
	"github.com/Jordan-J-Brown/casino-go/pkg/errorx"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/stress-test/internal/games"
	"github.com/shopspring/decimal"
)

type Session struct {
	client     *casinoclient.Client
	logger     logsx.Logger
	controller games.Controller
	engineID   string
	sessionID  string

	username  string
	balance   int64
	payout    decimal.Decimal
	lastRound *responses.Round
	state     games.GameState
}

type Config struct {
	EngineID       string
	Balance        int64
	Client         *casinoclient.Client
	Logger         logsx.Logger
	GameController games.Controller
}

func New(ctx context.Context, cfg Config) (*Session, error) {
	username := generateUsername()

	authRes, err := cfg.Client.Auth(requests.Auth{
		Username: username,
		Balance:  cfg.Balance,
	})
	if err != nil {
		return nil, err
	}
	if authRes.IsError {
		return nil, &authRes.Error
	}

	setupRes, err := cfg.Client.Setup(requests.Setup{
		SessionID: authRes.Result.SessionID,
		GameID:    cfg.EngineID,
		MetaData:  nil,
	})
	if err != nil {
		return nil, err
	}
	if setupRes.IsError {
		return nil, &setupRes.Error
	}

	if err := cfg.GameController.ProcessSetup(ctx, setupRes.Result); err != nil {
		return nil, err
	}

	return &Session{
		client:     cfg.Client,
		logger:     cfg.Logger,
		controller: cfg.GameController,
		engineID:   cfg.EngineID,
		sessionID:  authRes.Result.SessionID,
		username:   username,
		balance:    cfg.Balance,
	}, nil
}

func (session *Session) RunGame(ctx context.Context, bet int64) error {
	round, err := session.Play(ctx, bet)
	if err != nil {
		return err
	}

	if err := session.controller.ParseRound(ctx, games.GameStatePlay, round); err != nil {
		return err
	}

	for !session.controller.GameIsClosed(ctx) {
		round, err = session.Next(ctx)
		if err != nil {
			return err
		}

		if err := session.controller.ParseRound(ctx, games.GameStateNext, round); err != nil {
			return err
		}
	}

	_, err = session.Close(ctx)
	if err != nil {
		return err
	}

	session.controller.Reset(ctx)

	return nil
}

func (session *Session) Play(ctx context.Context, bet int64) (*responses.Round, error) {
	if session.state != games.GameStateReady && session.state != games.GameStateClosed {
		return nil, errorx.Newf("invalid game state: %s", session.state)
	}

	gameData, err := session.controller.GetPlayGameData(ctx)
	if err != nil {
		return nil, err
	}

	r, err := session.client.Play(requests.Play{
		SessionID: session.sessionID,
		GameID:    session.engineID,
		Bet:       bet,
		GameData:  gameData,
	})
	if err != nil {
		return nil, err
	}
	if r.IsError {
		return nil, &r.Error
	}

	payout, err := decimal.NewFromString(r.Result.PayoutMultiplier)
	if err != nil {
		return nil, err
	}

	session.balance = r.Result.Balance
	session.payout = payout
	session.lastRound = &r.Result
	session.state = games.GameStatePlay

	return &r.Result, nil
}

func (session *Session) Next(ctx context.Context) (*responses.Round, error) {
	gameData, err := session.controller.GetNextGameData(ctx)
	if err != nil {
		return nil, err
	}

	r, err := session.client.Next(requests.Next{
		SessionID: session.sessionID,
		GameID:    session.engineID,
		GameData:  gameData,
	})
	if err != nil {
		return nil, err
	}
	if r.IsError {
		return nil, &r.Error
	}

	payout, err := decimal.NewFromString(r.Result.PayoutMultiplier)
	if err != nil {
		return nil, err
	}

	session.balance = r.Result.Balance
	session.payout = payout
	session.lastRound = &r.Result
	session.state = games.GameStateNext

	return &r.Result, nil
}

func (session *Session) Close(ctx context.Context) (*responses.Close, error) {
	if !session.controller.GameIsClosed(ctx) {
		return nil, ErrGameNotClosed
	}

	r, err := session.client.Close(requests.Close{
		SessionID: session.sessionID,
		GameID:    session.engineID,
	})
	if err != nil {
		return nil, err
	}
	if r.IsError {
		return nil, &r.Error
	}

	session.balance = r.Result.Balance
	session.state = games.GameStateClosed

	return &r.Result, nil
}
