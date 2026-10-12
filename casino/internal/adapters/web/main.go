package web

import (
	"context"
	"net"
	"net/http"
	"reflect"
	"time"

	_ "github.com/Jordan-J-Brown/casino-go/casino/docs"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/adapters/web/handlers/authhandler"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/adapters/web/handlers/gamehandler"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/adapters/web/handlers/userhandler"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/adapters/web/middleware"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/services/gameservice"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/services/userservice"
	"github.com/Jordan-J-Brown/casino-go/pkg/config"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.opentelemetry.io/otel/trace"
)

// @title Casino API
// @version 1.0
// @description HTTP API for casino games and player accounts.
// @BasePath /api/v1
// @tag.name Health
// @tag.name Authentication
// @tag.name Games
// @tag.name Users
//

type ApiConfig struct {
	Env         config.Env
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
		// chimiddleware.Logger, //TODO: use the application's logger
		chimiddleware.RequestID,
		middleware.CorrelationID,
		chimiddleware.Recoverer,
		chimiddleware.Timeout(10*time.Second),
	)

	if !cfg.Env.IsProd() {
		router.Get("/swagger/*", httpSwagger.WrapHandler)
	}

	authHandler := authhandler.New(cfg.Logger, cfg.Tracer, cfg.UserService)
	gameHandler := gamehandler.New(cfg.Logger, cfg.Tracer, cfg.GameService)
	userHandler := userhandler.New(cfg.Logger, cfg.Tracer, cfg.UserService)

	router.Route("/api/v1", func(r chi.Router) {
		r.Head("/health", health)

		r.Post("/auth", authHandler.Authenticate)

		r.Route("/games", func(r chi.Router) {
			r.Use(authHandler.Middleware)

			r.Post("/setup", gameHandler.Setup)
			r.Post("/play", gameHandler.Play)
			r.Post("/next", gameHandler.Next)
			r.Post("/close", gameHandler.Close)
		})

		r.Route("/users", func(r chi.Router) {
			r.Use(authHandler.Middleware)

			r.Post("/balance", userHandler.Balance)
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
			if err == http.ErrServerClosed {
				return
			}
			panic("listen and serve error: " + err.Error())
		}
	}(cfg.Logger)

	return server.Shutdown
}

// Health reports that the API process is alive.
// @Tags Health
// @Summary Check API health
// @Description Returns no content when the API is running.
// @ID health
// @Success 204
// @Router /health [head]
func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func panicOnZero(value any, msg string) {
	rValue := reflect.ValueOf(value)
	if isValid := rValue.IsValid(); isValid && rValue.IsZero() || !isValid {
		panic(msg)
	}
}
