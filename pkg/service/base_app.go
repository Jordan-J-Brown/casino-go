package service

import (
	"context"
	"log"
	"os"
	"os/signal"
	"slices"

	"github.com/Jordan-J-Brown/casino-go/pkg/config"
	"github.com/Jordan-J-Brown/casino-go/pkg/errorx"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/pkg/telemetry"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// BaseApplication is responsible for initialising and storing core components that
// each service needs.
//
// Shutdown should be called to handle graceful shutdown.
type BaseApplication struct {
	shutdownChan  <-chan struct{}
	signalStop    context.CancelFunc
	shutdownFuncs []func(context.Context) error
	Logger        logsx.Logger
	Tracer        trace.Tracer
	NatsConn      *nats.Conn
}

type BaseAppConfig struct {
	ServiceName string
	NatsCfg     *config.Nats
	TracerCfg   *config.Tracer
	Logger      logsx.Logger
}

func NewBaseApplication(cfg BaseAppConfig) (ctx context.Context, base *BaseApplication, err error) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	baseApp := &BaseApplication{
		shutdownChan:  ctx.Done(),
		signalStop:    stop,
		shutdownFuncs: make([]func(context.Context) error, 0, 2),
		Logger:        cfg.Logger,
	}
	defer func() {
		if err != nil {
			baseApp.Logger.ErrorContext(ctx, "failed to create base application", "error", err)
			stop()
			baseApp.shutdown(ctx)
		}
	}()
	if cfg.Logger == nil {
		baseApp.Logger = logsx.NopLogger{}
	}

	//Setup tracing
	otelShutDown, err := telemetry.SetupOtelSDK(ctx, cfg.TracerCfg)
	if err != nil {
		return ctx, nil, err
	}
	baseApp.shutdownFuncs = append(baseApp.shutdownFuncs, otelShutDown)
	baseApp.Tracer = otel.Tracer(cfg.ServiceName)

	//Setup NATS
	nc, err := nats.Connect(cfg.NatsCfg.URL,
		nats.Name(cfg.ServiceName),
		nats.Timeout(cfg.NatsCfg.ConnectTimeout),
		nats.ReconnectWait(cfg.NatsCfg.ReconnectWait),
		nats.MaxReconnects(cfg.NatsCfg.MaxReconnects),
		nats.DrainTimeout(cfg.NatsCfg.DrainTimeout),
		nats.PingInterval(cfg.NatsCfg.PingInterval),
		nats.MaxPingsOutstanding(cfg.NatsCfg.MaxPingsOutstanding),
	)
	if err != nil {
		return ctx, nil, errorx.New("failed to create NATS connection", errorx.WithChildren(err))
	}
	baseApp.NatsConn = nc
	baseApp.shutdownFuncs = append(baseApp.shutdownFuncs, func(ctx context.Context) error {
		nc := baseApp.NatsConn
		if err := nc.Drain(); err != nil {
			nc.Close()
			return err
		}

		return nil
	})

	return ctx, baseApp, nil
}

func (base *BaseApplication) RegisterShutdownFunc(fn func(context.Context) error) {
	base.shutdownFuncs = append(base.shutdownFuncs, fn)
}

func (base *BaseApplication) WaitForShutdown(ctx context.Context) {
	<-ctx.Done()

	base.Logger.InfoContext(ctx, "\nGracefully shutting down...")

	base.shutdown(context.Background())
}

func (base *BaseApplication) shutdown(ctx context.Context) {
	for i, shutdown := range slices.Backward(base.shutdownFuncs) {
		if err := shutdown(ctx); err != nil {
			log.Fatalf("failed to run shutdown at index %d. Error: %s", i, err.Error())
		}
	}
}
