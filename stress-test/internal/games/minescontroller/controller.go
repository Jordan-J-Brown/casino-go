package minescontroller

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Jordan-J-Brown/casino-go/casino/pkg/responses"
	minesrequests "github.com/Jordan-J-Brown/casino-go/engines/mines/pkg/requests"
	minesresponses "github.com/Jordan-J-Brown/casino-go/engines/mines/pkg/responses"
	"github.com/Jordan-J-Brown/casino-go/pkg/logsx"
	"github.com/Jordan-J-Brown/casino-go/stress-test/internal/games"
)

type Controller struct {
	Mines     int
	Positions []int
	Logger    logsx.Logger
	Seed      *string

	currentIndex      int
	gameConfig        minesresponses.GameConfig
	gameState         string
	selectedPositions []int
}

func (controller *Controller) ProcessSetup(ctx context.Context, setup responses.Setup) error {
	var gc minesresponses.GameConfig
	if err := json.Unmarshal(setup.GameConfig, &gc); err != nil {
		return err
	}

	controller.gameConfig = gc

	return nil
}

func (controller *Controller) GetPlayGameData(ctx context.Context) (json.RawMessage, error) {
	var dev *minesrequests.DevData
	if controller.Seed != nil {
		dev = &minesrequests.DevData{
			Seed: *controller.Seed,
		}
	}

	return json.Marshal(minesrequests.PlayRequest{
		Mines: controller.Mines,
		Dev:   dev,
	})
}

func (controller *Controller) ParseRound(ctx context.Context, state games.GameState, round *responses.Round) error {
	switch state {
	case games.GameStatePlay:
		var r minesresponses.PlayResponse
		if err := json.Unmarshal(round.RoundData, &r); err != nil {
			return err
		}
		controller.gameState = r.State
	case games.GameStateNext:
		var r minesresponses.NextResponse
		if err := json.Unmarshal(round.RoundData, &r); err != nil {
			return err
		}
		controller.gameState = r.State
	default:
		return errors.New("invalid state to parse round: " + state.String())
	}

	return nil
}

func (controller *Controller) GameIsClosed(ctx context.Context) bool {
	return controller.gameState == "CASHED_OUT" || controller.gameState == "LOST"
}

func (controller *Controller) GetNextGameData(ctx context.Context) (json.RawMessage, error) {
	if controller.currentIndex < len(controller.Positions) {
		index := controller.Positions[controller.currentIndex]
		controller.currentIndex++

		controller.Logger.DebugContext(ctx, "Pick", "index", index)

		return json.Marshal(minesrequests.NextRequest{
			Index: index,
		})
	}

	return json.Marshal(minesrequests.NextRequest{
		Cashout: true,
	})
}

func (controller *Controller) Reset(ctx context.Context) {
	controller.currentIndex = 0
	controller.selectedPositions = nil
	controller.gameState = ""
}
