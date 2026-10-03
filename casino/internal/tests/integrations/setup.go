package integrations

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/casinoapp"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/config"
	baseconfig "github.com/Jordan-J-Brown/casino-go/pkg/config"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/pkg/pgdb"
	"github.com/Jordan-J-Brown/casino-go/pkg/retry"
	"github.com/Jordan-J-Brown/casino-go/pkg/testingx"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/nats-io/nats.go"
	"github.com/testcontainers/testcontainers-go"
	natstestcontainer "github.com/testcontainers/testcontainers-go/modules/nats"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
)

type SetupConfig struct {
	T            *testing.T
	CasinoConfig *config.Config
}

type Environment struct {
	CasinoApp            *casinoapp.Application
	NatsConnString       string
	OtelExporterEndpoint string
}

// Sets up the casino and makes core test dependencies.
//
// [DefaultCasinoConfig] can be used to set up the casino app if no additional configuration is needed.
func Bootstrap(ctx context.Context, t *testing.T, cfg SetupConfig) (*Environment, error) {
	if cfg.CasinoConfig == nil {
		cfg.CasinoConfig = DefaultCasinoConfig()
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	projectDir, err := testingx.ProjectRoot(workingDir)
	if err != nil {
		return nil, err
	}

	otelExporterEndpoint, err := setupOtelExporterContainer(ctx, t)
	if err != nil {
		return nil, err
	}
	cfg.CasinoConfig.Tracer.ExporterOtlpEndpoint = otelExporterEndpoint
	cfg.CasinoConfig.Tracer.ExporterOtlpInsecure = true

	natsConnectionString, err := setupNatsContainer(ctx, t)
	if err != nil {
		return nil, err
	}
	cfg.CasinoConfig.Nats.URL = natsConnectionString

	pgConnectionString, err := setupPostgresContainer(ctx, t)
	if err != nil {
		return nil, err
	}

	//Migrations
	m, err := migrate.New("file://"+filepath.Join(projectDir, "migrations"), pgConnectionString)
	if err != nil {
		return nil, errors.New("failed to create migrations instance: " + err.Error())
	}
	if err := m.Up(); err != nil {
		return nil, errors.New("failed to migrate up: " + err.Error())
	}

	//Casino
	casinoApp, err := startCasino(ctx, cfg.CasinoConfig, pgConnectionString)
	if err != nil {
		return nil, err
	}

	return &Environment{
		CasinoApp:            casinoApp,
		NatsConnString:       natsConnectionString,
		OtelExporterEndpoint: otelExporterEndpoint,
	}, nil
}

func setupNatsContainer(ctx context.Context, t *testing.T) (connString string, err error) {
	natsContainer, err := natstestcontainer.Run(ctx, "nats:2.10-alpine")
	if err != nil {
		return "", err
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(natsContainer); err != nil {
			t.Log("failed to terminate nats container:", err)
		}
	})

	natsConnString, err := natsContainer.ConnectionString(ctx)
	if err != nil {
		return "", errors.New("failed to get nats connection string: " + err.Error())
	}
	t.Log("NATS Connection String:", natsConnString)

	return natsConnString, nil
}

func setupPostgresContainer(ctx context.Context, t *testing.T) (connString string, err error) {
	pgContainer, err := pgcontainer.Run(ctx, "postgres:18-alpine",
		pgcontainer.WithPassword("password"),
	)
	if err != nil {
		return "", err
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(pgContainer); err != nil {
			t.Log("failed to terminate postgres container:", err)
		}
	})

	pgHealth := func() error {
		t.Log("Checking postgres health...")

		_, r, err := pgContainer.Exec(ctx, []string{"pg_isready", "-U", "postgres", "-d", "postgres"})
		if err != nil {
			return err
		}

		output, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		if strings.Contains(string(output), "no response") {
			return errors.New("not ready")
		}

		return nil
	}

	for i := 0; pgHealth() != nil; i++ {
		time.Sleep(1 * time.Second)
		if i == 2 {
			return "", errors.New("failed to check health of postgres container")
		}
	}

	t.Log("Postgres healthy")

	pgConnectionString, err := pgContainer.ConnectionString(ctx)
	if err != nil {
		return "", err
	}
	pgConnectionString += "sslmode=disable"

	t.Log("PG Connection String:", pgConnectionString)

	return pgConnectionString, nil
}

func setupOtelExporterContainer(ctx context.Context, t *testing.T) (exporterEndpoint string, err error) {
	container, err := testcontainers.Run(ctx, "cr.jaegertracing.io/jaegertracing/jaeger:2.20.0",
		testcontainers.WithExposedPorts("4317", "4318"),
	)
	if err != nil {
		return "", err
	}
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Log(err)
		}
	})

	return container.Endpoint(ctx, "http")
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
			BetSizes:            []int64{5, 10, 20, 50, 1_00, 2_00, 5_00, 10_00, 20_00, 100_00},
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
	pgConnString string,
) (app *casinoapp.Application, err error) {
	os.Setenv(pgdb.EnvPgConnectionString, pgConnString)

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
