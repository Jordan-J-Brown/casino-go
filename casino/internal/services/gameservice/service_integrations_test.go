package gameservice_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/gamemodels"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/tests/integrations/casinotestclient"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/tests/integrations/mockengine"
	"github.com/Jordan-J-Brown/casino-go/casino/pkg/requests"
	"github.com/Jordan-J-Brown/casino-go/pkg/chans"
	"github.com/stretchr/testify/assert"
)

func TestGameService_Two_Concurrent_Plays(t *testing.T) {
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

func TestGameService_Two_Concurrent_Next_Actions(t *testing.T) {
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

func TestGameService_Two_Concurrent_Close_Actions(t *testing.T) {
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
