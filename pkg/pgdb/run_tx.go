package pgdb

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CtxKey string

const (
	CtxKeyTx CtxKey = "tx"
)

func RunTx(ctx context.Context, pool *pgxpool.Pool, fn func(ctx context.Context, tx pgx.Tx) error) (err error) {
	var tx pgx.Tx
	if txFromCtx, ok := ctx.Value(CtxKeyTx).(pgx.Tx); ok {
		tx = txFromCtx
	} else {
		tx, err = pool.Begin(ctx)
		if err != nil {
			return err
		}
	}

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

	return fn(ctx, tx)
}
