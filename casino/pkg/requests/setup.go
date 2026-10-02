package requests

import "encoding/json"

type Setup struct {
	SessionHeader
	GameID   string          `json:"gameID"`
	MetaData json.RawMessage `json:"metaData"`
}
