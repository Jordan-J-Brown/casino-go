package requests

import "encoding/json"

type Play struct {
	SessionID string          `json:"sessionID"`
	GameID    string          `json:"gameID"`
	Bet       int64           `json:"bet"`
	GameData  json.RawMessage `json:"gameData"`
}
