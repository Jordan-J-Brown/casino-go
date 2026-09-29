package web

import (
	"context"
	"net"
	"net/http"
	"reflect"
	"time"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/adapters/web/handlers/authhandler"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/adapters/web/handlers/gamehandler"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/adapters/web/middleware"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/services/gameservice"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/services/userservice"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/trace"
)

//TODO: setup swagger

type ApiConfig struct {
	Port        string
	Logger      logsx.Logger
	Tracer      trace.Tracer
	UserService userservice.Service
	GameService gameservice.Service
}

func RunApiAsync(ctx context.Context, cfg ApiConfig) (shutdownFn func(ctx context.Context) error) {
	panicOnZero(cfg.Port, "port config is missing")
	panicOnZero(cfg.Logger, "logger is nil")
	panicOnZero(cfg.Tracer, "tracer is nil")
	panicOnZero(cfg.UserService, "user service is nil")
	panicOnZero(cfg.GameService, "game service is nil")

	router := chi.NewRouter()

	router.Use(
		chimiddleware.Logger, //TODO: use the application's logger
		chimiddleware.RequestID,
		middleware.CorrelationID,
		chimiddleware.Recoverer,
		chimiddleware.Timeout(10*time.Second),
	)

	authHandler := authhandler.New(
		cfg.Logger,
		cfg.Tracer,
		cfg.UserService,
	)
	gameHandler := gamehandler.New(
		cfg.Logger,
		cfg.Tracer,
		cfg.GameService,
	)

	router.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth", authHandler.Authenticate)

		r.Route("/games", func(r chi.Router) {
			r.Use(authHandler.Middleware)

			r.Post("/setup", gameHandler.Setup)
			r.Post("/play", gameHandler.Play)
			r.Post("/next", gameHandler.Next)
			r.Post("/close", gameHandler.Close)
		})
	})

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	go func(logger logsx.Logger) {
		err := server.ListenAndServe()
		if err != nil {
			panic("listen and serve error: " + err.Error())
		}
	}(cfg.Logger)

	return server.Shutdown
}

func panicOnZero(value any, msg string) {
	rValue := reflect.ValueOf(value)
	if isValid := rValue.IsValid(); isValid && rValue.IsZero() || !isValid {
		panic(msg)
	}
}
