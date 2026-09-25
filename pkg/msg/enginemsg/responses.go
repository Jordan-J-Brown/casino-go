package enginemsg

import "github.com/shopspring/decimal"

type RegisterEngineResponse struct {
	//TODO: is this needed?
}

type SetupResponse struct {
	GameConfig []byte `json:"gameConfig"`
	MetaData   []byte `json:"metaData"`
}

type RoundResponse struct {
	PayoutMultiplier decimal.Decimal `json:"payout_multiplier"`
	PrivateState     []byte          `json:"private_state"`
	PublicState      []byte          `json:"public_state"`
}

type CloseResponse struct {
	CanClose bool `json:"can_close"`
}
