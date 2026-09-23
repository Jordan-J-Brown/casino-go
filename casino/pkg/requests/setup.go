package requests

import "encoding/json"

type Setup struct {
	SessionID string          `json:"sessionID"`
	GameID    string          `json:"gameID"`
	MetaData  json.RawMessage `json:"metaData"`
}
