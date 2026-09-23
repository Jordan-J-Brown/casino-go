package userservice

import (
	"context"
	"time"
	"uuid"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/usermodels"
)

type UserWalletStore interface {
	GetUserByID(ctx context.Context, id string) (*usermodels.User, error)
	GetUserFromSessionID(ctx context.Context, sessionID string) (*usermodels.User, error)
	CreateUserWithBalance(ctx context.Context, username string, balance int64) (*usermodels.User, error)
	CreateSessionForUser(ctx context.Context, sessionID string, user *usermodels.User, expiresAt time.Time) error
}

type Service interface {
	GetUserByID(ctx context.Context, id string) (*usermodels.User, error)
	GetUserFromSessionID(ctx context.Context, sessionID string) (*usermodels.User, error)
	CreateUserWithBalance(ctx context.Context, username string, balance int64) (*usermodels.User, error)
	CreateSessionForUser(ctx context.Context, user *usermodels.User) (sessionID string, err error)
}

type service struct {
	cfg Config
}

type Config struct {
	SessionTTL time.Duration
	Store      UserWalletStore
}

func New(cfg Config) Service {
	if cfg.SessionTTL == 0 {
		cfg.SessionTTL = 30 * time.Minute
	}

	return &service{
		cfg: cfg,
	}
}

func (s *service) GetUserByID(ctx context.Context, id string) (*usermodels.User, error) {
	return s.cfg.Store.GetUserByID(ctx, id)
}

func (s *service) GetUserFromSessionID(ctx context.Context, sessionID string) (*usermodels.User, error) {
	return s.cfg.Store.GetUserFromSessionID(ctx, sessionID)
}

func (s *service) CreateUserWithBalance(ctx context.Context, username string, balance int64) (*usermodels.User, error) {
	if username == "" {
		return nil, usermodels.ErrInvalidUsername
	}
	if balance < 0 {
		return nil, usermodels.ErrInvalidBalance
	}

	return s.cfg.Store.CreateUserWithBalance(ctx, username, balance)
}

func (s *service) CreateSessionForUser(ctx context.Context, user *usermodels.User) (sessionID string, err error) {
	sessionID = uuid.New().String()
	err = s.cfg.Store.CreateSessionForUser(ctx, sessionID, user, time.Now().Add(s.cfg.SessionTTL))
	return
}
