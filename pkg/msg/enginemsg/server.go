package enginemsg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/pkg/mimex"
	"github.com/Jordan-J-Brown/casino-go/pkg/msg/casinomsg"
	"github.com/Jordan-J-Brown/casino-go/pkg/natsx"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Establishes subscriptions to handle RPC calls between the casino and engines.
//
// All engines must embed this in their application.
//
// [Server.Shutdown] should be called upon graceful shutdown to drain subscriptions.
type Server struct {
	nc     *nats.Conn
	client *natsx.Client
	logger logsx.Logger
	tracer trace.Tracer
	cfg    ServerConfig
	subs   []*nats.Subscription
}

type ServerConfig struct {
	EngineID     string
	RegisterFunc func(ctx context.Context) (*EngineRegistration, error)
	SetupHandler natsx.HandlerFunc[SetupRequest, SetupResponse]
	PlayHandler  natsx.HandlerFunc[PlayRequest, RoundResponse]
	NextHandler  natsx.HandlerFunc[NextRequest, RoundResponse]
	CloseHandler natsx.HandlerFunc[CloseRequest, CloseResponse]
}

func StartServer(
	ctx context.Context,
	nc *nats.Conn,
	logger logsx.Logger,
	tracer trace.Tracer,
	cfg ServerConfig,
) (*Server, error) {
	const engineQ = "engines"

	client, err := natsx.NewClient(ctx, natsx.ConnPoolConfig{
		URL: nc.ConnectedAddr(),
	})
	if err != nil {
		return nil, err
	}

	client.
		SetContentType(mimex.ContentTypeJSON).
		SetRetry(3, 100*time.Millisecond, 3*time.Second).
		SetTimeout(3 * time.Second)

	server := &Server{
		nc:     nc,
		client: client,
		logger: logger,
		tracer: tracer,
		cfg:    cfg,
		subs:   make([]*nats.Subscription, 0, 4),
	}

	if err := server.registerEngine(ctx); err != nil {
		return nil, err
	}

	setupSub, err := nc.QueueSubscribe(
		CreateEngineSubject(cfg.EngineID, "setup"), engineQ, natsx.HandleRPC(ctx, logger, tracer, cfg.SetupHandler),
	)
	if err != nil {
		return nil, err
	}
	server.subs = append(server.subs, setupSub)

	playSub, err := nc.QueueSubscribe(
		CreateEngineSubject(cfg.EngineID, "play"), engineQ, natsx.HandleRPC(ctx, logger, tracer, cfg.PlayHandler),
	)
	if err != nil {
		return nil, err
	}
	server.subs = append(server.subs, playSub)

	nextSub, err := nc.QueueSubscribe(
		CreateEngineSubject(cfg.EngineID, "next"), engineQ, natsx.HandleRPC(ctx, logger, tracer, cfg.NextHandler),
	)
	if err != nil {
		return nil, err
	}
	server.subs = append(server.subs, nextSub)

	closeSub, err := nc.QueueSubscribe(
		CreateEngineSubject(cfg.EngineID, "close"), engineQ, natsx.HandleRPC(ctx, logger, tracer, cfg.CloseHandler),
	)
	if err != nil {
		return nil, err
	}
	server.subs = append(server.subs, closeSub)

	go server.startHealthChecks(ctx)

	return &Server{
		nc:     nc,
		logger: logger,
		cfg:    cfg,
		client: client,
	}, nil
}

func (server *Server) Shutdown(ctx context.Context) error {
	if err := server.client.Close(); err != nil {
		server.logger.ErrorContext(ctx, "failed to close natsx client", "error", err)
	}

	for _, sub := range server.subs {
		err := sub.Drain()
		if err != nil {
			server.logger.ErrorContext(ctx, "failed to drain subscription",
				"subject", sub.Subject,
				"queue", sub.Queue,
				"error", err.Error(),
			)
		}
	}
	return nil
}

func (server *Server) registerEngine(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			err = errors.New("failed to register engine: " + err.Error())
		}
	}()

	ctx, span := server.tracer.Start(ctx, "engine-register", trace.WithAttributes(
		attribute.String("engine_id", server.cfg.EngineID),
	))
	defer span.End()

	header := nats.Header{}
	otel.GetTextMapPropagator().Inject(ctx, &natsx.NatsOtelHeaderCarrier{Header: header})

	registerData, err := server.cfg.RegisterFunc(ctx)
	if err != nil {
		return err
	}
	registerData.EngineID = server.cfg.EngineID

	var response RegisterEngineResponse
	var errResponse natsx.ErrorResponse
	req, err := server.client.R().
		SetHeaders(header).
		SetBody(&registerData).
		SetResult(&response).
		SetError(&errResponse).
		Exec(ctx, SubjectEnginesRegister)
	if err != nil {
		return err
	}
	if req.IsError() {
		return &errResponse
	}

	return nil
}

func (server *Server) checkCasinoHealth(ctx context.Context) error {
	server.logger.DebugContext(ctx, "Checking casino health...")
	_, err := server.client.R().Exec(ctx, casinomsg.SubjectHealth)
	if err == nil {
		server.logger.DebugContext(ctx, "Casino is healthy")
	}
	return err
}

func (server *Server) startHealthChecks(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var lostConnection bool
	for {
		select {
		case <-ticker.C:
			err := server.checkCasinoHealth(ctx)
			if err != nil {
				server.logger.ErrorContext(ctx, "failed to check casino health", "error", err)
				lostConnection = true
			}

			if lostConnection {
				server.logger.DebugContext(ctx, "Attempting to re-register engine...")
				err = server.registerEngine(ctx)
				if err != nil {
					server.logger.ErrorContext(ctx, "failed to re-register engine after disconnect", err)
					continue
				}

				lostConnection = false
			}
		case <-ctx.Done():
			return
		}
	}
}

func CreateEngineSubject(engineID, action string) string {
	return fmt.Sprintf("engines.%s.%s", engineID, action)
}
