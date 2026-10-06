package mineshooks

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Jordan-J-Brown/casino-go/casino/pkg/responses"
	minesrequests "github.com/Jordan-J-Brown/casino-go/engines/mines/pkg/requests"
	minesresponses "github.com/Jordan-J-Brown/casino-go/engines/mines/pkg/responses"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/stress-test/internal/casinosession"
)

type Hooks struct {
	Mines         int
	IndexesToPick []int
	Logger        logsx.Logger

	currentIndex      int
	gameConfig        minesresponses.GameConfig
	gameState         string
	selectedPositions []int
}

func (hooks *Hooks) ProcessSetup(ctx context.Context, setup responses.Setup) error {
	var gc minesresponses.GameConfig
	if err := json.Unmarshal(setup.GameConfig, &gc); err != nil {
		return err
	}

	hooks.gameConfig = gc

	return nil
}

func (hooks *Hooks) GetPlayGameData(ctx context.Context) (json.RawMessage, error) {
	return json.Marshal(minesrequests.PlayRequest{
		Mines: hooks.Mines,
	})
}

func (hooks *Hooks) ParseRound(ctx context.Context, state casinosession.GameState, round *responses.Round) error {
	switch state {
	case casinosession.GameStatePlay:
		var r minesresponses.PlayResponse
		if err := json.Unmarshal(round.RoundData, &r); err != nil {
			return err
		}
		hooks.gameState = r.State
	case casinosession.GameStateNext:
		var r minesresponses.NextResponse
		if err := json.Unmarshal(round.RoundData, &r); err != nil {
			return err
		}
		hooks.gameState = r.State
	default:
		return errors.New("invalid state to parse round: " + state.String())
	}

	return nil
}

func (hooks *Hooks) GameIsClosed(ctx context.Context) bool {
	return hooks.gameState == "CASHED_OUT" || hooks.gameState == "LOST"
}

func (hooks *Hooks) GetNextGameData(ctx context.Context) (json.RawMessage, error) {
	if hooks.currentIndex < len(hooks.IndexesToPick) {
		index := hooks.IndexesToPick[hooks.currentIndex]
		hooks.currentIndex++

		hooks.Logger.DebugContext(ctx, "Pick", "index", index)

		return json.Marshal(minesrequests.NextRequest{
			Index: index,
		})
	}

	return json.Marshal(minesrequests.NextRequest{
		Cashout: true,
	})
}

func (hooks *Hooks) Reset(ctx context.Context) {
	hooks.currentIndex = 0
	hooks.selectedPositions = nil
	hooks.gameState = ""
}
