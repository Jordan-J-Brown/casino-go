package app

import (
	"context"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/adapters/message"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/adapters/web"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/clients/engineclient"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/config"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/services/engineregistry"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/services/gameservice"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/services/userservice"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/stores/adapters/game_service_store_adapter"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/stores/engineregstore"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/stores/userwalletstorepg"
	baseconfig "github.com/Jordan-J-Brown/casino-go/pkg/config"
	"github.com/Jordan-J-Brown/casino-go/pkg/errorx"
	"github.com/Jordan-J-Brown/casino-go/pkg/pgdb"
	"github.com/Jordan-J-Brown/casino-go/pkg/service"
)

type Application struct {
	*service.BaseApplication
	Config   *AppConfig
	Services Services
}

type Services struct {
	GameService    gameservice.Service
	UserService    userservice.Service
	EngineRegistry engineregistry.Service
}

func (app *Application) initServices(ctx context.Context) error {
	pgPool, err := pgdb.Connect(ctx)
	if err != nil {
		return err
	}
	app.RegisterShutdownFunc(func(ctx context.Context) error {
		pgPool.Close()
		return nil
	})

	engineRegistry := engineregistry.New(ctx, app.Logger, engineregstore.New())
	app.Services.EngineRegistry = engineRegistry

	gameServiceStores := game_service_store_adapter.New(pgPool, app.Logger)
	app.Services.GameService = gameservice.New(gameservice.Config{
		MaxPayoutMultiplier: app.Config.Games.MaxPayoutMultiplier,
		EngineRegistry:      engineRegistry,
		EngineClient:        engineclient.New(app.NatsConn, app.Tracer, app.Config.Engines.ClientTimeoutSeconds),
		BetSizes:            app.Config.Games.BetSizes,
		Stores:              gameServiceStores,
		Logger:              app.Logger,
		CasinoWalletID:      app.Config.Config.CasinoWalletID,
	})

	app.Services.UserService = userservice.New(userservice.Config{
		SessionTTL: app.Config.Auth.SessionTTL,
		Store:      userwalletstorepg.New(pgPool, app.Logger),
	})

	return nil
}

type AppConfig struct {
	service.BaseAppConfig
	*config.Config
}

func Run(appConfig AppConfig) error {
	app := &Application{
		Config: &appConfig,
	}

	if appConfig.Config.Env == baseconfig.Local {
		errorx.OutputStack = true
	}

	ctx, base, err := service.NewBaseApplication(appConfig.BaseAppConfig)
	if err != nil {
		return errorx.New("failed to create base application", errorx.WithChildren(err))
	}
	app.BaseApplication = base

	if err := app.initServices(ctx); err != nil {
		return errorx.New("failed to initialise services", errorx.WithChildren(err))
	}

	msgAPI, err := message.RunApiAsync(ctx, app.NatsConn, app.Logger, app.Tracer, app.Services.EngineRegistry)
	if err != nil {
		return errorx.New("failed to start message server", errorx.WithChildren(err))
	}
	app.RegisterShutdownFunc(msgAPI.Shutdown)

	webApiShutdown := web.RunApiAsync(ctx, web.ApiConfig{
		Port:        app.Config.API.Port,
		Logger:      app.Logger,
		Tracer:      app.Tracer,
		UserService: app.Services.UserService,
		GameService: app.Services.GameService,
	})
	app.RegisterShutdownFunc(webApiShutdown)

	app.Logger.DebugContext(ctx, "Running...", "env", appConfig.Config.Env)

	app.WaitForShutdown(ctx)

	return nil
}
