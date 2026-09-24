package engineregistry

import (
	"context"
	"encoding/json"

	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/pkg/msg/enginemsg"
	"github.com/nats-io/nats.go"
)

type Store interface {
	GetRegistration(ctx context.Context, engineID string) (*enginemsg.EngineRegistraton, error)
	EngineIsRegistered(ctx context.Context, engineID string) (bool, error)
	PutRegistration(ctx context.Context, reg enginemsg.EngineRegistraton) error
}

type Service interface {
	GetEngineRegistration(ctx context.Context, engineID string) (*enginemsg.EngineRegistraton, error)
	Shutdown(ctx context.Context) error
}

type service struct {
	nc     *nats.Conn
	logger logsx.Logger
	store  Store
	sub    *nats.Subscription
}

func New(
	ctx context.Context,
	nc *nats.Conn,
	logger logsx.Logger,
	store Store,
) (Service, error) {
	service := &service{
		nc:     nc,
		logger: logger,
		store:  store,
	}

	sub, err := nc.Subscribe(enginemsg.SubjectEnginesRegister, service.handleRegistration(ctx))
	if err != nil {
		return nil, err
	}

	service.sub = sub

	return service, nil
}

func (service *service) GetEngineRegistration(ctx context.Context, engineID string) (*enginemsg.EngineRegistraton, error) {
	reg, err := service.store.GetRegistration(ctx, engineID)
	if err != nil {
		return nil, err
	}

	return reg, nil
}

func (service *service) Shutdown(ctx context.Context) error {
	err := service.sub.Drain()
	if err != nil {
		service.logger.ErrorContext(ctx, "failed to drain engine registration subscription", "error", err)
	}
	return nil
}

func (service *service) handleRegistration(ctx context.Context) nats.MsgHandler {
	return func(msg *nats.Msg) {
		var reg enginemsg.EngineRegistraton
		if err := json.Unmarshal(msg.Data, &reg); err != nil {
			service.logger.ErrorContext(ctx, "failed to unmarshal engine registration", "error", err)
			return
		}

		registered, err := service.store.EngineIsRegistered(ctx, reg.EngineID)
		if err != nil {
			service.logger.ErrorContext(ctx, "failed to check if engine is registered",
				"engine_id", reg.EngineID,
				"error", err,
			)
			return
		}

		if registered {
			return
		}

		err = service.store.PutRegistration(ctx, reg)
		if err != nil {
			service.logger.ErrorContext(ctx, "failed to register engine",
				"engine_id", reg.EngineID,
				"error", err,
			)
		}

		service.logger.DebugContext(ctx, "Engine_Registered", "engine_id", reg.EngineID)
	}
}
