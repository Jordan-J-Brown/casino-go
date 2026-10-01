package gameservice

import (
	"context"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/gamemodels"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/walletmodels"
)

type CasinoGameStoreProvider interface {
	UseStore(ctx context.Context, fn func(store CasinoGameStore) error) error
}

type CasinoGameStore interface {
	IsActiveGamePresent(ctx context.Context, userID string, engineID string) (bool, error)
	LoadGame(ctx context.Context, userID string, engineID string) (*gamemodels.Game, error)
	SaveGame(ctx context.Context, game *gamemodels.Game) error

	GetWallet(ctx context.Context, walletID string) (*walletmodels.Wallet, error)
	GetOrCreateCasinoWallet(ctx context.Context, walletID string, balance int64) (*walletmodels.Wallet, error)
	TransferFunds(ctx context.Context, args walletmodels.TransferArgs, transferFn func(src, dst *walletmodels.Wallet) error) error
}
