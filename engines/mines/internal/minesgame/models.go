package minesgame

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type State int8

const (
	StatePlaying State = iota
	StateLost
	StateCashedOut
)

func (s State) String() string {
	switch s {
	case StatePlaying:
		return "PLAYING"
	case StateLost:
		return "LOST"
	case StateCashedOut:
		return "CASHED_OUT"
	default:
		panic(fmt.Sprintf("unexpected minesgame.State: %#v", s))
	}
}

type Result struct {
	State             State
	PayoutMultiplier  decimal.Decimal
	SelectedPositions []int
}
