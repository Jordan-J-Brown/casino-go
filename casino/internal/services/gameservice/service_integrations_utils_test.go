//go:build integration

package gameservice_test

import (
	"context"
	"crypto/rand"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/tests/integrations"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/tests/integrations/mockengine"
	"github.com/Jordan-J-Brown/casino-go/casino/pkg/requests"
	"github.com/Jordan-J-Brown/casino-go/pkg/casinoclient"
	"github.com/Jordan-J-Brown/casino-go/pkg/config"
	"github.com/Jordan-J-Brown/casino-go/pkg/errorx"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/pkg/msg/enginemsg"
	"github.com/shopspring/decimal"
)

type environment struct {
	ctx context.Context
	*integrations.Environment
	Engine          *mockengine.Engine
	SessionID       string
	Username        string
	OriginalBalance int64
}

func (env *environment) Shutdown() {
	env.Engine.Shutdown()
	env.CasinoApp.Shutdown(env.ctx)
}

func setupEnv(ctx context.Context, t *testing.T, engineConfig ...enginemsg.ServerConfig) *environment {
	errorx.OutputStack = true

	baseEnv, err := integrations.Bootstrap(ctx, t, integrations.SetupConfig{
		T:            t,
		CasinoConfig: integrations.DefaultCasinoConfig(),
	})
	if err != nil {
		panic(err)
	}

	engine, err := setupEngine(ctx, baseEnv.NatsConnString, baseEnv.OtelExporterEndpoint, engineConfig...)
	if err != nil {
		panic(err)
	}
	t.Cleanup(engine.Shutdown)

	username := "player-" + rand.Text()[:8]

	balance := int64(1000)
	authResp, err := casinoclient.DefaultClient.Auth(requests.Auth{
		Username: username,
		Balance:  balance,
	})
	if err != nil {
		panic(err)
	}
	if authResp.IsError {
		panic(authResp.Error)
	}

	env := &environment{
		ctx:             ctx,
		Environment:     baseEnv,
		Engine:          engine,
		SessionID:       authResp.Result.SessionID,
		Username:        username,
		OriginalBalance: balance,
	}
	t.Cleanup(env.Shutdown)

	return env
}

func setupEngine(ctx context.Context, natsConnString, otlpEndpoint string, serverConfig ...enginemsg.ServerConfig) (*mockengine.Engine, error) {
	var srvConfig enginemsg.ServerConfig
	if len(serverConfig) == 1 {
		srvConfig = serverConfig[0]
	} else {
		srvConfig = enginemsg.ServerConfig{
			RegisterFunc: func(ctx context.Context) (*enginemsg.EngineRegistration, error) {
				return &enginemsg.EngineRegistration{
					MinBet: 10,
					MaxBet: 100,
				}, nil
			},
			SetupHandler: func(ctx context.Context, r enginemsg.SetupRequest) (*enginemsg.SetupResponse, *errorx.Error) {
				return &enginemsg.SetupResponse{
					GameConfig: []byte(`{}`),
					MetaData:   []byte(`{}`),
				}, nil
			},
			PlayHandler: func(ctx context.Context, r enginemsg.PlayRequest) (*enginemsg.RoundResponse, *errorx.Error) {
				return &enginemsg.RoundResponse{
					PayoutMultiplier: decimal.NewFromInt(1),
					PrivateState:     []byte(`{}`),
					PublicState:      []byte(`{}`),
				}, nil
			},
			NextHandler: func(ctx context.Context, r enginemsg.NextRequest) (*enginemsg.RoundResponse, *errorx.Error) {
				time.Sleep(10 * time.Millisecond) //Simulate a slow response so multiple next requests have chance to collide
				return &enginemsg.RoundResponse{
					PayoutMultiplier: decimal.NewFromInt(1),
					PrivateState:     []byte(`{}`),
					PublicState:      []byte(`{}`),
				}, nil
			},
			CloseHandler: func(ctx context.Context, r enginemsg.CloseRequest) (*enginemsg.CloseResponse, *errorx.Error) {
				time.Sleep(10 * time.Millisecond) //Simulate a slow response so multiple next requests have chance to collide
				return &enginemsg.CloseResponse{
					CanClose: true,
				}, nil
			},
		}
	}

	return mockengine.Start(ctx, mockengine.Config{
		Logging: config.Logging{
			Type:  logsx.LoggerTypeText,
			Level: slog.LevelDebug,
		},
		ServerConfig:   srvConfig,
		NatsConnString: natsConnString,
		OtlpEndpoint:   otlpEndpoint,
	})
}

type gameStep string

const (
	stepPlay  gameStep = "play"
	stepNext  gameStep = "next"
	stepClose gameStep = "close"
)

type gameWorker struct {
	sessionID string
	client    *casinoclient.Client
	errs      chan<- error
	steps     []gameStep
}

func (w *gameWorker) run() {
	for _, step := range w.steps {
		var err error
		switch step {
		case stepPlay:
			err = w.play()
		case stepNext:
			err = w.next()
		case stepClose:
			err = w.close()
		}
		if err != nil {
			w.errs <- err
			return
		}
	}
}

func (w *gameWorker) play() error {
	playRes, err := w.client.Play(requests.Play{
		SessionID: w.sessionID,
		GameID:    mockengine.EngineID,
		Bet:       100,
		GameData:  []byte(`{}`),
	})
	if err != nil {
		return err
	}

	if playRes.IsError {
		return errors.New(playRes.Error.Type)
	}

	return nil
}

func (w *gameWorker) next() error {
	nextRes, err := w.client.Next(requests.Next{
		SessionID: w.sessionID,
		GameID:    mockengine.EngineID,
		GameData:  []byte(`{}`),
	})
	if err != nil {
		return err
	}
	if nextRes.IsError {
		return errors.New(nextRes.Error.Type)
	}

	return nil
}

func (w *gameWorker) close() error {
	closeRes, err := w.client.Close(requests.Close{
		SessionID: w.sessionID,
		GameID:    mockengine.EngineID,
	})
	if err != nil {
		return err
	}

	if closeRes.IsError {
		return errors.New(closeRes.Error.Type)
	}

	return nil
}
