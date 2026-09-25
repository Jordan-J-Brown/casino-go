package engineregistry

import (
	"context"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/enginemodels"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/pkg/msg/enginemsg"
)

type Store interface {
	GetRegistration(ctx context.Context, engineID string) (*enginemsg.EngineRegistration, error)
	EngineIsRegistered(ctx context.Context, engineID string) (bool, error)
	PutRegistration(ctx context.Context, reg enginemsg.EngineRegistration) error
}

type Service interface {
	GetEngineRegistration(ctx context.Context, engineID string) (*enginemsg.EngineRegistration, error)
	RegisterEngine(ctx context.Context, reg enginemsg.EngineRegistration) error
}

type service struct {
	logger logsx.Logger
	store  Store
}

func New(
	ctx context.Context,
	logger logsx.Logger,
	store Store,
) Service {
	service := &service{
		logger: logger,
		store:  store,
	}

	return service
}

func (service *service) GetEngineRegistration(ctx context.Context, engineID string) (*enginemsg.EngineRegistration, error) {
	reg, err := service.store.GetRegistration(ctx, engineID)
	if err != nil {
		return nil, err
	}

	return reg, nil
}

func (service *service) RegisterEngine(ctx context.Context, reg enginemsg.EngineRegistration) error {
	err := service.store.PutRegistration(ctx, reg)
	if err == enginemodels.ErrEngineRegistered {
		return nil
	}

	return err
}
