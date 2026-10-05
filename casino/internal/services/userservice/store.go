package userservice

import (
	"context"
	"time"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/usermodels"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/walletmodels"
)

type Store interface {
	GetUserByID(ctx context.Context, id string) (*usermodels.User, error)
	GetUserFromSessionID(ctx context.Context, sessionIdHash string) (*usermodels.User, error)
	GetWallet(ctx context.Context, walletID string) (*walletmodels.Wallet, error)
	CreateUserWithBalance(ctx context.Context, username string, balance int64) (*usermodels.User, error)
	CreateSessionForUser(ctx context.Context, sessionIdHash string, user *usermodels.User, expiresAt time.Time) error
	RevokeSession(ctx context.Context, sessionIdHash string) error
}

type StoreProvider interface {
	UseStore(ctx context.Context, fn func(store Store) error) error
}
