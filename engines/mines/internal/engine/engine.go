package engine

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Jordan-J-Brown/casino-go/engines/mines/internal/config"
	"github.com/Jordan-J-Brown/casino-go/engines/mines/internal/minesgame"
	"github.com/Jordan-J-Brown/casino-go/engines/mines/internal/models/mathmodel"
	"github.com/Jordan-J-Brown/casino-go/engines/mines/pkg/requests"
	"github.com/Jordan-J-Brown/casino-go/engines/mines/pkg/responses"
	"github.com/Jordan-J-Brown/casino-go/pkg/errorx"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/pkg/msg/enginemsg"
	"github.com/Jordan-J-Brown/casino-go/pkg/randx"
	"github.com/Jordan-J-Brown/casino-go/pkg/service"
)

type Engine struct {
	*service.BaseApplication
	cfg       *config.Config
	mathModel *mathmodel.Model
	msgServer *enginemsg.Server
	logger    logsx.Logger
}

func Start(cfg *config.Config, logger logsx.Logger) error {
	engine := &Engine{
		cfg:    cfg,
		logger: logger,
	}

	mathModel, err := mathmodel.LoadFromFile()
	if err != nil {
		return errors.New("failed to load math model: " + err.Error())
	}
	engine.mathModel = mathModel

	ctx, baseApp, err := service.NewBaseApplication(service.BaseAppConfig{
		ServiceName:    cfg.ServiceName,
		ServiceVersion: cfg.ServiceVersion,
		NatsCfg:        &cfg.Nats,
		TracerCfg:      &cfg.Tracer,
		Logger:         logger,
	})
	if err != nil {
		return err
	}
	engine.BaseApplication = baseApp

	msgServer, err := enginemsg.StartServer(ctx, baseApp.NatsConn, baseApp.Logger, baseApp.Tracer, enginemsg.ServerConfig{
		EngineID:     cfg.EngineID,
		RegisterFunc: engine.GetRegistration,
		SetupHandler: engine.Setup,
		PlayHandler:  engine.Play,
		NextHandler:  engine.Next,
		CloseHandler: engine.Close,
	})
	if err != nil {
		return err
	}
	engine.msgServer = msgServer

	logger.InfoContext(ctx, "Engine running...")

	engine.WaitForShutdown(ctx)

	return nil
}

func (engine *Engine) GetRegistration(ctx context.Context) (*enginemsg.EngineRegistration, error) {
	return &enginemsg.EngineRegistration{
		MinBet: engine.mathModel.MinBet,
		MaxBet: engine.mathModel.MaxBet,
	}, nil
}

func (engine *Engine) Setup(ctx context.Context, r enginemsg.SetupRequest) (res *enginemsg.SetupResponse, errX *errorx.Error) {
	engine.logger.DebugContext(ctx, "SETUP")

	gameCfgBs, err := json.Marshal(responses.GameConfig{
		GridSize: engine.mathModel.GridSize,
	})
	if err != nil {
		return nil, errorx.New("failed to marshal game config", errorx.WithCode(errorx.ServerError), errorx.WithChildren(err))
	}

	metaDataBs, err := json.Marshal(responses.MetaData{
		Version: engine.cfg.EngineVersion,
		Math:    engine.mathModel.MetaData.Version,
	})
	if err != nil {
		return nil, errorx.New("failed to marshal meta data", errorx.WithCode(errorx.ServerError), errorx.WithChildren(err))
	}

	return &enginemsg.SetupResponse{
		GameConfig: gameCfgBs,
		MetaData:   metaDataBs,
	}, nil
}

