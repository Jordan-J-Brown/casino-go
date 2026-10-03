package gameservice

import (
	"context"
	"errors"
	"slices"
	"sync"
	"time"
	"uuid"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/clients/engineclient"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/gamemodels"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/walletmodels"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/services/engineregistry"
	"github.com/Jordan-J-Brown/casino-go/pkg/errorx"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/pkg/msg/enginemsg"
	"github.com/shopspring/decimal"
)

type Service interface {
	Setup(ctx context.Context, request SetupRequest) (*SetupResult, *errorx.Error)
	Play(ctx context.Context, request PlayRequest) (*RoundResult, *errorx.Error)
	Next(ctx context.Context, request NextRequest) (*RoundResult, *errorx.Error)
	Close(ctx context.Context, request CloseRequest) (*CloseResult, *errorx.Error)
}

type service struct {
	MaxPayoutMultiplier int64
	BetSizes            []int64
	StoreProvider       CasinoGameStoreProvider
	EngineRegistry      engineregistry.Service
	EngineClient        engineclient.Client
	Logger              logsx.Logger

	gameLocker *gameLocker
}

type Config struct {
	MaxPayoutMultiplier int64
	BetSizes            []int64
	StoreProvider       CasinoGameStoreProvider
	EngineRegistry      engineregistry.Service
	EngineClient        engineclient.Client
	Logger              logsx.Logger
}

func New(ctx context.Context, cfg Config) Service {
	return &service{
		MaxPayoutMultiplier: cfg.MaxPayoutMultiplier,
		BetSizes:            cfg.BetSizes,
		StoreProvider:       cfg.StoreProvider,
		EngineRegistry:      cfg.EngineRegistry,
		EngineClient:        cfg.EngineClient,
		Logger:              cfg.Logger,

		gameLocker: newGameLocker(),
	}
}

func (service *service) Setup(ctx context.Context, request SetupRequest) (*SetupResult, *errorx.Error) {
	bets, err := service.getBetsForEngine(ctx, request.EngineID)
	if err != nil {
		return nil, err
	}

	var game *gamemodels.Game
	storeErr := service.StoreProvider.UseStore(ctx, func(store CasinoGameStore) error {
		gameExists, err := store.IsActiveGamePresent(ctx, request.User.ID, request.EngineID)
		if err != nil {
			return err
		}
		if !gameExists {
			return nil
		}

		game, err = store.LoadGame(ctx, request.User.ID, request.EngineID)
		return err
	})
	if storeErr != nil {
		return nil, errorx.New("failed to check for active game from setup", errorx.WithChildren(err))
	}

	engineResponse, setupErr := service.EngineClient.Setup(ctx, request.EngineID, enginemsg.SetupRequest{
		MetaData: request.MetaData,
	})
	if setupErr != nil {
		return nil, setupErr
	}

	result := &SetupResult{
		Bets:       bets,
		GameConfig: engineResponse.GameConfig,
		MetaData:   engineResponse.MetaData,
	}

	if game != nil {
		result.PublicState = game.PublicState
	}

	return result, nil
}

func (service *service) Play(ctx context.Context, request PlayRequest) (*RoundResult, *errorx.Error) {
	if !service.gameLocker.tryLockGameForAction(actionPlay, request.User.ID, request.EngineID) {
		return nil, errorx.NewFromType(string(gamemodels.ErrTooManyRequests), errorx.WithCode(errorx.ClientError))
	}
	defer service.gameLocker.unlockGame(actionPlay, request.User.ID, request.EngineID)

	bets, err := service.getBetsForEngine(ctx, request.EngineID)
	if err != nil {
		return nil, err
	}

	if !slices.Contains(bets, request.Bet) {
		return nil, errorx.NewFromError(gamemodels.ErrInvalidBet)
	}

	var result *RoundResult
	storeErr := service.StoreProvider.UseStore(ctx, func(store CasinoGameStore) error {
		gameExists, err := store.IsActiveGamePresent(ctx, request.User.ID, request.EngineID)
		if err != nil {
			return errorx.New("failed to check if game exists", errorx.WithChildren(err))
		}
		if gameExists {
			return errorx.NewFromType(string(gamemodels.ErrGameInProgress), errorx.WithCode(errorx.ClientError))
		}

		game := &gamemodels.Game{
			GameID:   uuid.New().String(),
			EngineID: request.EngineID,
			UserID:   request.User.ID,
			Bet:      request.Bet,
			Status:   gamemodels.StatusPlaying,
		}

		newBalance, err := service.debitFunds(ctx, store, request.User.WalletID, request.Bet)
		if err != nil {
			if err == walletmodels.ErrInsufficientFunds {
				return errorx.NewFromType(string(gamemodels.ErrInsufficientFunds), errorx.WithCode(errorx.ClientError))
			}
			return errorx.New("failed to deduct bet", errorx.WithChildren(err))
		}

		engineResponse, playErr := service.EngineClient.Play(ctx, request.EngineID, enginemsg.PlayRequest{
			Bet:      request.Bet,
			MetaData: request.MetaData,
		})
		if playErr != nil {
			return playErr
		}
		game.PrivateState = engineResponse.PrivateState
		game.PublicState = engineResponse.PublicState
		payoutMultiplier, _ := service.capPayoutMultiplier(engineResponse.PayoutMultiplier)
		game.PayoutMultiplier = payoutMultiplier

		err = store.SaveGame(ctx, game)
		if err != nil {
			return errorx.New("failed to save game state", errorx.WithChildren(err))
		}

		result = &RoundResult{
			Balance:          newBalance,
			Bet:              game.Bet,
			PayoutMutliplier: service.getPublicStatePayoutMultiplier(game.PayoutMultiplier),
			PublicState:      game.PublicState,
		}

		return nil
	})
	if storeErr != nil {
		if errX, ok := storeErr.(*errorx.Error); ok {
			return nil, errX
		}
		return nil, errorx.NewFromError(storeErr)
	}

	return result, nil
}

