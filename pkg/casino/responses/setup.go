package responses

import "encoding/json"

type Setup struct {
	Bets       []int64         `json:"bets"`
	GameConfig json.RawMessage `json:"gameConfig"`
	MetaData   json.RawMessage `json:"metaData" swaggertype:"string" format:"json"`
	RoundData  json.RawMessage `json:"roundData,omitzero" swaggertype:"string" format:"json"`
}
