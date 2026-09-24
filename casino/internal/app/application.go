package app

import (
	"context"
	"time"

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
	GameService gameservice.Service
	UserService userservice.Service
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

	engineRegistry, err := engineregistry.New(ctx, app.NatsConn, app.Logger, engineregstore.New())
	if err != nil {
		return errorx.New("failed to create engine registry", errorx.WithChildren(err))
	}
	app.RegisterShutdownFunc(engineRegistry.Shutdown)

	gameServiceStores := game_service_store_adapter.New(pgPool, app.Logger)
	app.Services.GameService = gameservice.New(gameservice.Config{
		MaxPayoutMultiplier: app.Config.Games.MaxPayoutMultiplier,
		EngineRegistry:      engineRegistry,
		EngineClient:        engineclient.New(app.NatsConn, 3*time.Second), //TODO: add config for timeout
		BetSizes:            app.Config.Games.BetSizes,
		Stores:              gameServiceStores,
		Logger:              app.Logger,
		CasinoWalletID:      app.Config.Config.CasinoWalletID,
	})

	app.Services.UserService = userservice.New(userservice.Config{
		SessionTTL: 0, //TODO: add config
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

	apiShutdown := web.RunApiAsync(ctx, web.ApiConfig{
		Port:        app.Config.API.Port,
		Logger:      app.Logger,
		Tracer:      app.Tracer,
		UserService: app.Services.UserService,
		GameService: app.Services.GameService,
	})
	app.RegisterShutdownFunc(apiShutdown)

	app.Logger.DebugContext(ctx, "Running...", "env", appConfig.Config.Env)

	app.WaitForShutdown(ctx)

	return nil
}
