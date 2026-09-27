package minesgame

import "fmt"

type Error int8

const (
	ErrInvalidMinesCount Error = iota
	ErrInvalidGridSize
	ErrInvalidBet
	ErrInvalidGameState
	ErrInvalidIndex
	ErrInvalidHouseEdge
)

func (e Error) Error() string {
	switch e {
	case ErrInvalidBet:
		return "ERR_INVALID_BET"
	case ErrInvalidGridSize:
		return "ERR_INVALID_GRID_SIZE"
	case ErrInvalidMinesCount:
		return "ERR_INVALID_MINES_COUNT"
	case ErrInvalidGameState:
		return "ERR_INVALID_GAME_STATE"
	case ErrInvalidIndex:
		return "ERR_INVALID_INDEX"
	case ErrInvalidHouseEdge:
		return "ERR_INVALID_HOUSE_EDGE"
	default:
		panic(fmt.Sprintf("unexpected minesgame.Error: %#v", e))
	}
}
