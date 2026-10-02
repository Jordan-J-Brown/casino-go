package gamemodels

import (
	"github.com/shopspring/decimal"
)

type Status string

const (
	StatusPlaying  Status = "playing"
	StatusClosed   Status = "closed"
	StatusRefunded Status = "refunded"
)

type Game struct {
	GameID           string
	EngineID         string
	UserID           string
	Bet              int64
	Status           Status
	PayoutMultiplier decimal.Decimal
	PrivateState     []byte
	PublicState      []byte
}
