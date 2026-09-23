package game_service_store_adapter

import (
	"context"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/walletmodels"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/stores/userwalletstorepg"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/pkg/pgdb"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WalletAdapter struct {
	store *userwalletstorepg.Store
	Tx    pgx.Tx
}

func newWalletAdapter(pool *pgxpool.Pool, logger logsx.Logger) *WalletAdapter {
	return &WalletAdapter{
		store: userwalletstorepg.New(pool, logger),
	}
}

func (adapter *WalletAdapter) GetWallet(ctx context.Context, walletID string) (*walletmodels.Wallet, error) {
	return adapter.store.GetWallet(ctx, walletID)
}

func (adapter *WalletAdapter) TransferFunds(
	ctx context.Context,
	args walletmodels.TransferArgs,
	transferFn func(src, dst *walletmodels.Wallet) error,
) error {
	tx, ok := ctx.Value(pgdb.CtxKeyTx).(pgx.Tx)
	if !ok {
		return ErrInvalidTransaction
	}
	return adapter.store.TransferFunds(ctx, tx, args, transferFn)
}
