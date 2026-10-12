package minesgame

import (
	"slices"

	"github.com/Jordan-J-Brown/casino-go/pkg/randx"
	"github.com/shopspring/decimal"
)

type Game struct {
	State            State
	Bet              int64
	GridSize         int
	Mines            int
	Selected         []int
	PayoutMultiplier decimal.Decimal
	HouseEdge        decimal.Decimal
	Rng              randx.Generator
}

type Config struct {
	Rng       randx.Generator
	Bet       int64
	Mines     int
	GridSize  int
	HouseEdge float64
}

func New(cfg Config) (*Game, error) {
	houseEdge, err := getHouseEdgeDecimal(cfg.HouseEdge)
	if err != nil {
		return nil, err
	}
	if cfg.GridSize < 2 {
		return nil, ErrInvalidGridSize
	}
	if cfg.Mines <= 0 || cfg.Mines >= cfg.GridSize {
		return nil, ErrInvalidMinesCount
	}
	if cfg.Bet <= 0 {
		return nil, ErrInvalidBet
	}

	return &Game{
		Rng:              cfg.Rng,
		HouseEdge:        houseEdge,
		Bet:              cfg.Bet,
		State:            StatePlaying,
		GridSize:         cfg.GridSize,
		Mines:            cfg.Mines,
		PayoutMultiplier: decimal.NewFromInt(0),
	}, nil
}

func (game *Game) Pick(index int) (*Result, *Error) {
	if game.State != StatePlaying {
		return nil, new(ErrInvalidGameState)
	}
	if index < 0 || index >= game.GridSize {
		return nil, new(ErrInvalidIndex)
	}
	if slices.Contains(game.Selected, index) {
		return nil, new(ErrInvalidIndex)
	}

	remainingCells := int64(game.GridSize - len(game.Selected))
	remainingMines := int64(game.Mines)
	if game.Rng.Int(1, remainingCells) <= remainingMines {
		game.State = StateLost
		game.PayoutMultiplier = decimal.NewFromInt(0)
		return game.buildResult(), nil
	}

	game.Selected = append(game.Selected, index)
	game.PayoutMultiplier = game.calculatePayoutMultiplier()
	if len(game.Selected) == game.GridSize-game.Mines {
		game.State = StateCashedOut
	}

	return game.buildResult(), nil
}

func (game *Game) Cashout() (result *Result, err *Error) {
	if !game.IsPlaying() {
		return nil, new(ErrInvalidGameState)
	}

	game.State = StateCashedOut
	game.PayoutMultiplier = game.calculatePayoutMultiplier()

	return game.buildResult(), nil
}

func (game *Game) IsPlaying() bool {
	return game.State == StatePlaying
}

func (game *Game) calculatePayoutMultiplier() decimal.Decimal {
	if game.State == StateLost {
		return decimal.NewFromInt(0)
	}

	multiplier := decimal.NewFromInt(1).Sub(game.HouseEdge)
	mines := int64(game.Mines)
	for selected := int64(0); selected < int64(len(game.Selected)); selected++ {
		multiplier = multiplier.Mul(decimal.NewFromInt(int64(game.GridSize) - selected))
		multiplier = multiplier.Div(decimal.NewFromInt(int64(game.GridSize) - mines - selected))
	}
	return multiplier
}

func (game *Game) buildResult() *Result {
	result := &Result{
		State:             game.State,
		PayoutMultiplier:  game.PayoutMultiplier,
		SelectedPositions: game.Selected,
	}
	if !game.IsPlaying() {
	}
	return result
}

func getHouseEdgeDecimal(value float64) (decimal.Decimal, error) {
	houseEdge := decimal.NewFromFloat(value).Div(decimal.NewFromInt(100))
	if houseEdge.IsNegative() || houseEdge.GreaterThanOrEqual(decimal.NewFromInt(1)) {
		return decimal.Zero, ErrInvalidHouseEdge
	}
	return houseEdge, nil
}
