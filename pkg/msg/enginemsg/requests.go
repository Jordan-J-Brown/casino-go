package enginemsg

type EngineRegistration struct {
	EngineID string `json:"engine_id"`
	MinBet   int64  `json:"min_bet"`
	MaxBet   int64  `json:"max_bet"`
}

type SetupRequest struct {
	MetaData []byte `json:"meta_data"`
}

type PlayRequest struct {
	Bet      int64  `json:"bet"`
	MetaData []byte `json:"meta_data"`
}

type NextRequest struct {
	State    []byte `json:"state"`
	MetaData []byte `json:"meta_data"`
}

type CloseRequest struct {
	State []byte `json:"state"`
}
