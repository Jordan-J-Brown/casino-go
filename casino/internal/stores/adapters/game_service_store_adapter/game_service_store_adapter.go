package game_service_store_adapter

import (
	"context"
	"errors"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/services/gameservice"
	"github.com/Jordan-J-Brown/casino-go/pkg/errorx"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/pkg/pgdb"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidTransaction = errors.New("invalid transaction within context")
)

type GameServiceStoreAdapter struct {
	pool          *pgxpool.Pool
	walletAdapter *WalletAdapter
	gameAdapter   *GameAdapter
}

func New(pool *pgxpool.Pool, logger logsx.Logger) *GameServiceStoreAdapter {
	return &GameServiceStoreAdapter{
		pool:          pool,
		walletAdapter: newWalletAdapter(pool, logger),
		gameAdapter:   newGameAdapter(pool),
	}
}

func (g *GameServiceStoreAdapter) Transact(ctx context.Context, fn func(ctx context.Context, stores gameservice.Adapters) error) (err error) {
	tx, err := g.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			commitErr := tx.Commit(ctx)
			if commitErr != nil {
				err = errorx.New("failed to commit transaction", errorx.WithChildren(commitErr))
			}
		} else {
			rollBackErr := tx.Rollback(ctx)
			if rollBackErr != nil {
				err = errorx.New("failed to roll back transaction", errorx.WithChildren(err))
			}
		}
	}()

	ctx = context.WithValue(ctx, pgdb.CtxKeyTx, tx)

	return fn(ctx, gameservice.Adapters{
		GameStore:   g.gameAdapter,
		WalletStore: g.walletAdapter,
	})
}

var _ gameservice.TransactionProvider = &GameServiceStoreAdapter{}
