package stores

import (
	"context"

	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/gamemodels"
)

type GameStore interface {
	IsActiveGamePresent(ctx context.Context, userID string, engineID string) (bool, error)
	LoadGame(ctx context.Context, userID string, engineID string) (*gamemodels.Game, error)
	SaveGame(ctx context.Context, game *gamemodels.Game) error
}
