package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"slices"

	"github.com/Jordan-J-Brown/casino-go/engines/dice/internal/config"
	"github.com/Jordan-J-Brown/casino-go/engines/dice/internal/models"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/pkg/msg/enginemsg"
	"github.com/Jordan-J-Brown/casino-go/pkg/service"
	"github.com/shopspring/decimal"
)

//TODO: this is very placeholder, once communication with the casino is working, develop a real dice engine.

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	logger := logsx.NewLogger(os.Stdout)

	ctx, baseApp, err := service.NewBaseApplication(service.BaseAppConfig{
		ServiceName: "dice-engine",
		NatsCfg:     &cfg.Nats,
		TracerCfg:   &cfg.Tracer,
		Logger:      logger,
	})
	if err != nil {
		panic(err)
	}

	gameConfig := models.GameConfig{
		DiceOptions: []int{6, 10, 20},
	}

	casinoServer, err := enginemsg.StartServer(ctx, baseApp.NatsConn, baseApp.Logger, baseApp.Tracer, enginemsg.ServerConfig{
		EngineID: "dice",
		RegisterFunc: func(ctx context.Context) (*enginemsg.EngineRegistration, error) {
			return &enginemsg.EngineRegistration{
				MinBet: 100,
				MaxBet: 1000,
			}, nil
		},
		SetupHandler: func(ctx context.Context, r enginemsg.SetupRequest) (*enginemsg.SetupResponse, error) {
			logger.DebugContext(ctx, "SETUP")

			gameCfgBs, err := json.Marshal(gameConfig)
			if err != nil {
				return nil, err
			}

			metaDataBs, err := json.Marshal(models.MetaData{
				Version: "v0.1.0",
				Math:    "v0.1.0",
			})
			if err != nil {
				return nil, err
			}

			return &enginemsg.SetupResponse{
				GameConfig: gameCfgBs,
				MetaData:   metaDataBs,
			}, nil
		},
		PlayHandler: func(ctx context.Context, r enginemsg.PlayRequest) (*enginemsg.RoundResponse, error) {
			logger.DebugContext(ctx, "PLAY")

			var roundConfig models.RoundConfig
			if err := json.Unmarshal(r.MetaData, &roundConfig); err != nil {
				return nil, err
			}

			if !slices.Contains(gameConfig.DiceOptions, roundConfig.DiceSides) {
				return nil, errors.New("invalid dice option")
			}
			if roundConfig.Selection < 0 || roundConfig.Selection > int64(roundConfig.DiceSides) {
				return nil, errors.New("invalid selection")
			}

			//TODO: introduce rng package for random numbers, or dedicated service to use
			bigN, err := rand.Int(rand.Reader, big.NewInt(int64(roundConfig.DiceSides)))
			if err != nil {
				return nil, err
			}
			n := bigN.Int64() + 1

			//TODO: add real payoutMultiplier handling
			winnings := r.Bet * 2
			var payoutMultiplier decimal.Decimal
			if n == roundConfig.Selection {
				payoutMultiplier = decimal.NewFromInt(2)
			} else {
				payoutMultiplier = decimal.NewFromInt(0)
			}

			publicState, err := json.Marshal(models.RoundResult{
				Closed:   true,
				Rolled:   n,
				Winnings: winnings,
			})
			if err != nil {
				return nil, err
			}

			return &enginemsg.RoundResponse{
				PrivateState:     []byte{},
				PublicState:      publicState,
				PayoutMultiplier: payoutMultiplier,
			}, nil
		},
		NextHandler: func(ctx context.Context, r enginemsg.NextRequest) (*enginemsg.RoundResponse, error) {
			return nil, errors.New("next not available")
		},
		CloseHandler: func(ctx context.Context, r enginemsg.CloseRequest) (*enginemsg.CloseResponse, error) {
			logger.DebugContext(ctx, "CLOSE")

			return &enginemsg.CloseResponse{
				CanClose: true,
			}, nil
		},
	})
	if err != nil {
		panic(err)
	}
	baseApp.RegisterShutdownFunc(casinoServer.Shutdown)

	baseApp.WaitForShutdown(ctx)
}
