package responses

import "encoding/json"

type Round struct {
	Balance          int64           `json:"balance"`
	Bet              int64           `json:"bet"`
	PayoutMultiplier string          `json:"payoutMultiplier"`
	RoundData        json.RawMessage `json:"roundData" swaggertype:"string" format:"json"`
}
