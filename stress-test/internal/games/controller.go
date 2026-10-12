package games

import (
	"context"
	"encoding/json"

	"github.com/Jordan-J-Brown/casino-go/casino/pkg/responses"
)

type Controller interface {
	ProcessSetup(ctx context.Context, setup responses.Setup) error
	GetPlayGameData(ctx context.Context) (json.RawMessage, error)
	ParseRound(ctx context.Context, state GameState, round *responses.Round) error

	//All methods below will be called after ParseRound
	GameIsClosed(ctx context.Context) bool
	GetNextGameData(ctx context.Context) (json.RawMessage, error)
	Reset(ctx context.Context)
}
