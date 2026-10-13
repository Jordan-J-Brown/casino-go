package requests

import "encoding/json"

type Play struct {
	SessionHeader
	GameID   string          `json:"gameID"`
	Bet      int64           `json:"bet"`
	GameData json.RawMessage `json:"gameData" swaggertype:"string" format:"json"`
}
