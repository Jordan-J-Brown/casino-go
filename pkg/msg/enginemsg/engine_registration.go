package enginemsg

type EngineRegistraton struct {
	EngineID string `json:"engine_id"`
	MinBet   int64  `json:"min_bet"`
	MaxBet   int64  `json:"max_bet"`
}
