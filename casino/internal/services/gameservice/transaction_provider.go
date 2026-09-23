package gameservice

import (
	"context"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/gamemodels"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/walletmodels"
)

type TransactionProvider interface {
	Transact(ctx context.Context, fn func(ctx context.Context, stores Adapters) error) error
}

type Adapters struct {
	GameStore   GameStoreAdapter
	WalletStore WalletStoreAdapter
}

type GameStoreAdapter interface {
	IsActiveGamePresent(ctx context.Context, userID string, engineID string) (bool, error)
	LoadGame(ctx context.Context, userID string, engineID string) (*gamemodels.Game, error)
	SaveGame(ctx context.Context, game *gamemodels.Game) error
}

type WalletStoreAdapter interface {
	GetWallet(ctx context.Context, walletID string) (*walletmodels.Wallet, error)
	TransferFunds(ctx context.Context, args walletmodels.TransferArgs, transferFn func(src, dst *walletmodels.Wallet) error) error
}
