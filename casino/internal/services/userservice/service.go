package userservice

import (
	"context"
	"time"
	"uuid"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/usermodels"
)

type Service interface {
	GetUserByID(ctx context.Context, id string) (*usermodels.User, error)
	GetUserFromSessionID(ctx context.Context, sessionID string) (*usermodels.User, error)
	CreateUserWithBalance(ctx context.Context, username string, balance int64) (*usermodels.User, error)
	CreateSessionForUser(ctx context.Context, user *usermodels.User) (sessionID string, err error)
}

type service struct {
	SessionTTL    time.Duration
	StoreProvider StoreProvider
}

type Config struct {
	SessionTTL    time.Duration
	StoreProvider StoreProvider
}

func New(cfg Config) Service {
	if cfg.SessionTTL == 0 {
		cfg.SessionTTL = 30 * time.Minute
	}

	return &service{
		SessionTTL:    cfg.SessionTTL,
		StoreProvider: cfg.StoreProvider,
	}
}

func (s *service) GetUserByID(ctx context.Context, id string) (user *usermodels.User, err error) {
	err = s.StoreProvider.UseStore(ctx, func(store Store) error {
		user, err = store.GetUserByID(ctx, id)
		return err
	})
	return
}

func (s *service) GetUserFromSessionID(ctx context.Context, sessionID string) (user *usermodels.User, err error) {
	s.StoreProvider.UseStore(ctx, func(store Store) error {
		user, err = store.GetUserFromSessionID(ctx, sessionID)
		return err
	})
	return
}

func (s *service) CreateUserWithBalance(ctx context.Context, username string, balance int64) (user *usermodels.User, err error) {
	if !isValidUsername(username) {
		return nil, usermodels.ErrInvalidUsername
	}
	if balance < 0 {
		return nil, usermodels.ErrInvalidBalance
	}

	err = s.StoreProvider.UseStore(ctx, func(store Store) error {
		user, err = store.CreateUserWithBalance(ctx, username, balance)
		return err
	})

	return
}

func (s *service) CreateSessionForUser(ctx context.Context, user *usermodels.User) (sessionID string, err error) {
	err = s.StoreProvider.UseStore(ctx, func(store Store) error {
		sessionID = uuid.New().String()
		err = store.CreateSessionForUser(ctx, sessionID, user, time.Now().Add(s.SessionTTL))
		return err
	})
	return
}

func isValidUsername(username string) bool {
	if username == "" {
		return false
	}
	for _, char := range username {
		if !('0' <= char && char <= '9') &&
			!('a' <= char && char <= 'z') &&
			!('A' <= char && char <= 'Z') &&
			char != '-' && char != '.' && char != '_' {
			return false
		}
	}
	return true
}
