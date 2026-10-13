package userservice

import (
	"context"
	"time"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/usermodels"
	"github.com/Jordan-J-Brown/casino-go/pkg/errorx"
)

type Service interface {
	GetUserByID(ctx context.Context, id string) (*usermodels.User, *errorx.Error)
	GetUserByName(ctx context.Context, name string) (*usermodels.User, *errorx.Error)
	GetUserFromSessionID(ctx context.Context, sessionID string) (*usermodels.User, *errorx.Error)
	GetUserBalance(ctx context.Context, user *usermodels.User) (int64, *errorx.Error)
	CreateUserWithBalance(ctx context.Context, username string, balance int64) (*usermodels.User, *errorx.Error)
	CreateSessionForUser(ctx context.Context, user *usermodels.User) (sessionID string, err *errorx.Error)
	RevokeSession(ctx context.Context, sessionID string) *errorx.Error
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

func (s *service) GetUserByID(ctx context.Context, id string) (user *usermodels.User, err *errorx.Error) {
	storeErr := s.StoreProvider.UseStore(ctx, func(store Store) error {
		userFromStore, err := store.GetUserByID(ctx, id)
		user = userFromStore
		return err
	})
	if storeErr != nil {
		err = errorx.NewFromError(storeErr)
	}

	return
}

func (s *service) GetUserByName(ctx context.Context, name string) (*usermodels.User, *errorx.Error) {
	if !isValidUsername(name) {
		return nil, errorx.NewFromType(usermodels.ErrInvalidUsername.Error(), errorx.WithCode(errorx.ClientError))
	}

	var user *usermodels.User
	storeErr := s.StoreProvider.UseStore(ctx, func(store Store) error {
		userFromStore, err := store.GetUserByName(ctx, name)
		if err != nil {
			return err
		}

		user = userFromStore

		return nil
	})
	if storeErr != nil {
		if storeErr == usermodels.ErrUserNotFound {
			return nil, errorx.NewFromType(string(usermodels.ErrUserNotFound), errorx.WithCode(errorx.ClientError))
		}
		return nil, errorx.Newf("failed to retrieve user '%s' from store.", name, errorx.WithChildren(storeErr))
	}

	return user, nil
}

func (s *service) GetUserFromSessionID(ctx context.Context, sessionID string) (user *usermodels.User, err *errorx.Error) {
	rawSessionID, idErr := newSessionIdFromB64(sessionID)
	if idErr != nil {
		return nil, errorx.NewFromType(string(usermodels.ErrInvalidSessionID), errorx.WithCode(errorx.ClientError))
	}

	storeErr := s.StoreProvider.UseStore(ctx, func(store Store) error {
		userFromStore, err := store.GetUserFromSessionID(ctx, rawSessionID.Base64Sha256())
		user = userFromStore
		return err
	})
	if storeErr != nil {
		return nil, errorx.NewFromError(storeErr)
	}
	return
}

func (s *service) GetUserBalance(ctx context.Context, user *usermodels.User) (balance int64, err *errorx.Error) {
	if user == nil || user.ID == "" || user.WalletID == "" {
		return 0, errorx.NewFromType(usermodels.ErrBadArgs.Error(), errorx.WithCode(errorx.ClientError))
	}

	storeErr := s.StoreProvider.UseStore(ctx, func(store Store) error {
		wallet, err := store.GetWallet(ctx, user.WalletID)
		if err != nil {
			return err
		}

		balance = wallet.BalanceMinor

		return nil
	})
	if storeErr != nil {
		return 0, errorx.NewFromError(storeErr)
	}

	return
}

func (s *service) CreateUserWithBalance(ctx context.Context, username string, balance int64) (user *usermodels.User, err *errorx.Error) {
	if !isValidUsername(username) {
		return nil, errorx.NewFromType(usermodels.ErrInvalidUsername.Error(), errorx.WithCode(errorx.ClientError))
	}
	if balance < 0 {
		return nil, errorx.NewFromType(usermodels.ErrInvalidBalance.Error(), errorx.WithCode(errorx.ClientError))
	}

	storeErr := s.StoreProvider.UseStore(ctx, func(store Store) error {
		userFromStore, err := store.CreateUserWithBalance(ctx, username, balance)
		user = userFromStore
		return err
	})
	if storeErr != nil {
		if storeErr == usermodels.ErrUsernameExists {
			return nil, errorx.NewFromType(string(usermodels.ErrUsernameExists), errorx.WithCode(errorx.ClientError))
		}
		return nil, errorx.NewFromError(storeErr)
	}

	return
}

func (s *service) CreateSessionForUser(ctx context.Context, user *usermodels.User) (sessionID string, err *errorx.Error) {
	storeErr := s.StoreProvider.UseStore(ctx, func(store Store) error {
		rawSessionID := newSessionID()
		sessionID = rawSessionID.Base64()
		err := store.CreateSessionForUser(ctx, rawSessionID.Base64Sha256(), user, time.Now().Add(s.SessionTTL))
		return err
	})
	if storeErr != nil {
		return "", errorx.NewFromError(storeErr)
	}

	return
}

func (s *service) RevokeSession(ctx context.Context, sessionID string) *errorx.Error {
	rawSessionID, err := newSessionIdFromB64(sessionID)
	if err != nil {
		return errorx.NewFromType(string(usermodels.ErrInvalidSessionID), errorx.WithCode(errorx.ClientError))
	}

	storeErr := s.StoreProvider.UseStore(ctx, func(store Store) error {
		return store.RevokeSession(ctx, rawSessionID.Base64Sha256())
	})
	if storeErr != nil {
		return errorx.NewFromError(storeErr)
	}

	return nil
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
