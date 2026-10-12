package requests

type PlayRequest struct {
	Mines int      `json:"mines"`
	Dev   *DevData `json:"dev,omitzero"`
}

type DevData struct {
	Seed string `json:"seed"`
}

type NextRequest struct {
	Cashout bool `json:"cashout"`
	Index   int  `json:"index"`
}
