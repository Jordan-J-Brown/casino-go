package mockengine

import (
	"context"
	"log/slog"
	"os"
	"slices"
	"testing"

	"github.com/Jordan-J-Brown/casino-go/pkg/config"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/pkg/msg/enginemsg"
	"github.com/Jordan-J-Brown/casino-go/pkg/telemetry"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const (
	EngineID string = "mock-engine"
)

type Engine struct {
	t         *testing.T
	tracer    trace.Tracer
	logger    logsx.Logger
	nc        *nats.Conn
	srv       *enginemsg.Server
	shutdowns []func(context.Context) error
}

type Config struct {
	Logging      config.Logging
	ServerConfig enginemsg.ServerConfig
}

// Starts a mock engine using the provided config.
//
// [Engine.Shutdown] MUST be called upon completion of tests.
func Start(ctx context.Context, t *testing.T, cfg Config) (*Engine, error) {
	cfg.ServerConfig.EngineID = EngineID

	otelShutdown, err := telemetry.SetupOtelSDK(ctx, telemetry.OtelConfig{
		ServiceName:          EngineID,
		ServiceVersion:       "v0.0.0",
		ExporterOtlpEndpoint: "http://localhost:4317",
		ExporterOtlpInsecure: true,
	})
	if err != nil {
		return nil, err
	}

	tracer := otel.GetTracerProvider().Tracer(EngineID)

	logger := logsx.NewLogger(os.Stdout, logsx.Config{
		Type: cfg.Logging.Type,
		HandlerOpts: &slog.HandlerOptions{
			Level: cfg.Logging.Level,
		},
	})

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, err
	}

	srv, err := enginemsg.StartServer(ctx, nc, logger, tracer, cfg.ServerConfig)
	if err != nil {
		return nil, err
	}

	return &Engine{
		tracer: tracer,
		logger: logger,
		nc:     nc,
		srv:    srv,
		shutdowns: []func(context.Context) error{
			otelShutdown,
			srv.Shutdown,
		},
	}, nil
}

func (e *Engine) Shutdown() {
	for _, shutdown := range slices.Backward(e.shutdowns) {
		ctx := context.Background()
		err := shutdown(ctx)
		if err != nil {
			e.t.Log(err)
		}
	}
}
