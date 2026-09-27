package requests

import "encoding/json"

type Next struct {
	SessionID string          `json:"sessionID"`
	GameID    string          `json:"gameID"`
	GameData  json.RawMessage `json:"gameData"`
}
