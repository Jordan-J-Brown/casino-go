package tests

import (
	"context"
	"crypto/rand"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/gamemodels"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/tests/integrations"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/tests/integrations/casinotestclient"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/tests/integrations/mockengine"
	"github.com/Jordan-J-Brown/casino-go/casino/pkg/requests"
	"github.com/Jordan-J-Brown/casino-go/pkg/chans"
	"github.com/Jordan-J-Brown/casino-go/pkg/config"
	"github.com/Jordan-J-Brown/casino-go/pkg/errorx"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/pkg/msg/enginemsg"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCasino_Two_Concurrent_Plays(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	t.Cleanup(cancel)

	env, err := setup(ctx, t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(env.Shutdown)

	wg := new(sync.WaitGroup)
	errChan := make(chan error, 2)
	for range 2 {
		wg.Go(func() {
			(&gameWorker{
				sessionID: env.SessionID,
				client:    casinotestclient.NewDefaultClient(),
				errs:      errChan,
				steps:     []gameStep{stepPlay},
			}).run()
		})
	}
	wg.Wait()

	close(errChan)
	errs := chans.Collect(errChan)
	if !assert.Len(t, errs, 1) {
		return
	}

	errStr := errs[0].Error()
	if errStr != string(gamemodels.ErrTooManyRequests) && errStr != string(gamemodels.ErrGameInProgress) {
		t.Error("unexpected error type returned:", err)
	}
}

func TestCasino_Two_Concurrent_Next_Actions(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	t.Cleanup(cancel)

	env, err := setup(ctx, t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(env.Shutdown)

	res, err := casinotestclient.DefaultClient.Play(requests.Play{
		SessionID: env.SessionID,
		GameID:    mockengine.EngineID,
		Bet:       100,
		GameData:  []byte(`{}`),
	})
	if err != nil {
		t.Fatal("failed to send play request:", err)
	}
	if res.IsError {
		t.Fatal("error from play request:", res.Error.Error())
	}

	wg := new(sync.WaitGroup)
	errChan := make(chan error, 2)
	for range 2 {
		wg.Go(func() {
			(&gameWorker{
				sessionID: env.SessionID,
				client:    casinotestclient.NewDefaultClient(),
				errs:      errChan,
				steps:     []gameStep{stepNext},
			}).run()
		})
	}
	wg.Wait()

	close(errChan)
	errs := chans.Collect(errChan)
	if !assert.Len(t, errs, 1) {
		return
	}

	errStr := errs[0].Error()
	if errStr != string(gamemodels.ErrTooManyRequests) {
		t.Error("unexpected error type returned:", err)
	}
}

func TestCasino_Two_Concurrent_Close_Actions(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	t.Cleanup(cancel)

	env, err := setup(ctx, t)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(env.Shutdown)

	res, err := casinotestclient.DefaultClient.Play(requests.Play{
		SessionID: env.SessionID,
		GameID:    mockengine.EngineID,
		Bet:       100,
		GameData:  []byte(`{}`),
	})
	if err != nil {
		t.Fatal("failed to send play request:", err)
	}
	if res.IsError {
		t.Fatal("error from play request:", res.Error.Error())
	}

	wg := new(sync.WaitGroup)
	errChan := make(chan error, 2)
	for range 2 {
		wg.Go(func() {
			(&gameWorker{
				sessionID: env.SessionID,
				client:    casinotestclient.NewDefaultClient(),
				errs:      errChan,
				steps:     []gameStep{stepClose},
			}).run()
		})
	}
	wg.Wait()

	balanceRes, err := casinotestclient.DefaultClient.Balance(requests.Balance{
		SessionID: env.SessionID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if balanceRes.IsError {
		t.Fatal(balanceRes.Error.Error())
	}
	if !assert.Equal(t, int64(1000), balanceRes.Result.Balance) {
		return
	}

	close(errChan)
	errs := chans.Collect(errChan)
	if !assert.Len(t, errs, 1) {
		return
	}

	errStr := errs[0].Error()
	if errStr != string(gamemodels.ErrTooManyRequests) && errStr != string(gamemodels.ErrGameNotFound) {
		t.Error("unexpected error type returned:", err)
	}
}

type environment struct {
	*integrations.Environment
	Engine    *mockengine.Engine
	SessionID string
	Username  string
}

func (env *environment) Shutdown() {
	env.CasinoApp.Shutdown()
	env.Engine.Shutdown()
}

func setup(ctx context.Context, t *testing.T) (*environment, error) {
	env, err := integrations.Bootstrap(ctx, integrations.SetupConfig{
		T:            t,
		CasinoConfig: integrations.DefaultCasinoConfig(),
	})
	if err != nil {
		return nil, err
	}

	engine, err := setupEngine(ctx, t)
	if err != nil {
		return nil, err
	}

	username := "player-" + rand.Text()[:8]

	authResp, err := casinotestclient.DefaultClient.Auth(requests.Auth{
		Username: username,
		Balance:  1000,
	})
	if err != nil {
		return nil, err
	}
	if authResp.IsError {
		return nil, &authResp.Error
	}

	return &environment{
		Environment: env,
		Engine:      engine,
		SessionID:   authResp.Result.SessionID,
		Username:    username,
	}, nil
}

func setupEngine(ctx context.Context, t *testing.T) (*mockengine.Engine, error) {
	return mockengine.Start(ctx, t, mockengine.Config{
		Logging: config.Logging{
			Type:  logsx.LoggerTypeText,
			Level: slog.LevelDebug,
		},
		ServerConfig: enginemsg.ServerConfig{
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
		},
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
	client    *casinotestclient.Client
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
