package casinostorepg

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type tX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Store struct {
	tx     tX
	logger logsx.Logger
}

type StoreProvider[T any] struct {
	Pool   *pgxpool.Pool
	Logger logsx.Logger
}

func (s *StoreProvider[T]) UseStore(ctx context.Context, fn func(store T) error) (err error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
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

	store := &Store{
		tx:     tx,
		logger: s.Logger,
	}

	typedStore, ok := any(store).(T)
	if !ok {
		return fmt.Errorf("store does not implement: %s", reflect.TypeFor[T]().Name())
	}

	err = fn(typedStore)

	return
}