func (service *service) Next(ctx context.Context, request NextRequest) (*RoundResult, *errorx.Error) {
	if !service.gameLocker.tryLockGameForAction(actionNext, request.User.ID, request.EngineID) {
		return nil, errorx.NewFromType(string(gamemodels.ErrTooManyRequests), errorx.WithCode(errorx.ClientError))
	}
	defer service.gameLocker.unlockGame(actionNext, request.User.ID, request.EngineID)

	_, err := service.getEngineReg(ctx, request.EngineID)
	if err != nil {
		return nil, err
	}

	var result *RoundResult
	storeErr := service.StoreProvider.UseStore(ctx, func(store CasinoGameStore) error {
		game, err := store.LoadGame(ctx, request.User.ID, request.EngineID)
		if err != nil && err != gamemodels.ErrGameNotFound {
			return err
		}
		if game == nil {
			return errorx.NewFromType(string(gamemodels.ErrGameNotFound), errorx.WithCode(errorx.ClientError))
		}

		wallet, err := store.GetWallet(ctx, request.User.WalletID)
		if err != nil {
			return err
		}

		if _, capped := service.capPayoutMultiplier(game.PayoutMultiplier); capped {
			return errorx.NewFromType(string(gamemodels.ErrNextUnavailable), errorx.WithCode(errorx.ClientError))
		}

		response, nextErr := service.EngineClient.Next(ctx, request.EngineID, enginemsg.NextRequest{
			State:    game.PrivateState,
			MetaData: request.MetaData,
		})
		if nextErr != nil {
			return nextErr
		}
		game.PrivateState = response.PrivateState
		game.PublicState = response.PublicState
		payoutMultiplier, _ := service.capPayoutMultiplier(response.PayoutMultiplier)
		game.PayoutMultiplier = payoutMultiplier

		err = store.SaveGame(ctx, game)
		if err != nil {
			return err
		}

		result = &RoundResult{
			Balance:          wallet.BalanceMinor,
			Bet:              game.Bet,
			PayoutMutliplier: service.getPublicStatePayoutMultiplier(game.PayoutMultiplier),
			PublicState:      game.PublicState,
		}

		return nil
	})
	if storeErr != nil {
		if errX, ok := storeErr.(*errorx.Error); ok {
			return nil, errX
		}
		return nil, errorx.NewFromError(storeErr)
	}

	return result, nil
}

func (service *service) Close(ctx context.Context, request CloseRequest) (*CloseResult, *errorx.Error) {
	if !service.gameLocker.tryLockGameForAction(actionClose, request.User.ID, request.EngineID) {
		return nil, errorx.NewFromType(string(gamemodels.ErrTooManyRequests), errorx.WithCode(errorx.ClientError))
	}
	defer service.gameLocker.unlockGame(actionClose, request.User.ID, request.EngineID)

	var result *CloseResult
	transactErr := service.StoreProvider.UseStore(ctx, func(store CasinoGameStore) error {
		game, err := store.LoadGame(ctx, request.User.ID, request.EngineID)
		if err != nil && err != gamemodels.ErrGameNotFound {
			return errorx.NewFromError(err)
		}
		if game == nil {
			return errorx.NewFromType(string(gamemodels.ErrGameNotFound), errorx.WithCode(errorx.ClientError))
		}

		closeResponse, closeErr := service.EngineClient.Close(ctx, game.EngineID, enginemsg.CloseRequest{
			State: game.PrivateState,
		})
		if closeErr != nil {
			return closeErr
		}
		if !closeResponse.CanClose {
			return errorx.NewFromType(string(gamemodels.ErrCannotCloseGame), errorx.WithCode(errorx.ClientError))
		}

		payoutMultiplier, capped := service.capPayoutMultiplier(game.PayoutMultiplier)
		if capped {
			service.Logger.InfoContext(ctx, "game capped",
				"game_id", game.GameID,
				"multiplier", game.PayoutMultiplier.String(),
			)
		}

		var balance int64
		payout := decimal.NewFromInt(game.Bet).Mul(payoutMultiplier).IntPart()
		if payout > 0 {
			newBalance, err := service.creditFunds(ctx, store, request.User.WalletID, payout)
			if err != nil {
				return err
			}
			balance = newBalance
		} else {
			wallet, err := store.GetWallet(ctx, request.User.WalletID)
			if err != nil {
				return err
			}
			balance = wallet.BalanceMinor
		}

		game.Status = gamemodels.StatusClosed

		err = store.SaveGame(ctx, game)
		if err != nil {
			return err
		}

		result = &CloseResult{
			Balance: balance,
		}

		return nil
	})
	if transactErr != nil {
		if errX, ok := transactErr.(*errorx.Error); ok {
			return nil, errX
		}
		return nil, errorx.NewFromError(transactErr)
	}

	return result, nil
}

