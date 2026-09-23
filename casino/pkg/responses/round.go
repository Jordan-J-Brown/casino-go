package responses

import "encoding/json"

type Round struct {
	Balance          int64           `json:"balance"`
	PayoutMultiplier string          `json:"payoutMultiplier"`
	RoundData        json.RawMessage `json:"roundData"`
}
