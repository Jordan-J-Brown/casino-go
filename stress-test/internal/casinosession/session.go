package casinosession

import (
	"context"
	"encoding/json"

	"github.com/Jordan-J-Brown/casino-go/casino/pkg/requests"
	"github.com/Jordan-J-Brown/casino-go/casino/pkg/responses"
	"github.com/Jordan-J-Brown/casino-go/pkg/casinoclient"
	"github.com/Jordan-J-Brown/casino-go/pkg/errorx"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/shopspring/decimal"
)

type Session struct {
	client    *casinoclient.Client
	logger    logsx.Logger
	hooks     GameHooks
	engineID  string
	sessionID string

	username  string
	balance   int64
	payout    decimal.Decimal
	lastRound *responses.Round
	state     GameState
}

func New(
	ctx context.Context,
	engineID string,
	balance int64,
	client *casinoclient.Client,
	logger logsx.Logger,
	hooks GameHooks,
) (*Session, error) {
	username := generateUsername()

	authRes, err := client.Auth(requests.Auth{
		Username: username,
		Balance:  balance,
	})
	if err != nil {
		return nil, err
	}
	if authRes.IsError {
		return nil, &authRes.Error
	}

	setupRes, err := client.Setup(requests.Setup{
		SessionID: authRes.Result.SessionID,
		GameID:    engineID,
		MetaData:  nil,
	})
	if err != nil {
		return nil, err
	}
	if setupRes.IsError {
		return nil, &setupRes.Error
	}

	if err := hooks.ProcessSetup(ctx, setupRes.Result); err != nil {
		return nil, err
	}

	return &Session{
		client:    client,
		logger:    logger,
		hooks:     hooks,
		engineID:  engineID,
		sessionID: authRes.Result.SessionID,
		username:  username,
		balance:   balance,
	}, nil
}

func (session *Session) RunGame(ctx context.Context, bet int64) error {
	round, err := session.Play(ctx, bet)
	if err != nil {
		return err
	}

	if err := session.hooks.ParseRound(ctx, GameStatePlay, round); err != nil {
		return err
	}

	for !session.hooks.GameIsClosed(ctx) {
		round, err = session.Next(ctx)
		if err != nil {
			return err
		}

		if err := session.hooks.ParseRound(ctx, GameStateNext, round); err != nil {
			return err
		}
	}

	_, err = session.Close(ctx)
	if err != nil {
		return err
	}

	session.hooks.Reset(ctx)

	return nil
}

func (session *Session) Play(ctx context.Context, bet int64) (*responses.Round, error) {
	if session.state != GameStateReady && session.state != GameStateClosed {
		return nil, errorx.Newf("invalid game state: ")
	}

	gameData, err := session.hooks.GetPlayGameData(ctx)
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
	session.state = GameStatePlay

	return &r.Result, nil
}

func (session *Session) Next(ctx context.Context) (*responses.Round, error) {
	gameData, err := session.hooks.GetNextGameData(ctx)
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
	session.state = GameStateNext

	return &r.Result, nil
}

func (session *Session) Close(ctx context.Context) (*responses.Close, error) {
	if !session.hooks.GameIsClosed(ctx) {
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
	session.state = GameStateClosed

	return &r.Result, nil
}

type GameHooks interface {
	ProcessSetup(ctx context.Context, setup responses.Setup) error
	GetPlayGameData(ctx context.Context) (json.RawMessage, error)
	ParseRound(ctx context.Context, state GameState, round *responses.Round) error

	//All methods below will be called after ParseRound
	GameIsClosed(ctx context.Context) bool
	GetNextGameData(ctx context.Context) (json.RawMessage, error)
	Reset(ctx context.Context)
}
