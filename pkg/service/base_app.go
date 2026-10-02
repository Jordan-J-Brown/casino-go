package service

import (
	"context"
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
	shutdownFuncs []func(context.Context) error
	Logger        logsx.Logger
	Tracer        trace.Tracer
	NatsConn      *nats.Conn
}

type BaseAppConfig struct {
	ServiceName    string
	ServiceVersion string
	NatsCfg        *config.Nats
	TracerCfg      *config.Tracer
}

func NewBaseApplication(ctx context.Context, logger logsx.Logger, cfg BaseAppConfig) (base *BaseApplication, err error) {
	baseApp := &BaseApplication{
		shutdownChan:  ctx.Done(),
		shutdownFuncs: make([]func(context.Context) error, 0, 2),
		Logger:        logger,
	}
	defer func() {
		if err != nil {
			baseApp.Logger.ErrorContext(ctx, "failed to create base application", "error", err)
			baseApp.Shutdown()
		}
	}()
	if logger == nil {
		baseApp.Logger = logsx.NopLogger{}
	}

	//Setup tracing
	otelShutDown, err := telemetry.SetupOtelSDK(ctx, telemetry.OtelConfig{
		ServiceName:          cfg.ServiceName,
		ServiceVersion:       cfg.ServiceVersion,
		ExporterOtlpEndpoint: cfg.TracerCfg.ExporterOtlpEndpoint,
		ExporterOtlpInsecure: cfg.TracerCfg.ExporterOtlpInsecure,
	})
	if err != nil {
		return nil, err
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
		return nil, errorx.New("failed to create NATS connection", errorx.WithChildren(err))
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

	return baseApp, nil
}

func (base *BaseApplication) RegisterShutdownFunc(fn func(context.Context) error) {
	base.shutdownFuncs = append(base.shutdownFuncs, fn)
}

func (base *BaseApplication) Shutdown() {
	ctx := context.Background()
	for i, shutdown := range slices.Backward(base.shutdownFuncs) {
		if err := shutdown(ctx); err != nil {
			base.Logger.ErrorContext(ctx, "failed to run shutdown at index %d. Error: %s", i, err.Error())
		}
	}
}
