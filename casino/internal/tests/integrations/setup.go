package integrations

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/casinoapp"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/config"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/tests/integrations/mockengine"
	baseconfig "github.com/Jordan-J-Brown/casino-go/pkg/config"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/pkg/pgdb"
	"github.com/Jordan-J-Brown/casino-go/pkg/retry"
	"github.com/nats-io/nats.go"
)

type SetupConfig struct {
	T                *testing.T
	CasinoConfig     *config.Config
	MockEngineConfig *mockengine.Config //Optional. A mock engine will be created as part of the setup. EngineID does not need to be specified.
}

type Environment struct {
	CasinoApp *casinoapp.Application
}

// Sets up the casino and makes core test dependencies.
//
// [DefaultCasinoConfig] can be used to set up the casino app if no additional configuration is needed.
func Bootstrap(ctx context.Context, cfg SetupConfig) (*Environment, error) {
	casinoApp, err := startCasino(ctx, cfg.CasinoConfig)
	if err != nil {
		return nil, err
	}

	return &Environment{
		CasinoApp: casinoApp,
	}, nil
}

func DefaultCasinoConfig() *config.Config {
	return &config.Config{
		Base: &baseconfig.Base{
			ServiceName:    "casino",
			ServiceVersion: "v0.0.0",
			Env:            baseconfig.Production,
			Nats: baseconfig.Nats{
				URL: nats.DefaultURL,
			},
			Tracer: baseconfig.Tracer{
				ExporterOtlpEndpoint: "otlp://localhost:4317",
				ExporterOtlpInsecure: true,
			},
			Logging: baseconfig.Logging{
				Type:  logsx.LoggerTypeText,
				Level: slog.LevelDebug,
			},
		},
		API: config.API{
			Port: "10000",
		},
		Auth: config.Auth{
			SessionTTL: 5 * time.Minute,
		},
		Games: config.Games{
			BetSizes:            []int64{5, 10, 20, 50, 1_00, 2_00, 5_00, 10_00, 100_00},
			MaxPayoutMultiplier: 250_000,
		},
		Engines: config.Engines{
			ClientTimeoutSeconds: 5 * time.Second,
		},
	}
}

// Starts up a casino application for testing
//
// If casinoCfg is nil, DefaultConfig will be used.
func startCasino(
	ctx context.Context,
	cfg *config.Config,
) (app *casinoapp.Application, err error) {
	os.Setenv(pgdb.EnvPgConnectionString, "postgres://service:password@localhost:5432/service_db?sslmode=disable")

	if cfg == nil {
		cfg = DefaultCasinoConfig()
	}

	logger := logsx.NewLogger(os.Stdout, logsx.Config{
		Type: cfg.Logging.Type,
		HandlerOpts: &slog.HandlerOptions{
			Level: cfg.Logging.Level,
		},
	})

	app, err = casinoapp.Run(ctx, logger, cfg)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			app.Shutdown()
		}
	}()

	err = retry.Backoff(ctx, 3, 100*time.Millisecond, 3*time.Second, func(ctx context.Context) error {
		resp, err := http.DefaultClient.Head("http://localhost:10000/api/v1/health")
		if err == nil && resp.StatusCode == http.StatusNoContent {
			return nil
		}
		return err
	})
	if err != nil {
		return nil, err
	}

	return app, nil
}
