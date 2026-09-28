package gameservice

import (
	"github.com/Jordan-J-Brown/casino-go/casino/internal/domain/usermodels"
)

type SetupRequest struct {
	EngineID string
	User     *usermodels.User
	MetaData []byte
}

type SetupResult struct {
	Bets        []int64
	GameConfig  []byte
	MetaData    []byte
	PublicState []byte
}

type PlayRequest struct {
	EngineID string
	Bet      int64
	User     *usermodels.User
	MetaData []byte
}

type NextRequest struct {
	EngineID string
	User     *usermodels.User
	MetaData []byte
}

type RoundResult struct {
	Balance          int64
	Bet              int64
	PayoutMutliplier string
	PublicState      []byte
}

type CloseRequest struct {
	EngineID string
	User     *usermodels.User
}

type CloseResult struct {
	Balance int64
}
