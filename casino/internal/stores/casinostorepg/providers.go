package casinostorepg

import (
	"context"
	"errors"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/services/gameservice"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/services/userservice"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BaseStoreProvider struct {
	Pool   *pgxpool.Pool
	Logger logsx.Logger
}

func (b *BaseStoreProvider) runInTx(ctx context.Context, setTx func(tX), run func() error) (err error) {
	tx, err := b.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	setTx(tx)

	defer func() {
		if err == nil {
			if commitErr := tx.Commit(ctx); commitErr != nil {
				err = errors.Join(commitErr, err)
			}
			return
		}

		if rollBackErr := tx.Rollback(ctx); rollBackErr != nil {
			err = errors.Join(rollBackErr, err)
		}
	}()

	return run()
}

type CasinoGameStoreProvider struct {
	BaseStoreProvider
}

func (s *CasinoGameStoreProvider) UseStore(ctx context.Context, fn func(store gameservice.CasinoGameStore) error) (err error) {
	store := &Store{
		logger: s.Logger,
	}

	return s.runInTx(ctx, func(tx tX) { store.tx = tx }, func() error { return fn(store) })
}

type UserStoreProvider struct {
	BaseStoreProvider
}

func (s *UserStoreProvider) UseStore(ctx context.Context, fn func(store userservice.Store) error) (err error) {
	store := &Store{
		logger: s.Logger,
	}

	return s.runInTx(ctx, func(tx tX) { store.tx = tx }, func() error { return fn(store) })
}