func (engine *Engine) Play(ctx context.Context, r enginemsg.PlayRequest) (res *enginemsg.RoundResponse, errX *errorx.Error) {
	engine.logger.DebugContext(ctx, "PLAY")

	var request requests.PlayRequest
	if err := json.Unmarshal(r.MetaData, &request); err != nil {
		return nil, errorx.NewFromType("BAD_REQUEST", errorx.WithCode(errorx.ClientError), errorx.WithChildren(err))
	}

	game, err := minesgame.New(minesgame.Config{
		Rng:       randx.NewCryptoGenerator(),
		Bet:       r.Bet,
		Mines:     request.Mines,
		GridSize:  engine.mathModel.GridSize,
		HouseEdge: engine.mathModel.HouseEdge,
	})
	if err != nil {
		if gameErr, ok := err.(minesgame.Error); ok {
			return nil, errorx.NewFromType(gameErr.Error(), errorx.WithCode(errorx.ClientError))
		}
		return nil, errorx.New("failed to create game", errorx.WithCode(errorx.ServerError), errorx.WithChildren(err))
	}

	publicState, err := json.Marshal(responses.PlayResponse{
		State:            game.State.String(),
		PayoutMultiplier: game.PayoutMultiplier.String(),
		Mines:            game.Mines,
	})
	if err != nil {
		return nil, errorx.New("failed to marshal public state", errorx.WithCode(errorx.ServerError), errorx.WithChildren(err))
	}

	privateState, err := game.Serialise()
	if err != nil {
		return nil, errorx.New("failed to marshal private state", errorx.WithCode(errorx.ServerError), errorx.WithChildren(err))
	}

	return &enginemsg.RoundResponse{
		PrivateState:     privateState,
		PublicState:      publicState,
		PayoutMultiplier: game.PayoutMultiplier,
	}, nil
}

func (engine *Engine) Next(ctx context.Context, r enginemsg.NextRequest) (res *enginemsg.RoundResponse, errX *errorx.Error) {
	engine.logger.DebugContext(ctx, "NEXT")

	var request requests.NextRequest
	if err := json.Unmarshal(r.MetaData, &request); err != nil {
		engine.logger.ErrorContext(ctx, "failed to unmarshal next request", "error", err)
		return nil, errorx.NewFromType("BAD_REQUEST", errorx.WithCode(errorx.ClientError), errorx.WithChildren(err))
	}

	game := new(minesgame.Game)
	if err := game.Deserialise(r.State); err != nil {
		engine.logger.ErrorContext(ctx, "failed to unmarshal game state", "error", err)
		return nil, errorx.New("failed to deserialise game state", errorx.WithCode(errorx.ServerError), errorx.WithChildren(err))
	}

	var result *minesgame.Result
	var gameErr *minesgame.Error
	if request.Cashout {
		result, gameErr = game.Cashout()
	} else {
		result, gameErr = game.Pick(request.Index)
	}
	if gameErr != nil {
		engine.logger.ErrorContext(ctx, "failed to retrive result from game", "error", gameErr.Error())
		return nil, errorx.NewFromType(gameErr.Error(), errorx.WithCode(errorx.ClientError))
	}

	publicState, err := json.Marshal(responses.NextResponse{
		State:             result.State.String(),
		PayoutMultiplier:  result.PayoutMultiplier.String(),
		SelectedPositions: result.SelectedPositions,
	})
	if err != nil {
		return nil, errorx.New("failed to marshal public state", errorx.WithCode(errorx.ServerError), errorx.WithChildren(err))
	}

	privateState, err := game.Serialise()
	if err != nil {
		return nil, errorx.New("failed to marshal private state", errorx.WithCode(errorx.ServerError), errorx.WithChildren(err))
	}

	return &enginemsg.RoundResponse{
		PayoutMultiplier: game.PayoutMultiplier,
		PrivateState:     privateState,
		PublicState:      publicState,
	}, nil
}

func (engine *Engine) Close(ctx context.Context, r enginemsg.CloseRequest) (res *enginemsg.CloseResponse, errX *errorx.Error) {
	engine.logger.DebugContext(ctx, "CLOSE")

	game := new(minesgame.Game)
	if err := game.Deserialise(r.State); err != nil {
		return nil, errorx.New("failed to deserialise game state", errorx.WithCode(errorx.ServerError), errorx.WithChildren(err))
	}

	return &enginemsg.CloseResponse{
		CanClose: !game.IsPlaying(),
	}, nil
}
