package requests

import "encoding/json"

type Next struct {
	SessionHeader
	GameID   string          `json:"gameID"`
	GameData json.RawMessage `json:"gameData" swaggertype:"string" format:"json"`
}
