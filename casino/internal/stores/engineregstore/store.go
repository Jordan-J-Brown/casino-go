package engineregstore

import (
	"context"
	"sync"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/enginemodels"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/services/engineregistry"
	"github.com/Jordan-J-Brown/casino-go/pkg/msg/enginemsg"
)

type Store struct {
	cache sync.Map
}

func New() *Store {
	return &Store{
		cache: sync.Map{},
	}
}

func (s *Store) EngineIsRegistered(ctx context.Context, engineID string) (bool, error) {
	_, ok := s.cache.Load(engineID)
	return ok, nil
}

func (s *Store) GetRegistration(ctx context.Context, engineID string) (*enginemsg.EngineRegistration, error) {
	value, ok := s.cache.Load(engineID)
	if !ok {
		return nil, nil
	}

	reg := value.(enginemsg.EngineRegistration)

	return &reg, nil
}

func (s *Store) PutRegistration(ctx context.Context, reg enginemsg.EngineRegistration) error {
	if _, exists := s.cache.LoadOrStore(reg.EngineID, reg); exists {
		return enginemodels.ErrEngineRegistered	
	}
	return nil
}

var _ engineregistry.Store = &Store{}
