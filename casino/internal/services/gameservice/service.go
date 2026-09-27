package gameservice

import (
	"context"
	"errors"
	"slices"
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
	cfg Config
}

type Config struct {
	MaxPayoutMultiplier int64
	BetSizes            []int64
	Stores              TransactionProvider
	EngineRegistry      engineregistry.Service
	EngineClient        engineclient.Client
	Logger              logsx.Logger
	CasinoWalletID      string
}

func New(cfg Config) Service {
	return &service{
		cfg: cfg,
	}
}

func (service *service) capPayoutMultiplier(m decimal.Decimal) (mult decimal.Decimal, capped bool) {
	maxMult := decimal.NewFromInt(service.cfg.MaxPayoutMultiplier)
	if m.GreaterThanOrEqual(maxMult) {
		return maxMult, true
	}
	return m, false
}

func (service *service) transferFunds(ctx context.Context, store WalletStoreAdapter, args walletmodels.TransferArgs) (srcBalance, dstBalance int64, err error) {
	err = store.TransferFunds(ctx, args, func(src, dst *walletmodels.Wallet) error {
		if err = src.Debit(args.AmountMinor); err != nil {
			if err == walletmodels.ErrInsufficientFunds {
				return err
			}
			return errorx.New("failed to debit source wallet", errorx.WithChildren(err))
		}
		srcBalance = src.BalanceMinor

		if err = dst.Credit(args.AmountMinor); err != nil {
			return errorx.New("failed to credit destination wallet", errorx.WithChildren(err))
		}
		dstBalance = dst.BalanceMinor

		return nil
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
	engineReg, err := service.cfg.EngineRegistry.GetEngineRegistration(ctx, engineID)
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

	bets, limitErr := service.limitBets(engineReg.MinBet, engineReg.MaxBet, service.cfg.BetSizes)
	if limitErr != nil {
		return nil, errorx.NewFromError(limitErr)
	}

	return bets, nil
}

func (service *service) Setup(ctx context.Context, request SetupRequest) (*SetupResult, *errorx.Error) {
	bets, err := service.getBetsForEngine(ctx, request.EngineID)
	if err != nil {
		return nil, err
	}

	var game *gamemodels.Game
	transactErr := service.cfg.Stores.Transact(ctx, func(ctx context.Context, stores Adapters) error {
		gameExists, err := stores.GameStore.IsActiveGamePresent(ctx, request.User.ID, request.EngineID)
		if err != nil {
			return err
		}
		if !gameExists {
			return nil
		}

		game, err = stores.GameStore.LoadGame(ctx, request.User.ID, request.EngineID)
		return err
	})
	if transactErr != nil {
		return nil, errorx.New("failed to check for active game from setup", errorx.WithChildren(err))
	}

	engineResponse, setupErr := service.cfg.EngineClient.Setup(ctx, request.EngineID, enginemsg.SetupRequest{
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
	bets, err := service.getBetsForEngine(ctx, request.EngineID)
	if err != nil {
		return nil, err
	}

	if !slices.Contains(bets, request.Bet) {
		return nil, errorx.NewFromError(gamemodels.ErrInvalidBet)
	}

	result := new(RoundResult)

	transactErr := service.cfg.Stores.Transact(ctx, func(ctx context.Context, adapters Adapters) error {
		gameExists, err := adapters.GameStore.IsActiveGamePresent(ctx, request.User.ID, request.EngineID)
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

		srcBalance, _, err := service.transferFunds(ctx, adapters.WalletStore, walletmodels.TransferArgs{
			SourceWalletID:      request.User.WalletID,
			DestinationWalletID: service.cfg.CasinoWalletID,
			AmountMinor:         request.Bet,
			Timestamp:           time.Now(),
		})
		if err != nil {
			if err == walletmodels.ErrInsufficientFunds {
				return errorx.NewFromType(string(gamemodels.ErrInsufficientFunds), errorx.WithCode(errorx.ClientError))
			}
			return errorx.New("failed to transfer bet", errorx.WithChildren(err))
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		engineResponse, playErr := service.cfg.EngineClient.Play(ctx, request.EngineID, enginemsg.PlayRequest{
			Bet:      request.Bet,
			MetaData: request.MetaData,
		})
		if playErr != nil {
			return playErr
		}
		game.PrivateState = engineResponse.PrivateState
		game.PublicState = engineResponse.PublicState
		game.PayoutMultiplier = engineResponse.PayoutMultiplier

		err = adapters.GameStore.SaveGame(ctx, game)
		if err != nil {
			return errorx.New("failed to save game state", errorx.WithChildren(err))
		}

		result.Balance = srcBalance
		result.PayoutMutliplier = engineResponse.PayoutMultiplier.String()
		result.PublicState = engineResponse.PublicState

		return ctx.Err()
	})
	if transactErr != nil {
		if errX, ok := transactErr.(*errorx.Error); ok {
			return nil, errX
		}
		return nil, errorx.NewFromError(err)
	}

	return result, nil
}

func (service *service) Next(ctx context.Context, request NextRequest) (*RoundResult, *errorx.Error) {
	_, err := service.getEngineReg(ctx, request.EngineID)
	if err != nil {
		return nil, err
	}

	var result *RoundResult
	transactErr := service.cfg.Stores.Transact(ctx, func(ctx context.Context, stores Adapters) error {
		game, err := stores.GameStore.LoadGame(ctx, request.User.ID, request.EngineID)
		if err != nil && err != gamemodels.ErrGameNotFound {
			return err
		}
		if game == nil {
			return errorx.NewFromType(string(gamemodels.ErrGameNotFound), errorx.WithCode(errorx.ClientError))
		}

		wallet, err := stores.WalletStore.GetWallet(ctx, request.User.WalletID)
		if err != nil {
			return err
		}

		if _, capped := service.capPayoutMultiplier(game.PayoutMultiplier); capped {
			return errorx.NewFromType(string(gamemodels.ErrNextUnavailable), errorx.WithCode(errorx.ClientError))
		}

		response, nextErr := service.cfg.EngineClient.Next(ctx, request.EngineID, enginemsg.NextRequest{
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

		err = stores.GameStore.SaveGame(ctx, game)
		if err != nil {
			return err
		}

		result = &RoundResult{
			Balance:          wallet.BalanceMinor,
			PayoutMutliplier: game.PayoutMultiplier.String(),
			PublicState:      game.PublicState,
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

func (service *service) Close(ctx context.Context, request CloseRequest) (*CloseResult, *errorx.Error) {
	var result *CloseResult
	transactErr := service.cfg.Stores.Transact(ctx, func(ctx context.Context, stores Adapters) error {
		game, err := stores.GameStore.LoadGame(ctx, request.User.ID, request.EngineID)
		if err != nil && err != gamemodels.ErrGameNotFound {
			return errorx.NewFromError(err)
		}
		if game == nil {
			return errorx.NewFromType(string(gamemodels.ErrGameNotFound), errorx.WithCode(errorx.ClientError))
		}

		closeResponse, closeErr := service.cfg.EngineClient.Close(ctx, game.EngineID, enginemsg.CloseRequest{
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
			service.cfg.Logger.InfoContext(ctx, "game capped",
				"game_id", game.GameID,
				"multiplier", game.PayoutMultiplier.String(),
			)
		}

		var balance int64
		payout := decimal.NewFromInt(game.Bet).Mul(payoutMultiplier).IntPart()
		if payout > 0 {
			_, dstBalance, err := service.transferFunds(ctx, stores.WalletStore, walletmodels.TransferArgs{
				SourceWalletID:      service.cfg.CasinoWalletID,
				DestinationWalletID: request.User.WalletID,
				AmountMinor:         payout,
				Timestamp:           time.Now(),
			})
			if err != nil {
				return err
			}
			balance = dstBalance
		} else {
			wallet, err := stores.WalletStore.GetWallet(ctx, request.User.WalletID)
			if err != nil {
				return err
			}
			balance = wallet.BalanceMinor
		}

		game.Status = gamemodels.StatusClosed

		err = stores.GameStore.SaveGame(ctx, game)
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
