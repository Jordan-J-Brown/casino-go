package enginemsg

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/pkg/msg"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/trace"
)

// Establishes subscriptions to handle RPC calls between the casino and engines.
//
// All engines must embed this in their application.
//
// [Server.Shutdown] should be called upon graceful shutdown to drain subscriptions.
type Server struct {
	nc     *nats.Conn
	logger logsx.Logger
	tracer trace.Tracer
	cfg    ServerConfig
	subs   []*nats.Subscription
}

type ServerConfig struct {
	EngineID     string
	RegisterFunc func(ctx context.Context) (*EngineRegistraton, error)
	SetupHandler msg.HandlerFunc[SetupRequest, SetupResponse]
	PlayHandler  msg.HandlerFunc[PlayRequest, RoundResponse]
	NextHandler  msg.HandlerFunc[NextRequest, RoundResponse]
	CloseHandler msg.HandlerFunc[CloseRequest, CloseResponse]
}

func StartServer(
	ctx context.Context,
	nc *nats.Conn,
	logger logsx.Logger,
	tracer trace.Tracer,
	cfg ServerConfig,
) (*Server, error) {
	const engineQ = "engines"

	server := &Server{
		nc:     nc,
		logger: logger,
		tracer: tracer,
		cfg:    cfg,
		subs:   make([]*nats.Subscription, 0, 4),
	}

	if err := server.registerEngine(ctx); err != nil {
		return nil, err
	}

	setupSub, err := nc.QueueSubscribe(
		CreateEngineSubject(cfg.EngineID, "setup"), engineQ, msg.HandleRPC(ctx, logger, cfg.SetupHandler),
	)
	if err != nil {
		return nil, err
	}
	server.subs = append(server.subs, setupSub)

	playSub, err := nc.QueueSubscribe(
		CreateEngineSubject(cfg.EngineID, "play"), engineQ, msg.HandleRPC(ctx, logger, cfg.PlayHandler),
	)
	if err != nil {
		return nil, err
	}
	server.subs = append(server.subs, playSub)

	nextSub, err := nc.QueueSubscribe(
		CreateEngineSubject(cfg.EngineID, "next"), engineQ, msg.HandleRPC(ctx, logger, cfg.NextHandler),
	)
	if err != nil {
		return nil, err
	}
	server.subs = append(server.subs, nextSub)

	closeSub, err := nc.QueueSubscribe(
		CreateEngineSubject(cfg.EngineID, "close"), engineQ, msg.HandleRPC(ctx, logger, cfg.CloseHandler),
	)
	if err != nil {
		return nil, err
	}
	server.subs = append(server.subs, closeSub)

	return &Server{
		nc:     nc,
		logger: logger,
		cfg:    cfg,
	}, nil
}

func (server *Server) Shutdown(ctx context.Context) error {
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

func (server *Server) registerEngine(ctx context.Context) error {
	registerData, err := server.cfg.RegisterFunc(ctx)
	if err != nil {
		return err
	}
	registerData.EngineID = server.cfg.EngineID

	data, err := json.Marshal(registerData)
	if err != nil {
		return err
	}

	err = server.nc.Publish(SubjectEnginesRegister, data)
	if err != nil {
		return err
	}

	return nil
}

func CreateEngineSubject(engineID, action string) string {
	return fmt.Sprintf("engines.%s.%s", engineID, action)
}