func (service *service) capPayoutMultiplier(m decimal.Decimal) (mult decimal.Decimal, capped bool) {
	maxMult := decimal.NewFromInt(service.MaxPayoutMultiplier)
	if m.GreaterThanOrEqual(maxMult) {
		return maxMult, true
	}
	return m, false
}

func (service *service) debitFunds(
	ctx context.Context,
	store CasinoGameStore,
	walletID string,
	amount int64,
) (newBalance int64, err error) {
	err = store.ApplyWalletTransaction(ctx, walletID, func(wallet *walletmodels.Wallet) (*walletmodels.Transaction, error) {
		debitErr := wallet.Debit(amount)
		if debitErr != nil {
			return nil, debitErr
		}
		newBalance = wallet.BalanceMinor

		return &walletmodels.Transaction{
			ID:          uuid.New().String(),
			Timestamp:   time.Now(),
			Type:        walletmodels.TransactionTypeDebit,
			WalletID:    walletID,
			AmountMinor: amount,
		}, nil
	})
	return
}

func (service *service) creditFunds(
	ctx context.Context,
	store CasinoGameStore,
	walletID string,
	amount int64,
) (newBalance int64, err error) {
	err = store.ApplyWalletTransaction(ctx, walletID, func(wallet *walletmodels.Wallet) (*walletmodels.Transaction, error) {
		debitErr := wallet.Credit(amount)
		if debitErr != nil {
			return nil, err
		}
		newBalance = wallet.BalanceMinor

		return &walletmodels.Transaction{
			ID:          uuid.New().String(),
			Timestamp:   time.Now(),
			Type:        walletmodels.TransactionTypeCredit,
			WalletID:    walletID,
			AmountMinor: amount,
		}, nil
	})
	return
}

func (service *service) limitBets(minBet, maxBet int64, allBets []int64) ([]int64, error) {
	start, end := -1, -1
	for i := range allBets {
		switch allBets[i] {
		case minBet:
			start = i
		case maxBet:
			end = i + 1
		}
		if start > -1 && end > -1 {
			return allBets[start:end], nil
		}
	}

	return nil, errors.New("failed to limit bets")
}

func (service *service) getEngineReg(ctx context.Context, engineID string) (*enginemsg.EngineRegistration, *errorx.Error) {
	engineReg, err := service.EngineRegistry.GetEngineRegistration(ctx, engineID)
	if err != nil {
		return nil, errorx.NewFromError(err)
	}
	if engineReg == nil {
		return nil, errorx.NewFromType("ERR_ENGINE_NOT_FOUND", errorx.WithCode(errorx.ClientError))
	}
	return engineReg, nil
}

func (service *service) getBetsForEngine(ctx context.Context, engineID string) ([]int64, *errorx.Error) {
	engineReg, regErr := service.getEngineReg(ctx, engineID)
	if regErr != nil {
		return nil, regErr
	}

	bets, limitErr := service.limitBets(engineReg.MinBet, engineReg.MaxBet, service.BetSizes)
	if limitErr != nil {
		return nil, errorx.NewFromError(limitErr)
	}

	return bets, nil
}

func (service *service) getPublicStatePayoutMultiplier(multiplier decimal.Decimal) string {
	return multiplier.StringFixedBank(2)
}

type gameAction int

const (
	actionPlay gameAction = iota
	actionNext
	actionClose
)

type gameLocker struct {
	actionLocks [3]sync.Mutex
	gameLocks   map[string]struct{}
}

func newGameLocker() *gameLocker {
	return &gameLocker{
		actionLocks: [3]sync.Mutex{},
		gameLocks:   make(map[string]struct{}),
	}
}

func (g *gameLocker) tryLockGameForAction(action gameAction, userID, engineID string) (locked bool) {
	mu := &g.actionLocks[action]
	mu.Lock()
	defer mu.Unlock()

	key := userID + ":" + engineID
	_, exists := g.gameLocks[key]
	if exists {
		return false
	}

	g.gameLocks[key] = struct{}{}

	return true
}

func (g *gameLocker) unlockGame(action gameAction, userID, engineID string) {
	mu := &g.actionLocks[action]
	mu.Lock()
	defer mu.Unlock()

	key := userID + ":" + engineID

	delete(g.gameLocks, key)
}
