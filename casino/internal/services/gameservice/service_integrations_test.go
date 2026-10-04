//go:build integration

package gameservice_test

import (
	"context"
	"sync"
	"testing"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/gamemodels"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/tests/integrations/mockengine"
	"github.com/Jordan-J-Brown/casino-go/casino/pkg/requests"
	"github.com/Jordan-J-Brown/casino-go/pkg/casinoclient"
	"github.com/Jordan-J-Brown/casino-go/pkg/chans"
	"github.com/Jordan-J-Brown/casino-go/pkg/errorx"
	"github.com/Jordan-J-Brown/casino-go/pkg/msg/enginemsg"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestGameService_Two_Concurrent_Plays(t *testing.T) {
	ctx := t.Context()
	env := setupEnv(ctx, t)

	wg := new(sync.WaitGroup)
	errChan := make(chan error, 2)
	for range 2 {
		wg.Go(func() {
			(&gameWorker{
				sessionID: env.SessionID,
				client:    casinoclient.NewDefaultClient(),
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
		t.Error("unexpected error type returned:", errStr)
	}
}

func TestGameService_Two_Concurrent_Next_Actions(t *testing.T) {
	ctx := t.Context()
	env := setupEnv(ctx, t)

	res, err := casinoclient.DefaultClient.Play(requests.Play{
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
				client:    casinoclient.NewDefaultClient(),
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

func TestGameService_Two_Concurrent_Close_Actions(t *testing.T) {
	ctx := t.Context()
	env := setupEnv(ctx, t)

	res, err := casinoclient.DefaultClient.Play(requests.Play{
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
				client:    casinoclient.NewDefaultClient(),
				errs:      errChan,
				steps:     []gameStep{stepClose},
			}).run()
		})
	}
	wg.Wait()

	balanceRes := casinoclient.DefaultClient.MustGetBalance(requests.Balance{
		SessionID: env.SessionID,
	})
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

func TestGameService_Insufficient_Funds(t *testing.T) {
	ctx := t.Context()
	env := setupEnv(ctx, t, enginemsg.ServerConfig{
		RegisterFunc: func(ctx context.Context) (*enginemsg.EngineRegistration, error) {
			return &enginemsg.EngineRegistration{
				MinBet: 10,
				MaxBet: 20_00,
			}, nil
		},
	})

	client := casinoclient.DefaultClient
	playRes, err := client.Play(requests.Play{
		SessionID: env.SessionID,
		GameID:    mockengine.EngineID,
		Bet:       20_00,
		GameData:  []byte(`{}`),
	})
	if err != nil {
		t.Fatal("failed to send play request:", err)
	}
	if !playRes.IsError {
		t.Fatal("expected play to return an error")
	}

	assert.Equal(t, gamemodels.ErrInsufficientFunds.Error(), playRes.Error.Type)
}

func TestGameService_Play_Bet_Is_Not_Deducted_Upon_Error(t *testing.T) {
	ctx := t.Context()
	env := setupEnv(ctx, t, enginemsg.ServerConfig{
		RegisterFunc: func(ctx context.Context) (*enginemsg.EngineRegistration, error) {
			return &enginemsg.EngineRegistration{
				MinBet: 10,
				MaxBet: 1000,
			}, nil
		},
		PlayHandler: func(ctx context.Context, r enginemsg.PlayRequest) (*enginemsg.RoundResponse, *errorx.Error) {
			return nil, errorx.New("error from play")
		},
	})

	client := casinoclient.DefaultClient
	playRes, err := client.Play(requests.Play{
		SessionID: env.SessionID,
		GameID:    mockengine.EngineID,
		Bet:       100,
		GameData:  []byte(`{}`),
	})
	if err != nil {
		t.Fatal("failed to send play request:", err)
	}
	if !playRes.IsError {
		t.Fatal("expected play to return an error")
	}

	balanceRes := client.MustGetBalance(requests.Balance{SessionID: env.SessionID})

	assert.Equal(t, env.OriginalBalance, balanceRes.Result.Balance)
}

func TestGameService_Game_Is_Capped_In_Play(t *testing.T) {
	ctx := t.Context()
	payoutMultiplier := decimal.NewFromInt(250_001)
	env := setupEnv(ctx, t, enginemsg.ServerConfig{
		RegisterFunc: func(ctx context.Context) (*enginemsg.EngineRegistration, error) {
			return &enginemsg.EngineRegistration{
				MinBet: 10,
				MaxBet: 1000,
			}, nil
		},
		PlayHandler: func(ctx context.Context, r enginemsg.PlayRequest) (*enginemsg.RoundResponse, *errorx.Error) {
			return &enginemsg.RoundResponse{
				PayoutMultiplier: payoutMultiplier,
				PrivateState:     []byte(`{}`),
				PublicState:      []byte(`{}`),
			}, nil
		},
		CloseHandler: func(ctx context.Context, r enginemsg.CloseRequest) (*enginemsg.CloseResponse, *errorx.Error) {
			return &enginemsg.CloseResponse{
				CanClose: true,
			}, nil
		},
	})

	bet := int64(100)
	client := casinoclient.DefaultClient

	playRes := client.MustPlay(requests.Play{
		SessionID: env.SessionID,
		GameID:    mockengine.EngineID,
		Bet:       bet,
		GameData:  []byte(`{}`),
	})
	assert.Equal(t, "250000.00", playRes.Result.PayoutMultiplier)

	nextRes, err := client.Next(requests.Next{
		SessionID: env.SessionID,
		GameID:    mockengine.EngineID,
		GameData:  []byte(`{}`),
	})
	if err != nil {
		t.Fatal("failed to send next request:", err)
	}
	if !nextRes.IsError {
		t.Fatal("expected next request to fail due to capped multiplier")
	}
	assert.Equal(t, gamemodels.ErrNextUnavailable.Error(), nextRes.Error.Type)

	closeRes := client.MustClose(requests.Close{
		SessionID: env.SessionID,
		GameID:    mockengine.EngineID,
	})

	closingBalance := closeRes.Result.Balance
	balanceRes := client.MustGetBalance(requests.Balance{SessionID: env.SessionID})
	expectedBalance := (env.OriginalBalance - bet) + bet*250_000

	assert.Equal(t, expectedBalance, balanceRes.Result.Balance)
	assert.Equal(t, expectedBalance, closingBalance)
}

func TestGameService_Game_Is_Capped_In_Next(t *testing.T) {
	ctx := t.Context()
	env := setupEnv(ctx, t, enginemsg.ServerConfig{
		RegisterFunc: func(ctx context.Context) (*enginemsg.EngineRegistration, error) {
			return &enginemsg.EngineRegistration{
				MinBet: 10,
				MaxBet: 1000,
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
				PayoutMultiplier: decimal.NewFromInt(99999999),
				PrivateState:     []byte(`{}`),
				PublicState:      []byte(`{}`),
			}, nil
		},
		CloseHandler: func(ctx context.Context, r enginemsg.CloseRequest) (*enginemsg.CloseResponse, *errorx.Error) {
			return &enginemsg.CloseResponse{
				CanClose: true,
			}, nil
		},
	})

	bet := int64(100)
	client := casinoclient.DefaultClient

	client.MustPlay(requests.Play{
		SessionID: env.SessionID,
		GameID:    mockengine.EngineID,
		Bet:       bet,
		GameData:  []byte(`{}`),
	})

	nextRes := client.MustNext(requests.Next{
		SessionID: env.SessionID,
		GameID:    mockengine.EngineID,
		GameData:  []byte(`{}`),
	})
	assert.Equal(t, "250000.00", nextRes.Result.PayoutMultiplier)

	closeRes := client.MustClose(requests.Close{
		SessionID: env.SessionID,
		GameID:    mockengine.EngineID,
	})

	closingBalance := closeRes.Result.Balance
	balanceRes := client.MustGetBalance(requests.Balance{SessionID: env.SessionID})
	expectedBalance := (env.OriginalBalance - bet) + bet*250_000

	assert.Equal(t, expectedBalance, balanceRes.Result.Balance)
	assert.Equal(t, expectedBalance, closingBalance)
}
