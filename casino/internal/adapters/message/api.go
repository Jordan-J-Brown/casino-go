package message

import (
	"context"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/services/engineregistry"
	"github.com/Jordan-J-Brown/casino-go/pkg/errorx"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/pkg/msg/casinomsg"
	"github.com/Jordan-J-Brown/casino-go/pkg/msg/enginemsg"
	"github.com/Jordan-J-Brown/casino-go/pkg/natsx"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/trace"
)

type API struct {
	nc             *nats.Conn
	logger         logsx.Logger
	tracer         trace.Tracer
	engineRegistry engineregistry.Service
	subs           []*nats.Subscription
}

func RunApiAsync(
	ctx context.Context,
	nc *nats.Conn,
	logger logsx.Logger,
	tracer trace.Tracer,
	engineRegistry engineregistry.Service,
) (*API, error) {
	api := &API{
		nc:             nc,
		logger:         logger,
		tracer:         tracer,
		engineRegistry: engineRegistry,
	}

	regSub, err := nc.Subscribe(enginemsg.SubjectEnginesRegister, natsx.HandleRPC(ctx, logger, tracer, api.handleEngineRegistration))
	if err != nil {
		return nil, err
	}
	api.subs = append(api.subs, regSub)

	healthSub, err := nc.Subscribe(casinomsg.SubjectHealth, func(msg *nats.Msg) {
		api.handleHealthCheck(ctx, msg)
	})
	if err != nil {
		return nil, err
	}
	api.subs = append(api.subs, healthSub)

	return api, nil
}

func (api *API) Shutdown(ctx context.Context) error {
	for _, sub := range api.subs {
		err := sub.Drain()
		if err != nil {
			api.logger.ErrorContext(ctx, "failed to drain subscription",
				"subject", sub.Subject,
				"error", err,
			)
		}
	}
	return nil
}

func (api *API) handleEngineRegistration(ctx context.Context, r enginemsg.EngineRegistration) (*enginemsg.RegisterEngineResponse, *errorx.Error) {
	err := api.engineRegistry.RegisterEngine(ctx, r)
	if err != nil {
		return nil, errorx.New("failed to register engine", errorx.WithChildren(err))
	}

	return &enginemsg.RegisterEngineResponse{}, nil
}

func (api *API) handleHealthCheck(ctx context.Context, msg *nats.Msg) {
	if err := msg.Respond([]byte{}); err != nil {
		api.logger.ErrorContext(ctx, "failed to reply to health check", "error", err)
	}
}
