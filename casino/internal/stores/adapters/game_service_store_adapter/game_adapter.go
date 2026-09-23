package game_service_store_adapter

import (
	"context"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/gamemodels"
	"github.com/Jordan-J-Brown/casino-go/casino/internal/stores/gamestorepg"
	"github.com/Jordan-J-Brown/casino-go/pkg/pgdb"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GameAdapter struct {
	store *gamestorepg.Store
}

func newGameAdapter(pool *pgxpool.Pool) *GameAdapter {
	return &GameAdapter{
		store: gamestorepg.New(pool),
	}
}

func (g *GameAdapter) IsActiveGamePresent(ctx context.Context, userID string, engineID string) (bool, error) {
	return g.store.IsActiveGamePresent(ctx, userID, engineID)
}

func (g *GameAdapter) LoadGame(ctx context.Context, userID string, engineID string) (*gamemodels.Game, error) {
	tx, ok := ctx.Value(pgdb.CtxKeyTx).(pgx.Tx)
	if !ok {
		return nil, ErrInvalidTransaction
	}

	return g.store.LoadGame(ctx, tx, userID, engineID)
}

func (g *GameAdapter) SaveGame(ctx context.Context, game *gamemodels.Game) error {
	tx, ok := ctx.Value(pgdb.CtxKeyTx).(pgx.Tx)
	if !ok {
		return ErrInvalidTransaction
	}

	return g.store.SaveGame(ctx, tx, game)
}
