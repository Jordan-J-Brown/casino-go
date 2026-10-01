package casinostorepg

import (
	"context"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/gamemodels"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type gameStatus = string

const (
	gameStatusPlaying  gameStatus = "playing"
	gameStatusClosed   gameStatus = "closed"
	gameStatusRefunded gameStatus = "refunded"
)

func (store *Store) IsActiveGamePresent(ctx context.Context, userID string, engineID string) (active bool, err error) {
	const query = `SELECT EXISTS (SELECT 1 FROM service.games WHERE user_id=$1 AND engine_id=$2 AND game_status=$3);`

	err = store.tx.
		QueryRow(ctx, query, userID, engineID, gameStatusPlaying).
		Scan(&active)
	if err == pgx.ErrNoRows {
		return false, nil
	}

	return
}

func (store *Store) LoadGame(ctx context.Context, userID string, engineID string) (*gamemodels.Game, error) {
	const loadGameQuery = "SELECT * FROM service.games WHERE user_id=$1 AND engine_id=$2 AND game_status=$3;"

	game := new(gamemodels.Game)

	var (
		status           gameStatus
		payoutMultiplier string
	)

	err := store.tx.
		QueryRow(ctx, loadGameQuery, userID, engineID, gameStatusPlaying).
		Scan(
			&game.GameID,
			&game.UserID,
			&game.EngineID,
			&game.Bet,
			&status,
			&payoutMultiplier,
			&game.PrivateState,
			&game.PublicState,
		)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, gamemodels.ErrGameNotFound
		}
		return nil, err
	}

	switch status {
	case gameStatusPlaying:
		game.Status = gamemodels.StatusPlaying
	case gameStatusClosed:
		game.Status = gamemodels.StatusClosed
	case gameStatusRefunded:
		game.Status = gamemodels.StatusRefunded
	}

	game.PayoutMultiplier, err = decimal.NewFromString(payoutMultiplier)
	if err != nil {
		return nil, err
	}

	return game, nil
}

func (store *Store) SaveGame(ctx context.Context, game *gamemodels.Game) error {
	var gameExists bool
	err := store.tx.QueryRow(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM service.games WHERE game_id=$1 FOR UPDATE);`,
		game.GameID,
	).Scan(&gameExists)
	if err != nil {
		return err
	}

	if gameExists {
		_, err = store.tx.Exec(ctx, `
		UPDATE 
			service.games
		SET
			game_status=$1,
			payout_multiplier=$2,
			private_state=$3,
			public_state=$4
		WHERE
			game_id=$5
		;`,
			game.Status,
			game.PayoutMultiplier.StringFixed(2),
			game.PrivateState,
			game.PublicState,
			game.GameID,
		)
		return err
	}

	_, err = store.tx.Exec(ctx, `
	INSERT INTO service.games (
		game_id, 
		user_id, 
		engine_id, 
		bet, 
		game_status, 
		payout_multiplier, 
		private_state, 
		public_state
	) VALUES 
		($1, $2, $3, $4, $5, $6, $7, $8)
	;`,
		game.GameID,
		game.UserID,
		game.EngineID,
		game.Bet,
		game.Status,
		game.PayoutMultiplier.String(),
		game.PrivateState,
		game.PublicState,
	)

	return err
}
