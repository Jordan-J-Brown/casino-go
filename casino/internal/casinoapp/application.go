package casinoapp

import (
	"context"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/adapters/message"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/adapters/web"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/clients/engineclient"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/config"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/services/engineregistry"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/services/gameservice"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/services/userservice"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/stores/casinostorepg"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/stores/engineregstore"
	baseconfig "github.com/Jordan-J-Brown/casino-go/pkg/config"
	"github.com/Jordan-J-Brown/casino-go/pkg/errorx"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/pkg/pgdb"
	"github.com/Jordan-J-Brown/casino-go/pkg/service"
)

type Application struct {
	*service.BaseApplication
	config   *config.Config
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

	app.Services.GameService = gameservice.New(ctx, gameservice.Config{
		MaxPayoutMultiplier: app.config.Games.MaxPayoutMultiplier,
		EngineRegistry:      engineRegistry,
		EngineClient:        engineclient.New(app.NatsConn, app.Tracer, app.config.Engines.ClientTimeoutSeconds),
		BetSizes:            app.config.Games.BetSizes,
		StoreProvider: &casinostorepg.CasinoGameStoreProvider{
			Pool:   pgPool,
			Logger: app.Logger,
		},
		Logger: app.Logger,
	})

	app.Services.UserService = userservice.New(userservice.Config{
		SessionTTL: app.config.Auth.SessionTTL,
		StoreProvider: &casinostorepg.UserStoreProvider{
			Pool:   pgPool,
			Logger: app.Logger,
		},
	})

	return nil
}

// Runs the casino application.
//
// [Application.Shutdown] MUST be called and completed before exiting the program
func Run(ctx context.Context, logger logsx.Logger, cfg *config.Config) (app *Application, err error) {
	defer func() {
		if err != nil {
			app.Shutdown()
		}
	}()

	app = &Application{
		config: cfg,
	}

	if cfg.Env == baseconfig.Local {
		errorx.OutputStack = true
	}

	base, err := service.NewBaseApplication(ctx, logger, service.BaseAppConfig{
		ServiceName:    cfg.ServiceName,
		ServiceVersion: cfg.ServiceVersion,
		NatsCfg:        &cfg.Nats,
		TracerCfg:      &cfg.Tracer,
	})
	if err != nil {
		return app, errorx.New("failed to create base application", errorx.WithChildren(err))
	}
	app.BaseApplication = base

	if err := app.initServices(ctx); err != nil {
		return app, errorx.New("failed to initialise services", errorx.WithChildren(err))
	}

	msgAPI, err := message.RunApiAsync(ctx, app.NatsConn, app.Logger, app.Tracer, app.Services.EngineRegistry)
	if err != nil {
		return app, errorx.New("failed to start message server", errorx.WithChildren(err))
	}
	app.RegisterShutdownFunc(msgAPI.Shutdown)

	webApiShutdown := web.RunApiAsync(ctx, web.ApiConfig{
		Port:        app.config.API.Port,
		Logger:      app.Logger,
		Tracer:      app.Tracer,
		UserService: app.Services.UserService,
		GameService: app.Services.GameService,
	})
	app.RegisterShutdownFunc(webApiShutdown)

	app.Logger.DebugContext(ctx, "Running...", "env", cfg.Env)

	return
}
