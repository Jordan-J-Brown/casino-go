package telemetry

import (
	"context"
	"errors"
	"os"

	"github.com/Jordan-J-Brown/casino-go/pkg/ternary"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
)

const (
	EnvOtelServiceName          string = "OTEL_SERVICE_NAME"
	EnvOtelServiceVersion       string = "OTEL_SERVICE_VERSION"
	EnvOtelExporterOtlpEndpoint string = "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"
	EnvOtelExporterOtlpInsecure string = "OTEL_EXPORTER_OTLP_INSECURE"
)

type OtelConfig struct {
	ServiceName          string
	ServiceVersion       string
	ExporterOtlpEndpoint string
	ExporterOtlpInsecure bool
}

// SetupOtelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOtelSDK(ctx context.Context, cfg OtelConfig) (shutDownFunc func(context.Context) error, err error) {
	if err := setupEnv(cfg); err != nil {
		return nil, err
	}

	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := newTracerProvider(ctx)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	return shutdown, err
}

func setupEnv(cfg OtelConfig) error {
	if _, exists := os.LookupEnv(EnvOtelServiceName); !exists {
		if cfg.ServiceName == "" {
			return errors.New("otel service name config is missing")
		}
		os.Setenv(EnvOtelServiceName, cfg.ServiceName)
	}

	if _, exists := os.LookupEnv(EnvOtelServiceVersion); !exists {
		if cfg.ServiceVersion == "" {
			return errors.New("otel service version config is missing")
		}
		os.Setenv(EnvOtelServiceVersion, cfg.ServiceVersion)
	}

	if _, exists := os.LookupEnv(EnvOtelExporterOtlpEndpoint); !exists {
		if cfg.ExporterOtlpEndpoint == "" {
			return errors.New("otel exporter OTLP endpoint config is missing")
		}
		os.Setenv(EnvOtelExporterOtlpEndpoint, cfg.ExporterOtlpEndpoint)
	}

	if _, exists := os.LookupEnv(EnvOtelExporterOtlpInsecure); !exists {
		os.Setenv(EnvOtelExporterOtlpInsecure, ternary.Eval(cfg.ExporterOtlpInsecure, "true", "false"))
	}

	return nil
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
	)
}

func newTracerProvider(ctx context.Context) (*trace.TracerProvider, error) {
	exporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, err
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
	)

	return tracerProvider, nil
}
