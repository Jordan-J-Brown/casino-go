package tests

import (
	"context"
	"crypto/rand"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
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

func TestCasinoTwoConcurrentPlays(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	env, err := integrations.Bootstrap(ctx, integrations.SetupConfig{
		T:            t,
		CasinoConfig: integrations.DefaultCasinoConfig(),
	})
	if err != nil {
		t.Error(err)
		return
	}
	defer env.CasinoApp.Shutdown()

	engine, err := setupEngine(ctx, t)
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Shutdown()

	client := casinotestclient.DefaultClient
	username := "player-" + rand.Text()[:8]

	authResp, err := client.Auth(requests.Auth{
		Username: username,
		Balance:  1000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if authResp.IsError {
		t.Fatal(authResp.Error)
	}

	wg := new(sync.WaitGroup)
	start := make(chan struct{})
	playCounter := new(atomic.Int64)
	errChan := make(chan error, 2)
	for range 2 {
		wg.Go(func() {
			doPlay(
				playCounter,
				authResp.Result.SessionID,
				client,
				start,
				errChan,
			)
		})
	}
	close(start)

	wg.Wait()

	close(errChan)
	errs := chans.Collect(errChan)

	assert.Equal(t, []error{errors.New(string(gamemodels.ErrGameInProgress))}, errs)
	assert.Equal(t, int64(1), playCounter.Load())
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
				return &enginemsg.RoundResponse{
					PayoutMultiplier: decimal.NewFromInt(1),
					PrivateState:     []byte(`{}`),
					PublicState:      []byte(`{}`),
				}, nil
			},
			CloseHandler: func(ctx context.Context, r enginemsg.CloseRequest) (*enginemsg.CloseResponse, *errorx.Error) {
				return &enginemsg.CloseResponse{
					CanClose: true,
				}, nil
			},
		},
	})
}

func doPlay(
	counter *atomic.Int64,
	sessionID string,
	client *casinotestclient.Client,
	start <-chan struct{},
	errs chan<- error,
) {
	<-start

	playRes, err := client.Play(requests.Play{
		SessionID: sessionID,
		GameID:    mockengine.EngineID,
		Bet:       100,
		GameData:  []byte(`{}`),
	})
	if err != nil {
		errs <- err
		return
	}

	if playRes.IsError {
		errs <- errors.New(playRes.Error.Type)
		return
	}

	counter.Add(1)
}
